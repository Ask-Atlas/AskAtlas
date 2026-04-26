//go:build integration

// Integration test for the chunk + embedding store (ASK-219).
//
// Spins up the official pgvector image via testcontainers-go,
// applies a minimal slice of the production schema (just users +
// files + the chunk table), and exercises the sqlc-generated
// queries against real Postgres.
//
// Build-tagged so default `make test` stays Docker-free for unit
// runs. The `make test-integration` target + the `integration-test`
// CI job opt in via `-tags=integration`.
package db_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// minimalSchema brings up just enough tables to exercise the chunk
// store. We replicate the relevant slice from the migration history
// inline rather than running every migration so the test stays fast
// and self-contained. Schema kept in sync with
// migrations/sql/20260426024205_create_study_guide_file_chunks.up.sql.
const minimalSchema = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE files (
  id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE study_guide_file_chunks (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  file_id     UUID         NOT NULL REFERENCES files(id) ON DELETE CASCADE,
  chunk_idx   INTEGER      NOT NULL CHECK (chunk_idx >= 0),
  text        TEXT         NOT NULL,
  embedding   vector(1536) NOT NULL,
  page        INTEGER      CHECK (page IS NULL OR page > 0),
  heading     TEXT,
  tokens      INTEGER      NOT NULL CHECK (tokens >= 0),
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  UNIQUE (file_id, chunk_idx)
);

CREATE INDEX idx_study_guide_file_chunks_embedding_hnsw
  ON study_guide_file_chunks USING hnsw (embedding vector_cosine_ops);
`

// makeVec returns a 1536-dim vector with a single 1.0 at the given
// index and zeros elsewhere. Cosine distance between e_i and e_j is
// 1.0 for i != j and 0.0 for i == j, so a query of e_q ranks the
// chunk with embedding e_q first and the rest tied at 1.0.
func makeVec(idx int) pgvector.Vector {
	v := make([]float32, 1536)
	v[idx] = 1.0
	return pgvector.NewVector(v)
}

func newUUID(t *testing.T) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(uuid.NewString()); err != nil {
		t.Fatalf("uuid scan: %v", err)
	}
	return u
}

// bringUpPostgres starts the pgvector image, applies minimalSchema,
// and returns a pgxpool ready for use. Cleanup is registered via
// t.Cleanup so each test gets an isolated container.
func bringUpPostgres(t *testing.T) (*pgxpool.Pool, *db.Queries) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"pgvector/pgvector:pg17",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Terminate(context.Background())
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	poolCtx, poolCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer poolCancel()
	pool, err := pgxpool.New(poolCtx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(poolCtx, minimalSchema); err != nil {
		t.Fatalf("apply minimalSchema: %v", err)
	}
	return pool, db.New(pool)
}

func seedUserAndFile(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (pgtype.UUID, pgtype.UUID) {
	t.Helper()
	userID := newUUID(t)
	fileID := newUUID(t)
	if _, err := pool.Exec(ctx, "INSERT INTO users (id) VALUES ($1)", userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx, "INSERT INTO files (id, user_id) VALUES ($1, $2)", fileID, userID); err != nil {
		t.Fatalf("insert file: %v", err)
	}
	return userID, fileID
}

func TestSearchChunksByEmbedding_OrdersByDistance(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	_, fileID := seedUserAndFile(t, ctx, pool)

	// Insert four chunks with one-hot vectors at distinct indices.
	// Querying with the e_2 vector should rank chunk_idx=2 first
	// (distance 0); the others tie at distance 1.
	for i := 0; i < 4; i++ {
		_, err := queries.InsertStudyGuideFileChunk(ctx, db.InsertStudyGuideFileChunkParams{
			FileID:    fileID,
			ChunkIdx:  int32(i),
			Text:      "chunk " + string(rune('A'+i)),
			Embedding: makeVec(i),
			Page:      pgtype.Int4{Int32: int32(i + 1), Valid: true},
			Heading:   pgtype.Text{String: "h" + string(rune('a'+i)), Valid: true},
			Tokens:    10,
		})
		if err != nil {
			t.Fatalf("insert chunk %d: %v", i, err)
		}
	}

	rows, err := queries.SearchChunksByEmbedding(ctx, db.SearchChunksByEmbeddingParams{
		Embedding: makeVec(2),
		FileIds:   []pgtype.UUID{fileID},
		K:         3,
	})
	if err != nil {
		t.Fatalf("SearchChunksByEmbedding: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}
	if rows[0].ChunkIdx != 2 {
		t.Errorf("rows[0].ChunkIdx = %d, want 2 (closest match)", rows[0].ChunkIdx)
	}
	if rows[0].Distance > 0.0001 {
		t.Errorf("rows[0].Distance = %v, want ~0", rows[0].Distance)
	}
	for i := 1; i < len(rows); i++ {
		if rows[i].Distance < rows[i-1].Distance {
			t.Errorf("distances not monotonically non-decreasing at i=%d: %v < %v",
				i, rows[i].Distance, rows[i-1].Distance)
		}
	}
}

func TestStudyGuideFileChunks_CascadeDeleteOnFile(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	_, fileID := seedUserAndFile(t, ctx, pool)

	for i := 0; i < 3; i++ {
		if _, err := queries.InsertStudyGuideFileChunk(ctx, db.InsertStudyGuideFileChunkParams{
			FileID:    fileID,
			ChunkIdx:  int32(i),
			Text:      "x",
			Embedding: makeVec(i),
			Tokens:    1,
		}); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	got, err := queries.GetChunksByFile(ctx, fileID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("pre-delete chunks = %d, want 3", len(got))
	}

	if _, err := pool.Exec(ctx, "DELETE FROM files WHERE id = $1", fileID); err != nil {
		t.Fatalf("delete file: %v", err)
	}

	got, err = queries.GetChunksByFile(ctx, fileID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("post-delete chunks = %d, want 0 (cascade should have wiped them)", len(got))
	}
}

func TestSearchChunksByEmbedding_UsesHnswIndex(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	_, fileID := seedUserAndFile(t, ctx, pool)

	for i := 0; i < 64; i++ {
		if _, err := queries.InsertStudyGuideFileChunk(ctx, db.InsertStudyGuideFileChunkParams{
			FileID:    fileID,
			ChunkIdx:  int32(i),
			Text:      "x",
			Embedding: makeVec(i),
			Tokens:    1,
		}); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	if _, err := pool.Exec(ctx, "ANALYZE study_guide_file_chunks"); err != nil {
		t.Fatalf("analyze: %v", err)
	}

	// EXPLAIN the same shape as SearchChunksByEmbedding. We don't
	// strictly require "Index Scan using ..._hnsw" because the
	// planner can pick a bitmap variant in some configurations; we
	// just assert the HNSW index name appears in the plan.
	rows, err := pool.Query(ctx,
		`EXPLAIN (FORMAT TEXT)
		 SELECT id FROM study_guide_file_chunks
		 WHERE file_id = ANY($1::uuid[])
		 ORDER BY embedding <=> $2
		 LIMIT 5`,
		[]pgtype.UUID{fileID}, makeVec(7),
	)
	if err != nil {
		t.Fatalf("EXPLAIN: %v", err)
	}
	defer rows.Close()
	var plan strings.Builder
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			t.Fatalf("scan: %v", err)
		}
		plan.WriteString(line)
		plan.WriteString("\n")
	}
	if !strings.Contains(plan.String(), "idx_study_guide_file_chunks_embedding_hnsw") {
		t.Errorf("HNSW index not referenced in plan:\n%s", plan.String())
	}
	_ = queries
}

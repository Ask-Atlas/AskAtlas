//go:build integration

// Integration test for the ASK-221 chunk+embed worker. Same shape as
// extract_worker_integration_test.go (testcontainers Postgres) but
// with its own schema bootstrap because we also need pgvector + the
// study_guide_file_chunks table.
//
// The Embedder dependency is mocked in-process (no OpenAI traffic) so
// the test runs offline. The point is to verify the SQL paths the
// worker drives -- migration cohabits, sqlc encodes pgvector.Vector
// correctly, the delete-then-insert tx is atomic, and the cascade
// delete + cleanup run as expected against real Postgres.
package files_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// minimalChunkEmbedSchema mirrors the union of ASK-219 + ASK-220 +
// ASK-221 migrations -- enough to drive the chunk+embed worker
// end-to-end without running the full migration history. Kept inline
// so the test stays self-contained.
const minimalChunkEmbedSchema = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TYPE upload_status AS ENUM ('pending', 'complete', 'failed');
CREATE TYPE processing_status AS ENUM (
    'uploaded', 'extracting', 'extracted', 'embedding', 'ready', 'failed'
);

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid()
);

CREATE TABLE files (
  id                UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id           UUID              NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  s3_key            TEXT              NOT NULL,
  name              TEXT              NOT NULL DEFAULT '',
  mime_type         TEXT              NOT NULL,
  size              BIGINT            NOT NULL DEFAULT 0,
  status            upload_status     NOT NULL DEFAULT 'pending',
  processing_status processing_status NOT NULL DEFAULT 'uploaded',
  status_error      TEXT,
  deletion_status   TEXT,
  created_at        TIMESTAMPTZ       NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ       NOT NULL DEFAULT now()
);

CREATE TABLE files_extracted_text (
  file_id      UUID         PRIMARY KEY REFERENCES files(id) ON DELETE CASCADE,
  text         TEXT         NOT NULL,
  page_offsets INTEGER[],
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
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
`

func bringUpChunkEmbedPostgres(t *testing.T) (*pgxpool.Pool, *db.Queries) {
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

	if _, err := pool.Exec(poolCtx, minimalChunkEmbedSchema); err != nil {
		t.Fatalf("apply schema: %v", err)
	}
	return pool, db.New(pool)
}

func seedExtractedFile(t *testing.T, ctx context.Context, pool *pgxpool.Pool, text string, pageOffsets []int32) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	fileID := uuid.New()
	if _, err := pool.Exec(ctx, "INSERT INTO users (id) VALUES ($1)", userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO files (id, user_id, s3_key, mime_type, status, processing_status)
		 VALUES ($1, $2, $3, $4, 'complete', 'extracted')`,
		fileID, userID, "uploads/"+fileID.String(), "text/plain",
	); err != nil {
		t.Fatalf("insert file: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO files_extracted_text (file_id, text, page_offsets) VALUES ($1, $2, $3)`,
		fileID, text, pageOffsets,
	); err != nil {
		t.Fatalf("insert extracted text: %v", err)
	}
	return fileID
}

// TestChunkEmbedWorker_Integration_HappyPath drives the full pipeline
// against real Postgres + pgvector: extracted -> embedding -> ready,
// chunks persisted with their 1536-dim vectors, files_extracted_text
// row deleted afterwards.
func TestChunkEmbedWorker_Integration_HappyPath(t *testing.T) {
	pool, queries := bringUpChunkEmbedPostgres(t)
	ctx := context.Background()
	fileID := seedExtractedFile(t, ctx, pool,
		"First paragraph about recursion.\n\nSecond paragraph with more detail.\n\nThird paragraph closes.",
		nil)

	repo := files.NewChunkEmbedRepository(pool, queries)
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{tokens: 100})

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	var status string
	if err := pool.QueryRow(ctx,
		"SELECT processing_status::text FROM files WHERE id = $1", fileID,
	).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "ready" {
		t.Errorf("processing_status = %q, want %q", status, "ready")
	}

	var chunkCount int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM study_guide_file_chunks WHERE file_id = $1", fileID,
	).Scan(&chunkCount); err != nil {
		t.Fatalf("count chunks: %v", err)
	}
	if chunkCount == 0 {
		t.Errorf("expected at least one chunk persisted, got 0")
	}

	var extractedRows int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM files_extracted_text WHERE file_id = $1", fileID,
	).Scan(&extractedRows); err != nil {
		t.Fatalf("count extracted: %v", err)
	}
	if extractedRows != 0 {
		t.Errorf("files_extracted_text rows = %d, want 0 (cleanup should have run)", extractedRows)
	}
}

// TestChunkEmbedWorker_Integration_Idempotent re-runs Process after a
// successful first pass -- the second pass must no-op (already
// 'ready') AND not duplicate chunks.
func TestChunkEmbedWorker_Integration_Idempotent(t *testing.T) {
	pool, queries := bringUpChunkEmbedPostgres(t)
	ctx := context.Background()
	fileID := seedExtractedFile(t, ctx, pool, "Hello world.\n\nGoodbye world.", nil)

	repo := files.NewChunkEmbedRepository(pool, queries)
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{tokens: 10})

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("first Process: %v", err)
	}

	var firstCount int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM study_guide_file_chunks WHERE file_id = $1", fileID,
	).Scan(&firstCount); err != nil {
		t.Fatalf("count: %v", err)
	}

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("second Process: %v", err)
	}

	var secondCount int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM study_guide_file_chunks WHERE file_id = $1", fileID,
	).Scan(&secondCount); err != nil {
		t.Fatalf("count: %v", err)
	}
	if firstCount != secondCount {
		t.Errorf("chunk count changed across re-run: first=%d, second=%d", firstCount, secondCount)
	}
}

// TestChunkEmbedWorker_Integration_PersistAtomicity inserts a stale
// chunk row for the same file_id and verifies that PersistChunks
// wipes it before inserting -- the production worker's idempotency
// guarantee against partial prior runs.
func TestChunkEmbedWorker_Integration_PersistAtomicity(t *testing.T) {
	pool, queries := bringUpChunkEmbedPostgres(t)
	ctx := context.Background()
	fileID := seedExtractedFile(t, ctx, pool, "Some content here.", nil)

	// Build a 1536-dim "0,0,...,0" vector literal for the stale insert.
	zero := make([]string, 1536)
	for i := range zero {
		zero[i] = "0"
	}
	zeroVec := "[" + strings.Join(zero, ",") + "]"

	if _, err := pool.Exec(ctx,
		`INSERT INTO study_guide_file_chunks (file_id, chunk_idx, text, embedding, tokens)
		 VALUES ($1, 999, 'STALE CHUNK', $2::vector, 1)`,
		fileID, zeroVec,
	); err != nil {
		t.Fatalf("seed stale chunk: %v", err)
	}

	repo := files.NewChunkEmbedRepository(pool, queries)
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{tokens: 10})

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	var staleCount int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM study_guide_file_chunks WHERE file_id = $1 AND text = 'STALE CHUNK'`,
		fileID,
	).Scan(&staleCount); err != nil {
		t.Fatalf("count stale: %v", err)
	}
	if staleCount != 0 {
		t.Errorf("stale chunks remained after PersistChunks: %d", staleCount)
	}
}

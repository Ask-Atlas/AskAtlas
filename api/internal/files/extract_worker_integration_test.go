//go:build integration

// Integration test for the ASK-220 file extraction worker.
//
// Spins up Postgres via testcontainers-go, applies a minimal slice of
// the production schema (just the bits this ticket touches), and
// drives ExtractWorker.Process end-to-end against the real
// sqlc-generated queries. Catches the things unit tests can't:
//   - The migration's ENUM values + ALTER TABLE actually apply.
//   - sqlc-generated queries map ProcessingStatus values correctly.
//   - The UPSERT-on-conflict pattern in UpsertExtractedText is honored.
//   - The worker's idempotency switch reads the same enum value the
//     migration writes (catches casing / serialization bugs).
//
// Build-tagged so default `make test` stays Docker-free for unit
// runs. `make test-integration` opts in via `-tags=integration`. The
// integration-test CI job lands via PR #295 (ASK-219 follow-up).
package files_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// minimalSchema mirrors the ASK-220 migration plus its prerequisites.
// Kept inline rather than running the full migration history so the
// test stays fast + self-contained. If the production migration
// (20260426033127_add_file_processing_status.up.sql) drifts, this
// test should be updated in lockstep -- the deliberate redundancy
// catches schema/code drift the unit tests miss.
const minimalSchema = `
CREATE EXTENSION IF NOT EXISTS pgcrypto;

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
`

func bringUpPostgres(t *testing.T) (*pgxpool.Pool, *db.Queries) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	container, err := tcpostgres.Run(
		ctx,
		"postgres:17",
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

func seedFile(t *testing.T, ctx context.Context, pool *pgxpool.Pool, mime string) uuid.UUID {
	t.Helper()
	userID := uuid.New()
	fileID := uuid.New()
	if _, err := pool.Exec(ctx, "INSERT INTO users (id) VALUES ($1)", userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO files (id, user_id, s3_key, mime_type, status, processing_status)
		 VALUES ($1, $2, $3, $4, 'complete', 'uploaded')`,
		fileID, userID, "uploads/"+fileID.String(), mime,
	); err != nil {
		t.Fatalf("insert file: %v", err)
	}
	return fileID
}

// TestExtractWorker_Integration_PlainText exercises the happy path
// end-to-end: a real files row + the ASK-220 migration + sqlc-generated
// queries + the ExtractWorker. Asserts the row's processing_status
// flips to 'extracted' and a files_extracted_text row appears.
func TestExtractWorker_Integration_PlainText(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	fileID := seedFile(t, ctx, pool, "text/plain")

	repo := files.NewExtractRepository(queries)
	w := files.NewExtractWorker(repo, &fakeDownloader{body: []byte("Hello, integration.")})

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	var status, statusErr pgtype.Text
	if err := pool.QueryRow(ctx,
		"SELECT processing_status::text, status_error FROM files WHERE id = $1", fileID,
	).Scan(&status, &statusErr); err != nil {
		t.Fatalf("read processing_status: %v", err)
	}
	if status.String != "extracted" {
		t.Errorf("processing_status = %q, want %q", status.String, "extracted")
	}
	if statusErr.Valid {
		t.Errorf("status_error = %q, want NULL after success", statusErr.String)
	}

	var text string
	if err := pool.QueryRow(ctx,
		"SELECT text FROM files_extracted_text WHERE file_id = $1", fileID,
	).Scan(&text); err != nil {
		t.Fatalf("read extracted text: %v", err)
	}
	if text != "Hello, integration." {
		t.Errorf("extracted text = %q, want %q", text, "Hello, integration.")
	}
}

// TestExtractWorker_Integration_PDF runs the PDF parser against the
// real testdata fixture, real Postgres, real sqlc UPSERT. Verifies the
// per-page offsets array round-trips correctly through pgx as
// INTEGER[] -- caught at compile time as []int32, this catches it at
// runtime.
func TestExtractWorker_Integration_PDF(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	fileID := seedFile(t, ctx, pool, "application/pdf")

	body, err := os.ReadFile(filepath.Join("testdata", "sample.pdf"))
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	repo := files.NewExtractRepository(queries)
	w := files.NewExtractWorker(repo, &fakeDownloader{body: body})

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("Process: %v", err)
	}

	var offsets []int32
	if err := pool.QueryRow(ctx,
		"SELECT page_offsets FROM files_extracted_text WHERE file_id = $1", fileID,
	).Scan(&offsets); err != nil {
		t.Fatalf("read page_offsets: %v", err)
	}
	if len(offsets) == 0 {
		t.Errorf("page_offsets is empty; PDF extraction should yield at least one page")
	}
	if offsets[0] != 0 {
		t.Errorf("page_offsets[0] = %d, want 0 (first page starts at offset 0)", offsets[0])
	}
}

// TestExtractWorker_Integration_TerminalUnsupportedMime confirms the
// terminal-failure path lands in real Postgres: the row goes to
// processing_status='failed' with a populated status_error. This is
// the bug unit tests can't catch -- if MarkFileProcessingFailed
// serializes wrong, the unit tests don't notice.
func TestExtractWorker_Integration_TerminalUnsupportedMime(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	fileID := seedFile(t, ctx, pool, "application/zip")

	repo := files.NewExtractRepository(queries)
	w := files.NewExtractWorker(repo, &fakeDownloader{body: []byte("doesn't matter")})

	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("Process should swallow terminal failure; got: %v", err)
	}

	var status string
	var statusErr pgtype.Text
	if err := pool.QueryRow(ctx,
		"SELECT processing_status::text, status_error FROM files WHERE id = $1", fileID,
	).Scan(&status, &statusErr); err != nil {
		t.Fatalf("read row: %v", err)
	}
	if status != "failed" {
		t.Errorf("processing_status = %q, want %q", status, "failed")
	}
	if !statusErr.Valid || statusErr.String == "" {
		t.Errorf("status_error should be non-empty after terminal failure; got valid=%v str=%q",
			statusErr.Valid, statusErr.String)
	}
}

// TestExtractWorker_Integration_Idempotent confirms a second Process
// call against an already-extracted row no-ops -- no duplicate rows in
// files_extracted_text, no transition back to extracting. Unit tests
// cover the in-memory branch; this verifies the SQL UPSERT path in
// real Postgres.
func TestExtractWorker_Integration_Idempotent(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	fileID := seedFile(t, ctx, pool, "text/plain")

	repo := files.NewExtractRepository(queries)
	w := files.NewExtractWorker(repo, &fakeDownloader{body: []byte("first run")})
	if err := w.Process(ctx, fileID); err != nil {
		t.Fatalf("first Process: %v", err)
	}

	// Second run: the worker sees processing_status='extracted' and
	// no-ops. Even if it didn't, UPSERT keeps row count at 1.
	w2 := files.NewExtractWorker(repo, &fakeDownloader{body: []byte("ignored")})
	if err := w2.Process(ctx, fileID); err != nil {
		t.Fatalf("second Process: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM files_extracted_text WHERE file_id = $1", fileID,
	).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("files_extracted_text rows = %d, want 1", count)
	}

	// Text should still be the FIRST run's payload -- the second run
	// should have skipped, not overwritten with "ignored".
	var text string
	if err := pool.QueryRow(ctx,
		"SELECT text FROM files_extracted_text WHERE file_id = $1", fileID,
	).Scan(&text); err != nil {
		t.Fatalf("read text: %v", err)
	}
	if text != "first run" {
		t.Errorf("text = %q, want %q (second run should skip, not overwrite)", text, "first run")
	}
}

// TestExtractWorker_Integration_TransientPropagates confirms that
// transient I/O errors propagate up so QStash retries, and the row is
// left in 'extracting' (not 'failed'), so a retry can resume.
func TestExtractWorker_Integration_TransientPropagates(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	fileID := seedFile(t, ctx, pool, "text/plain")

	repo := files.NewExtractRepository(queries)
	w := files.NewExtractWorker(repo, &fakeDownloader{err: errors.New("connection refused")})

	if err := w.Process(ctx, fileID); err == nil {
		t.Fatal("expected transient error to propagate, got nil")
	}

	var status string
	if err := pool.QueryRow(ctx,
		"SELECT processing_status::text FROM files WHERE id = $1", fileID,
	).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "extracting" {
		t.Errorf("processing_status = %q, want %q (transient should leave row mid-pipeline)",
			status, "extracting")
	}
}

// TestExtractWorker_Integration_CascadeDeleteOnFile verifies that
// dropping a files row also wipes its files_extracted_text row -- the
// migration's ON DELETE CASCADE actually fires. Catches a class of
// missing-FK bugs that a unit test can't see.
func TestExtractWorker_Integration_CascadeDeleteOnFile(t *testing.T) {
	pool, queries := bringUpPostgres(t)
	ctx := context.Background()
	fileID := seedFile(t, ctx, pool, "text/plain")

	if err := queries.UpsertExtractedText(ctx, db.UpsertExtractedTextParams{
		FileID:      utils.UUID(fileID),
		Text:        "to be cascaded",
		PageOffsets: []int32{0},
	}); err != nil {
		t.Fatalf("UpsertExtractedText: %v", err)
	}

	if _, err := pool.Exec(ctx, "DELETE FROM files WHERE id = $1", fileID); err != nil {
		t.Fatalf("delete file: %v", err)
	}

	var count int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM files_extracted_text WHERE file_id = $1", fileID,
	).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Errorf("post-delete files_extracted_text rows = %d, want 0 (cascade should wipe)", count)
	}
}

package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

// Embedder is the slice of ai.Client the chunk+embed worker needs.
// Defined here so the worker package owns its dependency contract
// (and depends only on the single embedding method it uses). Test
// doubles still need the ai package because the method signature
// uses ai.EmbedRequest / ai.EmbedResponse -- a fully ai-free
// signature would mean re-implementing those request/response shapes
// inside files/, which is not worth the extra layer.
type Embedder interface {
	Embed(ctx context.Context, req ai.EmbedRequest) (ai.EmbedResponse, error)
}

// ChunkEmbedRepository is the narrow data-access surface the
// chunk+embed worker uses. Kept narrow so the test fake stays small;
// the production adapter wraps *db.Queries + a pgxpool for the
// transactional persist step.
type ChunkEmbedRepository interface {
	GetFileForExtraction(ctx context.Context, fileID uuid.UUID) (db.GetFileForExtractionRow, error)
	GetExtractedText(ctx context.Context, fileID uuid.UUID) (db.GetExtractedTextRow, error)
	SetFileProcessingStatus(ctx context.Context, fileID uuid.UUID, status db.ProcessingStatus) error
	MarkFileProcessingFailed(ctx context.Context, fileID uuid.UUID, statusError string) error
	// PersistChunks atomically wipes any existing chunks for fileID
	// then inserts the new set. Implementation runs under a tx so a
	// crash mid-write can't leave partial chunks.
	PersistChunks(ctx context.Context, fileID uuid.UUID, params []db.InsertStudyGuideFileChunkParams) error
	DeleteExtractedText(ctx context.Context, fileID uuid.UUID) error
}

// ChunkEmbedWorker drives one chunk+embed job (ASK-221). Stateless
// across calls -- every Process invocation re-resolves its row.
type ChunkEmbedWorker struct {
	repo     ChunkEmbedRepository
	embedder Embedder
}

// NewChunkEmbedWorker constructs the worker with its required deps.
func NewChunkEmbedWorker(repo ChunkEmbedRepository, embedder Embedder) *ChunkEmbedWorker {
	return &ChunkEmbedWorker{repo: repo, embedder: embedder}
}

// Process runs a single chunk+embed job. Idempotency contract:
//   - already past embedding (ready / failed): no-op success.
//   - vanished file (ErrNotFound on GetFileForExtraction): terminal
//     success so QStash stops retrying.
//   - missing extracted_text row (extract step ran but the cleanup
//     happened before us, or the row was wiped): mark file failed
//     with an explicit reason and return nil. Not retryable.
//
// Transient errors (DB write, OpenAI 5xx after retry budget) return
// the wrapped error so QStash retries. Terminal errors return nil.
func (w *ChunkEmbedWorker) Process(ctx context.Context, fileID uuid.UUID) error {
	row, err := w.repo.GetFileForExtraction(ctx, fileID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			slog.Warn("chunk-embed worker: file vanished; skipping",
				"file_id", fileID)
			return nil
		}
		return fmt.Errorf("ChunkEmbedWorker.Process: load file: %w", err)
	}

	switch row.ProcessingStatus {
	case db.ProcessingStatusReady, db.ProcessingStatusFailed:
		slog.Debug("chunk-embed worker: already past embed stage",
			"file_id", fileID, "processing_status", row.ProcessingStatus)
		return nil
	}

	extracted, err := w.repo.GetExtractedText(ctx, fileID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			// The extracted-text row is the handoff. If it's missing,
			// either the extract worker never ran or this worker
			// already cleaned up. Either way we can't proceed; mark
			// failed terminally so the user sees something actionable
			// instead of a row stuck in 'extracted'.
			reason := "chunk+embed: missing extracted text"
			if markErr := w.repo.MarkFileProcessingFailed(ctx, fileID, reason); markErr != nil {
				return fmt.Errorf("ChunkEmbedWorker.Process: mark failed: %w", markErr)
			}
			slog.Warn("chunk-embed worker: missing extracted text",
				"file_id", fileID)
			return nil
		}
		return fmt.Errorf("ChunkEmbedWorker.Process: load extracted text: %w", err)
	}

	if err := w.repo.SetFileProcessingStatus(ctx, fileID, db.ProcessingStatusEmbedding); err != nil {
		return fmt.Errorf("ChunkEmbedWorker.Process: mark embedding: %w", err)
	}

	chunks := Chunk(extracted.Text, extracted.PageOffsets)
	if len(chunks) == 0 {
		reason := "chunk+embed: chunker emitted zero chunks"
		if markErr := w.repo.MarkFileProcessingFailed(ctx, fileID, reason); markErr != nil {
			return fmt.Errorf("ChunkEmbedWorker.Process: mark failed: %w", markErr)
		}
		slog.Warn("chunk-embed worker: zero chunks", "file_id", fileID)
		return nil
	}

	ownerID, err := utils.PgxToGoogleUUID(row.UserID)
	if err != nil {
		return fmt.Errorf("ChunkEmbedWorker.Process: decode owner: %w", err)
	}

	inputs := make([]string, len(chunks))
	for i, c := range chunks {
		inputs[i] = c.Text
	}

	embed, err := w.embedder.Embed(ctx, ai.EmbedRequest{
		UserID:  ownerID,
		Feature: ai.FeatureEmbedding,
		Model:   ai.EmbeddingModelDefault,
		Inputs:  inputs,
	})
	if err != nil {
		// Treat upstream failure as transient: QStash retries; the
		// failure-callback path eventually marks failed if the budget
		// exhausts. We deliberately leave processing_status='embedding'
		// so the next attempt's idempotency check correctly resumes.
		return fmt.Errorf("ChunkEmbedWorker.Process: embed: %w", err)
	}
	if len(embed.Vectors) != len(chunks) {
		return fmt.Errorf("ChunkEmbedWorker.Process: vector count mismatch: got %d, want %d",
			len(embed.Vectors), len(chunks))
	}

	params := make([]db.InsertStudyGuideFileChunkParams, len(chunks))
	for i, c := range chunks {
		params[i] = db.InsertStudyGuideFileChunkParams{
			FileID:    utils.UUID(fileID),
			ChunkIdx:  c.ChunkIdx,
			Text:      c.Text,
			Embedding: pgvector.NewVector(embed.Vectors[i]),
			Page:      pgInt4Ptr(c.Page),
			Heading:   pgTextPtr(c.Heading),
			Tokens:    c.Tokens,
		}
	}

	if err := w.repo.PersistChunks(ctx, fileID, params); err != nil {
		return fmt.Errorf("ChunkEmbedWorker.Process: persist chunks: %w", err)
	}
	if err := w.repo.SetFileProcessingStatus(ctx, fileID, db.ProcessingStatusReady); err != nil {
		return fmt.Errorf("ChunkEmbedWorker.Process: mark ready: %w", err)
	}

	// Best-effort cleanup of the transient extracted-text row. A
	// failure here is non-fatal: the row will linger but doesn't
	// affect retrieval correctness, and a re-run of this worker is
	// idempotent (delete-then-insert chunks).
	if err := w.repo.DeleteExtractedText(ctx, fileID); err != nil {
		slog.Warn("chunk-embed worker: cleanup of extracted_text failed",
			"file_id", fileID, "error", err)
	}

	slog.Info("chunk-embed worker: done",
		"file_id", fileID,
		"chunks", len(chunks),
		"total_tokens", embed.Usage.InputTokens,
	)
	return nil
}

// MarkFailed is invoked from the QStash failure-callback path when
// retries are exhausted. Mirrors ExtractWorker.MarkFailed so the
// JobHandler can satisfy a single FileExtractor-like contract for
// both workers.
func (w *ChunkEmbedWorker) MarkFailed(ctx context.Context, fileID uuid.UUID, reason string) error {
	if err := w.repo.MarkFileProcessingFailed(ctx, fileID, reason); err != nil {
		return fmt.Errorf("ChunkEmbedWorker.MarkFailed: %w", err)
	}
	return nil
}

func pgInt4Ptr(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func pgTextPtr(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}

// chunkEmbedRepoAdapter bridges *db.Queries + a pgxpool to the
// narrow ChunkEmbedRepository surface. PersistChunks runs under a
// pgx tx so the delete-then-insert sequence is atomic.
type chunkEmbedRepoAdapter struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewChunkEmbedRepository constructs the production adapter.
func NewChunkEmbedRepository(pool *pgxpool.Pool, queries *db.Queries) ChunkEmbedRepository {
	return &chunkEmbedRepoAdapter{pool: pool, queries: queries}
}

func (r *chunkEmbedRepoAdapter) GetFileForExtraction(ctx context.Context, fileID uuid.UUID) (db.GetFileForExtractionRow, error) {
	row, err := r.queries.GetFileForExtraction(ctx, utils.UUID(fileID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetFileForExtractionRow{}, fmt.Errorf("GetFileForExtraction: %w", apperrors.ErrNotFound)
		}
		return db.GetFileForExtractionRow{}, fmt.Errorf("GetFileForExtraction: %w", err)
	}
	return row, nil
}

func (r *chunkEmbedRepoAdapter) GetExtractedText(ctx context.Context, fileID uuid.UUID) (db.GetExtractedTextRow, error) {
	row, err := r.queries.GetExtractedText(ctx, utils.UUID(fileID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetExtractedTextRow{}, fmt.Errorf("GetExtractedText: %w", apperrors.ErrNotFound)
		}
		return db.GetExtractedTextRow{}, fmt.Errorf("GetExtractedText: %w", err)
	}
	return row, nil
}

func (r *chunkEmbedRepoAdapter) SetFileProcessingStatus(ctx context.Context, fileID uuid.UUID, status db.ProcessingStatus) error {
	return r.queries.SetFileProcessingStatus(ctx, db.SetFileProcessingStatusParams{
		ProcessingStatus: status,
		FileID:           utils.UUID(fileID),
	})
}

func (r *chunkEmbedRepoAdapter) MarkFileProcessingFailed(ctx context.Context, fileID uuid.UUID, statusError string) error {
	return r.queries.MarkFileProcessingFailed(ctx, db.MarkFileProcessingFailedParams{
		StatusError: statusError,
		FileID:      utils.UUID(fileID),
	})
}

func (r *chunkEmbedRepoAdapter) PersistChunks(ctx context.Context, fileID uuid.UUID, params []db.InsertStudyGuideFileChunkParams) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("PersistChunks: begin tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Error("PersistChunks: rollback failed", "error", rollbackErr)
		}
	}()

	q := r.queries.WithTx(tx)
	if err := q.DeleteChunksByFile(ctx, utils.UUID(fileID)); err != nil {
		return fmt.Errorf("PersistChunks: delete existing: %w", err)
	}
	for _, p := range params {
		if _, err := q.InsertStudyGuideFileChunk(ctx, p); err != nil {
			return fmt.Errorf("PersistChunks: insert chunk %d: %w", p.ChunkIdx, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("PersistChunks: commit: %w", err)
	}
	return nil
}

func (r *chunkEmbedRepoAdapter) DeleteExtractedText(ctx context.Context, fileID uuid.UUID) error {
	return r.queries.DeleteExtractedText(ctx, utils.UUID(fileID))
}

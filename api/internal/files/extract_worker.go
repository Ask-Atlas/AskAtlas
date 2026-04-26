package files

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
)

// ExtractRepository is the slice of the files repository the extract
// worker needs. Keeping it narrow (4 methods) makes the test double
// trivial and decouples the worker from the much larger Repository
// interface that the request-path Service uses.
type ExtractRepository interface {
	GetFileForExtraction(ctx context.Context, fileID uuid.UUID) (db.GetFileForExtractionRow, error)
	SetFileProcessingStatus(ctx context.Context, fileID uuid.UUID, status db.ProcessingStatus) error
	MarkFileProcessingFailed(ctx context.Context, fileID uuid.UUID, statusError string) error
	UpsertExtractedText(ctx context.Context, fileID uuid.UUID, text string, pageOffsets []int32) error
}

// S3Downloader fetches an object's body bytes by key. The concrete
// s3client.Client satisfies this; a fake satisfies it in tests.
type S3Downloader interface {
	GetObject(ctx context.Context, key string) ([]byte, error)
}

// ExtractWorker drives a single extract job from the QStash callback.
// Stateless across calls -- every Process invocation resolves its own
// row, downloads, parses, and persists.
type ExtractWorker struct {
	repo       ExtractRepository
	downloader S3Downloader
}

// NewExtractWorker constructs the worker with its required collaborators.
func NewExtractWorker(repo ExtractRepository, downloader S3Downloader) *ExtractWorker {
	return &ExtractWorker{repo: repo, downloader: downloader}
}

// Process executes one extract job. Idempotency contract:
//   - already past extracting (extracted/embedding/ready): no-op success.
//   - failed: no-op success (do not retry indefinitely; ASK-222 surfaces
//     the failure and a re-upload restarts the pipeline).
//   - uploaded/extracting: do the work.
//
// On terminal parse failure (ErrUnsupportedMimeType, ErrEmptyExtraction)
// the row is marked failed with an explicit status_error. On transient
// I/O failure (S3 GetObject error, DB error) Process returns the error
// so QStash retries; the row stays in `extracting`.
func (w *ExtractWorker) Process(ctx context.Context, fileID uuid.UUID) error {
	row, err := w.repo.GetFileForExtraction(ctx, fileID)
	if err != nil {
		return fmt.Errorf("ExtractWorker.Process: load: %w", err)
	}

	switch row.ProcessingStatus {
	case db.ProcessingStatusExtracted,
		db.ProcessingStatusEmbedding,
		db.ProcessingStatusReady,
		db.ProcessingStatusFailed:
		slog.Debug("extract worker: skipping, already past extract stage",
			"file_id", fileID, "processing_status", row.ProcessingStatus)
		return nil
	}

	if err := w.repo.SetFileProcessingStatus(ctx, fileID, db.ProcessingStatusExtracting); err != nil {
		return fmt.Errorf("ExtractWorker.Process: mark extracting: %w", err)
	}

	body, err := w.downloader.GetObject(ctx, row.S3Key)
	if err != nil {
		// Transient: let QStash retry. We deliberately leave
		// processing_status='extracting' rather than flip back to
		// 'uploaded' -- a retry will re-enter Process and the existing
		// idempotency switch above lets the in-flight extracting row
		// continue.
		return fmt.Errorf("ExtractWorker.Process: download: %w", err)
	}

	doc, err := ExtractText(body, row.MimeType)
	if err != nil {
		// Terminal: mark failed and swallow the error so QStash does
		// not retry. The status_error field carries the cause for the
		// frontend (ASK-222) to render.
		if errors.Is(err, ErrUnsupportedMimeType) || errors.Is(err, ErrEmptyExtraction) {
			if markErr := w.repo.MarkFileProcessingFailed(ctx, fileID, err.Error()); markErr != nil {
				return fmt.Errorf("ExtractWorker.Process: mark failed: %w", markErr)
			}
			slog.Warn("extract worker: terminal failure",
				"file_id", fileID, "mime_type", row.MimeType, "error", err)
			return nil
		}
		// Otherwise (corrupt PDF, parser bug) -- treat as transient
		// once; QStash retries will eventually call extract-file-failed
		// which the failure handler maps to terminal `failed`.
		return fmt.Errorf("ExtractWorker.Process: parse: %w", err)
	}

	if err := w.repo.UpsertExtractedText(ctx, fileID, doc.Text, doc.PageOffsets); err != nil {
		return fmt.Errorf("ExtractWorker.Process: persist text: %w", err)
	}

	if err := w.repo.SetFileProcessingStatus(ctx, fileID, db.ProcessingStatusExtracted); err != nil {
		return fmt.Errorf("ExtractWorker.Process: mark extracted: %w", err)
	}

	// Note: the chunk+embed enqueue step lands in ASK-221. That ticket
	// will either add a publisher hook here or run a backfill scan over
	// processing_status='extracted' rows. We deliberately do not
	// publish to a topic with no consumer in this PR -- doing so would
	// trigger the QStash failure callback in a steady-state loop.

	slog.Info("extract worker: done",
		"file_id", fileID,
		"mime_type", row.MimeType,
		"text_length", len(doc.Text),
		"pages", len(doc.PageOffsets),
	)
	return nil
}

// MarkFailed is invoked from the QStash failure-callback path when
// retries are exhausted. Separate from the in-band terminal-failure
// path inside Process: at that point we know nothing structured about
// why retries failed (just that they did), so the message is generic.
func (w *ExtractWorker) MarkFailed(ctx context.Context, fileID uuid.UUID, reason string) error {
	if err := w.repo.MarkFileProcessingFailed(ctx, fileID, reason); err != nil {
		return fmt.Errorf("ExtractWorker.MarkFailed: %w", err)
	}
	return nil
}

// extractRepoAdapter adapts *db.Queries (the sqlc handle) to the
// narrow ExtractRepository the worker needs. Keeps the wide interface
// from leaking into the worker tests.
type extractRepoAdapter struct {
	queries *db.Queries
}

// NewExtractRepository wraps a *db.Queries (the sqlc handle) so the
// worker can call its narrow query set without going through the full
// files.Repository machinery (which is request-scoped).
func NewExtractRepository(queries *db.Queries) ExtractRepository {
	return &extractRepoAdapter{queries: queries}
}

func (r *extractRepoAdapter) GetFileForExtraction(ctx context.Context, fileID uuid.UUID) (db.GetFileForExtractionRow, error) {
	row, err := r.queries.GetFileForExtraction(ctx, utils.UUID(fileID))
	if err != nil {
		return db.GetFileForExtractionRow{}, fmt.Errorf("GetFileForExtraction: %w", err)
	}
	return row, nil
}

func (r *extractRepoAdapter) SetFileProcessingStatus(ctx context.Context, fileID uuid.UUID, status db.ProcessingStatus) error {
	return r.queries.SetFileProcessingStatus(ctx, db.SetFileProcessingStatusParams{
		ProcessingStatus: status,
		FileID:           utils.UUID(fileID),
	})
}

func (r *extractRepoAdapter) MarkFileProcessingFailed(ctx context.Context, fileID uuid.UUID, statusError string) error {
	return r.queries.MarkFileProcessingFailed(ctx, db.MarkFileProcessingFailedParams{
		StatusError: statusError,
		FileID:      utils.UUID(fileID),
	})
}

func (r *extractRepoAdapter) UpsertExtractedText(ctx context.Context, fileID uuid.UUID, text string, pageOffsets []int32) error {
	return r.queries.UpsertExtractedText(ctx, db.UpsertExtractedTextParams{
		FileID:      utils.UUID(fileID),
		Text:        text,
		PageOffsets: pageOffsets,
	})
}

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
)

// FileExtractor processes one extract job and (separately) marks a
// job terminally failed when QStash exhausts retries.
type FileExtractor interface {
	Process(ctx context.Context, fileID uuid.UUID) error
	MarkFailed(ctx context.Context, fileID uuid.UUID, reason string) error
}

// JobHandler handles async job requests from QStash.
type JobHandler struct {
	s3        *s3client.Client
	db        *db.Queries
	extractor FileExtractor
}

// NewJobHandler creates a JobHandler with S3, DB, and extract-worker
// dependencies. The extractor is required for the ASK-220 extract-file
// route; pass a real *files.ExtractWorker in main, a fake in tests.
func NewJobHandler(s3 *s3client.Client, queries *db.Queries, extractor FileExtractor) *JobHandler {
	return &JobHandler{s3: s3, db: queries, extractor: extractor}
}

// DeleteFileJob handles POST /jobs/delete-file.
// It deletes the S3 object and marks the file as deleted in the DB.
func (h *JobHandler) DeleteFileJob(w http.ResponseWriter, r *http.Request) {
	var body qstashclient.DeleteFileMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("invalid request body", nil))
		return
	}

	fileID, err := uuid.Parse(body.FileID)
	if err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("invalid file_id", nil))
		return
	}

	if err := h.s3.DeleteObject(r.Context(), body.S3Key); err != nil {
		slog.Error("DeleteFileJob: S3 deletion failed",
			"file_id", body.FileID,
			"s3_key", body.S3Key,
			"environment", body.Environment,
			"error", err,
		)
		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	if err := h.db.MarkFileDeleted(r.Context(), utils.UUID(fileID)); err != nil {
		slog.Error("DeleteFileJob: failed to mark file deleted in DB",
			"file_id", body.FileID,
			"environment", body.Environment,
			"error", err,
		)
		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteFileFailedJob handles POST /jobs/delete-file-failed.
// Called by QStash when the delete-file job fails after all retries.
func (h *JobHandler) DeleteFileFailedJob(w http.ResponseWriter, r *http.Request) {
	var body qstashclient.DeleteFileMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("DeleteFileFailedJob: failed to parse body", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	slog.Error("DeleteFileFailedJob: file S3 deletion failed after all retries",
		"file_id", body.FileID,
		"s3_key", body.S3Key,
		"user_id", body.UserID,
		"environment", body.Environment,
		"requested_at", body.RequestedAt,
	)

	w.WriteHeader(http.StatusOK)
}

// ExtractFileJob handles POST /jobs/extract-file (ASK-220). Decodes
// the QStash payload, delegates to the extract worker. A non-nil
// error from the worker yields a 500 so QStash retries; nil means
// either the job succeeded or hit a terminal-failure path that the
// worker already recorded -- either way QStash should not retry.
func (h *JobHandler) ExtractFileJob(w http.ResponseWriter, r *http.Request) {
	var body qstashclient.ExtractFileMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("invalid request body", nil))
		return
	}

	fileID, err := uuid.Parse(body.FileID)
	if err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("invalid file_id", nil))
		return
	}

	if err := h.extractor.Process(r.Context(), fileID); err != nil {
		slog.Error("ExtractFileJob: extract failed",
			"file_id", body.FileID,
			"s3_key", body.S3Key,
			"mime_type", body.MimeType,
			"environment", body.Environment,
			"error", err,
		)
		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ExtractFileFailedJob handles POST /jobs/extract-file-failed. Called
// by QStash when the extract-file job exhausts retries; we mark the
// row terminally failed so the user sees the failure in the file UI
// (ASK-222) instead of a permanently 'extracting' state.
func (h *JobHandler) ExtractFileFailedJob(w http.ResponseWriter, r *http.Request) {
	var body qstashclient.ExtractFileMessage
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("ExtractFileFailedJob: failed to parse body", "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	fileID, err := uuid.Parse(body.FileID)
	if err != nil {
		slog.Error("ExtractFileFailedJob: invalid file_id", "file_id", body.FileID, "error", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	reason := fmt.Sprintf("extract job exhausted retries (mime=%s)", body.MimeType)
	if err := h.extractor.MarkFailed(r.Context(), fileID, reason); err != nil {
		// We still 200 -- QStash has already given up; returning 5xx
		// just spawns more failure-callback retries. The error is
		// captured in logs for ops follow-up.
		slog.Error("ExtractFileFailedJob: failed to mark row failed",
			"file_id", body.FileID,
			"environment", body.Environment,
			"error", err,
		)
	}

	slog.Error("ExtractFileFailedJob: extract job exhausted retries",
		"file_id", body.FileID,
		"s3_key", body.S3Key,
		"user_id", body.UserID,
		"environment", body.Environment,
		"requested_at", body.RequestedAt,
	)

	w.WriteHeader(http.StatusOK)
}

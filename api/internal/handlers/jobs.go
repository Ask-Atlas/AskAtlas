package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	s3client "github.com/Ask-Atlas/AskAtlas/api/internal/s3"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
)

// JobHandler handles async job requests from QStash.
type JobHandler struct {
	s3 *s3client.Client
	db *db.Queries
}

// NewJobHandler creates a JobHandler with S3 and DB clients.
func NewJobHandler(s3 *s3client.Client, queries *db.Queries) *JobHandler {
	return &JobHandler{s3: s3, db: queries}
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

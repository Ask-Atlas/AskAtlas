package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
)

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

func parseGetFileParams(r *http.Request) (files.GetFileParams, *apperrors.AppError) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		return files.GetFileParams{}, appErr
	}

	rawID := r.PathValue("file_id")
	fileID, err := uuid.Parse(rawID)
	if err != nil {
		return files.GetFileParams{}, apperrors.NewBadRequest("invalid file_id", map[string]string{
			"file_id": "must be a valid UUID",
		})
	}

	return files.GetFileParams{
		ViewerID: viewerID,
		FileID:   fileID,
	}, nil
}

// parseListFilesParams delegates to the files package parser, injecting the
// viewer identity and ACL IDs sourced from the request context.
func parseListFilesParams(r *http.Request) (*files.ListFilesParams, *apperrors.AppError) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		return nil, appErr
	}

	// TODO: populate courseIDs and studyGuideIDs from ACL middleware context.
	return files.ParseListFilesParams(r, viewerID, nil, nil)
}

func viewerIDFromContext(r *http.Request) (uuid.UUID, *apperrors.AppError) {
	id, ok := authctx.UserIDFromContext(r.Context())
	if !ok {
		return uuid.UUID{}, apperrors.NewUnauthorized()
	}
	return id, nil
}

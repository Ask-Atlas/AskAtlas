package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
)

// respondJSON serializes the given payload to JSON and writes it to the response with the specified status code.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

// viewerIDFromContext extracts the authenticated user ID from the request context.
// It returns an Unauthorized AppError if the context does not contain a valid user ID.
func viewerIDFromContext(r *http.Request) (uuid.UUID, *apperrors.AppError) {
	id, ok := authctx.UserIDFromContext(r.Context())
	if !ok {
		return uuid.UUID{}, apperrors.NewUnauthorized()
	}
	return id, nil
}

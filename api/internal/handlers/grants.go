package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// GrantService defines the application logic required by the GrantHandler.
type GrantService interface {
	CreateGrant(ctx context.Context, params files.CreateGrantParams) (files.Grant, error)
	RevokeGrant(ctx context.Context, params files.RevokeGrantParams) error
}

// GrantHandler manages incoming HTTP requests relating to file grant operations.
type GrantHandler struct {
	service GrantService
}

// NewGrantHandler creates a new GrantHandler backed by the given GrantService.
func NewGrantHandler(service GrantService) *GrantHandler {
	return &GrantHandler{service: service}
}

// CreateGrant handles requests to create a permission grant on a file.
func (h *GrantHandler) CreateGrant(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.CreateGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := files.CreateGrantParams{
		FileID:      uuid.UUID(fileId),
		OwnerID:     viewerID,
		GranteeType: string(body.GranteeType),
		GranteeID:   uuid.UUID(body.GranteeId),
		Permission:  string(body.Permission),
	}

	grant, err := h.service.CreateGrant(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CreateGrant failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, toDTOGrantResponse(grant))
}

// RevokeGrant handles requests to revoke a permission grant on a file.
func (h *GrantHandler) RevokeGrant(w http.ResponseWriter, r *http.Request, fileId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.RevokeGrantRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := files.RevokeGrantParams{
		FileID:      uuid.UUID(fileId),
		OwnerID:     viewerID,
		GranteeType: string(body.GranteeType),
		GranteeID:   uuid.UUID(body.GranteeId),
		Permission:  string(body.Permission),
	}

	if err := h.service.RevokeGrant(r.Context(), params); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("RevokeGrant failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// toDTOGrantResponse converts a domain Grant object into the OpenAPI DTO GrantResponse format.
func toDTOGrantResponse(g files.Grant) api.GrantResponse {
	return api.GrantResponse{
		Id:          openapi_types.UUID(g.ID),
		FileId:      openapi_types.UUID(g.FileID),
		GranteeType: g.GranteeType,
		GranteeId:   openapi_types.UUID(g.GranteeID),
		Permission:  g.Permission,
		GrantedBy:   openapi_types.UUID(g.GrantedBy),
		CreatedAt:   g.CreatedAt,
	}
}

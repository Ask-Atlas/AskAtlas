package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// StudyGuideGrantService defines the application logic required by
// StudyGuideGrantHandler (ASK-211).
type StudyGuideGrantService interface {
	CreateGrant(ctx context.Context, params studyguides.CreateGrantParams) (studyguides.Grant, error)
	RevokeGrant(ctx context.Context, params studyguides.RevokeGrantParams) error
	ListGrants(ctx context.Context, params studyguides.ListGrantsParams) ([]studyguides.Grant, error)
}

// StudyGuideGrantHandler manages HTTP requests for the study-guide
// grants surface (ASK-211). Mirrors GrantHandler (file grants) but
// targets the study_guide_grants table.
type StudyGuideGrantHandler struct {
	service StudyGuideGrantService
}

// NewStudyGuideGrantHandler creates a new StudyGuideGrantHandler
// backed by the given StudyGuideGrantService.
func NewStudyGuideGrantHandler(service StudyGuideGrantService) *StudyGuideGrantHandler {
	return &StudyGuideGrantHandler{service: service}
}

// CreateStudyGuideGrant handles POST /study-guides/{study_guide_id}/grants.
func (h *StudyGuideGrantHandler) CreateStudyGuideGrant(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.CreateStudyGuideGrantJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := studyguides.CreateGrantParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
		GranteeType:  string(body.GranteeType),
		GranteeID:    uuid.UUID(body.GranteeId),
		Permission:   string(body.Permission),
	}

	grant, err := h.service.CreateGrant(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CreateStudyGuideGrant failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapStudyGuideGrantResponse(grant))
}

// RevokeStudyGuideGrant handles DELETE /study-guides/{study_guide_id}/grants.
func (h *StudyGuideGrantHandler) RevokeStudyGuideGrant(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.RevokeStudyGuideGrantJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := studyguides.RevokeGrantParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
		GranteeType:  string(body.GranteeType),
		GranteeID:    uuid.UUID(body.GranteeId),
		Permission:   string(body.Permission),
	}

	if err := h.service.RevokeGrant(r.Context(), params); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("RevokeStudyGuideGrant failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListStudyGuideGrants handles GET /study-guides/{study_guide_id}/grants.
func (h *StudyGuideGrantHandler) ListStudyGuideGrants(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	grants, err := h.service.ListGrants(r.Context(), studyguides.ListGrantsParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		ViewerID:     viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListStudyGuideGrants failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	out := make([]api.StudyGuideGrantResponse, 0, len(grants))
	for _, g := range grants {
		out = append(out, mapStudyGuideGrantResponse(g))
	}
	respondJSON(w, http.StatusOK, api.ListStudyGuideGrantsResponse{Grants: out})
}

// mapStudyGuideGrantResponse projects a domain studyguides.Grant onto
// the wire api.StudyGuideGrantResponse shape.
func mapStudyGuideGrantResponse(g studyguides.Grant) api.StudyGuideGrantResponse {
	return api.StudyGuideGrantResponse{
		Id:           openapi_types.UUID(g.ID),
		StudyGuideId: openapi_types.UUID(g.StudyGuideID),
		GranteeType:  g.GranteeType,
		GranteeId:    openapi_types.UUID(g.GranteeID),
		Permission:   g.Permission,
		GrantedBy:    openapi_types.UUID(g.GrantedBy),
		CreatedAt:    g.CreatedAt,
	}
}

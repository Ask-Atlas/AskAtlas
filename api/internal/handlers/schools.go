package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/schools"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// SchoolService defines the application logic required by the SchoolsHandler.
type SchoolService interface {
	ListSchools(ctx context.Context, params schools.ListSchoolsParams) (schools.ListSchoolsResult, error)
}

// SchoolsHandler manages incoming HTTP requests relating to school operations.
type SchoolsHandler struct {
	service SchoolService
}

// NewSchoolsHandler creates a new SchoolsHandler backed by the given SchoolService.
func NewSchoolsHandler(service SchoolService) *SchoolsHandler {
	return &SchoolsHandler{service: service}
}

// ListSchools handles GET /schools, returning a paginated list of schools
// optionally filtered by the q search term.
func (h *SchoolsHandler) ListSchools(w http.ResponseWriter, r *http.Request, params api.ListSchoolsParams) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var cursor *schools.Cursor
	if params.Cursor != nil && *params.Cursor != "" {
		decoded, err := schools.DecodeCursor(*params.Cursor)
		if err != nil {
			apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			}))
			return
		}
		cursor = &decoded
	}

	// Defense-in-depth: clamp before the int32 narrowing cast so a pathological
	// value can't wrap. The openapi validator caps page_limit at 100 at the HTTP
	// boundary and the service layer also clamps; this is the third line of
	// defense for in-process callers and any future routes bypassing the validator.
	var limit int32
	if params.PageLimit != nil {
		v := *params.PageLimit
		if v > int(schools.MaxPageLimit) {
			v = int(schools.MaxPageLimit)
		}
		limit = int32(v)
	}

	result, err := h.service.ListSchools(r.Context(), schools.ListSchoolsParams{
		Q:      params.Q,
		Limit:  limit,
		Cursor: cursor,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListSchools failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListSchoolsResponse(result))
}

// GetSchool is a temporary stub so SchoolsHandler satisfies the widened
// api.ServerInterface introduced by the new /schools/{school_id} operation.
// Replaced in a follow-up commit when the service path is wired.
func (*SchoolsHandler) GetSchool(w http.ResponseWriter, _ *http.Request, _ openapi_types.UUID) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

// mapListSchoolsResponse converts the domain ListSchoolsResult into the
// generated api.ListSchoolsResponse wire type.
func mapListSchoolsResponse(r schools.ListSchoolsResult) api.ListSchoolsResponse {
	out := make([]api.SchoolResponse, 0, len(r.Schools))
	for _, s := range r.Schools {
		out = append(out, api.SchoolResponse{
			Id:        openapi_types.UUID(s.ID),
			Name:      s.Name,
			Acronym:   s.Acronym,
			Domain:    s.Domain,
			Url:       s.URL,
			City:      s.City,
			State:     s.State,
			Country:   s.Country,
			CreatedAt: s.CreatedAt,
		})
	}
	return api.ListSchoolsResponse{
		Schools:    out,
		HasMore:    r.HasMore,
		NextCursor: r.NextCursor,
	}
}

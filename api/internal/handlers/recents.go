package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/recents"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

// RecentsService is the slice of the recents service surface this
// handler depends on. Defined here (where it is used) so the handler
// can be unit-tested with a mockery-generated mock without dragging
// in the full domain package.
type RecentsService interface {
	ListRecents(ctx context.Context, p recents.ListRecentsParams) (recents.ListRecentsResult, error)
}

// RecentsHandler serves GET /api/me/recents (ASK-145).
type RecentsHandler struct {
	service RecentsService
}

// NewRecentsHandler wires the handler over the given service.
func NewRecentsHandler(service RecentsService) *RecentsHandler {
	return &RecentsHandler{service: service}
}

// ListRecents handles GET /me/recents. The openapi layer enforces
// limit's [1, 30] bound and the integer type before this handler
// runs; the service still re-validates as defense in depth for
// internal Go callers that bypass the wrapper.
func (h *RecentsHandler) ListRecents(w http.ResponseWriter, r *http.Request, params api.ListRecentsParams) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	svcParams := recents.ListRecentsParams{ViewerID: viewerID}
	if params.Limit != nil {
		svcParams.Limit = *params.Limit
	}
	// service.ListRecents applies DefaultLimit when Limit is 0, so a
	// caller that omits the query param naturally lands on the spec's
	// default (10) without the handler hardcoding it twice.

	result, err := h.service.ListRecents(r.Context(), svcParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListRecents failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListRecentsResponse(result))
}

// mapListRecentsResponse projects the domain result onto the wire
// envelope. The recents slice is always non-nil (`"recents": []`
// rather than `"recents": null`) when the user has no view history.
func mapListRecentsResponse(result recents.ListRecentsResult) api.ListRecentsResponse {
	out := make([]api.RecentItem, 0, len(result.Recents))
	for _, item := range result.Recents {
		out = append(out, mapRecentItem(item))
	}
	return api.ListRecentsResponse{Recents: out}
}

// mapRecentItem projects a domain RecentItem onto the wire
// RecentItem. Exactly one of File, StudyGuide, or Course is set on
// the input; the corresponding pointer is populated on the output
// and the others stay nil so they are omitted by encoding/json (the
// schema declares each as optional, not nullable).
func mapRecentItem(item recents.RecentItem) api.RecentItem {
	out := api.RecentItem{
		EntityType: api.RecentItemEntityType(item.EntityType),
		EntityId:   item.EntityID,
		ViewedAt:   item.ViewedAt,
	}
	if item.File != nil {
		out.File = &api.RecentFileSummary{
			Id:       item.File.ID,
			Name:     item.File.Name,
			MimeType: item.File.MimeType,
		}
	}
	if item.StudyGuide != nil {
		out.StudyGuide = &api.RecentStudyGuideSummary{
			Id:               item.StudyGuide.ID,
			Title:            item.StudyGuide.Title,
			CourseDepartment: item.StudyGuide.CourseDepartment,
			CourseNumber:     item.StudyGuide.CourseNumber,
		}
	}
	if item.Course != nil {
		out.Course = &api.RecentCourseSummary{
			Id:         item.Course.ID,
			Department: item.Course.Department,
			Number:     item.Course.Number,
			Title:      item.Course.Title,
		}
	}
	return out
}

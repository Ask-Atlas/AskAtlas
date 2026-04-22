package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/favorites"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// FavoritesService is the slice of the favorites service surface
// this handler depends on. Defined here (where it is used) so the
// handler can be unit-tested with a mockery-generated mock without
// dragging in the full domain package.
type FavoritesService interface {
	ListFavorites(ctx context.Context, p favorites.ListFavoritesParams) (favorites.ListFavoritesResult, error)

	// ToggleFileFavorite / ToggleStudyGuideFavorite / ToggleCourseFavorite
	// power the per-entity favorite toggle endpoints (ASK-130 / ASK-156 / ASK-157).
	// Each returns 404 when the parent entity is missing or soft-deleted.
	ToggleFileFavorite(ctx context.Context, viewerID, fileID uuid.UUID) (favorites.ToggleFavoriteResult, error)
	ToggleStudyGuideFavorite(ctx context.Context, viewerID, studyGuideID uuid.UUID) (favorites.ToggleFavoriteResult, error)
	ToggleCourseFavorite(ctx context.Context, viewerID, courseID uuid.UUID) (favorites.ToggleFavoriteResult, error)
}

// FavoritesHandler serves GET /api/me/favorites (ASK-151).
type FavoritesHandler struct {
	service FavoritesService
}

// NewFavoritesHandler wires the handler over the given service.
func NewFavoritesHandler(service FavoritesService) *FavoritesHandler {
	return &FavoritesHandler{service: service}
}

// ListFavorites handles GET /me/favorites. The openapi layer
// enforces enum + integer bounds before this handler runs; the
// service still re-validates as defense in depth for internal Go
// callers that bypass the wrapper.
func (h *FavoritesHandler) ListFavorites(w http.ResponseWriter, r *http.Request, params api.ListFavoritesParams) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	svcParams := favorites.ListFavoritesParams{
		ViewerID: viewerID,
		Cursor:   params.Cursor,
	}
	if params.Limit != nil {
		svcParams.Limit = *params.Limit
	}
	if params.EntityType != nil {
		et := favorites.EntityType(*params.EntityType)
		svcParams.EntityType = &et
	}

	result, err := h.service.ListFavorites(r.Context(), svcParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListFavorites failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListFavoritesResponse(result))
}

// ToggleFileFavorite handles POST /api/files/{file_id}/favorite (ASK-130).
// Service enforces existence (404) and applies the toggle in a single
// CTE round trip; handler just maps domain -> wire envelope.
func (h *FavoritesHandler) ToggleFileFavorite(w http.ResponseWriter, r *http.Request, fileID openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}
	result, err := h.service.ToggleFileFavorite(r.Context(), viewerID, uuid.UUID(fileID))
	if err != nil {
		respondToggleErr(w, err, "ToggleFileFavorite")
		return
	}
	respondJSON(w, http.StatusOK, mapToggleFavoriteResponse(result))
}

// ToggleStudyGuideFavorite handles POST /api/me/study-guides/{study_guide_id}/favorite (ASK-156).
func (h *FavoritesHandler) ToggleStudyGuideFavorite(w http.ResponseWriter, r *http.Request, studyGuideID openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}
	result, err := h.service.ToggleStudyGuideFavorite(r.Context(), viewerID, uuid.UUID(studyGuideID))
	if err != nil {
		respondToggleErr(w, err, "ToggleStudyGuideFavorite")
		return
	}
	respondJSON(w, http.StatusOK, mapToggleFavoriteResponse(result))
}

// ToggleCourseFavorite handles POST /api/me/courses/{course_id}/favorite (ASK-157).
func (h *FavoritesHandler) ToggleCourseFavorite(w http.ResponseWriter, r *http.Request, courseID openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}
	result, err := h.service.ToggleCourseFavorite(r.Context(), viewerID, uuid.UUID(courseID))
	if err != nil {
		respondToggleErr(w, err, "ToggleCourseFavorite")
		return
	}
	respondJSON(w, http.StatusOK, mapToggleFavoriteResponse(result))
}

// respondToggleErr collapses the boilerplate shared by all 3 toggle
// handlers: log 5xx, then write the JSON error envelope.
func respondToggleErr(w http.ResponseWriter, err error, op string) {
	sysErr := apperrors.ToHTTPError(err)
	if sysErr.Code >= 500 {
		slog.Error(op+" failed", "error", err)
	}
	apperrors.RespondWithError(w, sysErr)
}

// mapToggleFavoriteResponse projects the domain ToggleFavoriteResult
// onto the wire envelope. FavoritedAt is *time.Time so the JSON
// renders as explicit null on unfavorite (matches the schema's
// required + nullable contract -- field is always present).
func mapToggleFavoriteResponse(result favorites.ToggleFavoriteResult) api.ToggleFavoriteResponse {
	return api.ToggleFavoriteResponse{
		Favorited:   result.Favorited,
		FavoritedAt: result.FavoritedAt,
	}
}

// mapListFavoritesResponse projects the domain result onto the
// wire envelope. The favorites slice is always non-nil so the
// JSON renders as `"favorites": []` (not null) when empty.
// NextCursor is *string so the wire field renders as explicit
// JSON null on the last page.
func mapListFavoritesResponse(result favorites.ListFavoritesResult) api.ListFavoritesResponse {
	out := make([]api.FavoriteItem, 0, len(result.Favorites))
	for _, item := range result.Favorites {
		out = append(out, mapFavoriteItem(item))
	}
	return api.ListFavoritesResponse{
		Favorites:  out,
		HasMore:    result.HasMore,
		NextCursor: result.NextCursor,
	}
}

// mapFavoriteItem projects a domain FavoriteItem onto the wire
// FavoriteItem. Exactly one of File, StudyGuide, or Course is set
// on the input; the corresponding pointer is populated on the
// output and the others stay nil so they are omitted by
// encoding/json (the schema declares each as optional, not nullable).
func mapFavoriteItem(item favorites.FavoriteItem) api.FavoriteItem {
	out := api.FavoriteItem{
		EntityType:  api.FavoriteItemEntityType(item.EntityType),
		EntityId:    item.EntityID,
		FavoritedAt: item.FavoritedAt,
	}
	if item.File != nil {
		out.File = &api.FavoriteFileSummary{
			Id:       item.File.ID,
			Name:     item.File.Name,
			MimeType: item.File.MimeType,
		}
	}
	if item.StudyGuide != nil {
		out.StudyGuide = &api.FavoriteStudyGuideSummary{
			Id:               item.StudyGuide.ID,
			Title:            item.StudyGuide.Title,
			CourseDepartment: item.StudyGuide.CourseDepartment,
			CourseNumber:     item.StudyGuide.CourseNumber,
		}
	}
	if item.Course != nil {
		out.Course = &api.FavoriteCourseSummary{
			Id:         item.Course.ID,
			Department: item.Course.Department,
			Number:     item.Course.Number,
			Title:      item.Course.Title,
		}
	}
	return out
}

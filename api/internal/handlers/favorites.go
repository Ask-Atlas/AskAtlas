package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/favorites"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

// FavoritesService is the slice of the favorites service surface
// this handler depends on. Defined here (where it is used) so the
// handler can be unit-tested with a mockery-generated mock without
// dragging in the full domain package.
type FavoritesService interface {
	ListFavorites(ctx context.Context, p favorites.ListFavoritesParams) (favorites.ListFavoritesResult, error)
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

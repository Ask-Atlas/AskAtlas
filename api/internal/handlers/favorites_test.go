package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/favorites"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// favoritesTestRouter wires the composite handler with mocked
// sibling services so /me/favorites requests resolve through the
// same routing the real binary uses. The FavoritesHandler under
// test is the only real (non-mock) handler.
func favoritesTestRouter(t *testing.T, fh *handlers.FavoritesHandler) chi.Router {
	fileH := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fileH, gh, sh, ch, sgh, qh, ssh, nil, fh, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestFavoritesHandler_ListFavorites_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)

	// No authctx -> handler must short-circuit with 401 before
	// touching the service mock.
	req := httptest.NewRequest(http.MethodGet, "/me/favorites", nil)
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFavoritesHandler_ListFavorites_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)

	viewer := uuid.New()
	fileID := uuid.New()
	guideID := uuid.New()
	courseID := uuid.New()
	now := time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.MatchedBy(func(p favorites.ListFavoritesParams) bool {
			return p.ViewerID == viewer && p.Limit == 25 && p.EntityType == nil
		})).
		Return(favorites.ListFavoritesResult{
			Favorites: []favorites.FavoriteItem{
				{
					EntityType:  favorites.EntityTypeStudyGuide,
					EntityID:    guideID,
					FavoritedAt: now,
					StudyGuide: &favorites.FavoriteStudyGuideSummary{
						ID:               guideID,
						Title:            "Binary Trees",
						CourseDepartment: "CPTS",
						CourseNumber:     "322",
					},
				},
				{
					EntityType:  favorites.EntityTypeFile,
					EntityID:    fileID,
					FavoritedAt: now.Add(-time.Hour),
					File: &favorites.FavoriteFileSummary{
						ID:       fileID,
						Name:     "midterm.pdf",
						MimeType: "application/pdf",
					},
				},
				{
					EntityType:  favorites.EntityTypeCourse,
					EntityID:    courseID,
					FavoritedAt: now.Add(-2 * time.Hour),
					Course: &favorites.FavoriteCourseSummary{
						ID:         courseID,
						Department: "CPTS",
						Number:     "322",
						Title:      "SE I",
					},
				},
			},
			HasMore:    false,
			NextCursor: nil,
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites?limit=25", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListFavoritesResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Favorites, 3)

	g := resp.Favorites[0]
	assert.Equal(t, api.FavoriteItemEntityType("study_guide"), g.EntityType)
	assert.Equal(t, guideID, uuid.UUID(g.EntityId))
	require.NotNil(t, g.StudyGuide)
	assert.Equal(t, "Binary Trees", g.StudyGuide.Title)
	assert.Nil(t, g.File)
	assert.Nil(t, g.Course)

	f := resp.Favorites[1]
	assert.Equal(t, api.FavoriteItemEntityType("file"), f.EntityType)
	require.NotNil(t, f.File)
	assert.Equal(t, "midterm.pdf", f.File.Name)
	assert.Nil(t, f.StudyGuide)
	assert.Nil(t, f.Course)

	c := resp.Favorites[2]
	assert.Equal(t, api.FavoriteItemEntityType("course"), c.EntityType)
	require.NotNil(t, c.Course)
	assert.Equal(t, "SE I", c.Course.Title)

	assert.False(t, resp.HasMore)
	assert.Nil(t, resp.NextCursor)
}

func TestFavoritesHandler_ListFavorites_NoLimitParam_OmitsLimitOnService(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()

	// When the caller omits ?limit, the handler must pass Limit=0
	// to the service (which then applies DefaultLimit). Single
	// source of truth for the default lives in the service.
	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.MatchedBy(func(p favorites.ListFavoritesParams) bool {
			return p.ViewerID == viewer && p.Limit == 0
		})).
		Return(favorites.ListFavoritesResult{Favorites: []favorites.FavoriteItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestFavoritesHandler_ListFavorites_EntityTypeFilter_PassedThrough(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()

	// ?entity_type=file -> handler must convert and pass a non-nil
	// EntityType pointer to the service.
	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.MatchedBy(func(p favorites.ListFavoritesParams) bool {
			if p.EntityType == nil {
				return false
			}
			return *p.EntityType == favorites.EntityTypeFile
		})).
		Return(favorites.ListFavoritesResult{Favorites: []favorites.FavoriteItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites?entity_type=file", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestFavoritesHandler_ListFavorites_CursorPassedThrough(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	cursor := favorites.EncodeCursor(50)

	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.MatchedBy(func(p favorites.ListFavoritesParams) bool {
			if p.Cursor == nil {
				return false
			}
			return *p.Cursor == cursor
		})).
		Return(favorites.ListFavoritesResult{Favorites: []favorites.FavoriteItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites?cursor="+cursor, nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestFavoritesHandler_ListFavorites_EmptyResult_RendersEmptyArrayAndNullCursor(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()

	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.Anything).
		Return(favorites.ListFavoritesResult{Favorites: []favorites.FavoriteItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Wire contract: `"favorites": []`, NOT `"favorites": null`.
	body := w.Body.String()
	assert.Contains(t, body, `"favorites":[]`)
	// next_cursor must always be present and explicit null on
	// last page (the schema has it as required + nullable).
	assert.Contains(t, body, `"next_cursor":null`)
}

func TestFavoritesHandler_ListFavorites_HasMoreShape(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	nextCursor := favorites.EncodeCursor(25)

	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.Anything).
		Return(favorites.ListFavoritesResult{
			Favorites:  []favorites.FavoriteItem{},
			HasMore:    true,
			NextCursor: &nextCursor,
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListFavoritesResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.True(t, resp.HasMore)
	require.NotNil(t, resp.NextCursor)
	assert.Equal(t, nextCursor, *resp.NextCursor)
}

func TestFavoritesHandler_ListFavorites_ServiceBadRequest_PropagatesAs400(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()

	svcErr := apperrors.NewBadRequest("Invalid query parameters", map[string]string{
		"cursor": "invalid cursor value",
	})
	mockSvc.EXPECT().ListFavorites(mock.Anything, mock.Anything).Return(favorites.ListFavoritesResult{}, svcErr)

	req := httptest.NewRequest(http.MethodGet, "/me/favorites?cursor=garbage", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var body api.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, 400, body.Code)
	require.NotNil(t, body.Details)
	assert.Equal(t, "invalid cursor value", (*body.Details)["cursor"])
}

func TestFavoritesHandler_ListFavorites_ServiceFails_Returns500(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()

	mockSvc.EXPECT().
		ListFavorites(mock.Anything, mock.Anything).
		Return(favorites.ListFavoritesResult{}, errors.New("db is down"))

	req := httptest.NewRequest(http.MethodGet, "/me/favorites", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ----------------------------------------------------------------------
// Toggle endpoints (ASK-130 / ASK-156 / ASK-157).
//
// Wire-level coverage. The service layer carries the bulk of the
// behavioral testing (existence -> toggle) -- handler tests verify
// that the right service method is called with the right viewer +
// path UUID, and that the wire envelope serializes
// favorited / favorited_at correctly (including explicit JSON null
// on the unfavorite branch).
// ----------------------------------------------------------------------

func TestFavoritesHandler_ToggleFileFavorite_Favorite_200(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	fileID := uuid.New()
	now := time.Now().UTC().Round(time.Microsecond)

	mockSvc.EXPECT().
		ToggleFileFavorite(mock.Anything, viewer, fileID).
		Return(favorites.ToggleFavoriteResult{
			Favorited:   true,
			FavoritedAt: &now,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body api.ToggleFavoriteResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.True(t, body.Favorited)
	require.NotNil(t, body.FavoritedAt)
	assert.True(t, body.FavoritedAt.Equal(now))
}

func TestFavoritesHandler_ToggleFileFavorite_Unfavorite_200(t *testing.T) {
	// Asserts the unfavorite path returns favorited=false +
	// favorited_at as explicit JSON null (the schema marks the field
	// required + nullable, so it must always be present).
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	fileID := uuid.New()

	mockSvc.EXPECT().
		ToggleFileFavorite(mock.Anything, viewer, fileID).
		Return(favorites.ToggleFavoriteResult{Favorited: false}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Raw-JSON inspection because *time.Time decodes a JSON null to
	// a nil pointer -- looks identical to absent. We want to assert
	// the field is explicitly null, not omitted.
	raw := w.Body.String()
	assert.Contains(t, raw, `"favorited":false`)
	assert.Contains(t, raw, `"favorited_at":null`)
}

func TestFavoritesHandler_ToggleFileFavorite_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	fileID := uuid.New()

	mockSvc.EXPECT().
		ToggleFileFavorite(mock.Anything, viewer, fileID).
		Return(favorites.ToggleFavoriteResult{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFavoritesHandler_ToggleFileFavorite_BadUUID_400(t *testing.T) {
	// chi/oapi-codegen reject the invalid UUID before the service is
	// reached -- the mock has no expectation, so any call would
	// fail the test.
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/files/not-a-uuid/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), uuid.New()))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFavoritesHandler_ToggleStudyGuideFavorite_Favorite_200(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	guideID := uuid.New()
	now := time.Now().UTC().Round(time.Microsecond)

	mockSvc.EXPECT().
		ToggleStudyGuideFavorite(mock.Anything, viewer, guideID).
		Return(favorites.ToggleFavoriteResult{
			Favorited:   true,
			FavoritedAt: &now,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/me/study-guides/"+guideID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body api.ToggleFavoriteResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.True(t, body.Favorited)
}

func TestFavoritesHandler_ToggleStudyGuideFavorite_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	guideID := uuid.New()

	mockSvc.EXPECT().
		ToggleStudyGuideFavorite(mock.Anything, viewer, guideID).
		Return(favorites.ToggleFavoriteResult{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/me/study-guides/"+guideID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFavoritesHandler_ToggleCourseFavorite_Favorite_200(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	courseID := uuid.New()
	now := time.Now().UTC().Round(time.Microsecond)

	mockSvc.EXPECT().
		ToggleCourseFavorite(mock.Anything, viewer, courseID).
		Return(favorites.ToggleFavoriteResult{
			Favorited:   true,
			FavoritedAt: &now,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/me/courses/"+courseID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body api.ToggleFavoriteResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.True(t, body.Favorited)
}

func TestFavoritesHandler_ToggleCourseFavorite_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	courseID := uuid.New()

	mockSvc.EXPECT().
		ToggleCourseFavorite(mock.Anything, viewer, courseID).
		Return(favorites.ToggleFavoriteResult{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/me/courses/"+courseID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFavoritesHandler_ToggleCourseFavorite_ServiceFails_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockFavoritesService(t)
	h := handlers.NewFavoritesHandler(mockSvc)
	viewer := uuid.New()
	courseID := uuid.New()

	mockSvc.EXPECT().
		ToggleCourseFavorite(mock.Anything, viewer, courseID).
		Return(favorites.ToggleFavoriteResult{}, errors.New("db connection lost"))

	req := httptest.NewRequest(http.MethodPost, "/me/courses/"+courseID.String()+"/favorite", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := favoritesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

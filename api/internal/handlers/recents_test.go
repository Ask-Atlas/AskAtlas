package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/recents"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// recentsTestRouter wires the composite handler with mocked sibling
// services so /me/recents requests resolve through the same routing
// the real binary uses. The RecentsHandler under test is the only
// real (non-mock) handler.
func recentsTestRouter(t *testing.T, rh *handlers.RecentsHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh, rh, nil, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestRecentsHandler_ListRecents_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockRecentsService(t)
	h := handlers.NewRecentsHandler(mockSvc)

	// No authctx -> handler must short-circuit with 401 before
	// touching the service mock (which would Fail the test if hit).
	req := httptest.NewRequest(http.MethodGet, "/me/recents", nil)
	w := httptest.NewRecorder()
	r := recentsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRecentsHandler_ListRecents_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockRecentsService(t)
	h := handlers.NewRecentsHandler(mockSvc)

	viewer := uuid.New()
	fileID := uuid.New()
	guideID := uuid.New()
	courseID := uuid.New()
	now := time.Date(2026, 4, 4, 14, 30, 0, 0, time.UTC)

	mockSvc.EXPECT().
		ListRecents(mock.Anything, mock.MatchedBy(func(p recents.ListRecentsParams) bool {
			return p.ViewerID == viewer && p.Limit == 5
		})).
		Return(recents.ListRecentsResult{Recents: []recents.RecentItem{
			{
				EntityType: recents.EntityTypeStudyGuide,
				EntityID:   guideID,
				ViewedAt:   now,
				StudyGuide: &recents.RecentStudyGuideSummary{
					ID:               guideID,
					Title:            "Binary Trees",
					CourseDepartment: "CPTS",
					CourseNumber:     "322",
				},
			},
			{
				EntityType: recents.EntityTypeFile,
				EntityID:   fileID,
				ViewedAt:   now.Add(-time.Hour),
				File: &recents.RecentFileSummary{
					ID:       fileID,
					Name:     "lecture.pdf",
					MimeType: "application/pdf",
				},
			},
			{
				EntityType: recents.EntityTypeCourse,
				EntityID:   courseID,
				ViewedAt:   now.Add(-2 * time.Hour),
				Course: &recents.RecentCourseSummary{
					ID:         courseID,
					Department: "CPTS",
					Number:     "322",
					Title:      "SE I",
				},
			},
		}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/recents?limit=5", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := recentsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListRecentsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Recents, 3)

	// First item: study guide. Verifies the discriminated-union
	// mapping landed StudyGuide and left File/Course nil.
	g := resp.Recents[0]
	assert.Equal(t, api.RecentItemEntityType("study_guide"), g.EntityType)
	assert.Equal(t, guideID, uuid.UUID(g.EntityId))
	require.NotNil(t, g.StudyGuide)
	assert.Equal(t, "Binary Trees", g.StudyGuide.Title)
	assert.Equal(t, "CPTS", g.StudyGuide.CourseDepartment)
	assert.Equal(t, "322", g.StudyGuide.CourseNumber)
	assert.Nil(t, g.File)
	assert.Nil(t, g.Course)

	// Second item: file.
	f := resp.Recents[1]
	assert.Equal(t, api.RecentItemEntityType("file"), f.EntityType)
	require.NotNil(t, f.File)
	assert.Equal(t, "lecture.pdf", f.File.Name)
	assert.Equal(t, "application/pdf", f.File.MimeType)
	assert.Nil(t, f.StudyGuide)
	assert.Nil(t, f.Course)

	// Third item: course.
	c := resp.Recents[2]
	assert.Equal(t, api.RecentItemEntityType("course"), c.EntityType)
	require.NotNil(t, c.Course)
	assert.Equal(t, "SE I", c.Course.Title)
	assert.Nil(t, c.File)
	assert.Nil(t, c.StudyGuide)
}

func TestRecentsHandler_ListRecents_NoLimitParam_OmitsLimitOnService(t *testing.T) {
	mockSvc := mock_handlers.NewMockRecentsService(t)
	h := handlers.NewRecentsHandler(mockSvc)
	viewer := uuid.New()

	// When the caller omits ?limit, the handler must pass Limit=0
	// to the service (which then applies DefaultLimit). This pins
	// the contract that the handler does NOT hardcode the default
	// value -- the service owns the default.
	mockSvc.EXPECT().
		ListRecents(mock.Anything, mock.MatchedBy(func(p recents.ListRecentsParams) bool {
			return p.ViewerID == viewer && p.Limit == 0
		})).
		Return(recents.ListRecentsResult{Recents: []recents.RecentItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/recents", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := recentsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestRecentsHandler_ListRecents_EmptyResult_RendersEmptyArrayNotNull(t *testing.T) {
	mockSvc := mock_handlers.NewMockRecentsService(t)
	h := handlers.NewRecentsHandler(mockSvc)
	viewer := uuid.New()

	mockSvc.EXPECT().
		ListRecents(mock.Anything, mock.Anything).
		Return(recents.ListRecentsResult{Recents: []recents.RecentItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/me/recents", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := recentsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// The wire contract requires `"recents": []`, NOT `"recents": null`.
	// Decoding into a typed struct can't catch the null-vs-array
	// difference because both unmarshal to a nil/empty slice -- the
	// raw body check is the only way to pin this.
	assert.Contains(t, w.Body.String(), `"recents":[]`)
}

func TestRecentsHandler_ListRecents_ServiceBadRequest_PropagatesAs400(t *testing.T) {
	mockSvc := mock_handlers.NewMockRecentsService(t)
	h := handlers.NewRecentsHandler(mockSvc)
	viewer := uuid.New()

	// Service-side limit validation (defense-in-depth path) returns
	// a typed AppError with details. Handler must propagate the 400
	// + payload unchanged.
	svcErr := apperrors.NewBadRequest("Invalid query parameters", map[string]string{
		"limit": "must be between 1 and 30",
	})
	mockSvc.EXPECT().ListRecents(mock.Anything, mock.Anything).Return(recents.ListRecentsResult{}, svcErr)

	req := httptest.NewRequest(http.MethodGet, "/me/recents?limit=15", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := recentsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var body api.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&body))
	assert.Equal(t, 400, body.Code)
	require.NotNil(t, body.Details)
	assert.Contains(t, (*body.Details)["limit"], "between")
}

func TestRecentsHandler_ListRecents_ServiceFails_Returns500(t *testing.T) {
	mockSvc := mock_handlers.NewMockRecentsService(t)
	h := handlers.NewRecentsHandler(mockSvc)
	viewer := uuid.New()

	mockSvc.EXPECT().
		ListRecents(mock.Anything, mock.Anything).
		Return(recents.ListRecentsResult{}, errors.New("db is down"))

	req := httptest.NewRequest(http.MethodGet, "/me/recents", nil)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewer))
	w := httptest.NewRecorder()
	r := recentsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

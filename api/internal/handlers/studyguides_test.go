package handlers_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// studyGuidesTestRouter wires the composite handler with mocked
// file/grant/schools/courses services so the chi route resolves
// through the same path the real binary uses.
func studyGuidesTestRouter(t *testing.T, sgh *handlers.StudyGuideHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestStudyGuidesHandler_List_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStudyGuidesHandler_List_EmptyCourse(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)
	mockSvc.EXPECT().
		ListStudyGuides(mock.Anything, mock.Anything).
		Return(studyguides.ListStudyGuidesResult{}, nil)

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListStudyGuidesResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.NotNil(t, resp.StudyGuides)
	assert.Empty(t, resp.StudyGuides)
	assert.False(t, resp.HasMore)
	assert.Nil(t, resp.NextCursor)
}

func TestStudyGuidesHandler_List_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	guideID := uuid.New()
	creatorID := uuid.New()
	courseID := uuid.New()
	desc := "A guide about trees."
	created := time.Date(2026, 3, 15, 8, 30, 0, 0, time.UTC)
	updated := time.Date(2026, 3, 28, 14, 22, 0, 0, time.UTC)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)
	mockSvc.EXPECT().
		ListStudyGuides(mock.Anything, mock.Anything).
		Return(studyguides.ListStudyGuidesResult{
			StudyGuides: []studyguides.StudyGuide{{
				ID:          guideID,
				Title:       "Binary Trees Cheat Sheet",
				Description: &desc,
				Tags:        []string{"trees", "data-structures", "midterm"},
				Creator: studyguides.Creator{
					ID: creatorID, FirstName: "Nathaniel", LastName: "Gaines",
				},
				CourseID:      courseID,
				VoteScore:     12,
				ViewCount:     87,
				IsRecommended: true,
				QuizCount:     2,
				CreatedAt:     created,
				UpdatedAt:     updated,
			}},
		}, nil)

	url := fmt.Sprintf("/courses/%s/study-guides", courseID)
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.ListStudyGuidesResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.StudyGuides, 1)
	g := resp.StudyGuides[0]
	assert.Equal(t, guideID, uuid.UUID(g.Id))
	assert.Equal(t, "Binary Trees Cheat Sheet", g.Title)
	require.NotNil(t, g.Description)
	assert.Equal(t, "A guide about trees.", *g.Description)
	assert.Equal(t, []string{"trees", "data-structures", "midterm"}, g.Tags)
	assert.Equal(t, creatorID, uuid.UUID(g.Creator.Id))
	assert.Equal(t, "Nathaniel", g.Creator.FirstName)
	assert.Equal(t, courseID, uuid.UUID(g.CourseId))
	assert.Equal(t, int64(12), g.VoteScore)
	assert.Equal(t, int64(87), g.ViewCount)
	assert.True(t, g.IsRecommended)
	assert.Equal(t, int64(2), g.QuizCount)
}

// Privacy + completeness regression: list response must NEVER include
// content (only the get-by-id endpoint exposes that), email, or
// clerk_id. Same wire-bytes pattern as the section roster test.
func TestStudyGuidesHandler_List_NoContentOrPIIInResponse(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)
	mockSvc.EXPECT().
		ListStudyGuides(mock.Anything, mock.Anything).
		Return(studyguides.ListStudyGuidesResult{
			StudyGuides: []studyguides.StudyGuide{{
				ID:       uuid.New(),
				Title:    "X",
				Tags:     []string{},
				Creator:  studyguides.Creator{ID: uuid.New(), FirstName: "A", LastName: "B"},
				CourseID: uuid.New(),
			}},
		}, nil)

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.NotContains(t, body, `"content"`)
	assert.NotContains(t, body, `"email"`)
	assert.NotContains(t, body, `"clerk_id"`)
	assert.NotContains(t, body, `"clerkId"`)
}

func TestStudyGuidesHandler_List_FiltersAndSortForwarded(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)
	mockSvc.EXPECT().
		ListStudyGuides(mock.Anything, mock.MatchedBy(func(p studyguides.ListStudyGuidesParams) bool {
			return p.Q != nil && *p.Q == "binary" &&
				len(p.Tags) == 2 && p.Tags[0] == "trees" && p.Tags[1] == "midterm" &&
				p.SortBy == studyguides.SortFieldNewest &&
				p.SortDir == studyguides.SortDirAsc &&
				p.Limit == 5
		})).
		Return(studyguides.ListStudyGuidesResult{}, nil)

	url := fmt.Sprintf("/courses/%s/study-guides?q=binary&tag=trees&tag=midterm&sort_by=newest&sort_dir=asc&page_limit=5", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudyGuidesHandler_List_BadCursor(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)

	url := fmt.Sprintf("/courses/%s/study-guides?cursor=!!!notbase64", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "invalid cursor value")
	mockSvc.AssertNotCalled(t, "ListStudyGuides", mock.Anything, mock.Anything)
}

func TestStudyGuidesHandler_List_CursorRoundTrip(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	vs := int64(7)
	cursorID := uuid.New()
	token, err := studyguides.EncodeCursor(studyguides.Cursor{ID: cursorID, VoteScore: &vs})
	require.NoError(t, err)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)
	mockSvc.EXPECT().
		ListStudyGuides(mock.Anything, mock.MatchedBy(func(p studyguides.ListStudyGuidesParams) bool {
			return p.Cursor != nil && p.Cursor.ID == cursorID &&
				p.Cursor.VoteScore != nil && *p.Cursor.VoteScore == 7
		})).
		Return(studyguides.ListStudyGuidesResult{}, nil)

	url := fmt.Sprintf("/courses/%s/study-guides?cursor=%s", uuid.NewString(), token)
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudyGuidesHandler_List_CourseNotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		AssertCourseExists(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Course not found")
	mockSvc.AssertNotCalled(t, "ListStudyGuides", mock.Anything, mock.Anything)
}

func TestStudyGuidesHandler_List_BadCourseUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := "/courses/not-a-uuid/study-guides"
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudyGuidesHandler_List_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().AssertCourseExists(mock.Anything, mock.Anything).Return(nil)
	mockSvc.EXPECT().
		ListStudyGuides(mock.Anything, mock.Anything).
		Return(studyguides.ListStudyGuidesResult{}, errors.New("db down"))

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

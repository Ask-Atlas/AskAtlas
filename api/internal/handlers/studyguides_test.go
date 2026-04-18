package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// ------------------------------------------------------------------------
// GetStudyGuide (ASK-114)
// ------------------------------------------------------------------------

func TestStudyGuidesHandler_Get_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStudyGuidesHandler_Get_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	guideID := uuid.New()
	creatorID := uuid.New()
	courseID := uuid.New()
	recID := uuid.New()
	quizID := uuid.New()
	resourceID := uuid.New()
	fileID := uuid.New()
	desc := "A guide about trees."
	content := "# Binary Trees\n\nA binary tree..."
	resDesc := "Interactive viz."
	created := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC)
	resCreated := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	up := studyguides.GuideVoteUp

	mockSvc.EXPECT().
		GetStudyGuide(mock.Anything, mock.MatchedBy(func(p studyguides.GetStudyGuideParams) bool {
			return p.StudyGuideID == guideID
		})).
		Return(studyguides.StudyGuideDetail{
			ID:          guideID,
			Title:       "Binary Trees Cheat Sheet",
			Description: &desc,
			Content:     &content,
			Tags:        []string{"trees", "midterm"},
			Creator:     studyguides.Creator{ID: creatorID, FirstName: "Tim", LastName: "Roughgarden"},
			Course: studyguides.GuideCourseSummary{
				ID: courseID, Department: "CS", Number: "161", Title: "Algorithms",
			},
			VoteScore:     7,
			UserVote:      &up,
			ViewCount:     87,
			IsRecommended: true,
			RecommendedBy: []studyguides.Creator{
				{ID: recID, FirstName: "Ananth", LastName: "Jillepalli"},
			},
			Quizzes: []studyguides.Quiz{
				{ID: quizID, Title: "Tree Traversal Quiz", QuestionCount: 10},
			},
			Resources: []studyguides.Resource{
				{
					ID: resourceID, Title: "Binary Trees Visual", URL: "https://visualgo.net/en/bst",
					Type: studyguides.ResourceTypeLink, Description: &resDesc, CreatedAt: resCreated,
				},
			},
			Files: []studyguides.GuideFile{
				{ID: fileID, Name: "Slides.pdf", MimeType: "application/pdf", Size: 2048000},
			},
			CreatedAt: created,
			UpdatedAt: updated,
		}, nil)

	url := fmt.Sprintf("/study-guides/%s", guideID)
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.StudyGuideDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, guideID, uuid.UUID(resp.Id))
	assert.Equal(t, "Binary Trees Cheat Sheet", resp.Title)
	require.NotNil(t, resp.Content)
	assert.Equal(t, content, *resp.Content)
	assert.Equal(t, "Tim", resp.Creator.FirstName)
	assert.Equal(t, courseID, uuid.UUID(resp.Course.Id))
	assert.Equal(t, int64(7), resp.VoteScore)
	assert.True(t, resp.IsRecommended)

	require.NotNil(t, resp.UserVote)
	assert.Equal(t, api.StudyGuideDetailResponseUserVote("up"), *resp.UserVote)

	require.Len(t, resp.RecommendedBy, 1)
	assert.Equal(t, "Ananth", resp.RecommendedBy[0].FirstName)
	require.Len(t, resp.Quizzes, 1)
	assert.Equal(t, int64(10), resp.Quizzes[0].QuestionCount)
	require.Len(t, resp.Resources, 1)
	assert.Equal(t, api.Link, resp.Resources[0].Type)
	require.Len(t, resp.Files, 1)
	assert.Equal(t, int64(2048000), resp.Files[0].Size)
}

// Critical wire-shape contract: viewer has not voted → user_vote
// literal null in the JSON bytes (not omitted) so the frontend can
// destructure safely.
func TestStudyGuidesHandler_Get_UserVoteNullWireShape(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		GetStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{
			ID:    uuid.New(),
			Title: "X",
			Tags:  []string{},
			Creator: studyguides.Creator{ID: uuid.New(), FirstName: "A", LastName: "B"},
			Course: studyguides.GuideCourseSummary{ID: uuid.New(), Department: "D", Number: "1", Title: "T"},
			// UserVote + all nested arrays nil/empty
		}, nil)

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"user_vote":null`, "user_vote must be emitted as literal JSON null")

	var resp api.StudyGuideDetailResponse
	require.NoError(t, json.NewDecoder(strings.NewReader(body)).Decode(&resp))
	assert.Nil(t, resp.UserVote)
	// nested arrays must be non-null empty slices
	assert.NotNil(t, resp.RecommendedBy)
	assert.NotNil(t, resp.Quizzes)
	assert.NotNil(t, resp.Resources)
	assert.NotNil(t, resp.Files)
}

// Privacy regression at the wire boundary: detail response must never
// include email, clerk_id, s3_key, checksum, or any other PII that's
// not in the documented schema. Pins the response bytes.
func TestStudyGuidesHandler_Get_NoPIIInResponse(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		GetStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{
			ID:    uuid.New(),
			Title: "X",
			Tags:  []string{},
			Creator: studyguides.Creator{ID: uuid.New(), FirstName: "A", LastName: "B"},
			Course: studyguides.GuideCourseSummary{ID: uuid.New(), Department: "D", Number: "1", Title: "T"},
			Files: []studyguides.GuideFile{
				{ID: uuid.New(), Name: "slides.pdf", MimeType: "application/pdf", Size: 100},
			},
		}, nil)

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	body := w.Body.String()
	assert.NotContains(t, body, `"email"`)
	assert.NotContains(t, body, `"clerk_id"`)
	assert.NotContains(t, body, `"clerkId"`)
	assert.NotContains(t, body, `"s3_key"`)
	assert.NotContains(t, body, `"checksum"`)
	assert.NotContains(t, body, `"user_id"`) // file owner id should not leak
}

func TestStudyGuidesHandler_Get_NotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		GetStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{}, apperrors.NewNotFound("Study guide not found"))

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Study guide not found")
}

func TestStudyGuidesHandler_Get_BadUUID(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := "/study-guides/not-a-uuid"
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudyGuidesHandler_Get_InternalError(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		GetStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{}, errors.New("db down"))

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------
// CreateStudyGuide (ASK-120)
// ---------------------------------------------------------------------

// createBody marshals a CreateStudyGuideRequest into a JSON body
// reader for use with httptest. Centralized so individual tests don't
// re-derive the JSON shape.
func createBody(t *testing.T, body api.CreateStudyGuideJSONRequestBody) *strings.Reader {
	t.Helper()
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return strings.NewReader(string(raw))
}

func TestStudyGuidesHandler_Create_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(`{"title":"T"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStudyGuidesHandler_Create_MalformedJSON_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, strings.NewReader(`{not json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudyGuidesHandler_Create_Success_201(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	guideID := uuid.New()
	courseID := uuid.New()
	creatorID := uuid.New()
	desc := "Cheat sheet."

	captured := &studyguides.CreateStudyGuideParams{}
	mockSvc.EXPECT().
		CreateStudyGuide(mock.Anything, mock.Anything).
		Run(func(_ context.Context, p studyguides.CreateStudyGuideParams) {
			*captured = p
		}).
		Return(studyguides.StudyGuideDetail{
			ID:          guideID,
			Title:       "Binary Trees",
			Description: &desc,
			Tags:        []string{"trees", "midterm"},
			Creator:     studyguides.Creator{ID: creatorID, FirstName: "Ada", LastName: "Lovelace"},
			Course: studyguides.GuideCourseSummary{
				ID: courseID, Department: "CS", Number: "161", Title: "Algorithms",
			},
			VoteScore:     0,
			ViewCount:     0,
			IsRecommended: false,
			RecommendedBy: []studyguides.Creator{},
			Quizzes:       []studyguides.Quiz{},
			Resources:     []studyguides.Resource{},
			Files:         []studyguides.GuideFile{},
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}, nil)

	tags := []string{"Trees", "midterm"}
	url := fmt.Sprintf("/courses/%s/study-guides", courseID)
	req := authedRequestMethod(t, http.MethodPost, url, createBody(t, api.CreateStudyGuideJSONRequestBody{
		Title:       "Binary Trees",
		Description: &desc,
		Tags:        &tags,
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp api.StudyGuideDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, guideID, uuid.UUID(resp.Id))
	assert.Equal(t, "Binary Trees", resp.Title)
	require.NotNil(t, resp.Description)
	assert.Equal(t, "Cheat sheet.", *resp.Description)
	assert.Nil(t, resp.UserVote)
	assert.NotNil(t, resp.RecommendedBy)
	assert.Empty(t, resp.RecommendedBy)
	assert.NotNil(t, resp.Quizzes)
	assert.Empty(t, resp.Quizzes)
	// Course id from path -> service params (mirrors how main wires the
	// generated wrapper). Tag normalization is the service's job; here
	// we just assert the raw input flows through to the service layer.
	assert.Equal(t, courseID, captured.CourseID)
	assert.Equal(t, []string{"Trees", "midterm"}, captured.Tags)
}

// Service-layer 400 (e.g. tag-too-long, title-empty) flows through the
// generic ToHTTPError mapping. The handler must NOT swallow detail
// fields -- they're how the frontend pinpoints the offending input.
func TestStudyGuidesHandler_Create_ServiceBadRequest_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		CreateStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
			"title": "must not be empty",
		}))

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, createBody(t, api.CreateStudyGuideJSONRequestBody{Title: ""}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var errResp apperrors.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Contains(t, errResp.Details, "title")
}

func TestStudyGuidesHandler_Create_CourseNotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		CreateStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{}, apperrors.NewNotFound("Course not found"))

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, createBody(t, api.CreateStudyGuideJSONRequestBody{Title: "T"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStudyGuidesHandler_Create_InternalError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		CreateStudyGuide(mock.Anything, mock.Anything).
		Return(studyguides.StudyGuideDetail{}, errors.New("db down"))

	url := fmt.Sprintf("/courses/%s/study-guides", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, createBody(t, api.CreateStudyGuideJSONRequestBody{Title: "T"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------
// DeleteStudyGuide (ASK-133)
// ---------------------------------------------------------------------

func TestStudyGuidesHandler_Delete_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestStudyGuidesHandler_Delete_Success_204(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	guideID := uuid.New()
	captured := &studyguides.DeleteStudyGuideParams{}
	mockSvc.EXPECT().
		DeleteStudyGuide(mock.Anything, mock.Anything).
		Run(func(_ context.Context, p studyguides.DeleteStudyGuideParams) {
			*captured = p
		}).
		Return(nil)

	url := fmt.Sprintf("/study-guides/%s", guideID)
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
	assert.Equal(t, guideID, captured.StudyGuideID)
	assert.NotEqual(t, uuid.Nil, captured.ViewerID)
}

func TestStudyGuidesHandler_Delete_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		DeleteStudyGuide(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Study guide not found"))

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStudyGuidesHandler_Delete_Forbidden_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		DeleteStudyGuide(mock.Anything, mock.Anything).
		Return(apperrors.NewForbidden())

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestStudyGuidesHandler_Delete_InternalError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockStudyGuideService(t)
	h := handlers.NewStudyGuideHandler(mockSvc)

	mockSvc.EXPECT().
		DeleteStudyGuide(mock.Anything, mock.Anything).
		Return(errors.New("db down"))

	url := fmt.Sprintf("/study-guides/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := studyGuidesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

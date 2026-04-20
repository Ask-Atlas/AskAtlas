package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/quizzes"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// quizzesTestRouter wires the composite handler with mocked
// file/grant/schools/courses/studyguides services so the chi route
// resolves through the same path the real binary uses. The
// QuizzesHandler under test is the only real (non-mock) handler.
func quizzesTestRouter(t *testing.T, qh *handlers.QuizzesHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh, nil, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

// validCreateQuizBody returns a wire-shaped CreateQuizRequest with one
// MCQ + one TF + one freeform question -- enough to exercise the
// happy-path JSON decode + service handoff in a single call.
func validCreateQuizBody() api.CreateQuizRequest {
	desc := "A quick test."
	hint := "Think hard."
	return api.CreateQuizRequest{
		Title:       "Mixed Quiz",
		Description: &desc,
		Questions: []api.CreateQuizQuestion{
			{
				Type:     api.CreateQuizQuestionTypeMultipleChoice,
				Question: "What is 2 + 2?",
				Options: &[]api.CreateQuizMCQOption{
					{Text: "3", IsCorrect: false},
					{Text: "4", IsCorrect: true},
					{Text: "5", IsCorrect: false},
				},
				Hint: &hint,
			},
			{
				Type:          api.CreateQuizQuestionTypeTrueFalse,
				Question:      "The sky is blue.",
				CorrectAnswer: true,
			},
			{
				Type:          api.CreateQuizQuestionTypeFreeform,
				Question:      "What is the capital of France?",
				CorrectAnswer: "Paris",
			},
		},
	}
}

func TestQuizzesHandler_Create_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	body, err := json.Marshal(validCreateQuizBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestQuizzesHandler_Create_InvalidJSON_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader([]byte("{not-json}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQuizzesHandler_Create_ServiceValidationError_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().CreateQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
			"questions[0].options": "exactly one option must be correct",
		}))

	body, err := json.Marshal(validCreateQuizBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, float64(400), resp["code"])
	assert.Contains(t, resp["details"], "questions[0].options")
}

func TestQuizzesHandler_Create_StudyGuideNotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().CreateQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, apperrors.NewNotFound("Study guide not found"))

	body, err := json.Marshal(validCreateQuizBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizzesHandler_Create_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().CreateQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, errors.New("connection refused"))

	body, err := json.Marshal(validCreateQuizBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestQuizzesHandler_Create_SortOrderOverflow_400 verifies the
// handler-side bounds check on sort_order int->int32 narrowing
// (copilot PR #147 feedback). Reject anything above MaxInt32 with
// a typed 400 keyed by the per-question field path so the frontend
// can highlight the offending input.
func TestQuizzesHandler_Create_SortOrderOverflow_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	body := validCreateQuizBody()
	overflow := math.MaxInt32 + 1
	body.Questions[0].SortOrder = &overflow

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Contains(t, resp["details"], "questions[0].sort_order")
}

// TestQuizzesHandler_Create_NegativeSortOrder_400 covers the
// handler-side rejection of negative sort_order values (copilot PR
// #147 feedback). Service-side defense-in-depth covers the same
// case for direct Go callers.
func TestQuizzesHandler_Create_NegativeSortOrder_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	body := validCreateQuizBody()
	negative := -1
	body.Questions[0].SortOrder = &negative

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Contains(t, resp["details"], "questions[0].sort_order")
}

// TestQuizzesHandler_Create_Success_FullPayload exercises the full
// happy-path: body decodes, service invoked, response renders the
// QuizDetailResponse wire shape with options for MCQ, boolean
// correct_answer for TF, and string correct_answer for freeform.
func TestQuizzesHandler_Create_Success_FullPayload(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quizID := uuid.New()
	mcqQID := uuid.New()
	tfQID := uuid.New()
	ffQID := uuid.New()
	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().CreateQuiz(mock.Anything, mock.MatchedBy(func(p quizzes.CreateQuizParams) bool {
		// Sanity-check that the handler decoded the wire body and
		// surfaced the path-param study guide id + 3 questions.
		return p.StudyGuideID == studyGuideID &&
			p.Title == "Mixed Quiz" &&
			len(p.Questions) == 3
	})).Return(quizzes.QuizDetail{
		ID:           quizID,
		StudyGuideID: studyGuideID,
		Title:        "Mixed Quiz",
		Description:  ptrStrQ("A quick test."),
		Creator: quizzes.Creator{
			ID:        creatorID,
			FirstName: "Ada",
			LastName:  "Lovelace",
		},
		Questions: []quizzes.Question{
			{
				ID:            mcqQID,
				Type:          quizzes.QuestionTypeMultipleChoice,
				Question:      "What is 2 + 2?",
				CorrectAnswer: "4",
				Options: []quizzes.MCQOption{
					{ID: uuid.New(), Text: "3", IsCorrect: false, SortOrder: 0},
					{ID: uuid.New(), Text: "4", IsCorrect: true, SortOrder: 1},
					{ID: uuid.New(), Text: "5", IsCorrect: false, SortOrder: 2},
				},
				SortOrder: 0,
			},
			{
				ID:            tfQID,
				Type:          quizzes.QuestionTypeTrueFalse,
				Question:      "The sky is blue.",
				CorrectAnswer: true,
				SortOrder:     1,
			},
			{
				ID:            ffQID,
				Type:          quizzes.QuestionTypeFreeform,
				Question:      "What is the capital of France?",
				CorrectAnswer: "Paris",
				SortOrder:     2,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	body, err := json.Marshal(validCreateQuizBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/study-guides/%s/quizzes", studyGuideID.String())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp api.QuizDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	assert.Equal(t, quizID, uuid.UUID(resp.Id))
	assert.Equal(t, studyGuideID, uuid.UUID(resp.StudyGuideId))
	assert.Equal(t, "Mixed Quiz", resp.Title)
	require.NotNil(t, resp.Description)
	assert.Equal(t, "A quick test.", *resp.Description)
	assert.Equal(t, creatorID, uuid.UUID(resp.Creator.Id))
	assert.Equal(t, "Ada", resp.Creator.FirstName)
	require.Len(t, resp.Questions, 3)

	mcq := resp.Questions[0]
	assert.Equal(t, mcqQID, uuid.UUID(mcq.Id))
	assert.Equal(t, api.QuizQuestionResponseType("multiple-choice"), mcq.Type)
	require.NotNil(t, mcq.Options)
	assert.Equal(t, []string{"3", "4", "5"}, *mcq.Options)
	assert.Equal(t, "4", mcq.CorrectAnswer)

	tf := resp.Questions[1]
	assert.Equal(t, api.QuizQuestionResponseType("true-false"), tf.Type)
	// Wire shape allows Options to be nil for non-MCQ; assert that.
	assert.Nil(t, tf.Options)
	assert.Equal(t, true, tf.CorrectAnswer)

	ff := resp.Questions[2]
	assert.Equal(t, api.QuizQuestionResponseType("freeform"), ff.Type)
	assert.Nil(t, ff.Options)
	assert.Equal(t, "Paris", ff.CorrectAnswer)
}

// ptrStrQ is a quizzes-test-local helper alias so this file does not
// collide with helper names defined in sibling _test.go files of the
// same package.
func ptrStrQ(s string) *string { return &s }

// ---- UpdateQuiz (ASK-153) ----

func TestQuizzesHandler_Update_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	body := bytes.NewReader([]byte(`{"title":"X"}`))
	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodPatch, url, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestQuizzesHandler_Update_InvalidJSON_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPatch, url, bytes.NewReader([]byte("{not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestQuizzesHandler_Update_NullDescription_TriState verifies the
// raw-body two-pass decode: a request with `description: null`
// must reach the service with ClearDescription=true and a nil
// Description pointer (drives the SQL CASE that NULLs the column).
// This is the keystone test for the tri-state JSON handling.
func TestQuizzesHandler_Update_NullDescription_TriState(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().UpdateQuiz(mock.Anything, mock.MatchedBy(func(p quizzes.UpdateQuizParams) bool {
		return p.ClearDescription && p.Description == nil && p.Title == nil
	})).Return(quizzes.QuizDetail{ID: uuid.New(), Creator: quizzes.Creator{}}, nil)

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPatch, url, bytes.NewReader([]byte(`{"description":null}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestQuizzesHandler_Update_AbsentDescription verifies the inverse
// of the above: a request with `{"title":"X"}` (no description
// key at all) must reach the service with ClearDescription=false.
func TestQuizzesHandler_Update_AbsentDescription(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().UpdateQuiz(mock.Anything, mock.MatchedBy(func(p quizzes.UpdateQuizParams) bool {
		return !p.ClearDescription && p.Title != nil && *p.Title == "X"
	})).Return(quizzes.QuizDetail{ID: uuid.New(), Creator: quizzes.Creator{}}, nil)

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPatch, url, bytes.NewReader([]byte(`{"title":"X"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestQuizzesHandler_Update_ValueDescription verifies a request
// with `description: "value"` reaches the service with
// ClearDescription=true and Description pointing to the value.
func TestQuizzesHandler_Update_ValueDescription(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().UpdateQuiz(mock.Anything, mock.MatchedBy(func(p quizzes.UpdateQuizParams) bool {
		return p.ClearDescription && p.Description != nil && *p.Description == "new desc"
	})).Return(quizzes.QuizDetail{ID: uuid.New(), Creator: quizzes.Creator{}}, nil)

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPatch, url, bytes.NewReader([]byte(`{"description":"new desc"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestQuizzesHandler_Update_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().UpdateQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, apperrors.NewNotFound("Quiz not found"))

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPatch, url, bytes.NewReader([]byte(`{"title":"X"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizzesHandler_Update_NotCreator_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().UpdateQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, apperrors.NewForbidden())

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPatch, url, bytes.NewReader([]byte(`{"title":"X"}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ---- DeleteQuiz (ASK-102) ----

func TestQuizzesHandler_Delete_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestQuizzesHandler_Delete_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().DeleteQuiz(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Quiz not found"))

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizzesHandler_Delete_NotCreator_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().DeleteQuiz(mock.Anything, mock.Anything).
		Return(apperrors.NewForbidden())

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestQuizzesHandler_Delete_DBError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().DeleteQuiz(mock.Anything, mock.Anything).
		Return(errors.New("connection refused"))

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestQuizzesHandler_Delete_Success verifies AC1: 204 No Content
// + empty body. Asserts on the path-param plumbing too -- the
// quiz id from the URL must reach the service unchanged.
func TestQuizzesHandler_Delete_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	quizID := uuid.New()
	mockSvc.EXPECT().DeleteQuiz(mock.Anything, mock.MatchedBy(func(p quizzes.DeleteQuizParams) bool {
		return p.QuizID == quizID
	})).Return(nil)

	url := fmt.Sprintf("/quizzes/%s", quizID.String())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String(), "204 must have no body")
}

// ---- ListQuizzes (ASK-136) ----

func TestQuizzesHandler_List_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestQuizzesHandler_List_StudyGuideNotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().ListQuizzes(mock.Anything, mock.Anything).
		Return(nil, apperrors.NewNotFound("Study guide not found"))

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizzesHandler_List_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().ListQuizzes(mock.Anything, mock.Anything).
		Return(nil, errors.New("connection refused"))

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestQuizzesHandler_List_EmptyResponseShape covers AC2: even when
// the service returns a nil slice, the wire response renders as
// `{"quizzes": []}` (not `{"quizzes": null}`).
func TestQuizzesHandler_List_EmptyResponseShape(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().ListQuizzes(mock.Anything, mock.Anything).Return(nil, nil)

	url := fmt.Sprintf("/study-guides/%s/quizzes", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Snapshot the body BEFORE decoding so the trailing string-shape
	// assertion still has data to inspect (json.NewDecoder consumes
	// the buffer).
	rawBody := w.Body.String()
	var resp api.ListQuizzesResponse
	require.NoError(t, json.Unmarshal([]byte(rawBody), &resp))
	assert.NotNil(t, resp.Quizzes)
	assert.Empty(t, resp.Quizzes)
	// Belt-and-braces: the raw JSON should contain `"quizzes":[]`,
	// not the `null` form. Catches a regression where the mapper
	// switches to a nil-allowing shape in the future.
	assert.Contains(t, rawBody, `"quizzes":[]`)
}

// TestQuizzesHandler_List_Success_FullPayload exercises the wire
// shape end-to-end: 200 + every required field per
// QuizListItemResponse, with the order preserved from the service.
func TestQuizzesHandler_List_Success_FullPayload(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	studyGuideID := uuid.New()
	creatorID := uuid.New()
	quiz1ID := uuid.New()
	quiz2ID := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().ListQuizzes(mock.Anything, mock.MatchedBy(func(p quizzes.ListQuizzesParams) bool {
		return p.StudyGuideID == studyGuideID
	})).Return([]quizzes.QuizListItem{
		{
			ID:            quiz1ID,
			Title:         "Newer Quiz",
			Description:   ptrStrQ("Recent material"),
			QuestionCount: 7,
			Creator:       quizzes.Creator{ID: creatorID, FirstName: "Ada", LastName: "Lovelace"},
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		{
			ID:            quiz2ID,
			Title:         "Older Quiz",
			Description:   nil,
			QuestionCount: 0,
			Creator:       quizzes.Creator{ID: creatorID, FirstName: "Ada", LastName: "Lovelace"},
			CreatedAt:     now.Add(-24 * time.Hour),
			UpdatedAt:     now.Add(-24 * time.Hour),
		},
	}, nil)

	url := fmt.Sprintf("/study-guides/%s/quizzes", studyGuideID.String())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListQuizzesResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Quizzes, 2)

	first := resp.Quizzes[0]
	assert.Equal(t, quiz1ID, uuid.UUID(first.Id))
	assert.Equal(t, "Newer Quiz", first.Title)
	require.NotNil(t, first.Description)
	assert.Equal(t, "Recent material", *first.Description)
	assert.Equal(t, int64(7), first.QuestionCount)
	assert.Equal(t, creatorID, uuid.UUID(first.Creator.Id))
	assert.Equal(t, "Ada", first.Creator.FirstName)

	second := resp.Quizzes[1]
	assert.Equal(t, "Older Quiz", second.Title)
	assert.Nil(t, second.Description, "nil description must serialize as JSON null per spec")
	assert.Equal(t, int64(0), second.QuestionCount)
}

// ---- AddQuizQuestion (ASK-115) ----

// validAddQuestionBody returns a wire-shaped CreateQuizQuestion (the
// AddQuizQuestion endpoint reuses that schema verbatim) with a
// well-formed MCQ. Per-test variants override individual fields to
// exercise specific edge cases.
func validAddQuestionBody() api.AddQuizQuestionJSONRequestBody {
	hint := "Think about which node comes first."
	return api.AddQuizQuestionJSONRequestBody{
		Type:     api.CreateQuizQuestionTypeMultipleChoice,
		Question: "Which traversal visits the root node first?",
		Options: &[]api.CreateQuizMCQOption{
			{Text: "In-order", IsCorrect: false},
			{Text: "Pre-order", IsCorrect: true},
			{Text: "Post-order", IsCorrect: false},
			{Text: "Level-order", IsCorrect: false},
		},
		Hint: &hint,
	}
}

func TestQuizzesHandler_AddQuestion_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	body, err := json.Marshal(validAddQuestionBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestQuizzesHandler_AddQuestion_InvalidJSON_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader([]byte("{not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestQuizzesHandler_AddQuestion_SortOrderOverflow_400 verifies the
// handler's int->int32 narrowing guard surfaces a 400 BEFORE the
// service is invoked. Same protection as CreateQuiz; the endpoint
// reuses convertCreateQuizQuestion.
func TestQuizzesHandler_AddQuestion_SortOrderOverflow_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	body := validAddQuestionBody()
	overflow := math.MaxInt32 + 1
	body.SortOrder = &overflow

	raw, err := json.Marshal(body)
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Detail key collapses to bare `sort_order` (no `questions[i]` prefix)
	// because the question is the entire request body.
	var appErr apperrors.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&appErr))
	require.Contains(t, appErr.Details, "sort_order")
}

func TestQuizzesHandler_AddQuestion_ServiceValidationError_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().AddQuestion(mock.Anything, mock.Anything).
		Return(quizzes.Question{}, apperrors.NewBadRequest("Validation failed", map[string]string{
			"options": "exactly one option must be correct",
		}))

	body, err := json.Marshal(validAddQuestionBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQuizzesHandler_AddQuestion_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().AddQuestion(mock.Anything, mock.Anything).
		Return(quizzes.Question{}, apperrors.NewNotFound("Quiz not found"))

	body, err := json.Marshal(validAddQuestionBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizzesHandler_AddQuestion_NotCreator_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().AddQuestion(mock.Anything, mock.Anything).
		Return(quizzes.Question{}, apperrors.NewForbidden())

	body, err := json.Marshal(validAddQuestionBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestQuizzesHandler_AddQuestion_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().AddQuestion(mock.Anything, mock.Anything).
		Return(quizzes.Question{}, errors.New("connection refused"))

	body, err := json.Marshal(validAddQuestionBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestQuizzesHandler_AddQuestion_Success exercises the full happy
// path: body decodes, the path-param quiz_id surfaces in the
// service params, the response renders the QuizQuestionResponse
// wire shape with the resolved correct_answer, options array, and
// non-null feedback envelope.
func TestQuizzesHandler_AddQuestion_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	quizID := uuid.New()
	questionID := uuid.New()

	mockSvc.EXPECT().AddQuestion(mock.Anything, mock.MatchedBy(func(p quizzes.AddQuestionParams) bool {
		return p.QuizID == quizID &&
			p.Question.Type == quizzes.QuestionTypeMultipleChoice &&
			p.Question.Question == "Which traversal visits the root node first?" &&
			len(p.Question.Options) == 4
	})).Return(quizzes.Question{
		ID:            questionID,
		Type:          quizzes.QuestionTypeMultipleChoice,
		Question:      "Which traversal visits the root node first?",
		CorrectAnswer: "Pre-order",
		Options: []quizzes.MCQOption{
			{ID: uuid.New(), Text: "In-order", IsCorrect: false, SortOrder: 0},
			{ID: uuid.New(), Text: "Pre-order", IsCorrect: true, SortOrder: 1},
			{ID: uuid.New(), Text: "Post-order", IsCorrect: false, SortOrder: 2},
			{ID: uuid.New(), Text: "Level-order", IsCorrect: false, SortOrder: 3},
		},
		Hint:      ptrStrQ("Think about which node comes first."),
		SortOrder: 5,
	}, nil)

	body, err := json.Marshal(validAddQuestionBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", quizID.String())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp api.QuizQuestionResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	assert.Equal(t, questionID, uuid.UUID(resp.Id))
	assert.Equal(t, api.QuizQuestionResponseType("multiple-choice"), resp.Type)
	assert.Equal(t, "Pre-order", resp.CorrectAnswer)
	require.NotNil(t, resp.Options)
	assert.Equal(t, []string{"In-order", "Pre-order", "Post-order", "Level-order"}, *resp.Options)
	assert.Equal(t, 5, resp.SortOrder)
	require.NotNil(t, resp.Hint)
}

// TestQuizzesHandler_AddQuestion_TF_Success verifies the TF wire
// shape: options is nil (no array on the response), correct_answer
// is the boolean.
func TestQuizzesHandler_AddQuestion_TF_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	quizID := uuid.New()
	questionID := uuid.New()

	mockSvc.EXPECT().AddQuestion(mock.Anything, mock.Anything).
		Return(quizzes.Question{
			ID:            questionID,
			Type:          quizzes.QuestionTypeTrueFalse,
			Question:      "Is BFS optimal in unweighted graphs?",
			CorrectAnswer: true,
			SortOrder:     2,
		}, nil)

	body, err := json.Marshal(api.AddQuizQuestionJSONRequestBody{
		Type:          api.CreateQuizQuestionTypeTrueFalse,
		Question:      "Is BFS optimal in unweighted graphs?",
		CorrectAnswer: true,
	})
	require.NoError(t, err)

	url := fmt.Sprintf("/quizzes/%s/questions", quizID.String())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp api.QuizQuestionResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	assert.Equal(t, api.QuizQuestionResponseType("true-false"), resp.Type)
	assert.Nil(t, resp.Options, "TF must not emit an options array on the wire")
	assert.Equal(t, true, resp.CorrectAnswer)
}

// ---- GetQuiz (ASK-142) ----

func TestQuizzesHandler_GetQuiz_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestQuizzesHandler_GetQuiz_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().GetQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, apperrors.NewNotFound("Quiz not found"))

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestQuizzesHandler_GetQuiz_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	mockSvc.EXPECT().GetQuiz(mock.Anything, mock.Anything).
		Return(quizzes.QuizDetail{}, errors.New("connection refused"))

	url := fmt.Sprintf("/quizzes/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestQuizzesHandler_GetQuiz_InvalidUUID_400 verifies the oapi-
// codegen path-param validator rejects a malformed UUID before the
// handler is ever invoked. The wrapper returns 400 itself; the
// service mock must NOT be called.
func TestQuizzesHandler_GetQuiz_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodGet, "/quizzes/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestQuizzesHandler_GetQuiz_Success exercises the full happy path:
// the handler reads the quiz_id path param, calls the service, and
// renders the QuizDetailResponse wire shape with options for MCQ,
// boolean correct_answer for TF, and string correct_answer for
// freeform -- the same projection used by CreateQuiz/UpdateQuiz so
// the practice player can render any of the three with one client
// mapper.
func TestQuizzesHandler_GetQuiz_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockQuizService(t)
	h := handlers.NewQuizzesHandler(mockSvc)

	quizID := uuid.New()
	studyGuideID := uuid.New()
	creatorID := uuid.New()
	mcqQID := uuid.New()
	tfQID := uuid.New()
	ffQID := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().GetQuiz(mock.Anything, mock.MatchedBy(func(p quizzes.GetQuizParams) bool {
		return p.QuizID == quizID
	})).Return(quizzes.QuizDetail{
		ID:           quizID,
		StudyGuideID: studyGuideID,
		Title:        "Tree Traversal Quiz",
		Description:  ptrStrQ("Test your knowledge."),
		Creator: quizzes.Creator{
			ID:        creatorID,
			FirstName: "Nathaniel",
			LastName:  "Gaines",
		},
		Questions: []quizzes.Question{
			{
				ID:            mcqQID,
				Type:          quizzes.QuestionTypeMultipleChoice,
				Question:      "What is the output of an in-order traversal of a BST?",
				CorrectAnswer: "Sorted ascending",
				Options: []quizzes.MCQOption{
					{ID: uuid.New(), Text: "Random order", IsCorrect: false, SortOrder: 0},
					{ID: uuid.New(), Text: "Sorted ascending", IsCorrect: true, SortOrder: 1},
					{ID: uuid.New(), Text: "Sorted descending", IsCorrect: false, SortOrder: 2},
					{ID: uuid.New(), Text: "Level order", IsCorrect: false, SortOrder: 3},
				},
				SortOrder: 0,
			},
			{
				ID:            tfQID,
				Type:          quizzes.QuestionTypeTrueFalse,
				Question:      "A complete binary tree is always a full binary tree.",
				CorrectAnswer: false,
				SortOrder:     1,
			},
			{
				ID:            ffQID,
				Type:          quizzes.QuestionTypeFreeform,
				Question:      "What is the time complexity of searching in a balanced BST?",
				CorrectAnswer: "O(log n)",
				SortOrder:     2,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	url := fmt.Sprintf("/quizzes/%s", quizID.String())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := quizzesTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp api.QuizDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))

	assert.Equal(t, quizID, uuid.UUID(resp.Id))
	assert.Equal(t, studyGuideID, uuid.UUID(resp.StudyGuideId))
	assert.Equal(t, "Tree Traversal Quiz", resp.Title)
	require.NotNil(t, resp.Description)
	assert.Equal(t, "Test your knowledge.", *resp.Description)
	assert.Equal(t, creatorID, uuid.UUID(resp.Creator.Id))
	require.Len(t, resp.Questions, 3)

	mcq := resp.Questions[0]
	assert.Equal(t, api.QuizQuestionResponseType("multiple-choice"), mcq.Type)
	require.NotNil(t, mcq.Options)
	assert.Equal(t, []string{"Random order", "Sorted ascending", "Sorted descending", "Level order"}, *mcq.Options)
	assert.Equal(t, "Sorted ascending", mcq.CorrectAnswer)

	tf := resp.Questions[1]
	assert.Equal(t, api.QuizQuestionResponseType("true-false"), tf.Type)
	assert.Nil(t, tf.Options, "TF must not emit an options array on the wire")
	assert.Equal(t, false, tf.CorrectAnswer)

	ff := resp.Questions[2]
	assert.Equal(t, api.QuizQuestionResponseType("freeform"), ff.Type)
	assert.Nil(t, ff.Options)
	assert.Equal(t, "O(log n)", ff.CorrectAnswer)
}

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
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh)
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

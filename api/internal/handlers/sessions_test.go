package handlers_test

import (
	"bytes"
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
	"github.com/Ask-Atlas/AskAtlas/api/internal/sessions"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// sessionsTestRouter wires the composite handler with mocked
// sibling services so chi routes through the same path the real
// binary uses. The SessionsHandler under test is the only real
// (non-mock) handler.
func sessionsTestRouter(t *testing.T, ssh *handlers.SessionsHandler) chi.Router {
	fh := handlers.NewFileHandler(mock_handlers.NewMockFileService(t), nil)
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestSessionsHandler_Start_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSessionsHandler_Start_InvalidUUID_400 verifies the oapi-codegen
// path-param validator rejects a malformed UUID before the handler
// is ever invoked. The mock service must not be called.
func TestSessionsHandler_Start_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodPost, "/quizzes/not-a-uuid/sessions", nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionsHandler_Start_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().StartSession(mock.Anything, mock.Anything).
		Return(sessions.StartSessionResult{}, apperrors.NewNotFound("Quiz not found"))

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSessionsHandler_Start_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().StartSession(mock.Anything, mock.Anything).
		Return(sessions.StartSessionResult{}, errors.New("connection refused"))

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestSessionsHandler_Start_NewSession_201 covers the create path:
// service returns Created=true, handler must render 201 with the
// PracticeSessionResponse wire shape (answers: [] for fresh
// sessions, completed_at: null).
func TestSessionsHandler_Start_NewSession_201(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	quizID := uuid.New()
	sessionID := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().StartSession(mock.Anything, mock.MatchedBy(func(p sessions.StartSessionParams) bool {
		return p.QuizID == quizID
	})).Return(sessions.StartSessionResult{
		Created: true,
		Session: sessions.SessionDetail{
			ID:             sessionID,
			QuizID:         quizID,
			StartedAt:      now,
			TotalQuestions: 5,
			CorrectAnswers: 0,
			Answers:        []sessions.AnswerSummary{},
		},
	}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", quizID.String())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code, "Created=true must render 201")
	var resp api.PracticeSessionResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, sessionID, uuid.UUID(resp.Id))
	assert.Equal(t, quizID, uuid.UUID(resp.QuizId))
	assert.Equal(t, 5, resp.TotalQuestions)
	assert.Equal(t, 0, resp.CorrectAnswers)
	assert.Nil(t, resp.CompletedAt, "fresh sessions must have completed_at: null")
	require.NotNil(t, resp.Answers)
	assert.Empty(t, resp.Answers, "fresh sessions must have answers: []")
}

// TestSessionsHandler_Start_Resume_200 covers the resume path:
// service returns Created=false with answers populated, handler
// must render 200 (NOT 201) and pass the answers through verbatim.
func TestSessionsHandler_Start_Resume_200(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	quizID := uuid.New()
	sessionID := uuid.New()
	q1 := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	answer := "Sorted ascending"
	correct := true

	mockSvc.EXPECT().StartSession(mock.Anything, mock.Anything).
		Return(sessions.StartSessionResult{
			Created: false,
			Session: sessions.SessionDetail{
				ID:             sessionID,
				QuizID:         quizID,
				StartedAt:      now,
				TotalQuestions: 10,
				CorrectAnswers: 1,
				Answers: []sessions.AnswerSummary{
					{
						QuestionID: &q1,
						UserAnswer: &answer,
						IsCorrect:  &correct,
						Verified:   true,
						AnsweredAt: now.Add(time.Minute),
					},
				},
			},
		}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", quizID.String())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Created=false must render 200")
	var resp api.PracticeSessionResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Answers, 1)
	require.NotNil(t, resp.Answers[0].QuestionId)
	assert.Equal(t, q1, uuid.UUID(*resp.Answers[0].QuestionId))
	require.NotNil(t, resp.Answers[0].UserAnswer)
	assert.Equal(t, "Sorted ascending", *resp.Answers[0].UserAnswer)
	require.NotNil(t, resp.Answers[0].IsCorrect)
	assert.True(t, *resp.Answers[0].IsCorrect)
	assert.True(t, resp.Answers[0].Verified)
}

// TestSessionsHandler_Start_NullableAnswers verifies the wire-side
// nullable-field rendering: a domain AnswerSummary with all three
// nil pointers (post ON DELETE SET NULL on question_id, NULL
// columns elsewhere) must serialize as JSON null on the wire so
// the openapi nullable: true contract holds.
func TestSessionsHandler_Start_NullableAnswers(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	quizID := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().StartSession(mock.Anything, mock.Anything).
		Return(sessions.StartSessionResult{
			Created: false,
			Session: sessions.SessionDetail{
				ID:             uuid.New(),
				QuizID:         quizID,
				StartedAt:      now,
				TotalQuestions: 1,
				Answers: []sessions.AnswerSummary{{
					// All three nullable fields are nil.
					Verified:   false,
					AnsweredAt: now,
				}},
			},
		}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", quizID.String())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Decode as map so we can verify the JSON literal "null" rendering
	// (api.PracticeAnswerResponse decodes nulls back to nil pointers,
	// which would also pass an assert.Nil but doesn't prove the wire
	// shape).
	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	answersAny, ok := raw["answers"].([]any)
	require.True(t, ok)
	require.Len(t, answersAny, 1)
	answer, ok := answersAny[0].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, answer["question_id"], "null question_id must serialize as JSON null")
	assert.Nil(t, answer["user_answer"], "null user_answer must serialize as JSON null")
	assert.Nil(t, answer["is_correct"], "null is_correct must serialize as JSON null")
	assert.Equal(t, false, answer["verified"])
}

// ---- SubmitPracticeAnswer (ASK-137) ----

// validSubmitAnswerBody returns a wire-shaped SubmitAnswerRequest.
func validSubmitAnswerBody() api.SubmitPracticeAnswerJSONRequestBody {
	return api.SubmitPracticeAnswerJSONRequestBody{
		QuestionId: openapi_types.UUID(uuid.New()),
		UserAnswer: "Sorted ascending",
	}
}

func TestSessionsHandler_SubmitAnswer_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	body, err := json.Marshal(validSubmitAnswerBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/sessions/%s/answers", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSessionsHandler_SubmitAnswer_InvalidJSON_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/sessions/%s/answers", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader([]byte("{not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionsHandler_SubmitAnswer_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	body, err := json.Marshal(validSubmitAnswerBody())
	require.NoError(t, err)

	req := authedRequestMethod(t, http.MethodPost, "/sessions/not-a-uuid/answers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionsHandler_SubmitAnswer_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().SubmitAnswer(mock.Anything, mock.Anything).
		Return(sessions.AnswerSummary{}, apperrors.NewNotFound("Session not found"))

	body, err := json.Marshal(validSubmitAnswerBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/sessions/%s/answers", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSessionsHandler_SubmitAnswer_NotOwner_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().SubmitAnswer(mock.Anything, mock.Anything).
		Return(sessions.AnswerSummary{}, apperrors.NewForbidden())

	body, err := json.Marshal(validSubmitAnswerBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/sessions/%s/answers", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSessionsHandler_SubmitAnswer_SessionCompleted_409(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().SubmitAnswer(mock.Anything, mock.Anything).
		Return(sessions.AnswerSummary{}, apperrors.NewConflict("Session already completed"))

	body, err := json.Marshal(validSubmitAnswerBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/sessions/%s/answers", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSessionsHandler_SubmitAnswer_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().SubmitAnswer(mock.Anything, mock.Anything).
		Return(sessions.AnswerSummary{}, errors.New("connection refused"))

	body, err := json.Marshal(validSubmitAnswerBody())
	require.NoError(t, err)

	url := fmt.Sprintf("/sessions/%s/answers", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestSessionsHandler_SubmitAnswer_Success exercises the happy
// path: handler decodes the body, plumbs UserID from JWT context,
// renders 201 with PracticeAnswerResponse including
// backend-determined is_correct + verified.
func TestSessionsHandler_SubmitAnswer_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	sessionID := uuid.New()
	questionID := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	answer := "Sorted ascending"
	correct := true

	mockSvc.EXPECT().SubmitAnswer(mock.Anything, mock.MatchedBy(func(p sessions.SubmitAnswerParams) bool {
		return p.SessionID == sessionID &&
			p.QuestionID == questionID &&
			p.UserAnswer == "Sorted ascending"
	})).Return(sessions.AnswerSummary{
		QuestionID: &questionID,
		UserAnswer: &answer,
		IsCorrect:  &correct,
		Verified:   true,
		AnsweredAt: now,
	}, nil)

	bodyBytes, err := json.Marshal(api.SubmitPracticeAnswerJSONRequestBody{
		QuestionId: openapi_types.UUID(questionID),
		UserAnswer: "Sorted ascending",
	})
	require.NoError(t, err)

	url := fmt.Sprintf("/sessions/%s/answers", sessionID.String())
	req := authedRequestMethod(t, http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp api.PracticeAnswerResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.QuestionId)
	assert.Equal(t, questionID, uuid.UUID(*resp.QuestionId))
	require.NotNil(t, resp.UserAnswer)
	assert.Equal(t, "Sorted ascending", *resp.UserAnswer)
	require.NotNil(t, resp.IsCorrect)
	assert.True(t, *resp.IsCorrect)
	assert.True(t, resp.Verified)
}

// ---- CompletePracticeSession (ASK-140) ----

func TestSessionsHandler_Complete_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/sessions/%s/complete", uuid.NewString())
	req := httptest.NewRequest(http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSessionsHandler_Complete_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodPost, "/sessions/not-a-uuid/complete", nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionsHandler_Complete_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().CompleteSession(mock.Anything, mock.Anything).
		Return(sessions.CompletedSessionDetail{}, apperrors.NewNotFound("Session not found"))

	url := fmt.Sprintf("/sessions/%s/complete", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSessionsHandler_Complete_NotOwner_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().CompleteSession(mock.Anything, mock.Anything).
		Return(sessions.CompletedSessionDetail{}, apperrors.NewForbidden())

	url := fmt.Sprintf("/sessions/%s/complete", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSessionsHandler_Complete_AlreadyCompleted_409(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().CompleteSession(mock.Anything, mock.Anything).
		Return(sessions.CompletedSessionDetail{}, apperrors.NewConflict("Session already completed"))

	url := fmt.Sprintf("/sessions/%s/complete", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSessionsHandler_Complete_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().CompleteSession(mock.Anything, mock.Anything).
		Return(sessions.CompletedSessionDetail{}, errors.New("connection refused"))

	url := fmt.Sprintf("/sessions/%s/complete", uuid.NewString())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestSessionsHandler_Complete_Success exercises the full happy
// path: 200 status, all fields plumb through, score_percentage
// renders as int.
func TestSessionsHandler_Complete_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	sessionID := uuid.New()
	quizID := uuid.New()
	startedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 1, 10, 15, 0, 0, time.UTC)

	mockSvc.EXPECT().CompleteSession(mock.Anything, mock.MatchedBy(func(p sessions.CompleteSessionParams) bool {
		return p.SessionID == sessionID
	})).Return(sessions.CompletedSessionDetail{
		ID:              sessionID,
		QuizID:          quizID,
		StartedAt:       startedAt,
		CompletedAt:     completedAt,
		TotalQuestions:  10,
		CorrectAnswers:  7,
		ScorePercentage: 70,
	}, nil)

	url := fmt.Sprintf("/sessions/%s/complete", sessionID.String())
	req := authedRequestMethod(t, http.MethodPost, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.CompletedSessionResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, sessionID, uuid.UUID(resp.Id))
	assert.Equal(t, quizID, uuid.UUID(resp.QuizId))
	assert.Equal(t, completedAt, resp.CompletedAt)
	assert.Equal(t, 10, resp.TotalQuestions)
	assert.Equal(t, 7, resp.CorrectAnswers)
	assert.Equal(t, 70, resp.ScorePercentage)
}

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
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh, nil, nil, nil)
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

// ---- GetPracticeSession (ASK-152) ----

func TestSessionsHandler_GetSession_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSessionsHandler_GetSession_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodGet, "/sessions/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionsHandler_GetSession_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().GetSession(mock.Anything, mock.Anything).
		Return(sessions.SessionDetail{}, apperrors.NewNotFound("Session not found"))

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSessionsHandler_GetSession_NotOwner_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().GetSession(mock.Anything, mock.Anything).
		Return(sessions.SessionDetail{}, apperrors.NewForbidden())

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSessionsHandler_GetSession_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().GetSession(mock.Anything, mock.Anything).
		Return(sessions.SessionDetail{}, errors.New("connection refused"))

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestSessionsHandler_GetSession_Completed_Success exercises the
// happy completed path: 200 with score_percentage populated.
func TestSessionsHandler_GetSession_Completed_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	sessionID := uuid.New()
	quizID := uuid.New()
	q1 := uuid.New()
	startedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 4, 1, 10, 15, 0, 0, time.UTC)
	score := int32(70)
	answer := "Sorted ascending"
	correct := true

	mockSvc.EXPECT().GetSession(mock.Anything, mock.MatchedBy(func(p sessions.GetSessionParams) bool {
		return p.SessionID == sessionID
	})).Return(sessions.SessionDetail{
		ID:              sessionID,
		QuizID:          quizID,
		StartedAt:       startedAt,
		CompletedAt:     &completedAt,
		TotalQuestions:  10,
		CorrectAnswers:  7,
		ScorePercentage: &score,
		Answers: []sessions.AnswerSummary{
			{QuestionID: &q1, UserAnswer: &answer, IsCorrect: &correct, Verified: true, AnsweredAt: startedAt.Add(time.Minute)},
		},
	}, nil)

	url := fmt.Sprintf("/sessions/%s", sessionID.String())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.SessionDetailResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, sessionID, uuid.UUID(resp.Id))
	assert.Equal(t, quizID, uuid.UUID(resp.QuizId))
	require.NotNil(t, resp.CompletedAt)
	assert.Equal(t, completedAt, *resp.CompletedAt)
	require.NotNil(t, resp.ScorePercentage)
	assert.Equal(t, 70, *resp.ScorePercentage)
	assert.Equal(t, 10, resp.TotalQuestions)
	assert.Equal(t, 7, resp.CorrectAnswers)
	require.Len(t, resp.Answers, 1)
}

// TestSessionsHandler_GetSession_InProgress_NullScore verifies
// the wire side renders nullable fields correctly: completed_at
// and score_percentage both serialize as JSON null while the
// session is in-progress.
func TestSessionsHandler_GetSession_InProgress_NullScore(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	sessionID := uuid.New()
	startedAt := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)

	mockSvc.EXPECT().GetSession(mock.Anything, mock.Anything).
		Return(sessions.SessionDetail{
			ID:             sessionID,
			QuizID:         uuid.New(),
			StartedAt:      startedAt,
			CompletedAt:    nil,
			TotalQuestions: 10,
			CorrectAnswers: 2,
			// ScorePercentage left nil
			Answers: []sessions.AnswerSummary{},
		}, nil)

	url := fmt.Sprintf("/sessions/%s", sessionID.String())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Decode as raw map to verify JSON literal nulls (typed
	// SessionDetailResponse decodes nulls back to nil pointers,
	// which would also pass an assert.Nil but doesn't prove the
	// wire shape).
	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	assert.Nil(t, raw["completed_at"], "in-progress: completed_at must be JSON null")
	assert.Nil(t, raw["score_percentage"], "in-progress: score_percentage must be JSON null")
	answersAny, ok := raw["answers"].([]any)
	require.True(t, ok)
	assert.Empty(t, answersAny, "empty answers must serialize as []")
}

// ============================================================
// ListPracticeSessions tests (ASK-149)
// ============================================================

func TestSessionsHandler_List_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSessionsHandler_List_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodGet, "/quizzes/not-a-uuid/sessions", nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSessionsHandler_List_LimitTooLarge_400 verifies the
// handler-side range check rejects limit > 50 (oapi-codegen also
// has the schema bound, so the wrapper may 400 first; either way
// the result must be 400 and the service is NOT invoked).
func TestSessionsHandler_List_LimitTooLarge_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/sessions?limit=51", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSessionsHandler_List_LimitZero_400: limit=0 must be rejected
// (spec: 1 <= limit <= 50).
func TestSessionsHandler_List_LimitZero_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/sessions?limit=0", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSessionsHandler_List_StatusUnknown_400: status=pending (or
// any value other than active/completed) must be rejected. The
// oapi-codegen wrapper enforces the enum at parse time, so we get
// a 400 before reaching the handler body.
func TestSessionsHandler_List_StatusUnknown_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/sessions?status=pending", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSessionsHandler_List_CursorInvalid_400: a garbled cursor
// (not valid base64) must be rejected with details.cursor.
func TestSessionsHandler_List_CursorInvalid_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/quizzes/%s/sessions?cursor=not-valid-base64$$$", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp api.AppError
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.NotNil(t, resp.Details)
	assert.Equal(t, "invalid cursor value", (*resp.Details)["cursor"])
}

func TestSessionsHandler_List_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().ListSessions(mock.Anything, mock.Anything).
		Return(sessions.ListSessionsResult{}, apperrors.NewNotFound("Quiz not found"))

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSessionsHandler_List_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().ListSessions(mock.Anything, mock.Anything).
		Return(sessions.ListSessionsResult{}, errors.New("connection refused"))

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestSessionsHandler_List_OK_DefaultParams covers the happy path
// with no query params: limit defaults to 10, status nil, cursor nil.
// The mock checks that the handler forwards those defaults to the
// service correctly. Response body is a non-empty page with no
// next cursor.
func TestSessionsHandler_List_OK_DefaultParams(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	quizID := uuid.New()
	sessionID := uuid.New()
	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	completed := now.Add(15 * time.Minute)
	score := int32(70)

	mockSvc.EXPECT().ListSessions(mock.Anything,
		mock.MatchedBy(func(p sessions.ListSessionsParams) bool {
			return p.QuizID == quizID && p.PageLimit == 10 && p.Status == nil && p.Cursor == nil
		})).Return(sessions.ListSessionsResult{
		Sessions: []sessions.SessionSummary{
			{
				ID:              sessionID,
				StartedAt:       now,
				CompletedAt:     &completed,
				TotalQuestions:  10,
				CorrectAnswers:  7,
				ScorePercentage: &score,
			},
		},
	}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", quizID.String())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListSessionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	require.Len(t, resp.Sessions, 1)
	assert.Equal(t, sessionID, uuid.UUID(resp.Sessions[0].Id))
	assert.Equal(t, 10, resp.Sessions[0].TotalQuestions)
	assert.Equal(t, 7, resp.Sessions[0].CorrectAnswers)
	require.NotNil(t, resp.Sessions[0].ScorePercentage)
	assert.Equal(t, 70, *resp.Sessions[0].ScorePercentage)
	assert.False(t, resp.HasMore)
	assert.Nil(t, resp.NextCursor)
}

// TestSessionsHandler_List_OK_HasMoreAndCursor verifies the wire
// shape when more pages exist: has_more=true and next_cursor set.
func TestSessionsHandler_List_OK_HasMoreAndCursor(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	cursor := "ZXhhbXBsZQ" // any non-nil string -- handler forwards verbatim
	mockSvc.EXPECT().ListSessions(mock.Anything, mock.Anything).
		Return(sessions.ListSessionsResult{
			Sessions:   []sessions.SessionSummary{},
			NextCursor: &cursor,
		}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListSessionsResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.True(t, resp.HasMore)
	require.NotNil(t, resp.NextCursor)
	assert.Equal(t, cursor, *resp.NextCursor)
}

// TestSessionsHandler_List_OK_EmptyRendersBracket verifies the
// empty-result wire shape: sessions:[] (NOT null), has_more:false,
// next_cursor:null. We unmarshal into a generic map to assert the
// raw JSON field types, not just the typed-decode result (which
// would happily accept a JSON null as a nil slice).
func TestSessionsHandler_List_OK_EmptyRendersBracket(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().ListSessions(mock.Anything, mock.Anything).
		Return(sessions.ListSessionsResult{Sessions: []sessions.SessionSummary{}}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	sessionsAny, ok := raw["sessions"].([]any)
	require.True(t, ok, "sessions must serialize as JSON array, not null")
	assert.Empty(t, sessionsAny)
	assert.Equal(t, false, raw["has_more"])
	assert.Nil(t, raw["next_cursor"])
}

// TestSessionsHandler_List_OK_FiltersForwarded verifies status +
// limit + cursor query params are forwarded to the service in
// their decoded form.
func TestSessionsHandler_List_OK_FiltersForwarded(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	cursorTime := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	cursorID := uuid.New()
	encoded, err := sessions.EncodeSessionsCursor(sessions.SessionsListCursor{
		StartedAt: cursorTime, ID: cursorID,
	})
	require.NoError(t, err)

	mockSvc.EXPECT().ListSessions(mock.Anything,
		mock.MatchedBy(func(p sessions.ListSessionsParams) bool {
			return p.PageLimit == 25 &&
				p.Status != nil && *p.Status == "completed" &&
				p.Cursor != nil &&
				p.Cursor.StartedAt.Equal(cursorTime) &&
				p.Cursor.ID == cursorID
		})).Return(sessions.ListSessionsResult{Sessions: []sessions.SessionSummary{}}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions?limit=25&status=completed&cursor=%s",
		uuid.NewString(), encoded)
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestSessionsHandler_List_OK_InProgressNullScore verifies the wire
// shape for an in-progress session: completed_at null AND
// score_percentage null. Both must serialize as JSON null (not
// missing keys, not zero values).
func TestSessionsHandler_List_OK_InProgressNullScore(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	now := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	mockSvc.EXPECT().ListSessions(mock.Anything, mock.Anything).
		Return(sessions.ListSessionsResult{
			Sessions: []sessions.SessionSummary{
				{
					ID:             uuid.New(),
					StartedAt:      now,
					TotalQuestions: 10,
					CorrectAnswers: 3,
					// CompletedAt + ScorePercentage left nil
				},
			},
		}, nil)

	url := fmt.Sprintf("/quizzes/%s/sessions", uuid.NewString())
	req := authedRequestMethod(t, http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	// Inspect raw JSON so we catch a missing-key bug (the typed
	// decode would happily accept either "missing" or "null").
	var raw map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &raw))
	sessionsAny, ok := raw["sessions"].([]any)
	require.True(t, ok)
	require.Len(t, sessionsAny, 1)
	first := sessionsAny[0].(map[string]any)
	assert.Nil(t, first["completed_at"], "in-progress: completed_at must be JSON null")
	assert.Nil(t, first["score_percentage"], "in-progress: score_percentage must be JSON null")
}

// ============================================================
// AbandonPracticeSession tests (ASK-144)
// ============================================================

func TestSessionsHandler_Abandon_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSessionsHandler_Abandon_InvalidUUID_400(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	req := authedRequestMethod(t, http.MethodDelete, "/sessions/not-a-uuid", nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSessionsHandler_Abandon_NotFound_404(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().AbandonSession(mock.Anything, mock.Anything).
		Return(apperrors.NewNotFound("Session not found"))

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSessionsHandler_Abandon_Forbidden_403(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().AbandonSession(mock.Anything, mock.Anything).
		Return(apperrors.NewForbidden())

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSessionsHandler_Abandon_Conflict_409(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().AbandonSession(mock.Anything, mock.Anything).
		Return(apperrors.NewConflict("Cannot delete a completed session"))

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSessionsHandler_Abandon_ServiceError_500(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	mockSvc.EXPECT().AbandonSession(mock.Anything, mock.Anything).
		Return(errors.New("connection refused"))

	url := fmt.Sprintf("/sessions/%s", uuid.NewString())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// TestSessionsHandler_Abandon_Success_204 covers the happy path:
// service returns nil, handler renders 204 No Content with no
// body. The mock matcher verifies the path session_id is
// forwarded into AbandonSessionParams. (UserID is stamped from
// the JWT context middleware -- the test request helper uses a
// fresh test user so the matcher only pins the field this test
// directly controls; copilot PR #159 review.)
func TestSessionsHandler_Abandon_Success_204(t *testing.T) {
	mockSvc := mock_handlers.NewMockSessionService(t)
	h := handlers.NewSessionsHandler(mockSvc)

	sessionID := uuid.New()
	mockSvc.EXPECT().AbandonSession(mock.Anything,
		mock.MatchedBy(func(p sessions.AbandonSessionParams) bool {
			return p.SessionID == sessionID && p.UserID != uuid.Nil
		})).Return(nil)

	url := fmt.Sprintf("/sessions/%s", sessionID.String())
	req := authedRequestMethod(t, http.MethodDelete, url, nil)
	w := httptest.NewRecorder()

	r := sessionsTestRouter(t, h)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code, "204 No Content per spec")
	assert.Empty(t, w.Body.String(), "204 must have no body")
}

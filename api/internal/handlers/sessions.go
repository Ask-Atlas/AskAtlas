package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/sessions"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// SessionService defines the application logic required by the
// SessionsHandler. Mirrors QuizService: small, defined at the
// consumer, and mocked via mockery for handler tests.
type SessionService interface {
	StartSession(ctx context.Context, params sessions.StartSessionParams) (sessions.StartSessionResult, error)
	SubmitAnswer(ctx context.Context, params sessions.SubmitAnswerParams) (sessions.AnswerSummary, error)
	CompleteSession(ctx context.Context, params sessions.CompleteSessionParams) (sessions.CompletedSessionDetail, error)
}

// SessionsHandler manages incoming HTTP requests for the practice-
// sessions surface. Embedded in CompositeHandler so a single
// instance satisfies the generated api.ServerInterface.
type SessionsHandler struct {
	service SessionService
}

// NewSessionsHandler creates a new SessionsHandler backed by the
// given SessionService.
func NewSessionsHandler(service SessionService) *SessionsHandler {
	return &SessionsHandler{service: service}
}

// StartPracticeSession handles POST /quizzes/{quiz_id}/sessions
// (ASK-128). The service does the heavy lifting: 404 dispatch on
// quiz/parent state, stale-cleanup of >7-day incomplete sessions,
// resume probe, and race-safe new-session creation with question
// snapshotting.
//
// The status code split between 200 (resumed) and 201 (created) is
// driven entirely by StartSessionResult.Created -- the wire
// PracticeSessionResponse shape is identical on both paths so the
// frontend only branches on the HTTP status, not on body shape.
func (h *SessionsHandler) StartPracticeSession(w http.ResponseWriter, r *http.Request, quizId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	result, err := h.service.StartSession(r.Context(), sessions.StartSessionParams{
		UserID: viewerID,
		QuizID: uuid.UUID(quizId),
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("StartPracticeSession failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	respondJSON(w, status, mapPracticeSessionResponse(result.Session))
}

// CompletePracticeSession handles POST /sessions/{session_id}/complete
// (ASK-140). All the gating + score-calc happens in
// service.CompleteSession; the handler is a thin auth + dispatch +
// render pass. Returns 200 on success with the finalized session
// payload (including server-computed score_percentage).
//
// 404 / 403 / 409 disambiguation lives in the service: 404 when
// the session is missing, 403 when the viewer isn't the owner,
// 409 when the session is already completed (this endpoint is NOT
// idempotent).
func (h *SessionsHandler) CompletePracticeSession(w http.ResponseWriter, r *http.Request, sessionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	detail, err := h.service.CompleteSession(r.Context(), sessions.CompleteSessionParams{
		SessionID: uuid.UUID(sessionId),
		UserID:    viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CompletePracticeSession failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapCompletedSessionResponse(detail))
}

// mapCompletedSessionResponse projects a domain
// CompletedSessionDetail onto the wire CompletedSessionResponse.
// All required fields populate non-null on success.
func mapCompletedSessionResponse(d sessions.CompletedSessionDetail) api.CompletedSessionResponse {
	return api.CompletedSessionResponse{
		Id:              openapi_types.UUID(d.ID),
		QuizId:          openapi_types.UUID(d.QuizID),
		StartedAt:       d.StartedAt,
		CompletedAt:     d.CompletedAt,
		TotalQuestions:  int(d.TotalQuestions),
		CorrectAnswers:  int(d.CorrectAnswers),
		ScorePercentage: int(d.ScorePercentage),
	}
}

// SubmitPracticeAnswer handles POST /sessions/{session_id}/answers
// (ASK-137). The backend determines is_correct + verified
// server-side -- the client never sends them. Per-type validation,
// ownership/completion checks, and the insert + counter-bump
// transaction all live in service.SubmitAnswer; the handler is a
// thin decode -> dispatch -> render pass.
//
// 201 on success carries the persisted PracticeAnswerResponse so
// the practice player can update its local state without a
// follow-up GET.
func (h *SessionsHandler) SubmitPracticeAnswer(w http.ResponseWriter, r *http.Request, sessionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.SubmitPracticeAnswerJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	answer, err := h.service.SubmitAnswer(r.Context(), sessions.SubmitAnswerParams{
		SessionID:  uuid.UUID(sessionId),
		UserID:     viewerID,
		QuestionID: uuid.UUID(body.QuestionId),
		UserAnswer: body.UserAnswer,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("SubmitPracticeAnswer failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapPracticeAnswerResponse(answer))
}

// mapPracticeSessionResponse projects a domain SessionDetail onto
// the wire PracticeSessionResponse shape. Always emits a non-nil
// Answers slice so the JSON renders `[]` (not `null`) on freshly-
// created sessions.
func mapPracticeSessionResponse(d sessions.SessionDetail) api.PracticeSessionResponse {
	answers := make([]api.PracticeAnswerResponse, 0, len(d.Answers))
	for _, a := range d.Answers {
		answers = append(answers, mapPracticeAnswerResponse(a))
	}
	return api.PracticeSessionResponse{
		Id:             openapi_types.UUID(d.ID),
		QuizId:         openapi_types.UUID(d.QuizID),
		StartedAt:      d.StartedAt,
		CompletedAt:    d.CompletedAt,
		TotalQuestions: int(d.TotalQuestions),
		CorrectAnswers: int(d.CorrectAnswers),
		Answers:        answers,
	}
}

// mapPracticeAnswerResponse projects an AnswerSummary onto the wire
// PracticeAnswerResponse. The three nullable wire fields
// (question_id, user_answer, is_correct) are pointer-typed on the
// domain side, so a nil pointer renders as JSON null per the
// openapi nullable: true declaration.
func mapPracticeAnswerResponse(a sessions.AnswerSummary) api.PracticeAnswerResponse {
	resp := api.PracticeAnswerResponse{
		UserAnswer: a.UserAnswer,
		IsCorrect:  a.IsCorrect,
		Verified:   a.Verified,
		AnsweredAt: a.AnsweredAt,
	}
	if a.QuestionID != nil {
		qid := openapi_types.UUID(*a.QuestionID)
		resp.QuestionId = &qid
	}
	return resp
}

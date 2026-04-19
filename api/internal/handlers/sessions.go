package handlers

import (
	"context"
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

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	GetSession(ctx context.Context, params sessions.GetSessionParams) (sessions.SessionDetail, error)
	ListSessions(ctx context.Context, params sessions.ListSessionsParams) (sessions.ListSessionsResult, error)
	AbandonSession(ctx context.Context, params sessions.AbandonSessionParams) error
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

// GetPracticeSession handles GET /sessions/{session_id} (ASK-152).
// Returns the session detail (metadata + answers + nullable score)
// for the authenticated owner. 404 / 403 dispatch lives in the
// service.
//
// The wire shape is the SessionDetailResponse schema -- distinct
// from PracticeSessionResponse (no score field) and from
// CompletedSessionResponse (no answers field). The score is
// nullable on the wire to distinguish in-progress sessions from
// completed ones cleanly.
func (h *SessionsHandler) GetPracticeSession(w http.ResponseWriter, r *http.Request, sessionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	detail, err := h.service.GetSession(r.Context(), sessions.GetSessionParams{
		SessionID: uuid.UUID(sessionId),
		UserID:    viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("GetPracticeSession failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapSessionDetailResponse(detail))
}

// mapSessionDetailResponse projects a domain SessionDetail onto
// the wire SessionDetailResponse. Always emits a non-nil Answers
// slice so the JSON renders `[]` (not `null`) for sessions with
// no submitted answers. ScorePercentage forwards as-is: nil
// pointer renders as JSON null per the openapi nullable: true
// declaration.
func mapSessionDetailResponse(d sessions.SessionDetail) api.SessionDetailResponse {
	answers := make([]api.PracticeAnswerResponse, 0, len(d.Answers))
	for _, a := range d.Answers {
		answers = append(answers, mapPracticeAnswerResponse(a))
	}
	resp := api.SessionDetailResponse{
		Id:             openapi_types.UUID(d.ID),
		QuizId:         openapi_types.UUID(d.QuizID),
		StartedAt:      d.StartedAt,
		CompletedAt:    d.CompletedAt,
		TotalQuestions: int(d.TotalQuestions),
		CorrectAnswers: int(d.CorrectAnswers),
		Answers:        answers,
	}
	if d.ScorePercentage != nil {
		score := int(*d.ScorePercentage)
		resp.ScorePercentage = &score
	}
	return resp
}

// AbandonPracticeSession handles DELETE /sessions/{session_id}
// (ASK-144). All gating (404 / 403 / 409) lives in
// service.AbandonSession; the handler is a thin auth + dispatch +
// 204-render pass.
//
// Returns 204 on success with no body (per spec). The endpoint is
// NOT idempotent: a second call on an already-abandoned session
// returns 404, which is the same shape the rest of the
// session-by-id endpoints (GET, POST /complete) return when the
// session is missing.
func (h *SessionsHandler) AbandonPracticeSession(w http.ResponseWriter, r *http.Request, sessionId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	err := h.service.AbandonSession(r.Context(), sessions.AbandonSessionParams{
		SessionID: uuid.UUID(sessionId),
		UserID:    viewerID,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("AbandonPracticeSession failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

// listSessionsDefaultLimit is the default page size when the caller
// omits the `limit` query param (ASK-149 spec). Min 1, max 50 are
// enforced below in mapListSessionsParams.
const listSessionsDefaultLimit = 10

// listSessionsMaxLimit is the cap on the `limit` query param per
// the ASK-149 spec. Anything > 50 is a 400.
const listSessionsMaxLimit = 50

// ListPracticeSessions handles GET /quizzes/{quiz_id}/sessions
// (ASK-149). Returns the authenticated user's sessions for a quiz,
// cursor-paginated and optionally status-filtered. The 404 dispatch
// for a non-live parent quiz lives in the service.
//
// Validation here is purely query-param shape (limit range, status
// enum, cursor decoding); ownership scoping is enforced inside the
// service by the JWT-derived UserID, not by the path param.
func (h *SessionsHandler) ListPracticeSessions(w http.ResponseWriter, r *http.Request, quizId openapi_types.UUID, params api.ListPracticeSessionsParams) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	domainParams, appErr := mapListSessionsParams(viewerID, uuid.UUID(quizId), params)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	result, err := h.service.ListSessions(r.Context(), domainParams)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListPracticeSessions failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListSessionsResponse(result))
}

// mapListSessionsParams converts the OpenAPI HTTP-layer parameters
// into the domain ListSessionsParams. Returns a typed 400 with the
// spec-mandated detail keys on any validation failure.
//
// Defaults:
//   - limit: missing -> 10
//   - status: missing -> nil (no filter)
//   - cursor: missing -> nil (first page)
//
// Invariants enforced:
//   - 1 <= limit <= 50 (spec)
//   - status, when present, is "active" or "completed"
//     (oapi-codegen also enforces the enum, but defense-in-depth
//     for the rare path where the wrapper passes an unknown value)
//   - cursor decodes successfully
func mapListSessionsParams(viewerID, quizID uuid.UUID, params api.ListPracticeSessionsParams) (sessions.ListSessionsParams, *apperrors.AppError) {
	p := sessions.ListSessionsParams{
		UserID:    viewerID,
		QuizID:    quizID,
		PageLimit: listSessionsDefaultLimit,
	}

	if params.Limit != nil {
		l := *params.Limit
		if l < 1 || l > listSessionsMaxLimit {
			return p, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"limit": fmt.Sprintf("must be between 1 and %d", listSessionsMaxLimit),
			})
		}
		p.PageLimit = l
	}

	if params.Status != nil {
		s := string(*params.Status)
		switch s {
		case sessions.SessionStatusActive, sessions.SessionStatusCompleted:
			p.Status = &s
		default:
			return p, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"status": "must be 'active' or 'completed'",
			})
		}
	}

	if params.Cursor != nil {
		c, err := sessions.DecodeSessionsCursor(*params.Cursor)
		if err != nil {
			return p, apperrors.NewBadRequest("Invalid query parameters", map[string]string{
				"cursor": "invalid cursor value",
			})
		}
		p.Cursor = &c
	}

	return p, nil
}

// mapListSessionsResponse projects a ListSessionsResult onto the
// wire ListSessionsResponse. Always emits a non-nil Sessions slice
// so the JSON renders `[]` (not `null`) on empty pages. HasMore is
// (NextCursor != nil) -- the service guarantees the two travel
// together (NextCursor non-nil iff more pages exist).
func mapListSessionsResponse(r sessions.ListSessionsResult) api.ListSessionsResponse {
	out := api.ListSessionsResponse{
		Sessions:   make([]api.SessionSummaryResponse, 0, len(r.Sessions)),
		HasMore:    r.NextCursor != nil,
		NextCursor: r.NextCursor,
	}
	for _, s := range r.Sessions {
		out.Sessions = append(out.Sessions, mapSessionSummaryResponse(s))
	}
	return out
}

// mapSessionSummaryResponse projects a domain SessionSummary onto
// the wire SessionSummaryResponse. The two nullable wire fields
// (completed_at, score_percentage) are pointer-typed on the domain
// side, so a nil pointer renders as JSON null. ScorePercentage is a
// pointer to int32 on the domain; the wire field is *int, so we
// allocate a new local int and forward its address.
func mapSessionSummaryResponse(s sessions.SessionSummary) api.SessionSummaryResponse {
	resp := api.SessionSummaryResponse{
		Id:             openapi_types.UUID(s.ID),
		StartedAt:      s.StartedAt,
		CompletedAt:    s.CompletedAt,
		TotalQuestions: int(s.TotalQuestions),
		CorrectAnswers: int(s.CorrectAnswers),
	}
	if s.ScorePercentage != nil {
		score := int(*s.ScorePercentage)
		resp.ScorePercentage = &score
	}
	return resp
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

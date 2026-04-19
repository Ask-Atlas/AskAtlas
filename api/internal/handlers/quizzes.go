package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/quizzes"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// QuizService defines the application logic required by the
// QuizzesHandler. Mirrors StudyGuideService: small, defined at the
// consumer, and mocked via mockery for handler tests.
type QuizService interface {
	CreateQuiz(ctx context.Context, params quizzes.CreateQuizParams) (quizzes.QuizDetail, error)
	GetQuiz(ctx context.Context, params quizzes.GetQuizParams) (quizzes.QuizDetail, error)
	ListQuizzes(ctx context.Context, params quizzes.ListQuizzesParams) ([]quizzes.QuizListItem, error)
	DeleteQuiz(ctx context.Context, params quizzes.DeleteQuizParams) error
	UpdateQuiz(ctx context.Context, params quizzes.UpdateQuizParams) (quizzes.QuizDetail, error)
	AddQuestion(ctx context.Context, params quizzes.AddQuestionParams) (quizzes.Question, error)
}

// QuizzesHandler manages incoming HTTP requests for the quizzes
// surface. Embedded in CompositeHandler so a single instance
// satisfies the generated api.ServerInterface.
type QuizzesHandler struct {
	service QuizService
}

// NewQuizzesHandler creates a new QuizzesHandler backed by the given
// QuizService.
func NewQuizzesHandler(service QuizService) *QuizzesHandler {
	return &QuizzesHandler{service: service}
}

// ListQuizzes handles GET /study-guides/{study_guide_id}/quizzes
// (ASK-136). Auth-only -- any authenticated user can list. The
// service runs the live-guide preflight + the list query; a missing
// or soft-deleted guide surfaces as 404. The response always emits
// a non-nil quizzes slice so the JSON shape is `[]` (not null) when
// the guide has no quizzes.
func (h *QuizzesHandler) ListQuizzes(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	items, err := h.service.ListQuizzes(r.Context(), quizzes.ListQuizzesParams{
		StudyGuideID: uuid.UUID(studyGuideId),
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("ListQuizzes failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapListQuizzesResponse(items))
}

// GetQuiz handles GET /quizzes/{quiz_id} (ASK-142). Auth-only -- the
// service trusts the JWT gate and returns the full QuizDetail with
// every question and per-type correct_answer. 404 covers quiz
// missing OR quiz soft-deleted OR parent guide soft-deleted; the
// service maps the underlying sql.ErrNoRows to apperrors.NewNotFound
// so all three surface identically (info-leak prevention -- a
// caller cannot distinguish "no such quiz" from "quiz exists but
// you can't see it because the parent guide was deleted").
//
// The wire response shape is the shared QuizDetailResponse used by
// CreateQuiz / UpdateQuiz, so the practice player can render any of
// the three quiz endpoints with the same client-side mapper.
func (h *QuizzesHandler) GetQuiz(w http.ResponseWriter, r *http.Request, quizId openapi_types.UUID) {
	if _, appErr := viewerIDFromContext(r); appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	detail, err := h.service.GetQuiz(r.Context(), quizzes.GetQuizParams{
		QuizID: uuid.UUID(quizId),
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("GetQuiz failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapQuizDetailResponse(detail))
}

// UpdateQuiz handles PATCH /quizzes/{quiz_id} (ASK-153).
// Decodes the raw body twice: once into a key-presence map so the
// service can distinguish 'description absent' from 'description
// explicitly null' (the openapi-generated CreateQuizRequest type
// uses *string with omitempty, which loses that distinction at
// the Go-struct level), once into the typed request body for
// title + description values. Builds UpdateQuizParams from both
// passes and delegates to service.UpdateQuiz.
//
// Creator-only authz + 404/403 dispatch lives in the service.
// 200 on success carries the freshly re-hydrated QuizDetail
// (same wire shape as CreateQuiz) so the frontend can patch its
// local state without a follow-up GET.
func (h *QuizzesHandler) UpdateQuiz(w http.ResponseWriter, r *http.Request, quizId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	// Two-pass decode: presence map for tri-state description,
	// typed struct for typed values.
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &keys); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}
	var body api.UpdateQuizJSONRequestBody
	if err := json.Unmarshal(rawBody, &body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	params := quizzes.UpdateQuizParams{
		QuizID:   uuid.UUID(quizId),
		ViewerID: viewerID,
		Title:    body.Title,
	}
	if _, present := keys["description"]; present {
		// Explicit clear: the JSON key was present (whether the
		// value was a string or null). The service uses
		// ClearDescription=true to drive the SQL CASE that
		// distinguishes "leave alone" from "set to null".
		params.ClearDescription = true
		params.Description = body.Description
	}

	detail, err := h.service.UpdateQuiz(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("UpdateQuiz failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusOK, mapQuizDetailResponse(detail))
}

// DeleteQuiz handles DELETE /quizzes/{quiz_id} (ASK-102).
// Creator-only -- the service runs the locked SELECT + creator
// check + soft-delete in a single transaction. 404 covers both
// 'never existed' and 'already deleted' (idempotent semantics);
// 403 covers viewer-is-not-creator. Returns 204 with no body on
// success.
func (h *QuizzesHandler) DeleteQuiz(w http.ResponseWriter, r *http.Request, quizId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	if err := h.service.DeleteQuiz(r.Context(), quizzes.DeleteQuizParams{
		QuizID:   uuid.UUID(quizId),
		ViewerID: viewerID,
	}); err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("DeleteQuiz failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateQuiz handles POST /study-guides/{study_guide_id}/quizzes.
// The body is decoded into the openapi-generated request type;
// the service layer applies the cross-field validation (per-type
// correct_answer typing, MCQ correct-count invariant) and runs the
// quiz + questions + options inserts inside one transaction. The
// creator id is always taken from the JWT -- the openapi schema
// explicitly forbids accepting one in the request body.
func (h *QuizzesHandler) CreateQuiz(w http.ResponseWriter, r *http.Request, studyGuideId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.CreateQuizJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	questions, convErr := convertCreateQuizQuestions(body.Questions)
	if convErr != nil {
		apperrors.RespondWithError(w, convErr)
		return
	}
	params := quizzes.CreateQuizParams{
		StudyGuideID: uuid.UUID(studyGuideId),
		CreatorID:    viewerID,
		Title:        body.Title,
		Description:  body.Description,
		Questions:    questions,
	}

	detail, err := h.service.CreateQuiz(r.Context(), params)
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("CreateQuiz failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapQuizDetailResponse(detail))
}

// AddQuizQuestion handles POST /quizzes/{quiz_id}/questions (ASK-115).
// Creator-only -- the service runs the locked SELECT + creator check
// + question-cap check + insert + updated_at touch in a single
// transaction. The wire request body is the same shape as one
// element of CreateQuizRequest.questions, so the per-question
// converter is shared with CreateQuiz; service-layer validation is
// also shared (validateQuestion is reused), so a question accepted
// on create is also accepted on add.
//
// 201 on success carries the freshly-hydrated QuizQuestionResponse
// (id + per-type correct_answer resolution + non-null feedback
// envelope) so the frontend can patch its local state without a
// follow-up GET on the parent quiz.
func (h *QuizzesHandler) AddQuizQuestion(w http.ResponseWriter, r *http.Request, quizId openapi_types.UUID) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.AddQuizQuestionJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	// Empty fieldPrefix: the question IS the whole body, so 400
	// detail keys collapse to e.g. `sort_order` rather than
	// `questions[0].sort_order`.
	question, convErr := convertCreateQuizQuestion(body, "")
	if convErr != nil {
		apperrors.RespondWithError(w, convErr)
		return
	}

	created, err := h.service.AddQuestion(r.Context(), quizzes.AddQuestionParams{
		QuizID:   uuid.UUID(quizId),
		ViewerID: viewerID,
		Question: question,
	})
	if err != nil {
		sysErr := apperrors.ToHTTPError(err)
		if sysErr.Code >= 500 {
			slog.Error("AddQuizQuestion failed", "error", err)
		}
		apperrors.RespondWithError(w, sysErr)
		return
	}

	respondJSON(w, http.StatusCreated, mapQuizQuestionResponse(created))
}

// convertCreateQuizQuestions projects the openapi-generated request
// type onto the domain CreateQuizQuestionInput slice. The per-question
// conversion (including sort_order overflow protection) is delegated
// to convertCreateQuizQuestion; this wrapper just iterates and
// stitches the per-index field key (e.g. `questions[3].sort_order`)
// for any 400 details.
func convertCreateQuizQuestions(in []api.CreateQuizQuestion) ([]quizzes.CreateQuizQuestionInput, *apperrors.AppError) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]quizzes.CreateQuizQuestionInput, len(in))
	for i, q := range in {
		input, err := convertCreateQuizQuestion(q, fmt.Sprintf("questions[%d]", i))
		if err != nil {
			return nil, err
		}
		out[i] = input
	}
	return out, nil
}

// convertCreateQuizQuestion projects a single openapi CreateQuizQuestion
// onto the domain CreateQuizQuestionInput. SortOrder crosses an int
// (from the wire decoder) -> int32 (DB column) bound here; on 64-bit
// platforms a >2^31 input would silently overflow into a negative
// value, so reject anything outside [0, MaxInt32] up-front with a
// typed 400 (copilot PR feedback on ASK-150).
//
// fieldPrefix is the dotted-path prefix for any 400 detail keys --
// e.g. "questions[3]" inside CreateQuiz, or "" inside AddQuizQuestion
// where the question is the entire body. An empty prefix collapses
// the field key to just `sort_order`.
//
// Returns a typed *apperrors.AppError on validation failure so the
// handler can RespondWithError without an extra unwrap.
func convertCreateQuizQuestion(q api.CreateQuizQuestion, fieldPrefix string) (quizzes.CreateQuizQuestionInput, *apperrors.AppError) {
	input := quizzes.CreateQuizQuestionInput{
		Type:              quizzes.QuestionType(q.Type),
		Question:          q.Question,
		CorrectAnswer:     q.CorrectAnswer,
		Hint:              q.Hint,
		FeedbackCorrect:   q.FeedbackCorrect,
		FeedbackIncorrect: q.FeedbackIncorrect,
	}
	if q.SortOrder != nil {
		s := *q.SortOrder
		if s < 0 || s > math.MaxInt32 {
			return quizzes.CreateQuizQuestionInput{}, apperrors.NewBadRequest("Invalid request body", map[string]string{
				prefixedKey(fieldPrefix, "sort_order"): fmt.Sprintf("must be between 0 and %d", math.MaxInt32),
			})
		}
		v := int32(s)
		input.SortOrder = &v
	}
	if q.Options != nil {
		opts := make([]quizzes.CreateQuizMCQOptionInput, len(*q.Options))
		for j, opt := range *q.Options {
			opts[j] = quizzes.CreateQuizMCQOptionInput{
				Text:      opt.Text,
				IsCorrect: opt.IsCorrect,
			}
		}
		input.Options = opts
	}
	return input, nil
}

// prefixedKey joins a dotted field path. An empty prefix collapses
// to just the field name so AddQuizQuestion (where the question is
// the whole body) renders `sort_order` rather than `.sort_order`.
func prefixedKey(prefix, field string) string {
	if prefix == "" {
		return field
	}
	return prefix + "." + field
}

// mapQuizDetailResponse projects the QuizDetail domain type onto
// the wire QuizDetailResponse. The questions slice is always non-nil
// so the JSON wire shape is `[]` (not `null`) when empty -- the
// minimum is 1 question per the spec, but the defensive non-nil
// guards future read-side endpoints that may legitimately load
// empty quizzes.
func mapQuizDetailResponse(d quizzes.QuizDetail) api.QuizDetailResponse {
	questions := make([]api.QuizQuestionResponse, 0, len(d.Questions))
	for _, q := range d.Questions {
		questions = append(questions, mapQuizQuestionResponse(q))
	}
	return api.QuizDetailResponse{
		Id:           openapi_types.UUID(d.ID),
		StudyGuideId: openapi_types.UUID(d.StudyGuideID),
		Title:        d.Title,
		Description:  d.Description,
		Creator:      mapCreatorSummaryFromQuizzes(d.Creator),
		Questions:    questions,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

// mapQuizQuestionResponse projects a domain Question onto the wire
// QuizQuestionResponse. For MCQ questions, options are emitted as
// the option text in display order; for non-MCQ types the options
// pointer stays nil so the field is omitted from JSON. The nested
// Feedback object's Correct/Incorrect fields are *string -- nil
// becomes JSON null per the spec contract.
func mapQuizQuestionResponse(q quizzes.Question) api.QuizQuestionResponse {
	resp := api.QuizQuestionResponse{
		Id:            openapi_types.UUID(q.ID),
		Type:          api.QuizQuestionResponseType(q.Type),
		Question:      q.Question,
		CorrectAnswer: q.CorrectAnswer,
		Hint:          q.Hint,
		SortOrder:     int(q.SortOrder),
	}
	resp.Feedback.Correct = q.FeedbackCorrect
	resp.Feedback.Incorrect = q.FeedbackIncorrect

	if q.Type == quizzes.QuestionTypeMultipleChoice && len(q.Options) > 0 {
		texts := make([]string, len(q.Options))
		for i, opt := range q.Options {
			texts[i] = opt.Text
		}
		resp.Options = &texts
	}
	return resp
}

// mapCreatorSummaryFromQuizzes projects the quizzes-package Creator
// onto the shared api.CreatorSummary wire shape. Defined as its own
// helper (rather than reusing the studyguides mapCreatorSummary) so
// the quizzes package can evolve its Creator type independently.
func mapCreatorSummaryFromQuizzes(c quizzes.Creator) api.CreatorSummary {
	return api.CreatorSummary{
		Id:        openapi_types.UUID(c.ID),
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}
}

// mapListQuizzesResponse projects a slice of QuizListItem domain
// values onto the wire ListQuizzesResponse. Always emits a non-nil
// Quizzes slice so the JSON wire shape is `[]` (not null) when the
// guide has no quizzes.
func mapListQuizzesResponse(items []quizzes.QuizListItem) api.ListQuizzesResponse {
	out := make([]api.QuizListItemResponse, 0, len(items))
	for _, q := range items {
		out = append(out, api.QuizListItemResponse{
			Id:            openapi_types.UUID(q.ID),
			Title:         q.Title,
			Description:   q.Description,
			QuestionCount: q.QuestionCount,
			Creator:       mapCreatorSummaryFromQuizzes(q.Creator),
			CreatedAt:     q.CreatedAt,
			UpdatedAt:     q.UpdatedAt,
		})
	}
	return api.ListQuizzesResponse{Quizzes: out}
}

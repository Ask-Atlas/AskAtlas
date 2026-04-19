package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	ListQuizzes(ctx context.Context, params quizzes.ListQuizzesParams) ([]quizzes.QuizListItem, error)
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

// convertCreateQuizQuestions projects the openapi-generated request
// type onto the domain CreateQuizQuestionInput slice. SortOrder
// crosses an int (from the wire decoder) -> int32 (DB column) bound
// here; on 64-bit platforms a >2^31 input would silently overflow
// into a negative value, so reject anything outside [0, MaxInt32]
// up-front with a typed 400 (copilot PR feedback).
//
// Returns a typed *apperrors.AppError on validation failure so the
// handler can RespondWithError without an extra unwrap.
func convertCreateQuizQuestions(in []api.CreateQuizQuestion) ([]quizzes.CreateQuizQuestionInput, *apperrors.AppError) {
	if len(in) == 0 {
		return nil, nil
	}
	out := make([]quizzes.CreateQuizQuestionInput, len(in))
	for i, q := range in {
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
				return nil, apperrors.NewBadRequest("Invalid request body", map[string]string{
					fmt.Sprintf("questions[%d].sort_order", i): fmt.Sprintf("must be between 0 and %d", math.MaxInt32),
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
		out[i] = input
	}
	return out, nil
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

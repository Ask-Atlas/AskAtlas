package quizzes

import "github.com/google/uuid"

// Field-length / collection caps. Each constant mirrors the
// equivalent openapi schema constraint -- duplicated here so the
// service layer can re-validate Go callers (including tests) that
// bypass the wrapper layer's enforcement.
const (
	// MaxTitleLength matches openapi CreateQuizRequest.title.maxLength.
	MaxTitleLength int = 500
	// MaxDescriptionLength matches openapi CreateQuizRequest.description.maxLength.
	MaxDescriptionLength int = 2000
	// MinQuestionsCount matches openapi CreateQuizRequest.questions.minItems.
	MinQuestionsCount int = 1
	// MaxQuestionsCount matches openapi CreateQuizRequest.questions.maxItems.
	MaxQuestionsCount int = 100

	// MaxQuestionLength matches openapi CreateQuizQuestion.question.maxLength.
	MaxQuestionLength int = 2000
	// MaxHintLength matches openapi CreateQuizQuestion.hint.maxLength.
	MaxHintLength int = 1000
	// MaxFeedbackLength matches openapi CreateQuizQuestion.feedback_correct/incorrect.maxLength.
	MaxFeedbackLength int = 1000

	// MinMCQOptions matches openapi CreateQuizQuestion.options.minItems for multiple-choice.
	MinMCQOptions int = 2
	// MaxMCQOptions matches openapi CreateQuizQuestion.options.maxItems for multiple-choice.
	MaxMCQOptions int = 10
	// MaxOptionTextLength matches openapi CreateQuizMCQOption.text.maxLength.
	MaxOptionTextLength int = 500

	// MaxFreeformAnswerLength caps the reference answer string for
	// freeform questions. Matches the spec's "max 500 chars" wording
	// (the openapi schema does not enforce it directly because
	// correct_answer is polymorphic; service-side enforcement is the
	// authoritative gate).
	MaxFreeformAnswerLength int = 500
)

// True/false canonical option labels. The write side inserts two
// quiz_answer_options rows with these texts (sort_order 0 and 1
// respectively); the read side identifies the canonical answer by
// matching the option text against TrueFalseOptionTrue. Defining
// the labels in one place keeps the write-side and read-side from
// drifting out of sync (a typo on either side would silently break
// correct-answer resolution at runtime).
const (
	TrueFalseOptionTrue  = "True"
	TrueFalseOptionFalse = "False"
)

// CreateQuizMCQOptionInput is the per-option payload on a single
// MCQ question in CreateQuizParams. The service trims each option's
// text and rejects empty-after-trim values with a per-question
// per-option 400.
type CreateQuizMCQOptionInput struct {
	Text      string
	IsCorrect bool
}

// CreateQuizQuestionInput is the per-question payload on
// CreateQuizParams. The polymorphic CorrectAnswer field carries the
// raw wire value (interface{}) and is type-asserted by the service:
// bool for true-false, string for freeform, ignored for
// multiple-choice. Options is meaningful only on MCQ.
//
// SortOrder is a pointer so the service can distinguish "client
// supplied an explicit value" (preserve) from "client omitted"
// (default to the array index). The CreateQuizQuestion openapi
// shape declares sort_order as optional; we want to honor the
// caller's explicit zero when given.
type CreateQuizQuestionInput struct {
	Type              QuestionType
	Question          string
	Options           []CreateQuizMCQOptionInput
	CorrectAnswer     any
	Hint              *string
	FeedbackCorrect   *string
	FeedbackIncorrect *string
	SortOrder         *int32
}

// CreateQuizParams is the input to Service.CreateQuiz. CreatorID is
// taken from the JWT in the handler -- the spec explicitly forbids
// accepting a creator id from the request body (would be a
// privilege-attribution forge vector, same rule as studyguides).
type CreateQuizParams struct {
	StudyGuideID uuid.UUID
	CreatorID    uuid.UUID
	Title        string
	Description  *string
	Questions    []CreateQuizQuestionInput
}

// ListQuizzesParams is the input to Service.ListQuizzes (ASK-136).
// No filters / pagination -- the endpoint returns every non-deleted
// quiz on the guide. ViewerID is intentionally absent: the spec
// has no per-viewer access control beyond authentication, and the
// privacy-floor creator info is uniform across all callers.
type ListQuizzesParams struct {
	StudyGuideID uuid.UUID
}

// DeleteQuizParams is the input to Service.DeleteQuiz (ASK-102).
// ViewerID drives the creator-only authorization gate; the service
// returns apperrors.NewForbidden if it doesn't match the row's
// creator_id.
type DeleteQuizParams struct {
	QuizID   uuid.UUID
	ViewerID uuid.UUID
}

// UpdateQuizParams is the input to Service.UpdateQuiz (ASK-153).
// Tri-state semantics for description require an explicit
// ClearDescription flag because Go cannot distinguish "field
// absent in JSON" from "field explicitly null" with a *string
// alone (both decode to nil pointers). The handler decodes the
// raw request body to detect the description key's presence and
// drives the params accordingly:
//
//   - Title nil                                -> column unchanged
//   - Title non-nil                            -> set to value (after trim)
//   - ClearDescription false                   -> column unchanged
//   - ClearDescription true, Description nil   -> column cleared (set to NULL)
//   - ClearDescription true, Description set   -> column set to value (after trim)
//
// The service rejects an all-nil-fields call as 400 'at least one
// field required' before SQL.
type UpdateQuizParams struct {
	QuizID           uuid.UUID
	ViewerID         uuid.UUID
	Title            *string
	ClearDescription bool
	Description      *string
}

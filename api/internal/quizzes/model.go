// Package quizzes hosts the domain types, params, mappers, and service
// logic for the quizzes surface. Mirrors the layering of
// internal/studyguides -- repository interface + sqlc-backed impl,
// service that owns transactions + per-type validation, pointer-free
// domain types that the handler projects onto the generated wire
// schema.
package quizzes

import (
	"time"

	"github.com/google/uuid"
)

// QuestionType is the domain enum for a quiz question's kind. Values
// are the kebab-case wire form (matching openapi); the mapper
// translates to the snake_case Postgres question_type enum at the
// SQL boundary.
type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple-choice"
	QuestionTypeTrueFalse      QuestionType = "true-false"
	QuestionTypeFreeform       QuestionType = "freeform"
)

// Creator is the compact user payload embedded in QuizDetail. Same
// privacy floor as studyguides.Creator: id + first_name + last_name
// only, no email or clerk_id. Defined here (not imported) so the
// quizzes package can evolve independently of the studyguides
// surface.
type Creator struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
}

// MCQOption is a single multiple-choice option as it lives in the
// database. The wire shape collapses this into a string slice
// (options) plus the resolved correct_answer string -- the mapper
// in handlers/quizzes.go handles that projection.
type MCQOption struct {
	ID        uuid.UUID
	Text      string
	IsCorrect bool
	SortOrder int32
}

// Question is the domain payload for a single quiz question. The
// CorrectAnswer field is intentionally polymorphic to mirror the
// wire shape (string for MCQ + freeform, bool for TF); the handler
// renders it directly into the api.QuizQuestionResponse interface{}
// field. Options is empty for non-MCQ types.
type Question struct {
	ID                uuid.UUID
	Type              QuestionType
	Question          string
	Options           []MCQOption
	CorrectAnswer     any
	Hint              *string
	FeedbackCorrect   *string
	FeedbackIncorrect *string
	SortOrder         int32
}

// QuizDetail is the domain payload for the full quiz returned by
// CreateQuiz (and the future GetQuiz endpoint). Mirrors the wire
// QuizDetailResponse shape one-for-one so the handler mapper is
// near-mechanical.
type QuizDetail struct {
	ID           uuid.UUID
	StudyGuideID uuid.UUID
	Title        string
	Description  *string
	Creator      Creator
	Questions    []Question
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

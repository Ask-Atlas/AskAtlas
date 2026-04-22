// Package sessions hosts the domain types, params, mappers, and
// service logic for the practice-sessions surface (ASK-128 onward).
// Mirrors the layering of internal/quizzes -- repository interface +
// sqlc-backed impl, service that owns transactions + the
// stale-cleanup / resume / create flow, pointer-free domain types
// that the handler projects onto the generated wire schema.
package sessions

import (
	"time"

	"github.com/google/uuid"
)

// AnswerSummary is the per-answer payload embedded in SessionDetail.
// Models the practice_answers row -- QuestionID, UserAnswer, and
// IsCorrect are pointers because the underlying columns are nullable:
//   - QuestionID becomes nil if the underlying quiz_questions row is
//     hard-deleted after the answer was submitted (ON DELETE SET
//     NULL on practice_answers.question_id).
//   - UserAnswer / IsCorrect track the schema's nullable columns.
//     The submit-answer endpoint (ASK-137, future) always writes
//     non-null values in practice; the pointers preserve the
//     ability to read historical rows that pre-dated stricter
//     server-side validation.
//
// Verified is true for server-validated answer types
// (multiple-choice, true-false) and false for freeform answers
// (string-match only). The submit endpoint sets it; this package
// only reads it.
type AnswerSummary struct {
	QuestionID *uuid.UUID
	UserAnswer *string
	IsCorrect  *bool
	Verified   bool
	AnsweredAt time.Time
}

// SessionDetail is the domain payload returned by Service.StartSession
// (ASK-128). Mirrors the wire PracticeSessionResponse shape
// one-for-one so the handler mapper is near-mechanical.
//
// Used for both the new-session (201) and resume (200) paths -- the
// shape is identical; the handler picks the status code based on
// StartSessionResult.Created.
type SessionDetail struct {
	ID             uuid.UUID
	QuizID         uuid.UUID
	StartedAt      time.Time
	CompletedAt    *time.Time
	TotalQuestions int32
	CorrectAnswers int32
	// ScorePercentage is set by GetSession (ASK-152) for completed
	// sessions and left nil for in-progress ones. StartSession
	// (ASK-128) always leaves it nil because PracticeSessionResponse
	// does not carry the field on the wire; the GetSession handler's
	// SessionDetailResponse mapper is the only consumer that reads
	// it. A pointer (vs int32 + IsValid bool) keeps the JSON wire
	// rendering trivial: nil -> null, value -> integer.
	ScorePercentage *int32
	Answers         []AnswerSummary
}

// CompletedSessionDetail is the domain payload returned by
// Service.CompleteSession (ASK-140). Distinct from SessionDetail
// because the wire shape is also distinct: no Answers slice
// (callers fetch them separately via GET /api/sessions/{id}),
// CompletedAt is non-nullable (the endpoint always sets it),
// and ScorePercentage is a server-computed derived field.
type CompletedSessionDetail struct {
	ID              uuid.UUID
	QuizID          uuid.UUID
	StartedAt       time.Time
	CompletedAt     time.Time
	TotalQuestions  int32
	CorrectAnswers  int32
	ScorePercentage int32
}

// SessionSummary is the compact per-session payload returned in the
// ListSessions listing (ASK-149). Distinct from SessionDetail: no
// Answers slice (callers fetch them via GET /sessions/{id}) and no
// QuizID (the listing is already scoped to one quiz, so the wire
// shape omits it -- see SessionSummaryResponse in the openapi spec).
//
// CompletedAt is a pointer because in-progress sessions have no
// completion timestamp; ScorePercentage is a pointer for the same
// reason -- the score is computed only when the session is
// finalised. Both render as JSON null on the wire when nil.
type SessionSummary struct {
	ID              uuid.UUID
	StartedAt       time.Time
	CompletedAt     *time.Time
	TotalQuestions  int32
	CorrectAnswers  int32
	ScorePercentage *int32
}

// ListSessionsResult bundles a page of SessionSummary rows with the
// optional next-page cursor produced by the service. NextCursor is
// nil when the caller is on the last page (no more rows beyond what
// Sessions carries). The handler converts (NextCursor != nil) into
// the wire `has_more` boolean and forwards the *string verbatim.
type ListSessionsResult struct {
	Sessions   []SessionSummary
	NextCursor *string
}

// StartSessionResult bundles the SessionDetail with a Created flag
// so the handler can choose 201 (created) vs 200 (resumed) without
// re-deriving the path from the session row's timestamps. Both
// paths return the same SessionDetail shape -- only the HTTP status
// code differs.
type StartSessionResult struct {
	Session SessionDetail
	Created bool
}

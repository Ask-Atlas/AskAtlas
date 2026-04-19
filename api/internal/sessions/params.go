package sessions

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// StaleSessionAge is the cutoff for an in-progress practice session
// to be considered abandoned and eligible for hard-delete on the
// next StartSession call (ASK-128 spec AC6: "stale incomplete
// session (started_at > 7 days ago) -> cleaned up + fresh 201").
//
// This is the Single Source of Truth for the threshold. The
// DeleteStaleIncompleteSessions sqlc query takes the value in
// seconds via stale_threshold_seconds and multiplies by
// `interval '1 second'` server-side -- so changing this constant
// is the only edit required to update the policy.
const StaleSessionAge = 7 * 24 * time.Hour

// SubmitAnswerParams is the input to Service.SubmitAnswer (ASK-137).
// UserID is taken from the JWT in the handler -- the spec forbids
// accepting a user id from the request body. UserAnswer is the
// raw client input; the backend determines is_correct + verified
// itself based on the question type, so neither field is part of
// this struct.
type SubmitAnswerParams struct {
	SessionID  uuid.UUID
	UserID     uuid.UUID
	QuestionID uuid.UUID
	UserAnswer string
}

// True/false canonical wire labels accepted on SubmitAnswer
// (lowercase per spec; "True"/"False" capitalised are 400s).
const (
	TrueFalseAnswerTrue  = "true"
	TrueFalseAnswerFalse = "false"
)

// True/false canonical option-text labels written by the
// create-quiz path. MUST stay in sync with
// quizzes.TrueFalseOptionTrue / quizzes.TrueFalseOptionFalse --
// the duplication is intentional to avoid a sessions->quizzes
// package import for a 6-byte string. SubmitAnswer's TF scoring
// reads quiz_answer_options.text and compares against
// trueFalseOptionTextTrue to derive the canonical boolean.
//
// If either side renames the labels, both constants must change.
// (Keeping these unexported -- the wire surface uses the
// lowercase TrueFalseAnswer* constants above.)
const (
	trueFalseOptionTextTrue  = "True"
	trueFalseOptionTextFalse = "False"
)

// Compile-time guard: ensure the lowercase wire labels and the
// option-text labels are case-insensitive matches. If anyone ever
// renames one without the other, this falls out at runtime in
// the package init below.
var _ = func() bool {
	if !strings.EqualFold(TrueFalseAnswerTrue, trueFalseOptionTextTrue) {
		panic("sessions: TF label drift -- TrueFalseAnswerTrue must case-insensitively match trueFalseOptionTextTrue")
	}
	if !strings.EqualFold(TrueFalseAnswerFalse, trueFalseOptionTextFalse) {
		panic("sessions: TF label drift -- TrueFalseAnswerFalse must case-insensitively match trueFalseOptionTextFalse")
	}
	return true
}()

// StartSessionParams is the input to Service.StartSession (ASK-128).
// UserID is taken from the JWT in the handler -- the spec does not
// allow accepting a user id from the request body. QuizID is the
// path parameter.
type StartSessionParams struct {
	UserID uuid.UUID
	QuizID uuid.UUID
}

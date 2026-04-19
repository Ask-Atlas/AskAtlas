package sessions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

// GetSessionParams is the input to Service.GetSession (ASK-152).
// UserID is taken from the JWT in the handler -- the spec forbids
// accepting a user id from the request. SessionID is the path
// parameter.
type GetSessionParams struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
}

// CompleteSessionParams is the input to Service.CompleteSession
// (ASK-140). UserID is taken from the JWT in the handler -- the
// spec forbids accepting a user id from the body. SessionID is
// the path parameter.
type CompleteSessionParams struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
}

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

// Init-time guard: ensure the lowercase wire labels and the
// option-text labels are case-insensitive matches. The check
// runs once at package init via the var initializer below; if
// the constants ever drift out of sync, the binary panics on
// startup with a clear message (NOT a compile-time error -- the
// strings are runtime values).
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

// SessionStatus filter values accepted by ListSessions (ASK-149).
// Match the lowercase enum values in the openapi schema verbatim
// so the handler can string-cast its way from
// api.ListPracticeSessionsParamsStatus to this domain type.
const (
	SessionStatusActive    = "active"
	SessionStatusCompleted = "completed"
)

// ListSessionsParams is the validated input to Service.ListSessions
// (ASK-149). UserID is taken from the JWT in the handler --
// users cannot list each other's sessions even if they spoof the
// quiz_id path param. QuizID is the path parameter; the service
// gates on its live status before running the list query (so a
// soft-deleted parent surfaces as 404, not as an empty list).
//
// Status is optional: nil means "both active + completed". A
// non-nil pointer to a value other than the SessionStatus*
// constants is rejected by the handler at parse time.
//
// PageLimit is the caller-requested page size (1-50 inclusive,
// default 10 -- enforced by the handler). The service passes
// PageLimit + 1 to the underlying sqlc query so it can detect
// has_more without a separate COUNT.
//
// Cursor is the opaque base64 token decoded by the handler. nil
// means "first page".
type ListSessionsParams struct {
	UserID    uuid.UUID
	QuizID    uuid.UUID
	Status    *string
	PageLimit int
	Cursor    *SessionsListCursor
}

// SessionsListCursor is the keyset-pagination payload encoded into
// the opaque base64 cursor token. The list query orders by
// (started_at DESC, id DESC) so the cursor must carry both fields
// to disambiguate ties on started_at.
//
// Both fields are non-pointer because both are always set when a
// cursor exists -- a nil *SessionsListCursor on ListSessionsParams
// means "no cursor", not "partially-empty cursor".
type SessionsListCursor struct {
	StartedAt time.Time `json:"started_at"`
	ID        uuid.UUID `json:"id"`
}

// EncodeSessionsCursor serializes a SessionsListCursor into a
// base64-URL-encoded opaque token. JSON wraps the timestamp + uuid
// so the token is self-describing and forward-compatible with
// future field additions.
func EncodeSessionsCursor(c SessionsListCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("encode sessions cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeSessionsCursor parses the base64-encoded token back into a
// SessionsListCursor. Errors here become HTTP 400 with the
// spec-mandated detail key {"cursor": "invalid cursor value"} --
// see the handler's validation pass.
//
// Validates that BOTH StartedAt and ID are populated. A cursor
// with a missing field would silently pass through and apply a
// no-op `(zero_time, nil_uuid) < (started_at, id)` predicate to
// the query (always true for any real row), effectively ignoring
// pagination -- so we reject it here as a 400 instead. coderabbit
// PR #158 feedback.
func DecodeSessionsCursor(s string) (SessionsListCursor, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return SessionsListCursor{}, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var c SessionsListCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return SessionsListCursor{}, fmt.Errorf("invalid cursor payload: %w", err)
	}
	if c.StartedAt.IsZero() || c.ID == (uuid.UUID{}) {
		return SessionsListCursor{}, fmt.Errorf("invalid cursor payload: started_at and id are both required")
	}
	return c, nil
}

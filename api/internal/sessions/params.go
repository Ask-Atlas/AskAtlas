package sessions

import (
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

// StartSessionParams is the input to Service.StartSession (ASK-128).
// UserID is taken from the JWT in the handler -- the spec does not
// allow accepting a user id from the request body. QuizID is the
// path parameter.
type StartSessionParams struct {
	UserID uuid.UUID
	QuizID uuid.UUID
}

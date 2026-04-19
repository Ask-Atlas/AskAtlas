package sessions

import "github.com/google/uuid"

// StartSessionParams is the input to Service.StartSession (ASK-128).
// UserID is taken from the JWT in the handler -- the spec does not
// allow accepting a user id from the request body. QuizID is the
// path parameter.
type StartSessionParams struct {
	UserID uuid.UUID
	QuizID uuid.UUID
}

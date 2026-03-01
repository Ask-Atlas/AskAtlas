package user

import (
	"github.com/google/uuid"
)

// User represents an internal system user mapped from an external authentication provider.
type User struct {
	ID         uuid.UUID
	ClerkID    string
	Email      string
	FirstName  string
	LastName   string
	MiddleName *string
	Metadata   map[string]interface{}
}

// UpsertUserPayload contains the requested data to insert or update a user.
type UpsertUserPayload struct {
	ClerkID    string                 `json:"clerk_id"`
	Email      string                 `json:"email"`
	FirstName  string                 `json:"first_name"`
	LastName   string                 `json:"last_name"`
	MiddleName *string                `json:"middle_name"`
	Metadata   map[string]interface{} `json:"metadata"`
}

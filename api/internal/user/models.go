package user

import (
	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	ClerkID    string
	Email      string
	FirstName  string
	LastName   string
	MiddleName *string
	Metadata   map[string]interface{}
}

type UpsertUserPayload struct {
	ClerkID    string                 `json:"clerk_id"`
	Email      string                 `json:"email"`
	FirstName  string                 `json:"first_name"`
	LastName   string                 `json:"last_name"`
	MiddleName *string                `json:"middle_name"`
	Metadata   map[string]interface{} `json:"metadata"`
}

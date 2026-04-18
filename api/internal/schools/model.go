package schools

import (
	"time"

	"github.com/google/uuid"
)

// School represents a university or college that hosts courses.
type School struct {
	ID        uuid.UUID
	Name      string
	Acronym   string
	Domain    *string
	URL       *string
	City      *string
	State     *string
	Country   *string
	CreatedAt time.Time
}

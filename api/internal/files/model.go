package files

import (
	"time"

	"github.com/google/uuid"
)

// File represents an uploaded file managed by the system.
type File struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Name         string
	Size         int64
	MimeType     string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FavoritedAt  *time.Time
	LastViewedAt *time.Time
}

// Grant represents a permission granted on a file to a specific grantee.
type Grant struct {
	ID          uuid.UUID
	FileID      uuid.UUID
	GranteeType string
	GranteeID   uuid.UUID
	Permission  string
	GrantedBy   uuid.UUID
	CreatedAt   time.Time
}

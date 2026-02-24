package files

import (
	"time"

	"github.com/google/uuid"
)

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

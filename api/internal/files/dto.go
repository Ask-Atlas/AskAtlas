package files

import (
	"time"

	"github.com/google/uuid"
)

type FileResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Size         int64      `json:"size"`
	MimeType     string     `json:"mime_type"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	FavoritedAt  *time.Time `json:"favorited_at,omitempty"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
}

type ListFilesResponse struct {
	Files      []FileResponse `json:"files"`
	NextCursor *string        `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

func ToFileResponse(f File) FileResponse {
	return FileResponse{
		ID:           f.ID,
		Name:         f.Name,
		Size:         f.Size,
		MimeType:     f.MimeType,
		Status:       f.Status,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
		FavoritedAt:  f.FavoritedAt,
		LastViewedAt: f.LastViewedAt,
	}
}

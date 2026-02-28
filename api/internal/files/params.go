package files

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SortField string
type SortDir string
type FileScope string

const (
	SortFieldUpdatedAt SortField = "updated_at"
	SortFieldCreatedAt SortField = "created_at"
	SortFieldName      SortField = "name"
	SortFieldSize      SortField = "size"
	SortFieldStatus    SortField = "status"
	SortFieldMimeType  SortField = "mime_type"

	SortDirAsc  SortDir = "asc"
	SortDirDesc SortDir = "desc"

	ScopeOwned      FileScope = "owned"
	ScopeCourse     FileScope = "course"
	ScopeStudyGuide FileScope = "study_guide"
	ScopeAccessible FileScope = "accessible"
)

// Cursor is the opaque pagination token. Only the field matching the active
// SortField is populated; all others are nil.
type Cursor struct {
	ID        uuid.UUID  `json:"id"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	NameLower *string    `json:"name_lower,omitempty"`
	Size      *int64     `json:"size,omitempty"`
	Status    *string    `json:"status,omitempty"`
	MimeType  *string    `json:"mime_type,omitempty"`
}

func EncodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func DecodeCursor(s string) (Cursor, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor encoding: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, fmt.Errorf("invalid cursor payload: %w", err)
	}
	return c, nil
}

// ListFilesParams is the fully validated input to Service.ListFiles.
type ListFilesParams struct {
	ViewerID      uuid.UUID
	OwnerID       uuid.UUID
	CourseIDs     []uuid.UUID
	StudyGuideIDs []uuid.UUID
	Scope         FileScope

	Status      *string
	MimeType    *string
	MinSize     *int64
	MaxSize     *int64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	UpdatedFrom *time.Time
	UpdatedTo   *time.Time
	Q           *string

	SortField SortField
	SortDir   SortDir
	PageLimit int
	Cursor    *Cursor
}

type GetFileParams struct {
	ViewerID      uuid.UUID
	FileID        uuid.UUID
	CourseIDs     []uuid.UUID
	StudyGuideIDs []uuid.UUID
}

package files

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
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

var validSortFields = map[SortField]struct{}{
	SortFieldUpdatedAt: {},
	SortFieldCreatedAt: {},
	SortFieldName:      {},
	SortFieldSize:      {},
	SortFieldStatus:    {},
	SortFieldMimeType:  {},
}

var validSortDirs = map[SortDir]struct{}{
	SortDirAsc:  {},
	SortDirDesc: {},
}

var validScopes = map[FileScope]struct{}{
	ScopeOwned:      {},
	ScopeCourse:     {},
	ScopeStudyGuide: {},
	ScopeAccessible: {},
}

var validStatuses = map[string]struct{}{
	string(db.UploadStatusPending):  {},
	string(db.UploadStatusComplete): {},
	string(db.UploadStatusFailed):   {},
}

var validMimeTypes = map[string]struct{}{
	string(db.MimeTypeImageJpeg):      {},
	string(db.MimeTypeImagePng):       {},
	string(db.MimeTypeImageWebp):      {},
	string(db.MimeTypeApplicationPdf): {},
}

const (
	DefaultPageLimit = 25
	MaxPageLimit     = 100
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

func encodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeCursor(s string) (Cursor, error) {
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

// ParseListFilesParams reads and validates query parameters from an HTTP request.
func ParseListFilesParams(
	r *http.Request,
	viewerID uuid.UUID,
	courseIDs []uuid.UUID,
	studyGuideIDs []uuid.UUID,
) (*ListFilesParams, *apperrors.AppError) {
	q := r.URL.Query()
	details := map[string]string{}

	scope := FileScope(strings.ToLower(q.Get("scope")))
	if scope == "" {
		scope = ScopeOwned
	}
	if _, ok := validScopes[scope]; !ok {
		details["scope"] = "must be one of: owned, course, study_guide, accessible"
	}

	sortField := SortField(strings.ToLower(q.Get("sort_by")))
	if sortField == "" {
		sortField = SortFieldUpdatedAt
	}
	if _, ok := validSortFields[sortField]; !ok {
		details["sort_by"] = "must be one of: updated_at, created_at, name, size, status, mime_type"
	}

	sortDir := SortDir(strings.ToLower(q.Get("sort_dir")))
	if sortDir == "" {
		sortDir = SortDirDesc
	}
	if _, ok := validSortDirs[sortDir]; !ok {
		details["sort_dir"] = "must be one of: asc, desc"
	}

	pageLimit := DefaultPageLimit
	if raw := q.Get("page_limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > MaxPageLimit {
			details["page_limit"] = fmt.Sprintf("must be an integer between 1 and %d", MaxPageLimit)
		} else {
			pageLimit = n
		}
	}

	var cursor *Cursor
	if raw := q.Get("cursor"); raw != "" {
		c, err := decodeCursor(raw)
		if err != nil {
			details["cursor"] = "invalid cursor value"
		} else {
			cursor = &c
		}
	}

	var status, mimeType, searchQ *string
	if raw := q.Get("status"); raw != "" {
		if _, ok := validStatuses[raw]; !ok {
			details["status"] = "must be one of: pending, complete, failed"
		} else {
			status = &raw
		}
	}
	if raw := q.Get("mime_type"); raw != "" {
		if _, ok := validMimeTypes[raw]; !ok {
			details["mime_type"] = "must be one of: image/jpeg, image/png, image/webp, application/pdf"
		} else {
			mimeType = &raw
		}
	}
	if raw := q.Get("q"); raw != "" {
		searchQ = &raw
	}

	var minSize, maxSize *int64
	if raw := q.Get("min_size"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 0 {
			details["min_size"] = "must be a non-negative integer (bytes)"
		} else {
			minSize = &n
		}
	}
	if raw := q.Get("max_size"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 0 {
			details["max_size"] = "must be a non-negative integer (bytes)"
		} else {
			maxSize = &n
		}
	}
	if minSize != nil && maxSize != nil && *minSize > *maxSize {
		details["min_size"] = "min_size cannot be greater than max_size"
	}

	parseTime := func(key string) *time.Time {
		raw := q.Get(key)
		if raw == "" {
			return nil
		}
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			details[key] = "must be RFC3339 format (e.g. 2024-01-15T00:00:00Z)"
			return nil
		}
		return &t
	}
	createdFrom := parseTime("created_from")
	createdTo := parseTime("created_to")
	updatedFrom := parseTime("updated_from")
	updatedTo := parseTime("updated_to")

	if createdFrom != nil && createdTo != nil && createdFrom.After(*createdTo) {
		details["created_from"] = "created_from cannot be after created_to"
	}
	if updatedFrom != nil && updatedTo != nil && updatedFrom.After(*updatedTo) {
		details["updated_from"] = "updated_from cannot be after updated_to"
	}

	if len(details) > 0 {
		return nil, apperrors.NewBadRequest("Invalid query parameters", details)
	}

	return &ListFilesParams{
		ViewerID:      viewerID,
		OwnerID:       viewerID,
		CourseIDs:     courseIDs,
		StudyGuideIDs: studyGuideIDs,
		Scope:         scope,
		Status:        status,
		MimeType:      mimeType,
		MinSize:       minSize,
		MaxSize:       maxSize,
		CreatedFrom:   createdFrom,
		CreatedTo:     createdTo,
		UpdatedFrom:   updatedFrom,
		UpdatedTo:     updatedTo,
		Q:             searchQ,
		SortField:     sortField,
		SortDir:       sortDir,
		PageLimit:     pageLimit,
		Cursor:        cursor,
	}, nil
}

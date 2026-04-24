package studyguides

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SortField identifies the column the ListStudyGuides query orders by.
type SortField string

// SortDir is the direction of a list query.
type SortDir string

const (
	SortFieldScore   SortField = "score"
	SortFieldViews   SortField = "views"
	SortFieldNewest  SortField = "newest"
	SortFieldUpdated SortField = "updated"

	SortDirAsc  SortDir = "asc"
	SortDirDesc SortDir = "desc"
)

const (
	// DefaultPageLimit matches the openapi.yaml default on page_limit.
	DefaultPageLimit int32 = 25
	// MaxPageLimit matches the openapi.yaml maximum on page_limit.
	MaxPageLimit int32 = 100
	// MaxSearchLength matches the openapi.yaml maxLength on q.
	MaxSearchLength int = 200
	// MaxTagLength matches the openapi.yaml items.maxLength on tag.
	MaxTagLength int = 50
)

// Cursor is the polymorphic keyset cursor for ListStudyGuides. Only the
// fields relevant to the active sort are populated on encode; the rest
// stay nil and are omitted from the JSON token.
//
// Score-sorted pages carry (VoteScore, ViewCount, UpdatedAt, ID) so the
// per-row aggregate vote_score is part of the strict total order. Views
// carries (ViewCount, UpdatedAt, ID). Newest carries (CreatedAt, ID).
// Updated carries (UpdatedAt, ID). ID is always the final tiebreaker.
type Cursor struct {
	ID        uuid.UUID  `json:"id"`
	VoteScore *int64     `json:"vote_score,omitempty"`
	ViewCount *int64     `json:"view_count,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// EncodeCursor serializes a Cursor into a base64-encoded JSON string.
func EncodeCursor(c Cursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("EncodeCursor: marshal: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor parses a base64-encoded JSON token back into a Cursor.
// Returns an error for malformed base64 or JSON; the handler maps that
// to a 400 VALIDATION_ERROR with the spec's "invalid cursor value".
func DecodeCursor(s string) (Cursor, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("DecodeCursor: base64: %w", err)
	}
	var c Cursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return Cursor{}, fmt.Errorf("DecodeCursor: json: %w", err)
	}
	return c, nil
}

// ListStudyGuidesParams is the input to Service.ListStudyGuides.
// ViewerID is the JWT viewer and drives the visibility predicate on
// the underlying SQL queries (ASK-211): public guides surface to
// everyone, private guides surface only to the creator, direct-user
// grantees, or course grantees enrolled via course_members.
type ListStudyGuidesParams struct {
	CourseID uuid.UUID
	ViewerID uuid.UUID
	Q        *string
	Tags     []string
	SortBy   SortField
	SortDir  SortDir
	Limit    int32
	Cursor   *Cursor
}

// ListStudyGuidesResult is the output of Service.ListStudyGuides.
type ListStudyGuidesResult struct {
	StudyGuides []StudyGuide
	HasMore     bool
	NextCursor  *string
}

// MySortField is the sort enum for GET /api/me/study-guides (ASK-131).
// Single direction per variant (the endpoint intentionally does not
// expose a sort_dir query param).
type MySortField string

const (
	// MySortFieldUpdated orders by updated_at DESC. Default.
	MySortFieldUpdated MySortField = "updated"
	// MySortFieldNewest orders by created_at DESC.
	MySortFieldNewest MySortField = "newest"
	// MySortFieldTitle orders by LOWER(title) ASC (case-insensitive).
	MySortFieldTitle MySortField = "title"
)

// MyCursor is the keyset cursor for ListMyStudyGuides (ASK-131).
// Only the field matching the active sort variant is populated; the
// rest stay nil and are omitted from the JSON token. ID is always
// the tiebreaker. Kept separate from studyguides.Cursor so the two
// encodings can evolve independently -- the course-scoped list uses
// score / views aggregates this one doesn't carry.
type MyCursor struct {
	ID         uuid.UUID  `json:"id"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	CreatedAt  *time.Time `json:"created_at,omitempty"`
	TitleLower *string    `json:"title_lower,omitempty"`
}

// EncodeMyCursor serializes a MyCursor into a base64-encoded JSON token.
func EncodeMyCursor(c MyCursor) (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("EncodeMyCursor: marshal: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeMyCursor parses a base64-encoded JSON token back into a
// MyCursor. The handler maps a decode error to 400 "invalid cursor
// value".
func DecodeMyCursor(s string) (MyCursor, error) {
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return MyCursor{}, fmt.Errorf("DecodeMyCursor: base64: %w", err)
	}
	var c MyCursor
	if err := json.Unmarshal(raw, &c); err != nil {
		return MyCursor{}, fmt.Errorf("DecodeMyCursor: json: %w", err)
	}
	return c, nil
}

// ListMyStudyGuidesParams is the input to Service.ListMyStudyGuides
// (ASK-131). ViewerID drives the creator-only scope (the endpoint
// always filters to the JWT viewer's own guides). CourseID is
// optional; when nil, returns guides across all courses.
type ListMyStudyGuidesParams struct {
	ViewerID uuid.UUID
	CourseID *uuid.UUID
	SortBy   MySortField
	Limit    int32
	Cursor   *MyCursor
}

// ListMyStudyGuidesResult is the output of Service.ListMyStudyGuides.
// NextCursor is *string so the wire field renders as explicit JSON
// null on the last page.
type ListMyStudyGuidesResult struct {
	StudyGuides []MyStudyGuide
	HasMore     bool
	NextCursor  *string
}

// GetStudyGuideParams is the input to Service.GetStudyGuide. ViewerID
// is used to fetch the user's own vote on the guide (user_vote in the
// response). A missing viewer vote ships as nil on StudyGuideDetail.
type GetStudyGuideParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
}

const (
	// MaxTitleLength matches openapi.yaml CreateStudyGuideRequest.title.maxLength.
	MaxTitleLength int = 500
	// MaxDescriptionLength matches openapi.yaml CreateStudyGuideRequest.description.maxLength.
	MaxDescriptionLength int = 2000
	// MaxContentLength matches openapi.yaml CreateStudyGuideRequest.content.maxLength.
	MaxContentLength int = 100000
	// MaxTagsCount matches openapi.yaml CreateStudyGuideRequest.tags.maxItems.
	MaxTagsCount int = 20
)

// CreateStudyGuideParams is the input to Service.CreateStudyGuide.
// CreatorID is taken from the JWT in the handler -- the spec
// explicitly forbids accepting a creator id from the request body
// (would be a privilege-attribution forge vector).
type CreateStudyGuideParams struct {
	CourseID    uuid.UUID
	CreatorID   uuid.UUID
	Title       string
	Description *string
	Content     *string
	Tags        []string
	// Visibility is optional -- nil means "use DB default (private)".
	// Valid values: "private", "public". Validated by the service.
	Visibility *string
}

// DeleteStudyGuideParams is the input to Service.DeleteStudyGuide.
// ViewerID drives the creator-only authorization check; the service
// returns apperrors.NewForbidden if it doesn't match the row's
// creator_id.
type DeleteStudyGuideParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
}

// UpdateStudyGuideParams is the input to Service.UpdateStudyGuide
// (ASK-129). Every updatable field is a pointer so a nil value
// reliably encodes "field absent in the request body" -- distinct
// from "field provided as empty/zero". The service rejects an
// all-nil-fields call as 400 'at least one field required' before
// SQL.
//
// Tag semantics: nil = don't touch existing tags; non-nil (even
// length 0) = REPLACE existing tags with the given list (after
// normalization).
type UpdateStudyGuideParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
	Title        *string
	Description  *string
	Content      *string
	Tags         *[]string
	// Visibility is optional -- nil preserves the current value via
	// COALESCE in the SQL. Valid values: "private", "public".
	Visibility *string
}

// CastVoteParams is the input to Service.CastVote (ASK-139). ViewerID
// is taken from the JWT in the handler. Vote is the GuideVote enum
// declared in model.go (mirrors the vote_direction Postgres enum).
type CastVoteParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
	Vote         GuideVote
}

// CastVoteResult is the output of Service.CastVote. Returns the
// post-upsert state so the handler can build CastVoteResponse without
// re-querying.
type CastVoteResult struct {
	Vote      GuideVote
	VoteScore int64
}

// RemoveVoteParams is the input to Service.RemoveVote (ASK-141).
// ViewerID identifies whose vote is being removed (always the JWT
// viewer; we never let one user remove another's vote).
type RemoveVoteParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
}

// RecommendStudyGuideParams is the input to Service.RecommendStudyGuide
// (ASK-147). ViewerID identifies the recommender (taken from the JWT;
// never accepted from the body).
type RecommendStudyGuideParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
}

// Recommendation is the output of Service.RecommendStudyGuide. The
// nested Recommender uses the same Creator privacy floor as everywhere
// else in this package (id + first_name + last_name only).
type Recommendation struct {
	StudyGuideID uuid.UUID
	Recommender  Creator
	CreatedAt    time.Time
}

// RemoveRecommendationParams is the input to
// Service.RemoveRecommendation (ASK-101). ViewerID identifies whose
// recommendation is being removed (always the JWT viewer).
type RemoveRecommendationParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
}

const (
	// MaxResourceTitleLength matches openapi.yaml AttachResourceRequest.title.maxLength.
	MaxResourceTitleLength int = 500
	// MaxResourceURLLength matches openapi.yaml AttachResourceRequest.url.maxLength.
	MaxResourceURLLength int = 2000
	// MaxResourceDescriptionLength matches openapi.yaml AttachResourceRequest.description.maxLength.
	MaxResourceDescriptionLength int = 1000
)

// AttachResourceParams is the input to Service.AttachResource (ASK-111).
// Title and URL are required. Description is optional. Type defaults
// to ResourceTypeLink server-side when nil/empty -- the openapi schema
// declares the same default.
//
// AttachedBy comes from the JWT viewer; the request body has no
// equivalent field (would be a privilege-attribution forge vector).
type AttachResourceParams struct {
	StudyGuideID uuid.UUID
	AttachedBy   uuid.UUID
	Title        string
	URL          string
	Type         ResourceType
	Description  *string
}

// DetachResourceParams is the input to Service.DetachResource
// (ASK-116). ViewerID drives the dual-authz check (guide-creator OR
// attached_by); the service maps a non-match to apperrors.NewForbidden.
type DetachResourceParams struct {
	StudyGuideID uuid.UUID
	ResourceID   uuid.UUID
	ViewerID     uuid.UUID
}

// AttachFileParams is the input to Service.AttachFile (ASK-121).
// AttacherID comes from the JWT viewer; the request body is empty so
// the only inputs are the two path params + viewer. The service
// rejects non-owner attachers with 403 (only the file owner can put
// their files on guides).
type AttachFileParams struct {
	StudyGuideID uuid.UUID
	FileID       uuid.UUID
	AttacherID   uuid.UUID
}

// FileAttachment is the output of Service.AttachFile. Mirrors the
// study_guide_files join row -- intentionally narrow (no expanded
// file metadata) so the wire shape matches the join table and the
// privacy floor for full file metadata stays on GET /study-guides/{id}.
type FileAttachment struct {
	FileID       uuid.UUID
	StudyGuideID uuid.UUID
	CreatedAt    time.Time
}

// DetachFileParams is the input to Service.DetachFile (ASK-124).
// ViewerID drives the dual-authz check (file owner OR guide creator);
// broader than POST (which is owner-only) so a guide creator can
// curate their guide's attached files without owning every file.
type DetachFileParams struct {
	StudyGuideID uuid.UUID
	FileID       uuid.UUID
	ViewerID     uuid.UUID
}

// CreateGrantParams is the input to Service.CreateGrant (ASK-211).
// ViewerID (the guide's creator -- enforced by the service) is the
// authz pivot; non-creators receive ErrForbidden.
//
// grantee_type must be "user" or "course" (study-guide grants do NOT
// accept "study_guide" as a grantee_type); permission must be one of
// "view", "share", "delete". Validated by the service before the DB
// round trip.
type CreateGrantParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
	GranteeType  string
	GranteeID    uuid.UUID
	Permission   string
}

// RevokeGrantParams is the input to Service.RevokeGrant (ASK-211).
// Same creator-only authz as CreateGrant. A 0-rows delete surfaces
// as 404 (not idempotent, mirrors file_grants).
type RevokeGrantParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
	GranteeType  string
	GranteeID    uuid.UUID
	Permission   string
}

// ListGrantsParams is the input to Service.ListGrants (ASK-211).
// Creator-only; non-creators receive ErrForbidden before the list
// query runs.
type ListGrantsParams struct {
	StudyGuideID uuid.UUID
	ViewerID     uuid.UUID
}

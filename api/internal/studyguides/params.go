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
type ListStudyGuidesParams struct {
	CourseID uuid.UUID
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

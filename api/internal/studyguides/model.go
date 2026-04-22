// Package studyguides hosts the domain types, params, mappers, and
// service logic for the study-guide surface. Mirrors the layering of
// internal/courses -- repository interface + sqlc-backed impl, service
// with per-sort-variant dispatch, pointer-free domain types that the
// handler projects onto the generated wire schema.
package studyguides

import (
	"time"

	"github.com/google/uuid"
)

// Creator is the compact user payload embedded in StudyGuide. Same
// privacy floor as SectionMember in the courses package: id +
// first_name + last_name only, no email or clerk_id.
type Creator struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
}

// StudyGuide is the list-row domain type returned by
// Service.ListStudyGuides. Excludes `content` (only on the get-by-id
// endpoint) to keep the list payload small.
type StudyGuide struct {
	ID            uuid.UUID
	Title         string
	Description   *string
	Tags          []string
	Creator       Creator
	CourseID      uuid.UUID
	VoteScore     int64
	ViewCount     int64
	IsRecommended bool
	QuizCount     int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// MyStudyGuide is the list-row domain type returned by
// Service.ListMyStudyGuides (ASK-131). Same projection as StudyGuide
// plus a nullable DeletedAt so the owner can see their own soft-
// deleted guides. Live guides have DeletedAt nil; soft-deleted
// guides carry a non-nil timestamp. The wire shape renders the
// field as `deleted_at: null` vs `deleted_at: "<timestamp>"` so
// the frontend can distinguish via `=== null` without an undefined
// case.
type MyStudyGuide struct {
	ID            uuid.UUID
	Title         string
	Description   *string
	Tags          []string
	Creator       Creator
	CourseID      uuid.UUID
	VoteScore     int64
	ViewCount     int64
	IsRecommended bool
	QuizCount     int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// GuideVote is the domain enum for a vote direction on a study guide.
// Mirrors the vote_direction Postgres enum and the openapi
// StudyGuideDetailResponse.user_vote enum.
type GuideVote string

const (
	GuideVoteUp   GuideVote = "up"
	GuideVoteDown GuideVote = "down"
)

// GuideCourseSummary is the compact course payload embedded in a
// StudyGuideDetail. Parallels EnrollmentCourse in the courses
// package but lives here so the two surfaces can evolve independently.
type GuideCourseSummary struct {
	ID         uuid.UUID
	Department string
	Number     string
	Title      string
}

// ResourceType mirrors the resource_type Postgres enum and the
// openapi ResourceSummary.type enum.
type ResourceType string

const (
	ResourceTypeLink    ResourceType = "link"
	ResourceTypeVideo   ResourceType = "video"
	ResourceTypeArticle ResourceType = "article"
	ResourceTypePDF     ResourceType = "pdf"
)

// Resource is the compact payload for an attached resource in the
// study-guide detail. No creator/uploader info -- the caller doesn't
// need to know who attached the resource.
type Resource struct {
	ID          uuid.UUID
	Title       string
	URL         string
	Type        ResourceType
	Description *string
	CreatedAt   time.Time
}

// Quiz is the compact payload for a quiz attached to a study guide.
// Only id + title + question_count -- no creator, no content, no
// scoring config. Detailed quiz fields are exposed by the quiz detail
// endpoint (future ticket ASK-142).
type Quiz struct {
	ID            uuid.UUID
	Title         string
	QuestionCount int64
}

// GuideFile is the compact payload for a file attached to a study
// guide. Privacy floor: id + name + mime_type + size only. No
// user_id, no s3_key, no checksum.
type GuideFile struct {
	ID       uuid.UUID
	Name     string
	MimeType string
	Size     int64
}

// StudyGuideDetail is the get-by-id domain type: the full guide with
// all embedded nested arrays + the viewer's own vote state.
// UserVote is *GuideVote so the wire shape can emit JSON null when
// the viewer has not voted, distinguishing "not voted" from "voted
// with empty string".
type StudyGuideDetail struct {
	ID            uuid.UUID
	Title         string
	Description   *string
	Content       *string
	Tags          []string
	Creator       Creator
	Course        GuideCourseSummary
	VoteScore     int64
	UserVote      *GuideVote
	ViewCount     int64
	IsRecommended bool
	RecommendedBy []Creator
	Quizzes       []Quiz
	Resources     []Resource
	Files         []GuideFile
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

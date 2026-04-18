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

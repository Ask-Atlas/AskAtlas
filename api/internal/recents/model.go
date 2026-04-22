// Package recents implements the GET /api/me/recents service surface
// (ASK-145): a per-viewer merged list of the most recently viewed
// files, study guides, and courses, used to power the sidebar
// "Recents" section.
//
// The merge happens in the Go service rather than SQL (a UNION over
// three differently-shaped rows would force a lowest-common-
// denominator projection or per-row casting), which keeps each query
// hitting its own focused (user_id, viewed_at) index and lets the
// service map types into a discriminated-union domain shape.
package recents

import (
	"time"

	"github.com/google/uuid"
)

// EntityType discriminates the populated payload on a RecentItem.
// The string values match the openapi enum exactly so the mapper
// can cast directly without a switch.
type EntityType string

const (
	EntityTypeFile       EntityType = "file"
	EntityTypeStudyGuide EntityType = "study_guide"
	EntityTypeCourse     EntityType = "course"
)

// RecentFileSummary is the compact file payload embedded in a
// RecentItem when EntityType == EntityTypeFile. Mirrors
// api.RecentFileSummary on the wire.
type RecentFileSummary struct {
	ID       uuid.UUID
	Name     string
	MimeType string
}

// RecentStudyGuideSummary is the compact study-guide payload
// embedded in a RecentItem when EntityType == EntityTypeStudyGuide.
// Includes the parent course's department + number so the sidebar
// can render "CPTS 322 -- Binary Trees Cheat Sheet" without a
// follow-up request. Mirrors api.RecentStudyGuideSummary on the wire.
type RecentStudyGuideSummary struct {
	ID               uuid.UUID
	Title            string
	CourseDepartment string
	CourseNumber     string
}

// RecentCourseSummary is the compact course payload embedded in a
// RecentItem when EntityType == EntityTypeCourse. Mirrors
// api.RecentCourseSummary on the wire.
type RecentCourseSummary struct {
	ID         uuid.UUID
	Department string
	Number     string
	Title      string
}

// RecentItem is one merged recent row. Exactly one of File,
// StudyGuide, or Course is non-nil; EntityType declares which one.
// EntityID mirrors the populated summary's ID so callers can route
// off the envelope without unpacking the per-type payload.
//
// Pointer fields (rather than a sealed-interface union) are chosen
// to match the existing codebase style and to round-trip cleanly
// through the api.RecentItem JSON struct (which uses *Summary
// fields with omitempty).
type RecentItem struct {
	EntityType EntityType
	EntityID   uuid.UUID
	ViewedAt   time.Time
	File       *RecentFileSummary
	StudyGuide *RecentStudyGuideSummary
	Course     *RecentCourseSummary
}

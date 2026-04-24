// Package favorites implements the GET /api/me/favorites service
// surface (ASK-151): a per-viewer paginated list of favorited
// files, study guides, and courses, used to power the sidebar
// "Starred" section and the /me/saved page.
//
// The merge happens in the Go service rather than SQL because a
// UNION over three differently-shaped rows would force a lowest-
// common-denominator projection or per-row casting, and offset
// pagination across a UNION is awkward to keep stable. Per-table
// queries each hit their focused (user_id, created_at) index;
// merge + offset/limit happen in-process.
//
// The shape mirrors internal/recents (ASK-145) deliberately --
// both are sidebar-data endpoints with the same discriminated-
// union envelope. Kept as separate packages rather than a shared
// `me/` package so each ticket evolves independently.
package favorites

import (
	"time"

	"github.com/google/uuid"
)

// EntityType discriminates the populated payload on a FavoriteItem.
// The string values match the openapi enum exactly so the mapper
// can cast directly.
type EntityType string

const (
	EntityTypeFile       EntityType = "file"
	EntityTypeStudyGuide EntityType = "study_guide"
	EntityTypeCourse     EntityType = "course"
)

// Valid reports whether s is a recognized EntityType. Used by the
// service to validate the optional entity_type query filter; the
// openapi wrapper enforces the enum at the HTTP boundary, this is
// defense in depth for internal Go callers.
func (e EntityType) Valid() bool {
	switch e {
	case EntityTypeFile, EntityTypeStudyGuide, EntityTypeCourse:
		return true
	}
	return false
}

// FavoriteFileSummary mirrors api.FavoriteFileSummary on the wire.
type FavoriteFileSummary struct {
	ID       uuid.UUID
	Name     string
	MimeType string
}

// FavoriteStudyGuideSummary mirrors api.FavoriteStudyGuideSummary
// on the wire. Includes the parent course's department + number so
// the sidebar can render "CPTS 322 -- <title>" without a follow-up.
type FavoriteStudyGuideSummary struct {
	ID               uuid.UUID
	Title            string
	CourseDepartment string
	CourseNumber     string
}

// FavoriteCourseSummary mirrors api.FavoriteCourseSummary on the wire.
type FavoriteCourseSummary struct {
	ID         uuid.UUID
	Department string
	Number     string
	Title      string
}

// FavoriteItem is one merged favorite row. Exactly one of File,
// StudyGuide, or Course is non-nil; EntityType declares which one.
// EntityID mirrors the populated summary's ID so callers can route
// off the envelope without unpacking the per-type payload.
type FavoriteItem struct {
	EntityType  EntityType
	EntityID    uuid.UUID
	FavoritedAt time.Time
	File        *FavoriteFileSummary
	StudyGuide  *FavoriteStudyGuideSummary
	Course      *FavoriteCourseSummary
}

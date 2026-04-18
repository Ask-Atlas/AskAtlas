package courses

import (
	"time"

	"github.com/google/uuid"
)

// Course is the list-view domain type: course metadata plus the embedded
// school summary needed to render a course card without a second round-trip.
type Course struct {
	ID          uuid.UUID
	School      SchoolSummary
	Department  string
	Number      string
	Title       string
	Description *string
	CreatedAt   time.Time
}

// SchoolSummary is the compact school payload embedded inside other resources
// (Course, future StudyGuide). Mirrors api.SchoolSummary on the wire.
type SchoolSummary struct {
	ID      uuid.UUID
	Name    string
	Acronym string
	City    *string
	State   *string
	Country *string
}

// Section is a course section as it appears in the detail view: term,
// optional section code + instructor, and a live roster count.
type Section struct {
	ID             uuid.UUID
	Term           string
	SectionCode    *string
	InstructorName *string
	MemberCount    int64
}

// CourseDetail is the get-by-id domain type: a Course plus the inline list
// of all its sections (always non-nil; empty slice when the course has none).
type CourseDetail struct {
	Course
	Sections []Section
}

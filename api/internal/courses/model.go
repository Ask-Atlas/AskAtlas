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

// MemberRole names the values stored in the course_role enum. The wire
// schema matches; the type alias keeps callers from hand-stringing the
// values and accidentally drifting from the DB enum.
type MemberRole string

const (
	MemberRoleStudent    MemberRole = "student"
	MemberRoleTA         MemberRole = "ta"
	MemberRoleInstructor MemberRole = "instructor"
)

// Membership is the domain representation of a user's enrollment in a
// course section. Returned by Service.JoinSection.
type Membership struct {
	UserID    uuid.UUID
	SectionID uuid.UUID
	Role      MemberRole
	JoinedAt  time.Time
}

// EnrollmentSection is the compact section payload embedded in an
// Enrollment. Mirrors api.EnrollmentSectionSummary on the wire.
type EnrollmentSection struct {
	ID             uuid.UUID
	Term           string
	SectionCode    *string
	InstructorName *string
}

// EnrollmentCourse is the compact course payload embedded in an
// Enrollment. No description/created_at -- the course detail endpoint
// is the source of truth for those.
type EnrollmentCourse struct {
	ID         uuid.UUID
	Department string
	Number     string
	Title      string
}

// EnrollmentSchool is the *very* compact school payload embedded in an
// Enrollment. Only id + acronym -- the dashboard renders an acronym
// badge per row, and pulling the full SchoolSummary would bloat the
// payload for users in many courses.
type EnrollmentSchool struct {
	ID      uuid.UUID
	Acronym string
}

// Enrollment is the dashboard row returned by Service.ListMyEnrollments:
// one section the viewer is enrolled in, with just enough course and
// school context to render a card without a follow-up request.
type Enrollment struct {
	Section  EnrollmentSection
	Course   EnrollmentCourse
	School   EnrollmentSchool
	Role     MemberRole
	JoinedAt time.Time
}

// MembershipCheck is the per-section enrolled/not-enrolled probe
// returned by Service.CheckMembership. Role and JoinedAt are pointer
// types because the wire shape requires explicit JSON nulls (not
// omitted) when the viewer is not enrolled.
type MembershipCheck struct {
	Enrolled bool
	Role     *MemberRole
	JoinedAt *time.Time
}

// SectionMember is the per-row payload returned by ListSectionMembers.
// Privacy floor: only the five fields the public schema exposes; no
// email, clerk_id, or other PII is reachable through this type.
type SectionMember struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
	Role      MemberRole
	JoinedAt  time.Time
}

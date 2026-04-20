// Package dashboard implements the GET /api/me/dashboard service
// surface (ASK-155): the aggregated home-page payload combining
// enrolled courses for the current term, recently updated study
// guides, practice stats + recent sessions, and file totals +
// recent files.
//
// The endpoint is deliberately a single round-trip with ~10
// underlying queries because the home page is the most-loaded
// authenticated surface and we want one network hop to render it.
// Each section is independent; a DB error in any one section
// fails the whole request (per spec) rather than returning
// partial data that the frontend would have to render around.
//
// Soft-deleted entities are excluded everywhere -- counts, lists,
// and aggregate sums all filter on the same predicates the rest
// of the API uses (study_guides.deleted_at IS NULL,
// files.deletion_status IS NULL).
package dashboard

import (
	"time"

	"github.com/google/uuid"
)

// MemberRole names the values stored in the course_role enum and
// surfaced via the dashboard's per-course summary. The string
// values match the openapi enum exactly so the mapper can cast
// directly without a switch.
type MemberRole string

const (
	MemberRoleStudent    MemberRole = "student"
	MemberRoleTA         MemberRole = "ta"
	MemberRoleInstructor MemberRole = "instructor"
)

// DashboardCourseSummary mirrors api.DashboardCourseSummary on the
// wire. Includes the viewer's role + the section's term so the
// home page can render "CPTS 322 (Spring 2026, student)" without
// follow-up requests.
type DashboardCourseSummary struct {
	ID          uuid.UUID
	Department  string
	Number      string
	Title       string
	Role        MemberRole
	SectionTerm string
}

// DashboardCoursesSection is the courses block of the dashboard
// payload. CurrentTerm is *string to encode "no enrollments at all"
// as JSON null; CurrentTerm being non-nil but EnrolledCount==0 is
// not a valid state (the term resolver only returns a term when
// the user has at least one enrollment).
type DashboardCoursesSection struct {
	EnrolledCount int32
	CurrentTerm   *string
	Courses       []DashboardCourseSummary
}

// DashboardStudyGuideSummary mirrors api.DashboardStudyGuideSummary
// on the wire.
type DashboardStudyGuideSummary struct {
	ID               uuid.UUID
	Title            string
	CourseDepartment string
	CourseNumber     string
	UpdatedAt        time.Time
}

// DashboardStudyGuidesSection is the study_guides block of the
// dashboard payload.
type DashboardStudyGuidesSection struct {
	CreatedCount int32
	Recent       []DashboardStudyGuideSummary
}

// DashboardSessionSummary mirrors api.DashboardSessionSummary on
// the wire. ScorePercentage is the rounded per-session accuracy
// (correct_answers / total_questions * 100) computed in Go so the
// SQL stays simple; if total_questions is 0 the percentage is 0
// rather than divide-by-zero (mirroring the overall_accuracy
// behavior).
type DashboardSessionSummary struct {
	ID              uuid.UUID
	QuizTitle       string
	StudyGuideTitle string
	ScorePercentage int32
	CompletedAt     time.Time
}

// DashboardPracticeSection is the practice block of the dashboard
// payload. All three integer fields are 0 when the user has no
// completed sessions; RecentSessions is an empty slice in that
// case (never nil).
type DashboardPracticeSection struct {
	SessionsCompleted      int32
	TotalQuestionsAnswered int32
	OverallAccuracy        int32
	RecentSessions         []DashboardSessionSummary
}

// DashboardFileSummary mirrors api.DashboardFileSummary on the wire.
type DashboardFileSummary struct {
	ID        uuid.UUID
	Name      string
	MimeType  string
	UpdatedAt time.Time
}

// DashboardFilesSection is the files block of the dashboard
// payload. TotalSize is bytes (int64) because file sizes can
// exceed int32 limits in aggregate.
type DashboardFilesSection struct {
	TotalCount int32
	TotalSize  int64
	Recent     []DashboardFileSummary
}

// DashboardData is the assembled payload returned by Service.GetDashboard.
// All four sections are always present (never nil); their inner
// list fields default to empty slices, count fields to 0.
type DashboardData struct {
	Courses     DashboardCoursesSection
	StudyGuides DashboardStudyGuidesSection
	Practice    DashboardPracticeSection
	Files       DashboardFilesSection
}

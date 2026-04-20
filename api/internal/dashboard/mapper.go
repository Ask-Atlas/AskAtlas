package dashboard

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
)

// mapCourse converts a sqlc ListEnrolledCoursesForTerm row into a
// DashboardCourseSummary.
func mapCourse(r db.ListEnrolledCoursesForTermRow) (DashboardCourseSummary, error) {
	id, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return DashboardCourseSummary{}, fmt.Errorf("mapCourse: course id: %w", err)
	}
	return DashboardCourseSummary{
		ID:          id,
		Department:  r.CourseDepartment,
		Number:      r.CourseNumber,
		Title:       r.CourseTitle,
		Role:        MemberRole(r.MemberRole),
		SectionTerm: r.SectionTerm,
	}, nil
}

// mapStudyGuide converts a sqlc ListRecentUserStudyGuides row into
// a DashboardStudyGuideSummary.
func mapStudyGuide(r db.ListRecentUserStudyGuidesRow) (DashboardStudyGuideSummary, error) {
	id, err := utils.PgxToGoogleUUID(r.StudyGuideID)
	if err != nil {
		return DashboardStudyGuideSummary{}, fmt.Errorf("mapStudyGuide: study guide id: %w", err)
	}
	return DashboardStudyGuideSummary{
		ID:               id,
		Title:            r.StudyGuideTitle,
		CourseDepartment: r.CourseDepartment,
		CourseNumber:     r.CourseNumber,
		UpdatedAt:        r.UpdatedAt.Time,
	}, nil
}

// mapSession converts a sqlc ListRecentUserSessions row into a
// DashboardSessionSummary. Computes per-session score percentage
// in Go: ROUND(100 * correct / total) when total > 0, else 0.
// Mirrors the overall_accuracy formula -- "no questions answered"
// is treated as 0% rather than NaN/undefined.
func mapSession(r db.ListRecentUserSessionsRow) (DashboardSessionSummary, error) {
	id, err := utils.PgxToGoogleUUID(r.SessionID)
	if err != nil {
		return DashboardSessionSummary{}, fmt.Errorf("mapSession: session id: %w", err)
	}
	return DashboardSessionSummary{
		ID:              id,
		QuizTitle:       r.QuizTitle,
		StudyGuideTitle: r.StudyGuideTitle,
		ScorePercentage: percentage(int64(r.CorrectAnswers), int64(r.TotalQuestions)),
		CompletedAt:     r.CompletedAt.Time,
	}, nil
}

// mapFile converts a sqlc ListRecentUserFiles row into a
// DashboardFileSummary.
func mapFile(r db.ListRecentUserFilesRow) (DashboardFileSummary, error) {
	id, err := utils.PgxToGoogleUUID(r.FileID)
	if err != nil {
		return DashboardFileSummary{}, fmt.Errorf("mapFile: file id: %w", err)
	}
	return DashboardFileSummary{
		ID:        id,
		Name:      r.FileName,
		MimeType:  r.FileMimeType,
		UpdatedAt: r.UpdatedAt.Time,
	}, nil
}

// percentage computes the rounded integer percentage of correct
// over total, returning 0 when total is 0 (division-by-zero
// guard). Mirrors the spec's
// `ROUND((SUM(correct) / NULLIF(SUM(total), 0)) * 100)` semantics.
// Used for both per-session ScorePercentage and the overall_accuracy
// computation in service.go.
func percentage(correct, total int64) int32 {
	if total <= 0 {
		return 0
	}
	// (correct*100 + total/2) / total is the standard "round-half-up"
	// integer-division trick, matching SQL ROUND() for positive
	// rationals. Using int64 throughout to avoid overflow on
	// correct*100 for power-user totals.
	return int32((correct*100 + total/2) / total)
}

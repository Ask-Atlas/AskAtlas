package recents

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
)

// mapRecentFile converts a sqlc ListRecentFiles row into a RecentItem
// with EntityType=file. The viewer's user_id is not on the row -- the
// query already filtered to the authed user, so the mapper does not
// need it.
func mapRecentFile(r db.ListRecentFilesRow) (RecentItem, error) {
	id, err := utils.PgxToGoogleUUID(r.FileID)
	if err != nil {
		return RecentItem{}, fmt.Errorf("mapRecentFile: file id: %w", err)
	}
	return RecentItem{
		EntityType: EntityTypeFile,
		EntityID:   id,
		ViewedAt:   r.ViewedAt.Time,
		File: &RecentFileSummary{
			ID:       id,
			Name:     r.FileName,
			MimeType: r.FileMimeType,
		},
	}, nil
}

// mapRecentStudyGuide converts a sqlc ListRecentStudyGuides row into
// a RecentItem with EntityType=study_guide. The (department, number)
// pair is sourced from the join onto courses (queries.sql) so the
// sidebar can render a "CPTS 322 -- <title>" label without a
// follow-up GET.
func mapRecentStudyGuide(r db.ListRecentStudyGuidesRow) (RecentItem, error) {
	id, err := utils.PgxToGoogleUUID(r.StudyGuideID)
	if err != nil {
		return RecentItem{}, fmt.Errorf("mapRecentStudyGuide: study guide id: %w", err)
	}
	return RecentItem{
		EntityType: EntityTypeStudyGuide,
		EntityID:   id,
		ViewedAt:   r.ViewedAt.Time,
		StudyGuide: &RecentStudyGuideSummary{
			ID:               id,
			Title:            r.StudyGuideTitle,
			CourseDepartment: r.CourseDepartment,
			CourseNumber:     r.CourseNumber,
		},
	}, nil
}

// mapRecentCourse converts a sqlc ListRecentCourses row into a
// RecentItem with EntityType=course.
func mapRecentCourse(r db.ListRecentCoursesRow) (RecentItem, error) {
	id, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return RecentItem{}, fmt.Errorf("mapRecentCourse: course id: %w", err)
	}
	return RecentItem{
		EntityType: EntityTypeCourse,
		EntityID:   id,
		ViewedAt:   r.ViewedAt.Time,
		Course: &RecentCourseSummary{
			ID:         id,
			Department: r.CourseDepartment,
			Number:     r.CourseNumber,
			Title:      r.CourseTitle,
		},
	}, nil
}

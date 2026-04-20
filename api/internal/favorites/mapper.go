package favorites

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
)

// mapFavoriteFile converts a sqlc ListFileFavorites row into a
// FavoriteItem with EntityType=file. The viewer's user_id is not
// on the row -- the query already filtered to the authed user.
func mapFavoriteFile(r db.ListFileFavoritesRow) (FavoriteItem, error) {
	id, err := utils.PgxToGoogleUUID(r.FileID)
	if err != nil {
		return FavoriteItem{}, fmt.Errorf("mapFavoriteFile: file id: %w", err)
	}
	return FavoriteItem{
		EntityType:  EntityTypeFile,
		EntityID:    id,
		FavoritedAt: r.FavoritedAt.Time,
		File: &FavoriteFileSummary{
			ID:       id,
			Name:     r.FileName,
			MimeType: r.FileMimeType,
		},
	}, nil
}

// mapFavoriteStudyGuide converts a sqlc ListStudyGuideFavorites row
// into a FavoriteItem with EntityType=study_guide. The
// (department, number) pair is sourced from the join onto courses
// so the sidebar can render a "CPTS 322 -- <title>" label without a
// follow-up GET.
func mapFavoriteStudyGuide(r db.ListStudyGuideFavoritesRow) (FavoriteItem, error) {
	id, err := utils.PgxToGoogleUUID(r.StudyGuideID)
	if err != nil {
		return FavoriteItem{}, fmt.Errorf("mapFavoriteStudyGuide: study guide id: %w", err)
	}
	return FavoriteItem{
		EntityType:  EntityTypeStudyGuide,
		EntityID:    id,
		FavoritedAt: r.FavoritedAt.Time,
		StudyGuide: &FavoriteStudyGuideSummary{
			ID:               id,
			Title:            r.StudyGuideTitle,
			CourseDepartment: r.CourseDepartment,
			CourseNumber:     r.CourseNumber,
		},
	}, nil
}

// mapFavoriteCourse converts a sqlc ListCourseFavorites row into a
// FavoriteItem with EntityType=course.
func mapFavoriteCourse(r db.ListCourseFavoritesRow) (FavoriteItem, error) {
	id, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return FavoriteItem{}, fmt.Errorf("mapFavoriteCourse: course id: %w", err)
	}
	return FavoriteItem{
		EntityType:  EntityTypeCourse,
		EntityID:    id,
		FavoritedAt: r.FavoritedAt.Time,
		Course: &FavoriteCourseSummary{
			ID:         id,
			Department: r.CourseDepartment,
			Number:     r.CourseNumber,
			Title:      r.CourseTitle,
		},
	}, nil
}

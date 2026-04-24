package refs

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
)

func mapStudyGuideRow(r db.ListStudyGuideRefSummariesRow) (Summary, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Summary{}, fmt.Errorf("refs.mapStudyGuideRow: %w", err)
	}
	qc := int(r.QuizCount)
	rec := r.IsRecommended
	return Summary{
		Type:  TypeStudyGuide,
		ID:    id,
		Title: r.Title,
		Course: &CourseInfo{
			Department: r.CourseDepartment,
			Number:     r.CourseNumber,
		},
		QuizCount:     &qc,
		IsRecommended: &rec,
	}, nil
}

func mapQuizRow(r db.ListQuizRefSummariesRow) (Summary, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Summary{}, fmt.Errorf("refs.mapQuizRow: %w", err)
	}
	qc := int(r.QuestionCount)
	return Summary{
		Type:          TypeQuiz,
		ID:            id,
		Title:         r.Title,
		QuestionCount: &qc,
		Creator: &CreatorInfo{
			FirstName: r.CreatorFirstName,
			LastName:  r.CreatorLastName,
		},
	}, nil
}

func mapFileRow(r db.ListFileRefSummariesRow) (Summary, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Summary{}, fmt.Errorf("refs.mapFileRow: %w", err)
	}
	sz := r.Size
	return Summary{
		Type:     TypeFile,
		ID:       id,
		Name:     r.Name,
		Size:     &sz,
		MimeType: r.MimeType,
		Status:   string(r.Status),
	}, nil
}

func mapCourseRow(r db.ListCourseRefSummariesRow) (Summary, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Summary{}, fmt.Errorf("refs.mapCourseRow: %w", err)
	}
	return Summary{
		Type:       TypeCourse,
		ID:         id,
		Title:      r.Title,
		Department: r.Department,
		Number:     r.Number,
		School: &SchoolInfo{
			Name:    r.SchoolName,
			Acronym: r.SchoolAcronym,
		},
	}, nil
}

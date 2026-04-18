package studyguides

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// sharedGuideRow holds the fields common to every sqlc-generated study
// guide list row (the 8 ListStudyGuides*Row variants). Each variant
// adapter below projects its typed row into this shared struct so
// mapStudyGuide can do the heavy lifting once.
type sharedGuideRow struct {
	ID               pgtype.UUID
	Title            string
	Description      pgtype.Text
	Tags             []string
	CourseID         pgtype.UUID
	ViewCount        int32
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	CreatorID        pgtype.UUID
	CreatorFirstName string
	CreatorLastName  string
	VoteScore        int64
	IsRecommended    bool
	QuizCount        int64
}

// mapStudyGuide converts a sharedGuideRow into the domain StudyGuide
// type. Privacy floor: only copies id + first_name + last_name from the
// creator -- any email/clerk_id present on a wider row type would be
// dropped here, but the SQL SELECT list is also the privacy floor so
// these fields shouldn't exist in the row anyway. The SQL-introspection
// guard test in commit 5 asserts that.
func mapStudyGuide(r sharedGuideRow) (StudyGuide, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return StudyGuide{}, fmt.Errorf("mapStudyGuide: id: %w", err)
	}
	courseID, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return StudyGuide{}, fmt.Errorf("mapStudyGuide: course id: %w", err)
	}
	creatorID, err := utils.PgxToGoogleUUID(r.CreatorID)
	if err != nil {
		return StudyGuide{}, fmt.Errorf("mapStudyGuide: creator id: %w", err)
	}
	return StudyGuide{
		ID: id,
		Creator: Creator{
			ID:        creatorID,
			FirstName: r.CreatorFirstName,
			LastName:  r.CreatorLastName,
		},
		CourseID:      courseID,
		Title:         r.Title,
		Description:   textPtr(r.Description),
		Tags:          append([]string(nil), r.Tags...),
		VoteScore:     r.VoteScore,
		ViewCount:     int64(r.ViewCount),
		IsRecommended: r.IsRecommended,
		QuizCount:     r.QuizCount,
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
	}, nil
}

// textPtr returns a *string for a nullable pgtype.Text column. Nil for
// SQL NULL so the handler emits JSON null (not the empty string).
func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

// Per-sort-variant adapters. Each projects the typed db row into the
// shared row struct so the rest of the package stays variant-agnostic.

func fromScoreDescRow(r db.ListStudyGuidesScoreDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromScoreAscRow(r db.ListStudyGuidesScoreAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromViewsDescRow(r db.ListStudyGuidesViewsDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromViewsAscRow(r db.ListStudyGuidesViewsAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromNewestDescRow(r db.ListStudyGuidesNewestDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromNewestAscRow(r db.ListStudyGuidesNewestAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromUpdatedDescRow(r db.ListStudyGuidesUpdatedDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromUpdatedAscRow(r db.ListStudyGuidesUpdatedAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

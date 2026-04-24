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
	Visibility       db.StudyGuideVisibility
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
		Description:   utils.TextPtr(r.Description),
		Tags:          append([]string(nil), r.Tags...),
		VoteScore:     r.VoteScore,
		ViewCount:     int64(r.ViewCount),
		IsRecommended: r.IsRecommended,
		QuizCount:     r.QuizCount,
		Visibility:    string(r.Visibility),
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
	}, nil
}

// mapStudyGuideDetail projects the main GetStudyGuideDetail row into
// the domain StudyGuideDetail type. The nested Quizzes, Resources,
// Files, RecommendedBy slices are fetched by separate queries and
// attached by the caller -- this helper only handles the row that
// comes out of GetStudyGuideDetail itself.
//
// Privacy floor: only copies id + first_name + last_name from creator
// fields -- any email/clerk_id present on a wider row type would be
// dropped here, but the SQL SELECT list is also the privacy floor so
// these fields shouldn't exist in the row anyway.
func mapStudyGuideDetail(r db.GetStudyGuideDetailRow) (StudyGuideDetail, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("mapStudyGuideDetail: id: %w", err)
	}
	courseID, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("mapStudyGuideDetail: course id: %w", err)
	}
	creatorID, err := utils.PgxToGoogleUUID(r.CreatorID)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("mapStudyGuideDetail: creator id: %w", err)
	}
	return StudyGuideDetail{
		ID: id,
		Creator: Creator{
			ID:        creatorID,
			FirstName: r.CreatorFirstName,
			LastName:  r.CreatorLastName,
		},
		Course: GuideCourseSummary{
			ID:         courseID,
			Department: r.CourseDepartment,
			Number:     r.CourseNumber,
			Title:      r.CourseTitle,
		},
		Title:         r.Title,
		Description:   utils.TextPtr(r.Description),
		Content:       utils.TextPtr(r.Content),
		Tags:          append([]string(nil), r.Tags...),
		VoteScore:     r.VoteScore,
		ViewCount:     int64(r.ViewCount),
		IsRecommended: r.IsRecommended,
		Visibility:    string(r.Visibility),
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
		// UserVote + nested arrays are attached by the service after
		// the sibling queries resolve.
	}, nil
}

// mapRecommender projects a ListGuideRecommenders row into the domain
// Creator type (the detail endpoint reuses the same privacy-floor
// payload for both creator and recommenders).
func mapRecommender(r db.ListGuideRecommendersRow) (Creator, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Creator{}, fmt.Errorf("mapRecommender: id: %w", err)
	}
	return Creator{
		ID:        id,
		FirstName: r.FirstName,
		LastName:  r.LastName,
	}, nil
}

// mapQuiz projects a ListGuideQuizzesWithQuestionCount row onto the
// domain Quiz type.
func mapQuiz(r db.ListGuideQuizzesWithQuestionCountRow) (Quiz, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Quiz{}, fmt.Errorf("mapQuiz: id: %w", err)
	}
	return Quiz{
		ID:            id,
		Title:         r.Title,
		QuestionCount: r.QuestionCount,
	}, nil
}

// mapResource projects a ListGuideResources row onto the domain
// Resource type.
func mapResource(r db.ListGuideResourcesRow) (Resource, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Resource{}, fmt.Errorf("mapResource: id: %w", err)
	}
	return Resource{
		ID:          id,
		Title:       r.Title,
		URL:         r.Url,
		Type:        ResourceType(r.Type),
		Description: utils.TextPtr(r.Description),
		CreatedAt:   r.CreatedAt.Time,
	}, nil
}

// mapGrant projects a sqlc db.StudyGuideGrant row onto the domain
// Grant type. Mirrors files.mapGrantRow.
func mapGrant(r db.StudyGuideGrant) (Grant, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrant: id: %w", err)
	}
	studyGuideID, err := utils.PgxToGoogleUUID(r.StudyGuideID)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrant: study guide id: %w", err)
	}
	granteeID, err := utils.PgxToGoogleUUID(r.GranteeID)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrant: grantee id: %w", err)
	}
	grantedBy, err := utils.PgxToGoogleUUID(r.GrantedBy)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrant: granted by: %w", err)
	}
	return Grant{
		ID:           id,
		StudyGuideID: studyGuideID,
		GranteeType:  string(r.GranteeType),
		GranteeID:    granteeID,
		Permission:   string(r.Permission),
		GrantedBy:    grantedBy,
		CreatedAt:    r.CreatedAt.Time,
	}, nil
}

// mapGuideFile projects a ListGuideFiles row onto the domain
// GuideFile type. Privacy floor -- no user_id, no s3_key, no checksum.
func mapGuideFile(r db.ListGuideFilesRow) (GuideFile, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return GuideFile{}, fmt.Errorf("mapGuideFile: id: %w", err)
	}
	return GuideFile{
		ID:       id,
		Name:     r.Name,
		MimeType: string(r.MimeType),
		Size:     r.Size,
	}, nil
}

// Per-sort-variant adapters. Each projects the typed db row into the
// shared row struct so the rest of the package stays variant-agnostic.

func fromScoreDescRow(r db.ListStudyGuidesScoreDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromScoreAscRow(r db.ListStudyGuidesScoreAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromViewsDescRow(r db.ListStudyGuidesViewsDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromViewsAscRow(r db.ListStudyGuidesViewsAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromNewestDescRow(r db.ListStudyGuidesNewestDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromNewestAscRow(r db.ListStudyGuidesNewestAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromUpdatedDescRow(r db.ListStudyGuidesUpdatedDescRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromUpdatedAscRow(r db.ListStudyGuidesUpdatedAscRow) sharedGuideRow {
	return sharedGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

// sharedMyGuideRow is the shape common to every
// ListMyStudyGuides*Row variant (ASK-131). Same fields as
// sharedGuideRow plus a nullable DeletedAt so the mapper can emit
// MyStudyGuide.DeletedAt as *time.Time.
type sharedMyGuideRow struct {
	ID               pgtype.UUID
	Title            string
	Description      pgtype.Text
	Tags             []string
	CourseID         pgtype.UUID
	ViewCount        int32
	Visibility       db.StudyGuideVisibility
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	CreatorID        pgtype.UUID
	CreatorFirstName string
	CreatorLastName  string
	VoteScore        int64
	IsRecommended    bool
	QuizCount        int64
}

// mapMyStudyGuide projects a sharedMyGuideRow into the domain
// MyStudyGuide. DeletedAt renders as *time.Time -- nil for live
// guides, non-nil for soft-deleted ones. Every other field mirrors
// mapStudyGuide's projection.
func mapMyStudyGuide(r sharedMyGuideRow) (MyStudyGuide, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return MyStudyGuide{}, fmt.Errorf("mapMyStudyGuide: id: %w", err)
	}
	courseID, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return MyStudyGuide{}, fmt.Errorf("mapMyStudyGuide: course id: %w", err)
	}
	creatorID, err := utils.PgxToGoogleUUID(r.CreatorID)
	if err != nil {
		return MyStudyGuide{}, fmt.Errorf("mapMyStudyGuide: creator id: %w", err)
	}
	return MyStudyGuide{
		ID: id,
		Creator: Creator{
			ID:        creatorID,
			FirstName: r.CreatorFirstName,
			LastName:  r.CreatorLastName,
		},
		CourseID:      courseID,
		Title:         r.Title,
		Description:   utils.TextPtr(r.Description),
		Tags:          append([]string(nil), r.Tags...),
		VoteScore:     r.VoteScore,
		ViewCount:     int64(r.ViewCount),
		IsRecommended: r.IsRecommended,
		QuizCount:     r.QuizCount,
		Visibility:    string(r.Visibility),
		CreatedAt:     r.CreatedAt.Time,
		UpdatedAt:     r.UpdatedAt.Time,
		DeletedAt:     utils.TimestamptzPtr(r.DeletedAt),
	}, nil
}

func fromMyUpdatedRow(r db.ListMyStudyGuidesUpdatedRow) sharedMyGuideRow {
	return sharedMyGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, DeletedAt: r.DeletedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromMyNewestRow(r db.ListMyStudyGuidesNewestRow) sharedMyGuideRow {
	return sharedMyGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, DeletedAt: r.DeletedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

func fromMyTitleRow(r db.ListMyStudyGuidesTitleRow) sharedMyGuideRow {
	return sharedMyGuideRow{
		ID: r.ID, Title: r.Title, Description: r.Description, Tags: r.Tags,
		CourseID: r.CourseID, ViewCount: r.ViewCount, Visibility: r.Visibility,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, DeletedAt: r.DeletedAt,
		CreatorID: r.CreatorID, CreatorFirstName: r.CreatorFirstName, CreatorLastName: r.CreatorLastName,
		VoteScore: r.VoteScore, IsRecommended: r.IsRecommended, QuizCount: r.QuizCount,
	}
}

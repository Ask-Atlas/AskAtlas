package courses

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// sharedCourseRow holds the common fields present on every sqlc-generated
// course row (GetCourseRow + the 8 ListCourses*Row variants). Each list
// variant gets its own adapter function below that projects the typed row
// into this shared struct so mapCourse can do the heavy lifting once.
type sharedCourseRow struct {
	ID          pgtype.UUID
	SchoolID    pgtype.UUID
	Department  string
	Number      string
	Title       string
	Description pgtype.Text
	CreatedAt   pgtype.Timestamptz
	SID         pgtype.UUID
	SName       string
	SAcronym    string
	SCity       pgtype.Text
	SState      pgtype.Text
	SCountry    pgtype.Text
}

// mapCourse converts a sharedCourseRow into the domain Course type. The same
// helper backs both the list endpoint and the get-by-id endpoint.
func mapCourse(r sharedCourseRow) (Course, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Course{}, fmt.Errorf("mapCourse: id: %w", err)
	}
	schoolID, err := utils.PgxToGoogleUUID(r.SID)
	if err != nil {
		return Course{}, fmt.Errorf("mapCourse: school id: %w", err)
	}
	return Course{
		ID: id,
		School: SchoolSummary{
			ID:      schoolID,
			Name:    r.SName,
			Acronym: r.SAcronym,
			City:    textPtr(r.SCity),
			State:   textPtr(r.SState),
			Country: textPtr(r.SCountry),
		},
		Department:  r.Department,
		Number:      r.Number,
		Title:       r.Title,
		Description: textPtr(r.Description),
		CreatedAt:   r.CreatedAt.Time,
	}, nil
}

// mapSection converts a sqlc-generated section row into the domain Section.
func mapSection(r db.ListCourseSectionsRow) (Section, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Section{}, fmt.Errorf("mapSection: id: %w", err)
	}
	return Section{
		ID:             id,
		Term:           r.Term,
		SectionCode:    textPtr(r.SectionCode),
		InstructorName: textPtr(r.InstructorName),
		MemberCount:    r.MemberCount,
	}, nil
}

// textPtr returns a *string for a nullable pgtype.Text column. nil for SQL NULL.
func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

// Per-sort-variant adapter functions. Each projects the typed db row into
// sharedCourseRow so the rest of the package can stay variant-agnostic.

func fromGetRow(r db.GetCourseRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromDepartmentAscRow(r db.ListCoursesDepartmentAscRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromDepartmentDescRow(r db.ListCoursesDepartmentDescRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromNumberAscRow(r db.ListCoursesNumberAscRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromNumberDescRow(r db.ListCoursesNumberDescRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromTitleAscRow(r db.ListCoursesTitleAscRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromTitleDescRow(r db.ListCoursesTitleDescRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromCreatedAtAscRow(r db.ListCoursesCreatedAtAscRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

func fromCreatedAtDescRow(r db.ListCoursesCreatedAtDescRow) sharedCourseRow {
	return sharedCourseRow{
		ID: r.ID, SchoolID: r.SchoolID,
		Department: r.Department, Number: r.Number, Title: r.Title,
		Description: r.Description, CreatedAt: r.CreatedAt,
		SID: r.SID, SName: r.SName, SAcronym: r.SAcronym,
		SCity: r.SCity, SState: r.SState, SCountry: r.SCountry,
	}
}

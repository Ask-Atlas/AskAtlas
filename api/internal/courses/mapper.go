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
			City:    utils.TextPtr(r.SCity),
			State:   utils.TextPtr(r.SState),
			Country: utils.TextPtr(r.SCountry),
		},
		Department:  r.Department,
		Number:      r.Number,
		Title:       r.Title,
		Description: utils.TextPtr(r.Description),
		CreatedAt:   r.CreatedAt.Time,
	}, nil
}

// mapMembership converts a sqlc-generated CourseMember row into the domain
// Membership type. Returned by JoinSection so the handler can render the
// 201 wire response without leaking pgtype.UUID into the HTTP layer.
func mapMembership(m db.CourseMember) (Membership, error) {
	userID, err := utils.PgxToGoogleUUID(m.UserID)
	if err != nil {
		return Membership{}, fmt.Errorf("mapMembership: user id: %w", err)
	}
	sectionID, err := utils.PgxToGoogleUUID(m.SectionID)
	if err != nil {
		return Membership{}, fmt.Errorf("mapMembership: section id: %w", err)
	}
	return Membership{
		UserID:    userID,
		SectionID: sectionID,
		Role:      MemberRole(m.Role),
		JoinedAt:  m.JoinedAt.Time,
	}, nil
}

// mapEnrollment converts a sqlc-generated enrollment row into the
// domain Enrollment type used by the dashboard.
func mapEnrollment(r db.ListMyEnrollmentsRow) (Enrollment, error) {
	sectionID, err := utils.PgxToGoogleUUID(r.SectionID)
	if err != nil {
		return Enrollment{}, fmt.Errorf("mapEnrollment: section id: %w", err)
	}
	courseID, err := utils.PgxToGoogleUUID(r.CourseID)
	if err != nil {
		return Enrollment{}, fmt.Errorf("mapEnrollment: course id: %w", err)
	}
	schoolID, err := utils.PgxToGoogleUUID(r.SchoolID)
	if err != nil {
		return Enrollment{}, fmt.Errorf("mapEnrollment: school id: %w", err)
	}
	return Enrollment{
		Section: EnrollmentSection{
			ID:             sectionID,
			Term:           r.SectionTerm,
			SectionCode:    utils.TextPtr(r.SectionSectionCode),
			InstructorName: utils.TextPtr(r.SectionInstructorName),
		},
		Course: EnrollmentCourse{
			ID:         courseID,
			Department: r.CourseDepartment,
			Number:     r.CourseNumber,
			Title:      r.CourseTitle,
		},
		School: EnrollmentSchool{
			ID:      schoolID,
			Acronym: r.SchoolAcronym,
		},
		Role:     MemberRole(r.MemberRole),
		JoinedAt: r.MemberJoinedAt.Time,
	}, nil
}

// mapMembershipCheckRow converts the GetMembership sqlc row into the
// enrolled-true MembershipCheck. The not-enrolled case is constructed
// directly in the service (sql.ErrNoRows branch) since there is no row
// to map.
func mapMembershipCheckRow(r db.GetMembershipRow) MembershipCheck {
	role := MemberRole(r.Role)
	joinedAt := r.JoinedAt.Time
	return MembershipCheck{
		Enrolled: true,
		Role:     &role,
		JoinedAt: &joinedAt,
	}
}

// mapSectionMember converts a sqlc-generated section-member row into
// the domain SectionMember type. Privacy floor: any change to either
// side of this mapping must keep email + clerk_id off the wire.
func mapSectionMember(r db.ListSectionMembersRow) (SectionMember, error) {
	userID, err := utils.PgxToGoogleUUID(r.UserID)
	if err != nil {
		return SectionMember{}, fmt.Errorf("mapSectionMember: user id: %w", err)
	}
	return SectionMember{
		UserID:    userID,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Role:      MemberRole(r.Role),
		JoinedAt:  r.JoinedAt.Time,
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
		SectionCode:    utils.TextPtr(r.SectionCode),
		InstructorName: utils.TextPtr(r.InstructorName),
		MemberCount:    r.MemberCount,
	}, nil
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

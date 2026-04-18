package courses_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/courses"
	mock_courses "github.com/Ask-Atlas/AskAtlas/api/internal/courses/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Each list-sort-variant has its own row type but identical field layout, so
// the tests use small per-variant fixture builders instead of one polymorphic
// helper -- typing the fixture forces the test to assert the dispatch landed
// on the correct repo method.

func deptAscFixture(t *testing.T, dept, num, title string) db.ListCoursesDepartmentAscRow {
	t.Helper()
	return db.ListCoursesDepartmentAscRow{
		ID:         utils.UUID(uuid.New()),
		SchoolID:   utils.UUID(uuid.New()),
		Department: dept,
		Number:     num,
		Title:      title,
		CreatedAt:  pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		SID:        utils.UUID(uuid.New()),
		SName:      "Washington State University",
		SAcronym:   "WSU",
	}
}

func titleAscFixture(t *testing.T, title string) db.ListCoursesTitleAscRow {
	t.Helper()
	return db.ListCoursesTitleAscRow{
		ID:         utils.UUID(uuid.New()),
		SchoolID:   utils.UUID(uuid.New()),
		Department: "CPTS",
		Number:     "322",
		Title:      title,
		CreatedAt:  pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		SID:        utils.UUID(uuid.New()),
		SName:      "Washington State University",
		SAcronym:   "WSU",
	}
}

func createdAtAscFixture(t *testing.T, when time.Time) db.ListCoursesCreatedAtAscRow {
	t.Helper()
	return db.ListCoursesCreatedAtAscRow{
		ID:         utils.UUID(uuid.New()),
		SchoolID:   utils.UUID(uuid.New()),
		Department: "CPTS",
		Number:     "322",
		Title:      "Software Engineering Principles I",
		CreatedAt:  pgtype.Timestamptz{Time: when, Valid: true},
		SID:        utils.UUID(uuid.New()),
		SName:      "Washington State University",
		SAcronym:   "WSU",
	}
}

func TestService_ListCourses_Empty(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.Anything).
		Return(nil, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10})

	require.NoError(t, err)
	assert.Empty(t, got.Courses)
	assert.False(t, got.HasMore)
	assert.Nil(t, got.NextCursor)
}

func TestService_ListCourses_DefaultSortIsDepartmentAsc(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	// No expectations on the other 7 variants -- if dispatch went to the wrong
	// variant, mockery's AssertExpectations cleanup would catch it.
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.Anything).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10})
	require.NoError(t, err)
}

func TestService_ListCourses_OverLimitTriggersCompositeNextCursor(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	limit := int32(2)
	rows := []db.ListCoursesDepartmentAscRow{
		deptAscFixture(t, "CPTS", "121", "Intro to Computing"),
		deptAscFixture(t, "CPTS", "322", "Software Engineering Principles I"),
		deptAscFixture(t, "CPTS", "355", "Database Systems"), // would be page 2
	}
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return arg.PageLimit == limit+1
		})).
		Return(rows, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: limit})

	require.NoError(t, err)
	assert.Len(t, got.Courses, int(limit))
	assert.True(t, got.HasMore)
	require.NotNil(t, got.NextCursor)

	// Composite cursor for the department sort encodes (department, number, id).
	decoded, err := courses.DecodeCursor(*got.NextCursor)
	require.NoError(t, err)
	require.NotNil(t, decoded.Department)
	require.NotNil(t, decoded.Number)
	assert.Equal(t, "CPTS", *decoded.Department)
	assert.Equal(t, "322", *decoded.Number)
	assert.Equal(t, got.Courses[1].ID, decoded.ID)
	assert.Nil(t, decoded.Title, "Title should not be set when sort_by=department")
	assert.Nil(t, decoded.CreatedAt, "CreatedAt should not be set when sort_by=department")
}

func TestService_ListCourses_TitleSortNextCursorOnlyHasTitle(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	limit := int32(1)
	rows := []db.ListCoursesTitleAscRow{
		titleAscFixture(t, "Algorithms"),
		titleAscFixture(t, "Software Engineering"), // would be page 2
	}
	repo.EXPECT().
		ListCoursesTitleAsc(mock.Anything, mock.Anything).
		Return(rows, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{
		Limit:   limit,
		SortBy:  courses.SortFieldTitle,
		SortDir: courses.SortDirAsc,
	})

	require.NoError(t, err)
	require.True(t, got.HasMore)
	require.NotNil(t, got.NextCursor)

	decoded, err := courses.DecodeCursor(*got.NextCursor)
	require.NoError(t, err)
	require.NotNil(t, decoded.Title)
	assert.Equal(t, "Algorithms", *decoded.Title)
	assert.Nil(t, decoded.Department)
	assert.Nil(t, decoded.Number)
	assert.Nil(t, decoded.CreatedAt)
}

func TestService_ListCourses_CreatedAtSortNextCursorOnlyHasCreatedAt(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	limit := int32(1)
	first := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	second := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	rows := []db.ListCoursesCreatedAtAscRow{
		createdAtAscFixture(t, first),
		createdAtAscFixture(t, second), // would be page 2
	}
	repo.EXPECT().
		ListCoursesCreatedAtAsc(mock.Anything, mock.Anything).
		Return(rows, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{
		Limit:   limit,
		SortBy:  courses.SortFieldCreatedAt,
		SortDir: courses.SortDirAsc,
	})

	require.NoError(t, err)
	require.True(t, got.HasMore)
	require.NotNil(t, got.NextCursor)

	decoded, err := courses.DecodeCursor(*got.NextCursor)
	require.NoError(t, err)
	require.NotNil(t, decoded.CreatedAt)
	assert.True(t, decoded.CreatedAt.Equal(first))
	assert.Nil(t, decoded.Department)
	assert.Nil(t, decoded.Title)
}

func TestService_ListCourses_DefaultLimitWhenZero(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return arg.PageLimit == courses.DefaultPageLimit+1
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 0})
	require.NoError(t, err)
}

func TestService_ListCourses_ClampsLimitAboveMax(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return arg.PageLimit == courses.MaxPageLimit+1
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10_000})
	require.NoError(t, err)
}

func TestService_ListCourses_FiltersForwarded(t *testing.T) {
	schoolID := uuid.New()
	dept := "CPTS"
	q := "software"

	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return arg.SchoolID.Valid && arg.SchoolID.Bytes == schoolID &&
				arg.Department.Valid && arg.Department.String == dept &&
				arg.Q.Valid && arg.Q.String == q
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{
		Limit:      10,
		SchoolID:   &schoolID,
		Department: &dept,
		Q:          &q,
	})
	require.NoError(t, err)
}

func TestService_ListCourses_QWildcardEscaped(t *testing.T) {
	q := "50%_off"
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return arg.Q.Valid && arg.Q.String == `50\%\_off`
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10, Q: &q})
	require.NoError(t, err)
}

func TestService_ListCourses_EmptyQTreatedAsNil(t *testing.T) {
	q := "   " // whitespace-only
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return !arg.Q.Valid
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10, Q: &q})
	require.NoError(t, err)
}

func TestService_ListCourses_CursorForwardedToRepo(t *testing.T) {
	cursorID := uuid.New()
	dept := "CPTS"
	num := "322"
	cur := courses.Cursor{ID: cursorID, Department: &dept, Number: &num}

	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.MatchedBy(func(arg db.ListCoursesDepartmentAscParams) bool {
			return arg.CursorDepartment.Valid && arg.CursorDepartment.String == "CPTS" &&
				arg.CursorNumber.Valid && arg.CursorNumber.String == "322" &&
				arg.CursorID.Valid && arg.CursorID.Bytes == cursorID
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10, Cursor: &cur})
	require.NoError(t, err)
}

func TestService_ListCourses_RepoErrorWrapped(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListCoursesDepartmentAsc(mock.Anything, mock.Anything).
		Return(nil, errors.New("boom"))

	svc := courses.NewService(repo)
	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{Limit: 10})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ListCourses")
	assert.Contains(t, err.Error(), "boom")
}

func TestService_ListCourses_UnsupportedSortReturnsError(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	svc := courses.NewService(repo)

	_, err := svc.ListCourses(context.Background(), courses.ListCoursesParams{
		Limit:  10,
		SortBy: courses.SortField("not_a_real_field"),
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported sort")
}

func TestService_GetCourse_Success(t *testing.T) {
	courseID := uuid.New()
	schoolID := uuid.New()
	desc := "Intro to software engineering."
	created := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		GetCourse(mock.Anything, mock.MatchedBy(func(id pgtype.UUID) bool {
			return id.Valid && id.Bytes == courseID
		})).
		Return(db.GetCourseRow{
			ID:          utils.UUID(courseID),
			SchoolID:    utils.UUID(schoolID),
			Department:  "CPTS",
			Number:      "322",
			Title:       "Software Engineering Principles I",
			Description: pgtype.Text{String: desc, Valid: true},
			CreatedAt:   pgtype.Timestamptz{Time: created, Valid: true},
			SID:         utils.UUID(schoolID),
			SName:       "Washington State University",
			SAcronym:    "WSU",
			SCity:       pgtype.Text{String: "Pullman", Valid: true},
			SState:      pgtype.Text{String: "WA", Valid: true},
			SCountry:    pgtype.Text{String: "US", Valid: true},
		}, nil)

	repo.EXPECT().
		ListCourseSections(mock.Anything, mock.MatchedBy(func(id pgtype.UUID) bool {
			return id.Valid && id.Bytes == courseID
		})).
		Return([]db.ListCourseSectionsRow{
			{
				ID:             utils.UUID(uuid.New()),
				Term:           "Spring 2026",
				SectionCode:    pgtype.Text{String: "01", Valid: true},
				InstructorName: pgtype.Text{String: "Dr. Ananth Jillepalli", Valid: true},
				MemberCount:    34,
			},
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.GetCourse(context.Background(), courses.GetCourseParams{CourseID: courseID})

	require.NoError(t, err)
	assert.Equal(t, courseID, got.ID)
	assert.Equal(t, "CPTS", got.Department)
	assert.Equal(t, "322", got.Number)
	require.NotNil(t, got.Description)
	assert.Equal(t, desc, *got.Description)
	assert.Equal(t, schoolID, got.School.ID)
	assert.Equal(t, "WSU", got.School.Acronym)
	require.NotNil(t, got.School.City)
	assert.Equal(t, "Pullman", *got.School.City)

	require.Len(t, got.Sections, 1)
	assert.Equal(t, "Spring 2026", got.Sections[0].Term)
	require.NotNil(t, got.Sections[0].SectionCode)
	assert.Equal(t, "01", *got.Sections[0].SectionCode)
	assert.Equal(t, int64(34), got.Sections[0].MemberCount)
}

func TestService_GetCourse_NoSectionsReturnsEmptySlice(t *testing.T) {
	courseID := uuid.New()
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		GetCourse(mock.Anything, mock.Anything).
		Return(db.GetCourseRow{
			ID:         utils.UUID(courseID),
			SchoolID:   utils.UUID(uuid.New()),
			Department: "CPTS", Number: "490",
			Title:     "Special Topics",
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			SID:       utils.UUID(uuid.New()),
			SName:     "Washington State University",
			SAcronym:  "WSU",
		}, nil)
	repo.EXPECT().
		ListCourseSections(mock.Anything, mock.Anything).
		Return(nil, nil) // no sections

	svc := courses.NewService(repo)
	got, err := svc.GetCourse(context.Background(), courses.GetCourseParams{CourseID: courseID})

	require.NoError(t, err)
	require.NotNil(t, got.Sections, "Sections must be a non-nil slice so JSON encodes as [] not null")
	assert.Empty(t, got.Sections)
}

func TestService_GetCourse_NotFoundPropagated(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		GetCourse(mock.Anything, mock.Anything).
		Return(db.GetCourseRow{}, fmt.Errorf("GetCourse: %w", apperrors.ErrNotFound))

	svc := courses.NewService(repo)
	_, err := svc.GetCourse(context.Background(), courses.GetCourseParams{CourseID: uuid.New()})

	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound), "error must wrap apperrors.ErrNotFound")
}

func TestService_GetCourse_RepoErrorWrapped(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		GetCourse(mock.Anything, mock.Anything).
		Return(db.GetCourseRow{}, errors.New("connection refused"))

	svc := courses.NewService(repo)
	_, err := svc.GetCourse(context.Background(), courses.GetCourseParams{CourseID: uuid.New()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GetCourse")
	assert.Contains(t, err.Error(), "connection refused")
}

func TestService_GetCourse_SectionsErrorWrapped(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		GetCourse(mock.Anything, mock.Anything).
		Return(db.GetCourseRow{
			ID:         utils.UUID(uuid.New()),
			SchoolID:   utils.UUID(uuid.New()),
			Department: "CPTS", Number: "322",
			Title:     "Software Engineering Principles I",
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			SID:       utils.UUID(uuid.New()),
			SName:     "WSU",
			SAcronym:  "WSU",
		}, nil)
	repo.EXPECT().
		ListCourseSections(mock.Anything, mock.Anything).
		Return(nil, errors.New("query timeout"))

	svc := courses.NewService(repo)
	_, err := svc.GetCourse(context.Background(), courses.GetCourseParams{CourseID: uuid.New()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "GetCourse")
	assert.Contains(t, err.Error(), "list sections")
	assert.Contains(t, err.Error(), "query timeout")
}

func TestCursor_RoundTrip(t *testing.T) {
	dept := "CPTS"
	num := "322"
	original := courses.Cursor{ID: uuid.New(), Department: &dept, Number: &num}

	token, err := courses.EncodeCursor(original)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	decoded, err := courses.DecodeCursor(token)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	require.NotNil(t, decoded.Department)
	require.NotNil(t, decoded.Number)
	assert.Equal(t, "CPTS", *decoded.Department)
	assert.Equal(t, "322", *decoded.Number)
}

func TestDecodeCursor_BadInput(t *testing.T) {
	_, err := courses.DecodeCursor("!!!not-base64!!!")
	require.Error(t, err)
}

// =====================================================================
// JoinSection / LeaveSection (ASK-132 / ASK-138)
// =====================================================================

// expectAssertHelpers wires the two preflight existence probes used by
// both JoinSection and LeaveSection. Pulling this out keeps the per-AC
// tests focused on the case under test instead of repeating six lines of
// repo wiring.
func expectAssertOK(repo *mock_courses.MockRepository, courseID, sectionID uuid.UUID) {
	repo.EXPECT().
		CourseExists(mock.Anything, utils.UUID(courseID)).
		Return(true, nil)
	repo.EXPECT().
		SectionInCourseExists(mock.Anything, db.SectionInCourseExistsParams{
			SectionID: utils.UUID(sectionID),
			CourseID:  utils.UUID(courseID),
		}).
		Return(true, nil)
}

func TestService_JoinSection_Success(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	userID := uuid.New()
	joinedAt := time.Now().UTC()

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		JoinSection(mock.Anything, db.JoinSectionParams{
			UserID:    utils.UUID(userID),
			SectionID: utils.UUID(sectionID),
		}).
		Return(db.CourseMember{
			UserID:    utils.UUID(userID),
			SectionID: utils.UUID(sectionID),
			Role:      db.CourseRoleStudent,
			JoinedAt:  pgtype.Timestamptz{Time: joinedAt, Valid: true},
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.JoinSection(context.Background(), courses.JoinSectionParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    userID,
	})
	require.NoError(t, err)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, sectionID, got.SectionID)
	assert.Equal(t, courses.MemberRoleStudent, got.Role)
	assert.True(t, got.JoinedAt.Equal(joinedAt))
}

func TestService_JoinSection_AlreadyMember(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	userID := uuid.New()

	expectAssertOK(repo, courseID, sectionID)
	// ON CONFLICT DO NOTHING returns no row -> sql.ErrNoRows from RETURNING.
	repo.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(db.CourseMember{}, sql.ErrNoRows)

	svc := courses.NewService(repo)
	_, err := svc.JoinSection(context.Background(), courses.JoinSectionParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    userID,
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
	assert.Equal(t, "Already a member of this section", appErr.Message)
}

func TestService_JoinSection_CourseNotFound(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	repo.EXPECT().
		CourseExists(mock.Anything, utils.UUID(courseID)).
		Return(false, nil)

	svc := courses.NewService(repo)
	_, err := svc.JoinSection(context.Background(), courses.JoinSectionParams{
		CourseID:  courseID,
		SectionID: uuid.New(),
		UserID:    uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Course not found", appErr.Message)
}

func TestService_JoinSection_SectionNotFoundOrCrossCourse(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	repo.EXPECT().
		CourseExists(mock.Anything, mock.Anything).
		Return(true, nil)
	repo.EXPECT().
		SectionInCourseExists(mock.Anything, mock.Anything).
		Return(false, nil)

	svc := courses.NewService(repo)
	_, err := svc.JoinSection(context.Background(), courses.JoinSectionParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Section not found", appErr.Message)
}

func TestService_JoinSection_RepoErrorPropagates(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	expectAssertOK(repo, uuid.Nil, uuid.Nil)
	// Override the assert wiring to use Anything so the test does not over-couple to ID values.
	// (mockery records all expectations; the explicit matchers above accept the zero IDs.)

	repo.EXPECT().
		JoinSection(mock.Anything, mock.Anything).
		Return(db.CourseMember{}, errors.New("boom"))

	svc := courses.NewService(repo)
	_, err := svc.JoinSection(context.Background(), courses.JoinSectionParams{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestService_LeaveSection_Success(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	userID := uuid.New()

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		LeaveSection(mock.Anything, db.LeaveSectionParams{
			UserID:    utils.UUID(userID),
			SectionID: utils.UUID(sectionID),
		}).
		Return(utils.UUID(userID), nil)

	svc := courses.NewService(repo)
	err := svc.LeaveSection(context.Background(), courses.LeaveSectionParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    userID,
	})
	require.NoError(t, err)
}

func TestService_LeaveSection_NotMember(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		LeaveSection(mock.Anything, mock.Anything).
		Return(pgtype.UUID{}, sql.ErrNoRows)

	svc := courses.NewService(repo)
	err := svc.LeaveSection(context.Background(), courses.LeaveSectionParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Not a member of this section", appErr.Message)
}

func TestService_LeaveSection_CourseNotFound(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(false, nil)

	svc := courses.NewService(repo)
	err := svc.LeaveSection(context.Background(), courses.LeaveSectionParams{
		CourseID:  uuid.New(),
		SectionID: uuid.New(),
		UserID:    uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Course not found", appErr.Message)
}

func TestService_LeaveSection_SectionNotFound(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(false, nil)

	svc := courses.NewService(repo)
	err := svc.LeaveSection(context.Background(), courses.LeaveSectionParams{
		CourseID:  uuid.New(),
		SectionID: uuid.New(),
		UserID:    uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Section not found", appErr.Message)
}

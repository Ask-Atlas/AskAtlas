package courses_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// expectAssertOK wires the two preflight existence probes used by both
// JoinSection and LeaveSection. Pulling this out keeps the per-AC tests
// focused on the case under test instead of repeating six lines of repo
// wiring.
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

// Covers ASK-132 AC #4 (section does not exist) and AC #5 (section
// exists under a *different* course than the URL's course_id). Both
// surface the same way at the repo seam: SectionInCourseExists returns
// false. The 404 message is intentionally identical for both cases so
// the URL path can't be used to probe for sections under unrelated
// courses.
func TestService_JoinSection_SectionNotFoundOrCrossCoursePath(t *testing.T) {
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
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(true, nil)
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

// =====================================================================
// ListMyEnrollments / CheckMembership (ASK-154 / ASK-148)
// =====================================================================

func enrollmentRow(t *testing.T, term, dept, num, role string) db.ListMyEnrollmentsRow {
	t.Helper()
	return db.ListMyEnrollmentsRow{
		SectionID:             utils.UUID(uuid.New()),
		SectionTerm:           term,
		SectionSectionCode:    pgtype.Text{String: "01", Valid: true},
		SectionInstructorName: pgtype.Text{String: "Dr. Test", Valid: true},
		CourseID:              utils.UUID(uuid.New()),
		CourseDepartment:      dept,
		CourseNumber:          num,
		CourseTitle:           "Some Title",
		SchoolID:              utils.UUID(uuid.New()),
		SchoolAcronym:         "WSU",
		MemberRole:            db.CourseRole(role),
		MemberJoinedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
}

func TestService_ListMyEnrollments_Empty(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListMyEnrollments(mock.Anything, mock.MatchedBy(func(arg db.ListMyEnrollmentsParams) bool {
			return !arg.Term.Valid && !arg.Role.Valid
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListMyEnrollments(context.Background(), courses.ListMyEnrollmentsParams{UserID: uuid.New()})
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestService_ListMyEnrollments_Success(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().
		ListMyEnrollments(mock.Anything, mock.Anything).
		Return([]db.ListMyEnrollmentsRow{
			enrollmentRow(t, "Spring 2026", "CPTS", "322", "student"),
			enrollmentRow(t, "Spring 2026", "CPTS", "355", "ta"),
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListMyEnrollments(context.Background(), courses.ListMyEnrollmentsParams{UserID: uuid.New()})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Spring 2026", got[0].Section.Term)
	assert.Equal(t, "CPTS", got[0].Course.Department)
	assert.Equal(t, "WSU", got[0].School.Acronym)
	assert.Equal(t, courses.MemberRoleStudent, got[0].Role)
	assert.Equal(t, courses.MemberRoleTA, got[1].Role)
}

func TestService_ListMyEnrollments_FiltersForwarded(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	term := "Spring 2026"
	role := courses.MemberRoleInstructor

	repo.EXPECT().
		ListMyEnrollments(mock.Anything, mock.MatchedBy(func(arg db.ListMyEnrollmentsParams) bool {
			return arg.Term.Valid && arg.Term.String == "Spring 2026" &&
				arg.Role.Valid && arg.Role.CourseRole == db.CourseRoleInstructor
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListMyEnrollments(context.Background(), courses.ListMyEnrollmentsParams{
		UserID: uuid.New(),
		Term:   &term,
		Role:   &role,
	})
	require.NoError(t, err)
}

func TestService_ListMyEnrollments_RejectsBadRole(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	bad := courses.MemberRole("admin")

	svc := courses.NewService(repo)
	_, err := svc.ListMyEnrollments(context.Background(), courses.ListMyEnrollmentsParams{
		UserID: uuid.New(),
		Role:   &bad,
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestService_ListMyEnrollments_RejectsTermTooLong(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	tooLong := strings.Repeat("a", courses.MaxTermLength+1)

	svc := courses.NewService(repo)
	_, err := svc.ListMyEnrollments(context.Background(), courses.ListMyEnrollmentsParams{
		UserID: uuid.New(),
		Term:   &tooLong,
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestService_ListMyEnrollments_EmptyTermTreatedAsNoFilter(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	empty := "   "

	repo.EXPECT().
		ListMyEnrollments(mock.Anything, mock.MatchedBy(func(arg db.ListMyEnrollmentsParams) bool {
			return !arg.Term.Valid
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListMyEnrollments(context.Background(), courses.ListMyEnrollmentsParams{
		UserID: uuid.New(),
		Term:   &empty,
	})
	require.NoError(t, err)
}

func TestService_CheckMembership_Enrolled(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	userID := uuid.New()
	joinedAt := time.Now().UTC()

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		GetMembership(mock.Anything, db.GetMembershipParams{
			UserID:    utils.UUID(userID),
			SectionID: utils.UUID(sectionID),
		}).
		Return(db.GetMembershipRow{
			Role:     db.CourseRoleStudent,
			JoinedAt: pgtype.Timestamptz{Time: joinedAt, Valid: true},
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.CheckMembership(context.Background(), courses.CheckMembershipParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    userID,
	})
	require.NoError(t, err)
	assert.True(t, got.Enrolled)
	require.NotNil(t, got.Role)
	assert.Equal(t, courses.MemberRoleStudent, *got.Role)
	require.NotNil(t, got.JoinedAt)
	assert.True(t, got.JoinedAt.Equal(joinedAt))
}

func TestService_CheckMembership_NotEnrolledIs200(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		GetMembership(mock.Anything, mock.Anything).
		Return(db.GetMembershipRow{}, sql.ErrNoRows)

	svc := courses.NewService(repo)
	got, err := svc.CheckMembership(context.Background(), courses.CheckMembershipParams{
		CourseID:  courseID,
		SectionID: sectionID,
		UserID:    uuid.New(),
	})
	require.NoError(t, err)
	assert.False(t, got.Enrolled)
	assert.Nil(t, got.Role)
	assert.Nil(t, got.JoinedAt)
}

// Cascade race: section is deleted between the preflight (which still
// sees the section) and the GetMembership lookup (where the membership
// row has cascade-vanished). The service must re-probe and return 404
// "Section not found" -- not 200 enrolled=false -- per the ASK-148
// spec table. Cannot use expectAssertOK here because we need the second
// SectionInCourseExists call to return a different value than the first.
func TestService_CheckMembership_SectionCascadedDuringRequest(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()

	repo.EXPECT().
		CourseExists(mock.Anything, utils.UUID(courseID)).
		Return(true, nil).Once()
	// First probe (in assertCourseAndSection): section still there.
	repo.EXPECT().
		SectionInCourseExists(mock.Anything, db.SectionInCourseExistsParams{
			SectionID: utils.UUID(sectionID),
			CourseID:  utils.UUID(courseID),
		}).
		Return(true, nil).Once()
	repo.EXPECT().
		GetMembership(mock.Anything, mock.Anything).
		Return(db.GetMembershipRow{}, sql.ErrNoRows).Once()
	// Re-probe (after ErrNoRows): section is gone.
	repo.EXPECT().
		SectionInCourseExists(mock.Anything, db.SectionInCourseExistsParams{
			SectionID: utils.UUID(sectionID),
			CourseID:  utils.UUID(courseID),
		}).
		Return(false, nil).Once()

	svc := courses.NewService(repo)
	_, err := svc.CheckMembership(context.Background(), courses.CheckMembershipParams{
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

// If the cascade re-probe itself errors, the error must propagate as a
// 500 (wrapped, not turned into a misleading 404 or 200). Pins the
// branch where the re-probe is the failure source rather than the
// original GetMembership.
func TestService_CheckMembership_CascadeReprobeErrorPropagates(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil).Once()
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(true, nil).Once()
	repo.EXPECT().GetMembership(mock.Anything, mock.Anything).Return(db.GetMembershipRow{}, sql.ErrNoRows).Once()
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(false, errors.New("db gone")).Once()

	svc := courses.NewService(repo)
	_, err := svc.CheckMembership(context.Background(), courses.CheckMembershipParams{
		CourseID:  uuid.New(),
		SectionID: uuid.New(),
		UserID:    uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cascade re-probe")
	assert.Contains(t, err.Error(), "db gone")
}

func TestService_CheckMembership_CourseNotFound(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(false, nil)

	svc := courses.NewService(repo)
	_, err := svc.CheckMembership(context.Background(), courses.CheckMembershipParams{
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

func TestService_CheckMembership_SectionNotFound(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(false, nil)

	svc := courses.NewService(repo)
	_, err := svc.CheckMembership(context.Background(), courses.CheckMembershipParams{
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

// =====================================================================
// ListSectionMembers (ASK-143)
// =====================================================================

func memberRow(t *testing.T, first, last, role string, joinedAt time.Time) db.ListSectionMembersRow {
	t.Helper()
	return db.ListSectionMembersRow{
		UserID:    utils.UUID(uuid.New()),
		FirstName: first,
		LastName:  last,
		Role:      db.CourseRole(role),
		JoinedAt:  pgtype.Timestamptz{Time: joinedAt, Valid: true},
	}
}

func TestService_ListSectionMembers_Empty(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(arg db.ListSectionMembersParams) bool {
			return arg.SectionID.Valid && arg.SectionID.Bytes == sectionID && !arg.Role.Valid
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  courseID,
		SectionID: sectionID,
	})
	require.NoError(t, err)
	assert.Empty(t, got.Members)
	assert.False(t, got.HasMore)
	assert.Nil(t, got.NextCursor)
}

func TestService_ListSectionMembers_Success(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	t1 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.Anything).
		Return([]db.ListSectionMembersRow{
			memberRow(t, "Ananth", "Jillepalli", "instructor", t1),
			memberRow(t, "David", "Del Val", "student", t2),
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  courseID,
		SectionID: sectionID,
		Limit:     25,
	})
	require.NoError(t, err)
	require.Len(t, got.Members, 2)
	assert.Equal(t, "Ananth", got.Members[0].FirstName)
	assert.Equal(t, courses.MemberRoleInstructor, got.Members[0].Role)
	assert.Equal(t, "David", got.Members[1].FirstName)
	assert.Equal(t, courses.MemberRoleStudent, got.Members[1].Role)
	assert.False(t, got.HasMore)
	assert.Nil(t, got.NextCursor)
}

// limit=2 with 3 returned rows triggers the n+1 trick: trims to 2 and
// emits next_cursor. The cursor must encode the LAST visible row's
// (joined_at, user_id), not the trimmed-off third row.
func TestService_ListSectionMembers_HasMoreEmitsCursor(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	t1 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 2, 3, 9, 30, 0, 0, time.UTC)

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(arg db.ListSectionMembersParams) bool {
			return arg.PageLimit == 3 // 2 + 1 for has_more detection
		})).
		Return([]db.ListSectionMembersRow{
			memberRow(t, "A", "A", "student", t1),
			memberRow(t, "B", "B", "student", t2),
			memberRow(t, "C", "C", "student", t3), // trimmed off
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  courseID,
		SectionID: sectionID,
		Limit:     2,
	})
	require.NoError(t, err)
	require.Len(t, got.Members, 2)
	assert.True(t, got.HasMore)
	require.NotNil(t, got.NextCursor)

	// Round-trip the encoded cursor and verify it points at the last
	// VISIBLE row (B, joined at t2), not the trimmed third row.
	decoded, err := courses.DecodeMemberCursor(*got.NextCursor)
	require.NoError(t, err)
	assert.True(t, decoded.JoinedAt.Equal(t2))
	assert.Equal(t, got.Members[1].UserID, decoded.UserID)
}

// Full client-journey test: page 1 returns has_more=true with a
// next_cursor, then a follow-up call using that cursor as input must
// reach the SQL with the right (joined_at, user_id) tuple. This
// catches sign-flip bugs (e.g., a > swapped to <) that the per-step
// HasMoreEmitsCursor + CursorForwarded tests would miss because they
// verify encode + decode in isolation rather than as a round-trip.
func TestService_ListSectionMembers_PaginationRoundTrip(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	t1 := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 2, 3, 9, 30, 0, 0, time.UTC)
	page1Last := memberRow(t, "B", "B", "student", t2)

	// Page 1: assertOK, then ListSectionMembers returns 3 rows for limit=2.
	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(arg db.ListSectionMembersParams) bool {
			// First call should have no cursor.
			return arg.PageLimit == 3 && !arg.CursorJoinedAt.Valid && !arg.CursorUserID.Valid
		})).
		Return([]db.ListSectionMembersRow{
			memberRow(t, "A", "A", "student", t1),
			page1Last,
			memberRow(t, "C", "C", "student", t3),
		}, nil).Once()

	svc := courses.NewService(repo)
	page1, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID: courseID, SectionID: sectionID, Limit: 2,
	})
	require.NoError(t, err)
	require.True(t, page1.HasMore)
	require.NotNil(t, page1.NextCursor)
	page1LastUserID := page1.Members[len(page1.Members)-1].UserID

	// Page 2: decode the cursor the service emitted, feed it back.
	decoded, err := courses.DecodeMemberCursor(*page1.NextCursor)
	require.NoError(t, err)

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(arg db.ListSectionMembersParams) bool {
			// Second call must carry forward exactly the page-1 boundary.
			return arg.PageLimit == 3 &&
				arg.CursorJoinedAt.Valid && arg.CursorJoinedAt.Time.Equal(page1Last.JoinedAt.Time) &&
				arg.CursorUserID.Valid && arg.CursorUserID.Bytes == page1LastUserID
		})).
		Return(nil, nil).Once()

	_, err = svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID: courseID, SectionID: sectionID, Limit: 2,
		Cursor: &decoded,
	})
	require.NoError(t, err)
}

func TestService_ListSectionMembers_RoleFilterForwarded(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	role := courses.MemberRoleTA

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(arg db.ListSectionMembersParams) bool {
			return arg.Role.Valid && arg.Role.CourseRole == db.CourseRoleTa
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  courseID,
		SectionID: sectionID,
		Role:      &role,
	})
	require.NoError(t, err)
}

func TestService_ListSectionMembers_CursorForwarded(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	courseID := uuid.New()
	sectionID := uuid.New()
	cursorUserID := uuid.New()
	cursorJoinedAt := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	expectAssertOK(repo, courseID, sectionID)
	repo.EXPECT().
		ListSectionMembers(mock.Anything, mock.MatchedBy(func(arg db.ListSectionMembersParams) bool {
			return arg.CursorJoinedAt.Valid && arg.CursorJoinedAt.Time.Equal(cursorJoinedAt) &&
				arg.CursorUserID.Valid && arg.CursorUserID.Bytes == cursorUserID
		})).
		Return(nil, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  courseID,
		SectionID: sectionID,
		Cursor:    &courses.MemberCursor{JoinedAt: cursorJoinedAt, UserID: cursorUserID},
	})
	require.NoError(t, err)
}

func TestService_ListSectionMembers_RejectsBadRole(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(true, nil)
	bad := courses.MemberRole("admin")

	svc := courses.NewService(repo)
	_, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		Role: &bad,
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestService_ListSectionMembers_CourseNotFound(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(false, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  uuid.New(),
		SectionID: uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Course not found", appErr.Message)
}

func TestService_ListSectionMembers_SectionNotFoundOrCrossCoursePath(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().SectionInCourseExists(mock.Anything, mock.Anything).Return(false, nil)

	svc := courses.NewService(repo)
	_, err := svc.ListSectionMembers(context.Background(), courses.ListSectionMembersParams{
		CourseID:  uuid.New(),
		SectionID: uuid.New(),
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Section not found", appErr.Message)
}

func TestMemberCursor_RoundTrip(t *testing.T) {
	original := courses.MemberCursor{
		JoinedAt: time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC),
		UserID:   uuid.New(),
	}
	token, err := courses.EncodeMemberCursor(original)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	decoded, err := courses.DecodeMemberCursor(token)
	require.NoError(t, err)
	assert.True(t, original.JoinedAt.Equal(decoded.JoinedAt))
	assert.Equal(t, original.UserID, decoded.UserID)
}

func TestDecodeMemberCursor_BadInput(t *testing.T) {
	_, err := courses.DecodeMemberCursor("!!!not-base64!!!")
	require.Error(t, err)
}

// Regression guard: the ListSectionMembers SQL must filter out
// soft-deleted users (users.deleted_at IS NOT NULL). Soft-delete is the
// codebase's convention -- partial indexes idx_users_active_email and
// idx_users_deleted_at both treat deleted_at IS NULL as the live-user
// filter -- and a public-by-design roster must respect it. This reads
// the source SQL file rather than mocking, so a future maintainer who
// removes the predicate triggers an immediate test failure they can't
// route around with a service-layer mock change.
func TestListSectionMembersSQL_ExcludesSoftDeletedUsers(t *testing.T) {
	// Resolve the SQL relative to the courses test file so the test
	// works whether go test runs from the package dir or the api root.
	sql, err := os.ReadFile(filepath.Join("..", "..", "db", "queries", "courses.sql"))
	require.NoError(t, err)

	src := string(sql)
	// Find the ListSectionMembers query block and verify the predicate
	// appears within it (a global grep would let an unrelated query
	// satisfy the test).
	startMarker := "-- name: ListSectionMembers :many"
	startIdx := strings.Index(src, startMarker)
	require.NotEqual(t, -1, startIdx, "ListSectionMembers query block not found")

	// Block ends at the next named query or EOF.
	endIdx := strings.Index(src[startIdx+len(startMarker):], "-- name: ")
	var block string
	if endIdx == -1 {
		block = src[startIdx:]
	} else {
		block = src[startIdx : startIdx+len(startMarker)+endIdx]
	}

	assert.Contains(t, block, "u.deleted_at IS NULL",
		"ListSectionMembers must filter soft-deleted users (privacy convention)")
}

// ============================================================
// ListCourseSections tests (ASK-127)
// ============================================================

// fixtureSectionsTime is a stable timestamp for ASK-127 fixtures so
// assertions don't drift across runs.
var fixtureSectionsTime = time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

// sectionsRow is a shorthand builder for db.ListSectionsForCourseRow
// fixtures. SectionCode and instructor are passed as plain pointers
// (nil -> NULL on the wire) for readable test bodies.
func sectionsRow(id, courseID uuid.UUID, term string, sectionCode, instructor *string, members int64) db.ListSectionsForCourseRow {
	row := db.ListSectionsForCourseRow{
		ID:          utils.UUID(id),
		CourseID:    utils.UUID(courseID),
		Term:        term,
		CreatedAt:   pgtype.Timestamptz{Time: fixtureSectionsTime, Valid: true},
		MemberCount: members,
	}
	if sectionCode != nil {
		row.SectionCode = pgtype.Text{String: *sectionCode, Valid: true}
	}
	if instructor != nil {
		row.InstructorName = pgtype.Text{String: *instructor, Valid: true}
	}
	return row
}

// TestListCourseSections_AC1_NoFilter_ReturnsAll: a valid course
// with N sections returns all of them when no term filter is set.
// Verifies the absence of StatusFilter (sqlc.narg unset) is
// forwarded as a non-Valid pgtype.Text.
func TestListCourseSections_AC1_NoFilter_ReturnsAll(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, utils.UUID(courseID)).Return(true, nil)
	id1, id2 := uuid.New(), uuid.New()
	code1, code2 := "01", "02"
	instructor := "Dr. Ananth Jillepalli"
	repo.EXPECT().ListSectionsForCourse(mock.Anything,
		mock.MatchedBy(func(arg db.ListSectionsForCourseParams) bool {
			return arg.CourseID == utils.UUID(courseID) && !arg.Term.Valid
		})).Return([]db.ListSectionsForCourseRow{
		sectionsRow(id1, courseID, "Spring 2026", &code1, &instructor, 34),
		sectionsRow(id2, courseID, "Spring 2026", &code2, nil, 12),
	}, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID,
	})
	require.NoError(t, err)
	require.Len(t, got.Sections, 2)
	assert.Equal(t, id1, got.Sections[0].ID)
	assert.Equal(t, courseID, got.Sections[0].CourseID)
	assert.Equal(t, "Spring 2026", got.Sections[0].Term)
	require.NotNil(t, got.Sections[0].SectionCode)
	assert.Equal(t, "01", *got.Sections[0].SectionCode)
	require.NotNil(t, got.Sections[0].InstructorName)
	assert.Equal(t, "Dr. Ananth Jillepalli", *got.Sections[0].InstructorName)
	assert.Equal(t, int64(34), got.Sections[0].MemberCount)
	assert.True(t, got.Sections[0].CreatedAt.Equal(fixtureSectionsTime))
	// Second section: nullable instructor stays nil through the mapper.
	assert.Nil(t, got.Sections[1].InstructorName)
}

// TestListCourseSections_AC2_TermFilter: an exact-match term
// filter is forwarded to sqlc as a Valid pgtype.Text.
func TestListCourseSections_AC2_TermFilter(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, utils.UUID(courseID)).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything,
		mock.MatchedBy(func(arg db.ListSectionsForCourseParams) bool {
			return arg.Term.Valid && arg.Term.String == "Spring 2026"
		})).Return([]db.ListSectionsForCourseRow{}, nil)

	term := "Spring 2026"
	svc := courses.NewService(repo)
	got, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID, Term: &term,
	})
	require.NoError(t, err)
	require.NotNil(t, got.Sections, "empty slice must be non-nil")
	assert.Empty(t, got.Sections)
}

// TestListCourseSections_AC3_TermFilter_NoMatch: a term that
// matches no sections returns an empty slice (NOT 404). The
// course-existence preflight passes; the query just returns 0 rows.
func TestListCourseSections_AC3_TermFilter_NoMatch(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, utils.UUID(courseID)).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything, mock.Anything).
		Return([]db.ListSectionsForCourseRow{}, nil)

	term := "Summer 2099"
	svc := courses.NewService(repo)
	got, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID, Term: &term,
	})
	require.NoError(t, err)
	assert.Empty(t, got.Sections)
}

// TestListCourseSections_AC4_CourseNotFound_404: missing course ->
// 404 with "Course not found" message.
func TestListCourseSections_AC4_CourseNotFound_404(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, utils.UUID(courseID)).Return(false, nil)
	// ListSectionsForCourse must NOT be called.

	svc := courses.NewService(repo)
	_, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusNotFound, sysErr.Code)
	assert.Equal(t, "Course not found", sysErr.Message)
}

// TestListCourseSections_AC7_NullableFields: a section with NULL
// section_code and NULL instructor_name surfaces as nil pointers
// in the domain payload (which the wire mapper renders as JSON null).
func TestListCourseSections_AC7_NullableFields(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()
	sectionID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything, mock.Anything).Return(
		[]db.ListSectionsForCourseRow{
			sectionsRow(sectionID, courseID, "Spring 2026", nil, nil, 0),
		}, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID,
	})
	require.NoError(t, err)
	require.Len(t, got.Sections, 1)
	assert.Nil(t, got.Sections[0].SectionCode)
	assert.Nil(t, got.Sections[0].InstructorName)
	assert.Equal(t, int64(0), got.Sections[0].MemberCount)
}

// TestListCourseSections_TermTooLong_400: a term longer than
// MaxTermLength chars surfaces as a typed 400 with details.term.
// The HTTP layer enforces the same bound via openapi maxLength: 30,
// but this test guards the service-side defense-in-depth check.
func TestListCourseSections_TermTooLong_400(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)

	tooLong := strings.Repeat("a", courses.MaxTermLength+1)
	svc := courses.NewService(repo)
	_, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: uuid.New(), Term: &tooLong,
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusBadRequest, sysErr.Code)
	require.NotNil(t, sysErr.Details)
	assert.Contains(t, sysErr.Details["term"], "30")
}

// TestListCourseSections_MultiByteTerm_CountsRunesNotBytes pins
// the rune-vs-byte fix from gemini PR #160 review. A 30-character
// multi-byte CJK term is 90 bytes but 30 runes, so a byte-count
// check would (incorrectly) reject it. This test would fail
// under len(trimmed) but pass under utf8.RuneCountInString --
// making the expected behavior an explicit contract.
func TestListCourseSections_MultiByteTerm_CountsRunesNotBytes(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything, mock.Anything).
		Return([]db.ListSectionsForCourseRow{}, nil)

	// "春" is 3 bytes in UTF-8; 30 of them = 90 bytes / 30 runes.
	multibyte := strings.Repeat("春", courses.MaxTermLength)
	svc := courses.NewService(repo)
	_, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID, Term: &multibyte,
	})
	require.NoError(t, err, "30-rune multi-byte term must not be rejected as too long")
}

// TestListCourseSections_EmptyTerm_TreatedAsNoFilter: a pointer
// to an empty/whitespace string collapses to "no filter" rather
// than being passed to the query as an empty string (which would
// match nothing per the exact-equality SQL).
func TestListCourseSections_EmptyTerm_TreatedAsNoFilter(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	courseID := uuid.New()

	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything,
		mock.MatchedBy(func(arg db.ListSectionsForCourseParams) bool {
			// Empty string -> Valid=false (no filter), NOT Valid=true String="".
			return !arg.Term.Valid
		})).Return([]db.ListSectionsForCourseRow{}, nil)

	empty := "   "
	svc := courses.NewService(repo)
	_, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: courseID, Term: &empty,
	})
	require.NoError(t, err)
}

// TestListCourseSections_CourseExistsError_500: a DB failure on
// the existence probe surfaces as 500 (not 404 -- a transport
// error must not be disguised as a missing resource).
func TestListCourseSections_CourseExistsError_500(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).
		Return(false, errors.New("connection refused"))

	svc := courses.NewService(repo)
	_, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestListCourseSections_QueryError_500: a DB failure on the list
// query surfaces as 500.
func TestListCourseSections_QueryError_500(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything, mock.Anything).
		Return(nil, errors.New("query timeout"))

	svc := courses.NewService(repo)
	_, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: uuid.New(),
	})
	require.Error(t, err)
	sysErr := apperrors.ToHTTPError(err)
	assert.Equal(t, http.StatusInternalServerError, sysErr.Code)
}

// TestListCourseSections_EmptyResult_NonNilSlice: a course with 0
// sections returns an empty slice (not nil) so the wire renders []
// per spec.
func TestListCourseSections_EmptyResult_NonNilSlice(t *testing.T) {
	repo := mock_courses.NewMockRepository(t)
	repo.EXPECT().CourseExists(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().ListSectionsForCourse(mock.Anything, mock.Anything).
		Return([]db.ListSectionsForCourseRow{}, nil)

	svc := courses.NewService(repo)
	got, err := svc.ListCourseSections(context.Background(), courses.ListCourseSectionsParams{
		CourseID: uuid.New(),
	})
	require.NoError(t, err)
	require.NotNil(t, got.Sections, "empty slice must be non-nil so wire renders []")
	assert.Empty(t, got.Sections)
}

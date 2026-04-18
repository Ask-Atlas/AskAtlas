package courses_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/courses"
	mock_courses "github.com/Ask-Atlas/AskAtlas/api/internal/courses/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
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

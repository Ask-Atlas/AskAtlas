package recents_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/recents"
	mock_recents "github.com/Ask-Atlas/AskAtlas/api/internal/recents/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fileRow / guideRow / courseRow build minimal sqlc-generated row
// fixtures with only the fields the mappers read.

func fileRow(t *testing.T, viewedAt time.Time) db.ListRecentFilesRow {
	t.Helper()
	return db.ListRecentFilesRow{
		FileID:       utils.UUID(uuid.New()),
		ViewedAt:     pgtype.Timestamptz{Time: viewedAt, Valid: true},
		FileName:     "lecture.pdf",
		FileMimeType: "application/pdf",
	}
}

func guideRow(t *testing.T, viewedAt time.Time) db.ListRecentStudyGuidesRow {
	t.Helper()
	return db.ListRecentStudyGuidesRow{
		StudyGuideID:     utils.UUID(uuid.New()),
		ViewedAt:         pgtype.Timestamptz{Time: viewedAt, Valid: true},
		StudyGuideTitle:  "Binary Trees Cheat Sheet",
		CourseDepartment: "CPTS",
		CourseNumber:     "322",
	}
}

func courseRow(t *testing.T, viewedAt time.Time) db.ListRecentCoursesRow {
	t.Helper()
	return db.ListRecentCoursesRow{
		CourseID:         utils.UUID(uuid.New()),
		ViewedAt:         pgtype.Timestamptz{Time: viewedAt, Valid: true},
		CourseDepartment: "CPTS",
		CourseNumber:     "322",
		CourseTitle:      "Software Engineering Principles I",
	}
}

func TestListRecents_DefaultLimitWhenZero(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)
	viewer := uuid.New()

	// Service must apply DefaultLimit=10 when caller passes Limit=0.
	repo.EXPECT().
		ListRecentFiles(mock.Anything, mock.MatchedBy(func(p db.ListRecentFilesParams) bool {
			return p.PageLimit == recents.DefaultLimit
		})).
		Return(nil, nil)
	repo.EXPECT().
		ListRecentStudyGuides(mock.Anything, mock.MatchedBy(func(p db.ListRecentStudyGuidesParams) bool {
			return p.PageLimit == recents.DefaultLimit
		})).
		Return(nil, nil)
	repo.EXPECT().
		ListRecentCourses(mock.Anything, mock.MatchedBy(func(p db.ListRecentCoursesParams) bool {
			return p.PageLimit == recents.DefaultLimit
		})).
		Return(nil, nil)

	result, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: viewer,
		Limit:    0,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Recents)
}

func TestListRecents_RejectsLimitOverMax(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)

	_, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: uuid.New(),
		Limit:    recents.MaxLimit + 1,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["limit"], "between")
}

func TestListRecents_MergesAndSortsByViewedAtDesc(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)
	viewer := uuid.New()

	now := time.Now().UTC()
	// Seed each per-type query so the merged output's ordering is
	// determined entirely by the ViewedAt timestamps -- otherwise a
	// silent regression in the sort step would still pass any test
	// that only checked one entity type.
	files := []db.ListRecentFilesRow{
		fileRow(t, now.Add(-30*time.Minute)), // oldest of the three
	}
	guides := []db.ListRecentStudyGuidesRow{
		guideRow(t, now.Add(-10*time.Minute)), // newest
	}
	courses := []db.ListRecentCoursesRow{
		courseRow(t, now.Add(-20*time.Minute)), // middle
	}
	repo.EXPECT().ListRecentFiles(mock.Anything, mock.Anything).Return(files, nil)
	repo.EXPECT().ListRecentStudyGuides(mock.Anything, mock.Anything).Return(guides, nil)
	repo.EXPECT().ListRecentCourses(mock.Anything, mock.Anything).Return(courses, nil)

	result, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: viewer,
		Limit:    10,
	})
	require.NoError(t, err)
	require.Len(t, result.Recents, 3)
	assert.Equal(t, recents.EntityTypeStudyGuide, result.Recents[0].EntityType)
	assert.Equal(t, recents.EntityTypeCourse, result.Recents[1].EntityType)
	assert.Equal(t, recents.EntityTypeFile, result.Recents[2].EntityType)
}

func TestListRecents_TruncatesToLimit(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)

	now := time.Now().UTC()
	// Three files (newer), two guides (older). limit=2 must keep the
	// two newest items, both files. This pins the post-merge truncate.
	files := []db.ListRecentFilesRow{
		fileRow(t, now.Add(-1*time.Minute)),
		fileRow(t, now.Add(-2*time.Minute)),
		fileRow(t, now.Add(-3*time.Minute)),
	}
	guides := []db.ListRecentStudyGuidesRow{
		guideRow(t, now.Add(-10*time.Minute)),
		guideRow(t, now.Add(-11*time.Minute)),
	}
	repo.EXPECT().ListRecentFiles(mock.Anything, mock.Anything).Return(files, nil)
	repo.EXPECT().ListRecentStudyGuides(mock.Anything, mock.Anything).Return(guides, nil)
	repo.EXPECT().ListRecentCourses(mock.Anything, mock.Anything).Return(nil, nil)

	result, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: uuid.New(),
		Limit:    2,
	})
	require.NoError(t, err)
	require.Len(t, result.Recents, 2)
	assert.Equal(t, recents.EntityTypeFile, result.Recents[0].EntityType)
	assert.Equal(t, recents.EntityTypeFile, result.Recents[1].EntityType)
}

func TestListRecents_TieBreakOnEntityID(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)

	now := time.Now().UTC()
	// Two rows with identical ViewedAt -- the sort tie-breaker
	// (EntityID lexicographic ASC) must put the smaller UUID first
	// or the output is non-deterministic across runs.
	a := fileRow(t, now)
	b := fileRow(t, now)
	repo.EXPECT().ListRecentFiles(mock.Anything, mock.Anything).Return(
		[]db.ListRecentFilesRow{a, b}, nil)
	repo.EXPECT().ListRecentStudyGuides(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListRecentCourses(mock.Anything, mock.Anything).Return(nil, nil)

	result, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: uuid.New(),
		Limit:    10,
	})
	require.NoError(t, err)
	require.Len(t, result.Recents, 2)
	// The smaller ID lex-string must come first regardless of the
	// order the mock returned them in.
	assert.True(t, result.Recents[0].EntityID.String() < result.Recents[1].EntityID.String())
}

func TestListRecents_FilesQueryFails_ReturnsError(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)

	boom := errors.New("connection lost")
	repo.EXPECT().ListRecentFiles(mock.Anything, mock.Anything).Return(nil, boom)

	_, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: uuid.New(),
		Limit:    10,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

func TestListRecents_EmptyAcrossAll_ReturnsEmptySlice(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)

	repo.EXPECT().ListRecentFiles(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListRecentStudyGuides(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListRecentCourses(mock.Anything, mock.Anything).Return(nil, nil)

	result, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: uuid.New(),
		Limit:    10,
	})
	require.NoError(t, err)
	// Non-nil empty slice -- the handler relies on this so the JSON
	// renders as `"recents": []` not `"recents": null`.
	require.NotNil(t, result.Recents)
	assert.Empty(t, result.Recents)
}

func TestListRecents_PerQueryLimitMatchesRequest(t *testing.T) {
	repo := mock_recents.NewMockRepository(t)
	svc := recents.NewService(repo)

	// limit=5 -- each per-type query must receive PageLimit=5 so the
	// per-table window matches the post-merge cap.
	repo.EXPECT().
		ListRecentFiles(mock.Anything, mock.MatchedBy(func(p db.ListRecentFilesParams) bool {
			return p.PageLimit == 5
		})).
		Return(nil, nil)
	repo.EXPECT().
		ListRecentStudyGuides(mock.Anything, mock.MatchedBy(func(p db.ListRecentStudyGuidesParams) bool {
			return p.PageLimit == 5
		})).
		Return(nil, nil)
	repo.EXPECT().
		ListRecentCourses(mock.Anything, mock.MatchedBy(func(p db.ListRecentCoursesParams) bool {
			return p.PageLimit == 5
		})).
		Return(nil, nil)

	_, err := svc.ListRecents(context.Background(), recents.ListRecentsParams{
		ViewerID: uuid.New(),
		Limit:    5,
	})
	require.NoError(t, err)
}

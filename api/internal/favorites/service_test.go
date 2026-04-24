package favorites_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/favorites"
	mock_favorites "github.com/Ask-Atlas/AskAtlas/api/internal/favorites/mocks"
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

func fileRow(t *testing.T, favoritedAt time.Time) db.ListFileFavoritesRow {
	t.Helper()
	return db.ListFileFavoritesRow{
		FileID:       utils.UUID(uuid.New()),
		FavoritedAt:  pgtype.Timestamptz{Time: favoritedAt, Valid: true},
		FileName:     "midterm.pdf",
		FileMimeType: "application/pdf",
	}
}

func guideRow(t *testing.T, favoritedAt time.Time) db.ListStudyGuideFavoritesRow {
	t.Helper()
	return db.ListStudyGuideFavoritesRow{
		StudyGuideID:     utils.UUID(uuid.New()),
		FavoritedAt:      pgtype.Timestamptz{Time: favoritedAt, Valid: true},
		StudyGuideTitle:  "Binary Trees Cheat Sheet",
		CourseDepartment: "CPTS",
		CourseNumber:     "322",
	}
}

func courseRow(t *testing.T, favoritedAt time.Time) db.ListCourseFavoritesRow {
	t.Helper()
	return db.ListCourseFavoritesRow{
		CourseID:         utils.UUID(uuid.New()),
		FavoritedAt:      pgtype.Timestamptz{Time: favoritedAt, Valid: true},
		CourseDepartment: "CPTS",
		CourseNumber:     "322",
		CourseTitle:      "Software Engineering Principles I",
	}
}

func TestEncodeDecodeCursor_RoundTrip(t *testing.T) {
	cases := []int32{0, 1, 25, 100, favorites.MaxOffset}
	for _, want := range cases {
		s := favorites.EncodeCursor(want)
		got, err := favorites.DecodeCursor(s)
		require.NoError(t, err, "round-trip offset=%d", want)
		assert.Equal(t, want, got, "round-trip offset=%d", want)
	}
}

func TestDecodeCursor_RejectsMalformed(t *testing.T) {
	cases := []string{
		"!!!notbase64!!!",
		"abc",                           // base64 of "i\xb7" -- not an integer
		favorites.EncodeCursor(0) + "x", // valid prefix + junk
	}
	for _, c := range cases {
		_, err := favorites.DecodeCursor(c)
		assert.Error(t, err, "input=%q", c)
	}
}

func TestDecodeCursor_RejectsOverMaxOffset(t *testing.T) {
	s := favorites.EncodeCursor(favorites.MaxOffset + 1)
	_, err := favorites.DecodeCursor(s)
	require.Error(t, err)
}

func TestListFavorites_DefaultLimitWhenZero(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	// Limit=0 -> service applies DefaultLimit=25, which means each
	// per-table query asks for (0 + 25 + 1) = 26 rows.
	repo.EXPECT().
		ListFileFavorites(mock.Anything, mock.MatchedBy(func(p db.ListFileFavoritesParams) bool {
			return p.PageLimit == favorites.DefaultLimit+1 && p.PageOffset == 0
		})).
		Return(nil, nil)
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(nil, nil)

	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    0,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Favorites)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

func TestListFavorites_RejectsLimitOverMax(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	_, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    favorites.MaxLimit + 1,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["limit"], "between")
}

func TestListFavorites_RejectsInvalidCursor(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	bad := "!!!notbase64!!!"
	_, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    10,
		Cursor:   &bad,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 400, appErr.Code)
	assert.Equal(t, "invalid cursor value", appErr.Details["cursor"])
}

func TestListFavorites_MergesAndSortsByFavoritedAtDesc(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	now := time.Now().UTC()
	files := []db.ListFileFavoritesRow{fileRow(t, now.Add(-30*time.Minute))}         // oldest
	guides := []db.ListStudyGuideFavoritesRow{guideRow(t, now.Add(-10*time.Minute))} // newest
	courses := []db.ListCourseFavoritesRow{courseRow(t, now.Add(-20*time.Minute))}   // middle
	repo.EXPECT().ListFileFavorites(mock.Anything, mock.Anything).Return(files, nil)
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(guides, nil)
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(courses, nil)

	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    10,
	})
	require.NoError(t, err)
	require.Len(t, result.Favorites, 3)
	assert.Equal(t, favorites.EntityTypeStudyGuide, result.Favorites[0].EntityType)
	assert.Equal(t, favorites.EntityTypeCourse, result.Favorites[1].EntityType)
	assert.Equal(t, favorites.EntityTypeFile, result.Favorites[2].EntityType)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

func TestListFavorites_HasMoreWhenOverflow(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	now := time.Now().UTC()
	// 3 files (newest), 0 guides, 0 courses. limit=2 -> we should
	// return 2 files and signal has_more = true with next_cursor
	// pointing to offset 2.
	files := []db.ListFileFavoritesRow{
		fileRow(t, now.Add(-1*time.Minute)),
		fileRow(t, now.Add(-2*time.Minute)),
		fileRow(t, now.Add(-3*time.Minute)),
	}
	repo.EXPECT().ListFileFavorites(mock.Anything, mock.Anything).Return(files, nil)
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(nil, nil)

	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    2,
	})
	require.NoError(t, err)
	require.Len(t, result.Favorites, 2)
	assert.True(t, result.HasMore)
	require.NotNil(t, result.NextCursor)
	// Cursor should round-trip to offset 2.
	got, err := favorites.DecodeCursor(*result.NextCursor)
	require.NoError(t, err)
	assert.Equal(t, int32(2), got)
}

func TestListFavorites_OffsetCursorAdvancesPage(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	now := time.Now().UTC()
	// 4 files; limit=2 cursor=offset(2) -> page should be the
	// last 2 files, has_more=false.
	files := []db.ListFileFavoritesRow{
		fileRow(t, now.Add(-1*time.Minute)),
		fileRow(t, now.Add(-2*time.Minute)),
		fileRow(t, now.Add(-3*time.Minute)),
		fileRow(t, now.Add(-4*time.Minute)),
	}
	// Service requests perTableLimit = offset(2) + limit(2) + 1 = 5.
	repo.EXPECT().
		ListFileFavorites(mock.Anything, mock.MatchedBy(func(p db.ListFileFavoritesParams) bool {
			return p.PageLimit == 5 && p.PageOffset == 0
		})).
		Return(files, nil)
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(nil, nil)

	cursor := favorites.EncodeCursor(2)
	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    2,
		Cursor:   &cursor,
	})
	require.NoError(t, err)
	require.Len(t, result.Favorites, 2)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

func TestListFavorites_SingleEntityTypeFilter_Files(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)
	viewer := uuid.New()

	// EntityType=file -> only file query runs (no guides, no courses).
	// SQL applies OFFSET/LIMIT directly with limit+1 for has_more.
	now := time.Now().UTC()
	rows := []db.ListFileFavoritesRow{
		fileRow(t, now.Add(-1*time.Minute)),
		fileRow(t, now.Add(-2*time.Minute)),
		fileRow(t, now.Add(-3*time.Minute)),
	}
	repo.EXPECT().
		ListFileFavorites(mock.Anything, mock.MatchedBy(func(p db.ListFileFavoritesParams) bool {
			return p.PageLimit == 3 && p.PageOffset == 0 // limit(2)+1=3
		})).
		Return(rows, nil)
	// Critically: NO expectations on ListStudyGuideFavorites or
	// ListCourseFavorites -- the filter path must not query them.

	et := favorites.EntityTypeFile
	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID:   viewer,
		Limit:      2,
		EntityType: &et,
	})
	require.NoError(t, err)
	require.Len(t, result.Favorites, 2)
	for _, f := range result.Favorites {
		assert.Equal(t, favorites.EntityTypeFile, f.EntityType)
	}
	assert.True(t, result.HasMore)
}

func TestListFavorites_SingleEntityTypeFilter_StudyGuide(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	rows := []db.ListStudyGuideFavoritesRow{guideRow(t, time.Now().UTC())}
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(rows, nil)

	et := favorites.EntityTypeStudyGuide
	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID:   uuid.New(),
		Limit:      10,
		EntityType: &et,
	})
	require.NoError(t, err)
	require.Len(t, result.Favorites, 1)
	assert.Equal(t, favorites.EntityTypeStudyGuide, result.Favorites[0].EntityType)
}

func TestListFavorites_SingleEntityTypeFilter_Course(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	rows := []db.ListCourseFavoritesRow{courseRow(t, time.Now().UTC())}
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(rows, nil)

	et := favorites.EntityTypeCourse
	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID:   uuid.New(),
		Limit:      10,
		EntityType: &et,
	})
	require.NoError(t, err)
	require.Len(t, result.Favorites, 1)
	assert.Equal(t, favorites.EntityTypeCourse, result.Favorites[0].EntityType)
}

func TestListFavorites_RejectsInvalidEntityType(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	bogus := favorites.EntityType("quiz")
	_, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID:   uuid.New(),
		Limit:      10,
		EntityType: &bogus,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["entity_type"], "must be")
}

func TestListFavorites_FilesQueryFails_PropagatesError(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	boom := errors.New("connection lost")
	repo.EXPECT().ListFileFavorites(mock.Anything, mock.Anything).Return(nil, boom)

	_, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    10,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

func TestListFavorites_EmptyAcrossAll_ReturnsEmptySlice(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	repo.EXPECT().ListFileFavorites(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(nil, nil)

	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    10,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Favorites) // Empty slice, not nil
	assert.Empty(t, result.Favorites)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

func TestListFavorites_OffsetBeyondAllResults_ReturnsEmpty(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	// Only 2 files exist but offset=50 -> the merge yields 2 items,
	// offset(50) > total(2) -> empty page, has_more=false.
	now := time.Now().UTC()
	repo.EXPECT().ListFileFavorites(mock.Anything, mock.Anything).Return(
		[]db.ListFileFavoritesRow{fileRow(t, now), fileRow(t, now.Add(-time.Hour))}, nil)
	repo.EXPECT().ListStudyGuideFavorites(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListCourseFavorites(mock.Anything, mock.Anything).Return(nil, nil)

	cursor := favorites.EncodeCursor(50)
	result, err := svc.ListFavorites(context.Background(), favorites.ListFavoritesParams{
		ViewerID: uuid.New(),
		Limit:    10,
		Cursor:   &cursor,
	})
	require.NoError(t, err)
	assert.Empty(t, result.Favorites)
	assert.False(t, result.HasMore)
	assert.Nil(t, result.NextCursor)
}

// ----------------------------------------------------------------------
// Toggle endpoints (ASK-130 / ASK-156 / ASK-157).
//
// All three follow the same shape: existence probe -> toggle CTE.
// The cases below exercise:
//   * favorite path  -- row inserted, returns Favorited=true + timestamp
//   * unfavorite path -- row deleted,  returns Favorited=false + nil
//   * 404 from existence check propagates without touching the toggle
//   * repo error from the toggle wraps as a 500-class error
// ----------------------------------------------------------------------

func TestService_ToggleFileFavorite_Favorite(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	viewerID := uuid.New()
	fileID := uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().CheckFileExists(mock.Anything, utils.UUID(fileID)).Return(nil)
	repo.EXPECT().
		ToggleFileFavorite(mock.Anything, mock.MatchedBy(func(arg db.ToggleFileFavoriteParams) bool {
			return arg.UserID == utils.UUID(viewerID) && arg.FileID == utils.UUID(fileID)
		})).
		Return(db.ToggleFileFavoriteRow{
			Favorited:   true,
			FavoritedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

	res, err := svc.ToggleFileFavorite(context.Background(), viewerID, fileID)
	require.NoError(t, err)
	assert.True(t, res.Favorited)
	require.NotNil(t, res.FavoritedAt)
	assert.True(t, res.FavoritedAt.Equal(now))
}

func TestService_ToggleFileFavorite_Unfavorite(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	viewerID := uuid.New()
	fileID := uuid.New()

	repo.EXPECT().CheckFileExists(mock.Anything, utils.UUID(fileID)).Return(nil)
	repo.EXPECT().
		ToggleFileFavorite(mock.Anything, mock.Anything).
		Return(db.ToggleFileFavoriteRow{
			Favorited:   false,
			FavoritedAt: pgtype.Timestamptz{}, // SELECT (created_at FROM inserted) is NULL on delete path
		}, nil)

	res, err := svc.ToggleFileFavorite(context.Background(), viewerID, fileID)
	require.NoError(t, err)
	assert.False(t, res.Favorited)
	assert.Nil(t, res.FavoritedAt)
}

func TestService_ToggleFileFavorite_NotFound(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	viewerID := uuid.New()
	fileID := uuid.New()
	repo.EXPECT().CheckFileExists(mock.Anything, utils.UUID(fileID)).
		Return(apperrors.ErrNotFound)
	// ToggleFileFavorite must NOT be called -- if it were, mockery
	// would fail the test on the unexpected invocation.

	_, err := svc.ToggleFileFavorite(context.Background(), viewerID, fileID)
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_ToggleFileFavorite_RepoError(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	repo.EXPECT().CheckFileExists(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().ToggleFileFavorite(mock.Anything, mock.Anything).
		Return(db.ToggleFileFavoriteRow{}, errors.New("connection lost"))

	_, err := svc.ToggleFileFavorite(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection lost")
}

func TestService_ToggleStudyGuideFavorite_Favorite(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	viewerID := uuid.New()
	guideID := uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().CheckStudyGuideExists(mock.Anything, utils.UUID(guideID)).Return(nil)
	repo.EXPECT().
		ToggleStudyGuideFavorite(mock.Anything, mock.MatchedBy(func(arg db.ToggleStudyGuideFavoriteParams) bool {
			return arg.UserID == utils.UUID(viewerID) && arg.StudyGuideID == utils.UUID(guideID)
		})).
		Return(db.ToggleStudyGuideFavoriteRow{
			Favorited:   true,
			FavoritedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

	res, err := svc.ToggleStudyGuideFavorite(context.Background(), viewerID, guideID)
	require.NoError(t, err)
	assert.True(t, res.Favorited)
	require.NotNil(t, res.FavoritedAt)
}

func TestService_ToggleStudyGuideFavorite_Unfavorite(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	repo.EXPECT().CheckStudyGuideExists(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().ToggleStudyGuideFavorite(mock.Anything, mock.Anything).
		Return(db.ToggleStudyGuideFavoriteRow{Favorited: false}, nil)

	res, err := svc.ToggleStudyGuideFavorite(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.False(t, res.Favorited)
	assert.Nil(t, res.FavoritedAt)
}

func TestService_ToggleStudyGuideFavorite_NotFound(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	repo.EXPECT().CheckStudyGuideExists(mock.Anything, mock.Anything).
		Return(apperrors.ErrNotFound)

	_, err := svc.ToggleStudyGuideFavorite(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_ToggleCourseFavorite_Favorite(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	viewerID := uuid.New()
	courseID := uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().CheckCourseExists(mock.Anything, utils.UUID(courseID)).Return(nil)
	repo.EXPECT().
		ToggleCourseFavorite(mock.Anything, mock.MatchedBy(func(arg db.ToggleCourseFavoriteParams) bool {
			return arg.UserID == utils.UUID(viewerID) && arg.CourseID == utils.UUID(courseID)
		})).
		Return(db.ToggleCourseFavoriteRow{
			Favorited:   true,
			FavoritedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

	res, err := svc.ToggleCourseFavorite(context.Background(), viewerID, courseID)
	require.NoError(t, err)
	assert.True(t, res.Favorited)
	require.NotNil(t, res.FavoritedAt)
}

func TestService_ToggleCourseFavorite_Unfavorite(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	repo.EXPECT().CheckCourseExists(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().ToggleCourseFavorite(mock.Anything, mock.Anything).
		Return(db.ToggleCourseFavoriteRow{Favorited: false}, nil)

	res, err := svc.ToggleCourseFavorite(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.False(t, res.Favorited)
	assert.Nil(t, res.FavoritedAt)
}

func TestService_ToggleCourseFavorite_NotFound(t *testing.T) {
	repo := mock_favorites.NewMockRepository(t)
	svc := favorites.NewService(repo)

	repo.EXPECT().CheckCourseExists(mock.Anything, mock.Anything).
		Return(apperrors.ErrNotFound)

	_, err := svc.ToggleCourseFavorite(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

package studyguides_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/studyguides"
	mock_studyguides "github.com/Ask-Atlas/AskAtlas/api/internal/studyguides/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// scoreDescFixture builds a minimal score-desc row. Each field maps 1:1
// to a sqlc-generated column from study_guides.sql.
func scoreDescFixture(t *testing.T, title string, voteScore int64, viewCount int32) db.ListStudyGuidesScoreDescRow {
	t.Helper()
	return db.ListStudyGuidesScoreDescRow{
		ID:               utils.UUID(uuid.New()),
		Title:            title,
		Tags:             []string{"tag-a"},
		CourseID:         utils.UUID(uuid.New()),
		ViewCount:        viewCount,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		CreatorID:        utils.UUID(uuid.New()),
		CreatorFirstName: "Ada",
		CreatorLastName:  "Lovelace",
		VoteScore:        voteScore,
		IsRecommended:    false,
		QuizCount:        0,
	}
}

func TestService_ListStudyGuides_DefaultsToScoreDesc(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.Anything).
		Return([]db.ListStudyGuidesScoreDescRow{
			scoreDescFixture(t, "A", 5, 10),
		}, nil)

	svc := studyguides.NewService(repo)
	got, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(),
		Limit:    25,
	})
	require.NoError(t, err)
	require.Len(t, got.StudyGuides, 1)
	assert.Equal(t, "A", got.StudyGuides[0].Title)
	assert.Equal(t, int64(5), got.StudyGuides[0].VoteScore)
	assert.False(t, got.HasMore)
	assert.Nil(t, got.NextCursor)
}

func TestService_ListStudyGuides_DispatchesAllSortVariants(t *testing.T) {
	cases := []struct {
		name  string
		by    studyguides.SortField
		dir   studyguides.SortDir
		setup func(*mock_studyguides.MockRepository)
	}{
		{"score-desc", studyguides.SortFieldScore, studyguides.SortDirDesc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesScoreDesc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"score-asc", studyguides.SortFieldScore, studyguides.SortDirAsc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesScoreAsc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"views-desc", studyguides.SortFieldViews, studyguides.SortDirDesc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesViewsDesc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"views-asc", studyguides.SortFieldViews, studyguides.SortDirAsc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesViewsAsc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"newest-desc", studyguides.SortFieldNewest, studyguides.SortDirDesc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesNewestDesc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"newest-asc", studyguides.SortFieldNewest, studyguides.SortDirAsc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesNewestAsc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"updated-desc", studyguides.SortFieldUpdated, studyguides.SortDirDesc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesUpdatedDesc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
		{"updated-asc", studyguides.SortFieldUpdated, studyguides.SortDirAsc, func(r *mock_studyguides.MockRepository) {
			r.EXPECT().ListStudyGuidesUpdatedAsc(mock.Anything, mock.Anything).Return(nil, nil)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := mock_studyguides.NewMockRepository(t)
			tc.setup(repo)
			svc := studyguides.NewService(repo)
			_, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
				CourseID: uuid.New(),
				SortBy:   tc.by,
				SortDir:  tc.dir,
				Limit:    25,
			})
			require.NoError(t, err)
		})
	}
}

// n+1 trick: 3 rows returned for limit=2 -> trim to 2 + has_more=true,
// emitted cursor encodes the LAST visible row (not the trimmed extra).
func TestService_ListStudyGuides_HasMoreEmitsCursor(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	r2VoteScore := int64(7)
	r2ViewCount := int32(20)
	r2UpdatedAt := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	row2ID := uuid.New()

	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.MatchedBy(func(arg db.ListStudyGuidesScoreDescParams) bool {
			return arg.PageLimit == 3 // 2 + 1 for has_more detection
		})).
		Return([]db.ListStudyGuidesScoreDescRow{
			scoreDescFixture(t, "A", 10, 30),
			{
				ID:               utils.UUID(row2ID),
				Title:            "B",
				CourseID:         utils.UUID(uuid.New()),
				CreatorID:        utils.UUID(uuid.New()),
				CreatorFirstName: "X",
				CreatorLastName:  "Y",
				ViewCount:        r2ViewCount,
				VoteScore:        r2VoteScore,
				CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
				UpdatedAt:        pgtype.Timestamptz{Time: r2UpdatedAt, Valid: true},
			},
			scoreDescFixture(t, "C", 5, 10), // trimmed off
		}, nil)

	svc := studyguides.NewService(repo)
	got, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(),
		Limit:    2,
	})
	require.NoError(t, err)
	require.Len(t, got.StudyGuides, 2)
	assert.True(t, got.HasMore)
	require.NotNil(t, got.NextCursor)

	decoded, err := studyguides.DecodeCursor(*got.NextCursor)
	require.NoError(t, err)
	require.NotNil(t, decoded.VoteScore)
	assert.Equal(t, r2VoteScore, *decoded.VoteScore)
	require.NotNil(t, decoded.ViewCount)
	assert.Equal(t, int64(r2ViewCount), *decoded.ViewCount)
	require.NotNil(t, decoded.UpdatedAt)
	assert.True(t, decoded.UpdatedAt.Equal(r2UpdatedAt))
	assert.Equal(t, row2ID, decoded.ID)
}

// Page-iteration round-trip: page 1 emits next_cursor; feeding it back
// must reach the SQL with the right keyset tuple. Catches sign-flip
// bugs (e.g. > swapped to <) that the per-step HasMoreEmitsCursor +
// filter-forwarding tests would miss because they verify encode +
// decode in isolation rather than as a round-trip.
func TestService_ListStudyGuides_PaginationRoundTrip(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	page1Row := scoreDescFixture(t, "A", 10, 30)
	page1RowID, _ := utils.PgxToGoogleUUID(page1Row.ID)

	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.MatchedBy(func(arg db.ListStudyGuidesScoreDescParams) bool {
			return !arg.CursorID.Valid && arg.PageLimit == 2
		})).
		Return([]db.ListStudyGuidesScoreDescRow{page1Row, scoreDescFixture(t, "extra", 0, 0)}, nil).Once()

	svc := studyguides.NewService(repo)
	page1, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(), Limit: 1,
	})
	require.NoError(t, err)
	require.True(t, page1.HasMore)
	require.NotNil(t, page1.NextCursor)

	decoded, err := studyguides.DecodeCursor(*page1.NextCursor)
	require.NoError(t, err)

	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.MatchedBy(func(arg db.ListStudyGuidesScoreDescParams) bool {
			return arg.CursorID.Valid && arg.CursorID.Bytes == page1RowID &&
				arg.CursorVoteScore.Valid && arg.CursorVoteScore.Int64 == page1Row.VoteScore &&
				arg.CursorViewCount.Valid && arg.CursorViewCount.Int64 == int64(page1Row.ViewCount)
		})).
		Return(nil, nil).Once()

	_, err = svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(), Limit: 1, Cursor: &decoded,
	})
	require.NoError(t, err)
}

func TestService_ListStudyGuides_TagAndQFiltersForwarded(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	q := "binary"

	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.MatchedBy(func(arg db.ListStudyGuidesScoreDescParams) bool {
			return arg.Q.Valid && arg.Q.String == "binary" &&
				len(arg.Tags) == 2 && arg.Tags[0] == "trees" && arg.Tags[1] == "midterm"
		})).
		Return(nil, nil)

	svc := studyguides.NewService(repo)
	_, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(),
		Q:        &q,
		Tags:     []string{"trees", "midterm"},
		Limit:    25,
	})
	require.NoError(t, err)
}

// Q with %, _, \ must be escaped before reaching the SQL ESCAPE '\'
// clause -- otherwise a user supplying "50%_off" would match anything
// containing "50" + any char + "off".
func TestService_ListStudyGuides_QEscapesLikeWildcards(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	q := `50%_off\`

	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.MatchedBy(func(arg db.ListStudyGuidesScoreDescParams) bool {
			return arg.Q.Valid && arg.Q.String == `50\%\_off\\`
		})).
		Return(nil, nil)

	svc := studyguides.NewService(repo)
	_, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(), Q: &q, Limit: 25,
	})
	require.NoError(t, err)
}

func TestService_ListStudyGuides_RejectsLongQ(t *testing.T) {
	tooLong := strings.Repeat("a", studyguides.MaxSearchLength+1)
	svc := studyguides.NewService(mock_studyguides.NewMockRepository(t))
	_, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(), Q: &tooLong, Limit: 25,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestService_ListStudyGuides_RejectsLongTag(t *testing.T) {
	tooLong := strings.Repeat("a", studyguides.MaxTagLength+1)
	svc := studyguides.NewService(mock_studyguides.NewMockRepository(t))
	_, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(), Tags: []string{"ok", tooLong}, Limit: 25,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
}

func TestService_ListStudyGuides_RepoErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().
		ListStudyGuidesScoreDesc(mock.Anything, mock.Anything).
		Return(nil, errors.New("db down"))

	svc := studyguides.NewService(repo)
	_, err := svc.ListStudyGuides(context.Background(), studyguides.ListStudyGuidesParams{
		CourseID: uuid.New(), Limit: 25,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestService_AssertCourseExists_OK(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
	svc := studyguides.NewService(repo)
	require.NoError(t, svc.AssertCourseExists(context.Background(), uuid.New()))
}

func TestService_AssertCourseExists_NotFound(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(false, nil)
	svc := studyguides.NewService(repo)
	err := svc.AssertCourseExists(context.Background(), uuid.New())
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Course not found", appErr.Message)
}

func TestCursor_RoundTrip(t *testing.T) {
	vs := int64(42)
	original := studyguides.Cursor{ID: uuid.New(), VoteScore: &vs}
	token, err := studyguides.EncodeCursor(original)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	decoded, err := studyguides.DecodeCursor(token)
	require.NoError(t, err)
	assert.Equal(t, original.ID, decoded.ID)
	require.NotNil(t, decoded.VoteScore)
	assert.Equal(t, int64(42), *decoded.VoteScore)
}

func TestDecodeCursor_BadInput(t *testing.T) {
	_, err := studyguides.DecodeCursor("!!!not-base64!!!")
	require.Error(t, err)
}

// Regression guard: every ListStudyGuides* SQL variant must filter out
// soft-deleted guides AND soft-deleted users. Same SQL-text
// introspection pattern PR #135 established for ListSectionMembers --
// reads the source SQL file rather than mocking, so a future
// maintainer who removes either predicate triggers an immediate test
// failure they can't route around with a service-layer mock change.
func TestListStudyGuidesSQL_ExcludesSoftDeletedGuidesAndUsers(t *testing.T) {
	sql, err := os.ReadFile(filepath.Join("..", "..", "db", "queries", "study_guides.sql"))
	require.NoError(t, err)
	src := string(sql)

	variants := []string{
		"-- name: ListStudyGuidesScoreDesc :many",
		"-- name: ListStudyGuidesScoreAsc :many",
		"-- name: ListStudyGuidesViewsDesc :many",
		"-- name: ListStudyGuidesViewsAsc :many",
		"-- name: ListStudyGuidesNewestDesc :many",
		"-- name: ListStudyGuidesNewestAsc :many",
		"-- name: ListStudyGuidesUpdatedDesc :many",
		"-- name: ListStudyGuidesUpdatedAsc :many",
	}

	for _, marker := range variants {
		t.Run(marker, func(t *testing.T) {
			startIdx := strings.Index(src, marker)
			require.NotEqual(t, -1, startIdx, "variant block missing")

			rest := src[startIdx+len(marker):]
			endIdx := strings.Index(rest, "-- name: ")
			var block string
			if endIdx == -1 {
				block = src[startIdx:]
			} else {
				block = src[startIdx : startIdx+len(marker)+endIdx]
			}

			assert.Contains(t, block, "sg.deleted_at IS NULL",
				"%s must filter soft-deleted guides", marker)
			assert.Contains(t, block, "u.deleted_at IS NULL",
				"%s must filter soft-deleted users (privacy convention)", marker)
		})
	}
}

// =====================================================================
// GetStudyGuide (ASK-114) — detail endpoint tests
// =====================================================================

// detailFixture builds a minimal GetStudyGuideDetail row for the happy
// path. The fields map 1:1 to the sqlc-generated columns.
func detailFixture(t *testing.T, id, courseID, creatorID uuid.UUID) db.GetStudyGuideDetailRow {
	t.Helper()
	return db.GetStudyGuideDetailRow{
		ID:               utils.UUID(id),
		Title:            "Binary Trees Cheat Sheet",
		Description:      pgtype.Text{String: "Tree traversals + balancing.", Valid: true},
		Content:          pgtype.Text{String: "# Binary Trees\n...", Valid: true},
		Tags:             []string{"trees", "midterm"},
		ViewCount:        87,
		CreatedAt:        pgtype.Timestamptz{Time: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC), Valid: true},
		CourseID:         utils.UUID(courseID),
		CourseDepartment: "CS",
		CourseNumber:     "161",
		CourseTitle:      "Design and Analysis of Algorithms",
		CreatorID:        utils.UUID(creatorID),
		CreatorFirstName: "Tim",
		CreatorLastName:  "Roughgarden",
		VoteScore:        7,
		IsRecommended:    true,
	}
}

func TestService_GetStudyGuide_Success(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)

	guideID := uuid.New()
	courseID := uuid.New()
	creatorID := uuid.New()
	viewerID := uuid.New()

	repo.EXPECT().
		GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, courseID, creatorID), nil)
	repo.EXPECT().
		GetUserVoteForGuide(mock.Anything, mock.MatchedBy(func(arg db.GetUserVoteForGuideParams) bool {
			return arg.StudyGuideID.Valid && arg.StudyGuideID.Bytes == guideID &&
				arg.ViewerID.Valid && arg.ViewerID.Bytes == viewerID
		})).
		Return(db.VoteDirectionUp, nil)

	recID := uuid.New()
	repo.EXPECT().
		ListGuideRecommenders(mock.Anything, mock.Anything).
		Return([]db.ListGuideRecommendersRow{{
			ID: utils.UUID(recID), FirstName: "Ananth", LastName: "Jillepalli",
		}}, nil)

	quizID := uuid.New()
	repo.EXPECT().
		ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).
		Return([]db.ListGuideQuizzesWithQuestionCountRow{{
			ID: utils.UUID(quizID), Title: "Tree Traversal Quiz", QuestionCount: 10,
		}}, nil)

	resourceID := uuid.New()
	repo.EXPECT().
		ListGuideResources(mock.Anything, mock.Anything).
		Return([]db.ListGuideResourcesRow{{
			ID:    utils.UUID(resourceID),
			Title: "Binary Trees Visual", Url: "https://visualgo.net/en/bst",
			Type: db.ResourceTypeLink, Description: pgtype.Text{String: "Interactive viz.", Valid: true},
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}}, nil)

	fileID := uuid.New()
	repo.EXPECT().
		ListGuideFiles(mock.Anything, mock.Anything).
		Return([]db.ListGuideFilesRow{{
			ID: utils.UUID(fileID), Name: "Lecture Slides - Week 7.pdf",
			MimeType: "application/pdf", Size: 2048000,
		}}, nil)

	svc := studyguides.NewService(repo)
	got, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: viewerID,
	})
	require.NoError(t, err)
	assert.Equal(t, guideID, got.ID)
	assert.Equal(t, "Binary Trees Cheat Sheet", got.Title)
	require.NotNil(t, got.Content)
	assert.Equal(t, "# Binary Trees\n...", *got.Content)
	assert.Equal(t, "Tim", got.Creator.FirstName)
	assert.Equal(t, courseID, got.Course.ID)
	assert.Equal(t, "CS", got.Course.Department)
	assert.Equal(t, int64(7), got.VoteScore)
	assert.True(t, got.IsRecommended)

	// user_vote branch: viewer voted up
	require.NotNil(t, got.UserVote)
	assert.Equal(t, studyguides.GuideVoteUp, *got.UserVote)

	// nested arrays populated
	require.Len(t, got.RecommendedBy, 1)
	assert.Equal(t, "Ananth", got.RecommendedBy[0].FirstName)
	require.Len(t, got.Quizzes, 1)
	assert.Equal(t, int64(10), got.Quizzes[0].QuestionCount)
	require.Len(t, got.Resources, 1)
	assert.Equal(t, studyguides.ResourceTypeLink, got.Resources[0].Type)
	require.Len(t, got.Files, 1)
	assert.Equal(t, int64(2048000), got.Files[0].Size)
}

// Viewer has not voted: GetUserVoteForGuide returns sql.ErrNoRows, the
// service must map that to nil UserVote (not an error).
func TestService_GetStudyGuide_UserVoteNil(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()

	repo.EXPECT().
		GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().
		GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows)
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil)

	svc := studyguides.NewService(repo)
	got, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.NoError(t, err)
	assert.Nil(t, got.UserVote)
}

// Empty nested arrays must still be non-nil slices so the JSON output
// is '[]', not null.
func TestService_GetStudyGuide_EmptyNestedArraysStayNonNil(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()

	repo.EXPECT().
		GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().
		GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows)
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil)

	svc := studyguides.NewService(repo)
	got, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.NoError(t, err)
	assert.NotNil(t, got.RecommendedBy)
	assert.NotNil(t, got.Quizzes)
	assert.NotNil(t, got.Resources)
	assert.NotNil(t, got.Files)
	assert.Empty(t, got.RecommendedBy)
	assert.Empty(t, got.Quizzes)
	assert.Empty(t, got.Resources)
	assert.Empty(t, got.Files)
}

// Missing or soft-deleted guide: GetStudyGuideDetail returns
// sql.ErrNoRows, service maps to 404 AppError.
func TestService_GetStudyGuide_NotFound(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().
		GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideDetailRow{}, sql.ErrNoRows)

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)

	// Subsequent queries must not fire -- proving the short-circuit.
	repo.AssertNotCalled(t, "GetUserVoteForGuide", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "ListGuideRecommenders", mock.Anything, mock.Anything)
}

func TestService_GetStudyGuide_DetailErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().
		GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideDetailRow{}, errors.New("db down"))

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

// Regression guard: the detail query must filter soft-deleted guides +
// soft-deleted creators. Extends the same SQL-text introspection
// pattern to the new GetStudyGuideDetail block.
func TestGetStudyGuideDetailSQL_ExcludesSoftDeleted(t *testing.T) {
	sql, err := os.ReadFile(filepath.Join("..", "..", "db", "queries", "study_guides.sql"))
	require.NoError(t, err)
	src := string(sql)

	startMarker := "-- name: GetStudyGuideDetail :one"
	startIdx := strings.Index(src, startMarker)
	require.NotEqual(t, -1, startIdx, "GetStudyGuideDetail block missing")

	rest := src[startIdx+len(startMarker):]
	endIdx := strings.Index(rest, "-- name: ")
	var block string
	if endIdx == -1 {
		block = src[startIdx:]
	} else {
		block = src[startIdx : startIdx+len(startMarker)+endIdx]
	}

	assert.Contains(t, block, "sg.deleted_at IS NULL",
		"GetStudyGuideDetail must filter soft-deleted guides (404)")
	assert.Contains(t, block, "u.deleted_at IS NULL",
		"GetStudyGuideDetail must filter soft-deleted creators (privacy)")
}

// Regression guard: ListGuideRecommenders must filter soft-deleted
// recommender users so a user who deleted their account disappears
// from the "recommended by" list.
func TestListGuideRecommendersSQL_ExcludesSoftDeletedUsers(t *testing.T) {
	sql, err := os.ReadFile(filepath.Join("..", "..", "db", "queries", "study_guides.sql"))
	require.NoError(t, err)
	src := string(sql)

	startMarker := "-- name: ListGuideRecommenders :many"
	startIdx := strings.Index(src, startMarker)
	require.NotEqual(t, -1, startIdx, "ListGuideRecommenders block missing")

	rest := src[startIdx+len(startMarker):]
	endIdx := strings.Index(rest, "-- name: ")
	var block string
	if endIdx == -1 {
		block = src[startIdx:]
	} else {
		block = src[startIdx : startIdx+len(startMarker)+endIdx]
	}

	assert.Contains(t, block, "u.deleted_at IS NULL",
		"ListGuideRecommenders must filter soft-deleted users")
}

// Regression guard: file list must filter files whose upload has not
// completed. A pending or failed file exposed in the detail payload
// would give the frontend a row it can't download. Pins f.status =
// 'complete' in the ListGuideFiles query block.
func TestListGuideFilesSQL_FiltersByCompleteStatus(t *testing.T) {
	sql, err := os.ReadFile(filepath.Join("..", "..", "db", "queries", "study_guides.sql"))
	require.NoError(t, err)
	src := string(sql)

	startMarker := "-- name: ListGuideFiles :many"
	startIdx := strings.Index(src, startMarker)
	require.NotEqual(t, -1, startIdx, "ListGuideFiles block missing")

	rest := src[startIdx+len(startMarker):]
	endIdx := strings.Index(rest, "-- name: ")
	var block string
	if endIdx == -1 {
		block = src[startIdx:]
	} else {
		block = src[startIdx : startIdx+len(startMarker)+endIdx]
	}

	assert.Contains(t, block, "f.status = 'complete'",
		"ListGuideFiles must filter files whose upload is not complete")
}

// Regression guard: quiz list must filter soft-deleted quizzes so
// deleted quizzes don't leak back into the detail payload.
func TestListGuideQuizzesSQL_ExcludesSoftDeleted(t *testing.T) {
	sql, err := os.ReadFile(filepath.Join("..", "..", "db", "queries", "study_guides.sql"))
	require.NoError(t, err)
	src := string(sql)

	startMarker := "-- name: ListGuideQuizzesWithQuestionCount :many"
	startIdx := strings.Index(src, startMarker)
	require.NotEqual(t, -1, startIdx, "ListGuideQuizzesWithQuestionCount block missing")

	rest := src[startIdx+len(startMarker):]
	endIdx := strings.Index(rest, "-- name: ")
	var block string
	if endIdx == -1 {
		block = src[startIdx:]
	} else {
		block = src[startIdx : startIdx+len(startMarker)+endIdx]
	}

	assert.Contains(t, block, "q.deleted_at IS NULL",
		"ListGuideQuizzesWithQuestionCount must filter soft-deleted quizzes")
}

// Each sibling-query error path: GetUserVoteForGuide with a non-
// ErrNoRows error, ListGuideRecommenders, ListGuideQuizzes...,
// ListGuideResources, ListGuideFiles. All 5 must surface as a 500
// (wrapped error) through GetStudyGuide, not a 200 with partial data.
//
// These were gaps in the original serial implementation's coverage
// (PR #137 review LOW 2). They matter more after the errgroup
// refactor: a future maintainer who accidentally swallows an error
// in one of the goroutines would ship a 200 with a missing/partial
// nested array. These tests pin the error propagation contract.

func TestService_GetStudyGuide_UserVoteErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), errors.New("boom"))
	// Siblings MAY or MAY NOT fire depending on goroutine scheduling
	// and errgroup ctx cancellation. .Maybe() tolerates either outcome.
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func TestService_GetStudyGuide_RecommendersErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows).Maybe()
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).
		Return(nil, errors.New("recommenders down"))
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "recommenders down")
}

func TestService_GetStudyGuide_QuizzesErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows).Maybe()
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).
		Return(nil, errors.New("quizzes down"))
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quizzes down")
}

func TestService_GetStudyGuide_ResourcesErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows).Maybe()
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).
		Return(nil, errors.New("resources down"))
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resources down")
}

func TestService_GetStudyGuide_FilesErrorPropagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), uuid.New()), nil)
	repo.EXPECT().GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows).Maybe()
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).
		Return(nil, errors.New("files down"))

	svc := studyguides.NewService(repo)
	_, err := svc.GetStudyGuide(context.Background(), studyguides.GetStudyGuideParams{
		StudyGuideID: guideID, ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "files down")
}

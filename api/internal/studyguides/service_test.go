package studyguides_test

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

// ---------------------------------------------------------------------
// CreateStudyGuide (ASK-120)
// ---------------------------------------------------------------------

// expectInsertReturning sets up the InsertStudyGuide mock to capture the
// resolved sqlc params and return a synthetic row with the given id.
// The capture is exposed via the returned pointer so individual tests
// can assert the params the service ended up sending to the DB (most
// importantly: the normalized tags).
func expectInsertReturning(t *testing.T, repo *mock_studyguides.MockRepository, guideID uuid.UUID) *db.InsertStudyGuideParams {
	t.Helper()
	captured := &db.InsertStudyGuideParams{}
	repo.EXPECT().
		InsertStudyGuide(mock.Anything, mock.Anything).
		Run(func(_ context.Context, arg db.InsertStudyGuideParams) {
			*captured = arg
		}).
		Return(db.InsertStudyGuideRow{
			ID:        utils.UUID(guideID),
			ViewCount: 0,
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)
	return captured
}

// hydrateRow returns a GetStudyGuideDetailRow shaped like a freshly
// inserted guide: vote_score=0, is_recommended=false, view_count=0.
func hydrateRow(t *testing.T, guideID, courseID, creatorID uuid.UUID, title string, tags []string) db.GetStudyGuideDetailRow {
	t.Helper()
	return db.GetStudyGuideDetailRow{
		ID:               utils.UUID(guideID),
		Title:            title,
		Tags:             tags,
		ViewCount:        0,
		CreatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		CourseID:         utils.UUID(courseID),
		CourseDepartment: "CS",
		CourseNumber:     "161",
		CourseTitle:      "Algorithms",
		CreatorID:        utils.UUID(creatorID),
		CreatorFirstName: "Ada",
		CreatorLastName:  "Lovelace",
		VoteScore:        0,
		IsRecommended:    false,
	}
}

func TestService_CreateStudyGuide_Success(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)

	guideID := uuid.New()
	courseID := uuid.New()
	creatorID := uuid.New()
	desc := "Cheat sheet."
	content := "# Trees"

	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
	captured := expectInsertReturning(t, repo, guideID)
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(hydrateRow(t, guideID, courseID, creatorID, "Binary Trees", []string{"trees", "midterm"}), nil)

	svc := studyguides.NewService(repo)
	got, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID:    courseID,
		CreatorID:   creatorID,
		Title:       "Binary Trees",
		Description: &desc,
		Content:     &content,
		Tags:        []string{"Trees", "MIDTERM"},
	})

	require.NoError(t, err)
	assert.Equal(t, guideID, got.ID)
	assert.Equal(t, "Binary Trees", got.Title)
	assert.Equal(t, int64(0), got.VoteScore)
	assert.False(t, got.IsRecommended)
	assert.Nil(t, got.UserVote)
	require.NotNil(t, got.RecommendedBy)
	assert.Empty(t, got.RecommendedBy)
	require.NotNil(t, got.Quizzes)
	assert.Empty(t, got.Quizzes)
	require.NotNil(t, got.Resources)
	assert.Empty(t, got.Resources)
	require.NotNil(t, got.Files)
	assert.Empty(t, got.Files)
	assert.Equal(t, []string{"trees", "midterm"}, captured.Tags)
	require.True(t, captured.Description.Valid)
	assert.Equal(t, "Cheat sheet.", captured.Description.String)
	require.True(t, captured.Content.Valid)
	assert.Equal(t, "# Trees", captured.Content.String)
}

func TestService_CreateStudyGuide_NormalizesAndDedupesTags(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, courseID, creatorID := uuid.New(), uuid.New(), uuid.New()

	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
	captured := expectInsertReturning(t, repo, guideID)
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(hydrateRow(t, guideID, courseID, creatorID, "T", nil), nil)

	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID:  courseID,
		CreatorID: creatorID,
		Title:     "T",
		Tags:      []string{"  Trees  ", "TREES", "midterm", "Midterm"},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"trees", "midterm"}, captured.Tags)
}

// PR #138 review M1: description + content are trimmed and treated as
// SQL NULL when empty after trim, mirroring how title is handled. A
// body of `{"description": "   "}` must not persist a whitespace-only
// string.
func TestService_CreateStudyGuide_DescriptionAndContent_TrimmedAndDroppedWhenWhitespace(t *testing.T) {
	cases := map[string]struct {
		in  string
		// expectValid: true means we expect a SQL NULL (Valid=false on
		// pgtype.Text). false means we expect a populated value.
		expectValid bool
		expectVal   string
	}{
		"whitespace_only":        {in: "   \t\n  ", expectValid: false},
		"empty_string":           {in: "", expectValid: false},
		"surrounded_by_spaces":   {in: "  hello  ", expectValid: true, expectVal: "hello"},
		"normal":                 {in: "no leading or trailing", expectValid: true, expectVal: "no leading or trailing"},
		"newlines_inside_kept":   {in: "line one\nline two", expectValid: true, expectVal: "line one\nline two"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := mock_studyguides.NewMockRepository(t)
			guideID, courseID, creatorID := uuid.New(), uuid.New(), uuid.New()

			repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
			captured := expectInsertReturning(t, repo, guideID)
			repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
				Return(hydrateRow(t, guideID, courseID, creatorID, "T", nil), nil)

			svc := studyguides.NewService(repo)
			in := tc.in
			_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
				CourseID: courseID, CreatorID: creatorID, Title: "T",
				Description: &in,
				Content:     &in,
			})
			require.NoError(t, err)

			if tc.expectValid {
				require.True(t, captured.Description.Valid, "description should be non-NULL")
				assert.Equal(t, tc.expectVal, captured.Description.String)
				require.True(t, captured.Content.Valid, "content should be non-NULL")
				assert.Equal(t, tc.expectVal, captured.Content.String)
			} else {
				assert.False(t, captured.Description.Valid, "description should be SQL NULL after trim")
				assert.False(t, captured.Content.Valid, "content should be SQL NULL after trim")
			}
		})
	}
}

// Tags must always land as a non-nil slice so the Postgres NOT NULL
// DEFAULT '{}' column doesn't trip on a NULL bind value.
func TestService_CreateStudyGuide_NilTagsLandsAsEmptySlice(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, courseID, creatorID := uuid.New(), uuid.New(), uuid.New()

	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
	captured := expectInsertReturning(t, repo, guideID)
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(hydrateRow(t, guideID, courseID, creatorID, "T", nil), nil)

	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: courseID, CreatorID: creatorID, Title: "T", Tags: nil,
	})
	require.NoError(t, err)
	require.NotNil(t, captured.Tags)
	assert.Empty(t, captured.Tags)
}

func TestService_CreateStudyGuide_TitleEmptyOrWhitespace_400(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"whitespace": "   \t\n  ",
	}
	for name, title := range cases {
		t.Run(name, func(t *testing.T) {
			repo := mock_studyguides.NewMockRepository(t)
			svc := studyguides.NewService(repo)
			_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
				CourseID: uuid.New(), CreatorID: uuid.New(), Title: title,
			})
			require.Error(t, err)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, http.StatusBadRequest, appErr.Code)
			assert.Contains(t, appErr.Details, "title")
		})
	}
}

func TestService_CreateStudyGuide_TitleTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(),
		Title: strings.Repeat("a", studyguides.MaxTitleLength+1),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "title")
}

func TestService_CreateStudyGuide_DescriptionTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	tooLong := strings.Repeat("a", studyguides.MaxDescriptionLength+1)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
		Description: &tooLong,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "description")
}

func TestService_CreateStudyGuide_ContentTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	tooLong := strings.Repeat("a", studyguides.MaxContentLength+1)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
		Content: &tooLong,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "content")
}

func TestService_CreateStudyGuide_TooManyTags_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	tags := make([]string, studyguides.MaxTagsCount+1)
	for i := range tags {
		tags[i] = "t" + strings.Repeat("x", i%5)
	}
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T", Tags: tags,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "tags")
}

func TestService_CreateStudyGuide_TagEmptyAfterTrim_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
		Tags: []string{"valid", "   "},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "tags")
}

func TestService_CreateStudyGuide_TagTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
		Tags: []string{strings.Repeat("a", studyguides.MaxTagLength+1)},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "tags")
}

func TestService_CreateStudyGuide_CourseNotFound_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(false, nil)

	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Course not found", appErr.Message)
	repo.AssertNotCalled(t, "InsertStudyGuide", mock.Anything, mock.Anything)
}

func TestService_CreateStudyGuide_InsertError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().InsertStudyGuide(mock.Anything, mock.Anything).
		Return(db.InsertStudyGuideRow{}, errors.New("insert blew up"))

	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert blew up")
}

func TestService_CreateStudyGuide_HydrateError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().CourseExistsForGuides(mock.Anything, mock.Anything).Return(true, nil)
	expectInsertReturning(t, repo, uuid.New())
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideDetailRow{}, errors.New("hydrate blew up"))

	svc := studyguides.NewService(repo)
	_, err := svc.CreateStudyGuide(context.Background(), studyguides.CreateStudyGuideParams{
		CourseID: uuid.New(), CreatorID: uuid.New(), Title: "T",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hydrate blew up")
}

// ---------------------------------------------------------------------
// DeleteStudyGuide (ASK-133)
// ---------------------------------------------------------------------

// inTxRunsFn wires the InTx mock to invoke the closure inline against
// the SAME repo, so SoftDeleteStudyGuide / SoftDeleteQuizzesForGuide /
// GetStudyGuideByIDForUpdate expectations land on the parent mock as
// they would in production after Queries.WithTx returns the same
// underlying connection. Returns the closure's error untouched so
// service-layer error mapping (404, 403, 500) flows through.
func inTxRunsFn(repo *mock_studyguides.MockRepository) {
	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(studyguides.Repository) error) error {
			return fn(repo)
		})
}

func TestService_DeleteStudyGuide_Success(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(guideID),
			CreatorID: utils.UUID(creatorID),
		}, nil)
	repo.EXPECT().SoftDeleteStudyGuide(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().SoftDeleteQuizzesForGuide(mock.Anything, mock.Anything).Return(nil)

	svc := studyguides.NewService(repo)
	err := svc.DeleteStudyGuide(context.Background(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: guideID,
		ViewerID:     creatorID,
	})
	require.NoError(t, err)
}

func TestService_DeleteStudyGuide_NotFound_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{}, sql.ErrNoRows)

	svc := studyguides.NewService(repo)
	err := svc.DeleteStudyGuide(context.Background(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
	repo.AssertNotCalled(t, "SoftDeleteStudyGuide", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "SoftDeleteQuizzesForGuide", mock.Anything, mock.Anything)
}

// Idempotent semantics: re-deleting an already-soft-deleted guide
// surfaces as 404 (desired state already reached), not 409.
func TestService_DeleteStudyGuide_AlreadyDeleted_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(uuid.New()),
			CreatorID: utils.UUID(uuid.New()),
			DeletedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)

	svc := studyguides.NewService(repo)
	err := svc.DeleteStudyGuide(context.Background(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	repo.AssertNotCalled(t, "SoftDeleteStudyGuide", mock.Anything, mock.Anything)
}

func TestService_DeleteStudyGuide_NotCreator_403(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	inTxRunsFn(repo)
	creatorID := uuid.New()
	otherViewer := uuid.New()
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(uuid.New()),
			CreatorID: utils.UUID(creatorID),
		}, nil)

	svc := studyguides.NewService(repo)
	err := svc.DeleteStudyGuide(context.Background(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: otherViewer,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.Code)
	repo.AssertNotCalled(t, "SoftDeleteStudyGuide", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "SoftDeleteQuizzesForGuide", mock.Anything, mock.Anything)
}

func TestService_DeleteStudyGuide_LockError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	inTxRunsFn(repo)
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{}, errors.New("lock blew up"))

	svc := studyguides.NewService(repo)
	err := svc.DeleteStudyGuide(context.Background(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "lock blew up")
}

// Quiz-cascade error returns from the closure -- production InTx then
// rolls the whole tx back. We can't observe the Postgres rollback in a
// mock-backed test, but we CAN pin the contract that the closure's
// error propagates to the service caller verbatim.
func TestService_DeleteStudyGuide_QuizCascadeError_Propagates(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID := uuid.New()
	creatorID := uuid.New()

	inTxRunsFn(repo)
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(guideID),
			CreatorID: utils.UUID(creatorID),
		}, nil)
	repo.EXPECT().SoftDeleteStudyGuide(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().SoftDeleteQuizzesForGuide(mock.Anything, mock.Anything).
		Return(errors.New("quiz cascade blew up"))

	svc := studyguides.NewService(repo)
	err := svc.DeleteStudyGuide(context.Background(), studyguides.DeleteStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "quiz cascade blew up")
}

// ---------------------------------------------------------------------
// CastVote (ASK-139)
// ---------------------------------------------------------------------

func TestService_CastVote_UpsertsAndReturnsScore(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, viewerID := uuid.New(), uuid.New()

	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
	captured := &db.UpsertStudyGuideVoteParams{}
	repo.EXPECT().
		UpsertStudyGuideVote(mock.Anything, mock.Anything).
		Run(func(_ context.Context, arg db.UpsertStudyGuideVoteParams) {
			*captured = arg
		}).Return(nil)
	repo.EXPECT().ComputeGuideVoteScore(mock.Anything, mock.Anything).Return(int64(7), nil)

	svc := studyguides.NewService(repo)
	got, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
		StudyGuideID: guideID, ViewerID: viewerID, Vote: studyguides.GuideVoteUp,
	})
	require.NoError(t, err)
	assert.Equal(t, studyguides.GuideVoteUp, got.Vote)
	assert.Equal(t, int64(7), got.VoteScore)
	assert.Equal(t, db.VoteDirection("up"), captured.Vote)
}

func TestService_CastVote_AcceptsBothDirections(t *testing.T) {
	for _, dir := range []studyguides.GuideVote{studyguides.GuideVoteUp, studyguides.GuideVoteDown} {
		t.Run(string(dir), func(t *testing.T) {
			repo := mock_studyguides.NewMockRepository(t)
			repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
			repo.EXPECT().UpsertStudyGuideVote(mock.Anything, mock.Anything).Return(nil)
			repo.EXPECT().ComputeGuideVoteScore(mock.Anything, mock.Anything).Return(int64(0), nil)

			svc := studyguides.NewService(repo)
			_, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
				StudyGuideID: uuid.New(), ViewerID: uuid.New(), Vote: dir,
			})
			require.NoError(t, err)
		})
	}
}

func TestService_CastVote_InvalidDirection_400(t *testing.T) {
	cases := []studyguides.GuideVote{"", "neutral", "UP", "Down"}
	for _, dir := range cases {
		t.Run(string(dir), func(t *testing.T) {
			repo := mock_studyguides.NewMockRepository(t)
			svc := studyguides.NewService(repo)
			_, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
				StudyGuideID: uuid.New(), ViewerID: uuid.New(), Vote: dir,
			})
			require.Error(t, err)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, http.StatusBadRequest, appErr.Code)
			assert.Contains(t, appErr.Details, "vote")
			repo.AssertNotCalled(t, "GuideExistsAndLive", mock.Anything, mock.Anything)
		})
	}
}

func TestService_CastVote_GuideMissing_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(false, nil)

	svc := studyguides.NewService(repo)
	_, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(), Vote: studyguides.GuideVoteUp,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
	repo.AssertNotCalled(t, "UpsertStudyGuideVote", mock.Anything, mock.Anything)
}

func TestService_CastVote_LiveCheckError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(false, errors.New("db down"))

	svc := studyguides.NewService(repo)
	_, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(), Vote: studyguides.GuideVoteUp,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestService_CastVote_UpsertError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().UpsertStudyGuideVote(mock.Anything, mock.Anything).Return(errors.New("upsert blew up"))

	svc := studyguides.NewService(repo)
	_, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(), Vote: studyguides.GuideVoteUp,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "upsert blew up")
}

func TestService_CastVote_ScoreError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().UpsertStudyGuideVote(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().ComputeGuideVoteScore(mock.Anything, mock.Anything).Return(int64(0), errors.New("score blew up"))

	svc := studyguides.NewService(repo)
	_, err := svc.CastVote(context.Background(), studyguides.CastVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(), Vote: studyguides.GuideVoteUp,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "score blew up")
}

// ---------------------------------------------------------------------
// RemoveVote (ASK-141)
// ---------------------------------------------------------------------

func TestService_RemoveVote_Success(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().DeleteStudyGuideVote(mock.Anything, mock.Anything).Return(int64(1), nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveVote(context.Background(), studyguides.RemoveVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.NoError(t, err)
}

// Guide-existence check runs before the delete: when both the guide
// is missing AND there's no vote, the more-specific "Study guide not
// found" message must win.
func TestService_RemoveVote_GuideMissing_404_GuideMessage(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(false, nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveVote(context.Background(), studyguides.RemoveVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
	repo.AssertNotCalled(t, "DeleteStudyGuideVote", mock.Anything, mock.Anything)
}

// Guide exists but the viewer has no vote -> 404 'Vote not found'.
func TestService_RemoveVote_NoExistingVote_404_VoteMessage(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().DeleteStudyGuideVote(mock.Anything, mock.Anything).Return(int64(0), nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveVote(context.Background(), studyguides.RemoveVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Vote not found", appErr.Message)
}

func TestService_RemoveVote_LiveCheckError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(false, errors.New("db down"))

	svc := studyguides.NewService(repo)
	err := svc.RemoveVote(context.Background(), studyguides.RemoveVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestService_RemoveVote_DeleteError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().GuideExistsAndLive(mock.Anything, mock.Anything).Return(true, nil)
	repo.EXPECT().DeleteStudyGuideVote(mock.Anything, mock.Anything).Return(int64(0), errors.New("delete blew up"))

	svc := studyguides.NewService(repo)
	err := svc.RemoveVote(context.Background(), studyguides.RemoveVoteParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete blew up")
}

// ---------------------------------------------------------------------
// RecommendStudyGuide (ASK-147)
// ---------------------------------------------------------------------

// gateRow returns a synthetic ViewerCanRecommendForGuideRow with the
// given booleans. Centralized so individual tests don't re-derive
// the row shape.
func gateRow(guideExists, hasRole bool) db.ViewerCanRecommendForGuideRow {
	return db.ViewerCanRecommendForGuideRow{GuideExists: guideExists, HasRole: hasRole}
}

func TestService_RecommendStudyGuide_Success_201Body(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, viewerID := uuid.New(), uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, true), nil)
	repo.EXPECT().InsertStudyGuideRecommendation(mock.Anything, mock.Anything).
		Return(db.InsertStudyGuideRecommendationRow{
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			FirstName: "Ananth",
			LastName:  "Jillepalli",
		}, nil)

	svc := studyguides.NewService(repo)
	got, err := svc.RecommendStudyGuide(context.Background(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: guideID, ViewerID: viewerID,
	})
	require.NoError(t, err)
	assert.Equal(t, guideID, got.StudyGuideID)
	assert.Equal(t, viewerID, got.Recommender.ID)
	assert.Equal(t, "Ananth", got.Recommender.FirstName)
	assert.Equal(t, "Jillepalli", got.Recommender.LastName)
	assert.WithinDuration(t, now, got.CreatedAt, time.Second)
}

func TestService_RecommendStudyGuide_GuideMissing_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(false, false), nil)

	svc := studyguides.NewService(repo)
	_, err := svc.RecommendStudyGuide(context.Background(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
	repo.AssertNotCalled(t, "InsertStudyGuideRecommendation", mock.Anything, mock.Anything)
}

// Guide exists but viewer is only a student / not enrolled at all
// (has_role=false). The 403 message is the recommend-side variant
// per the spec.
func TestService_RecommendStudyGuide_NotInstructorOrTA_403(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, false), nil)

	svc := studyguides.NewService(repo)
	_, err := svc.RecommendStudyGuide(context.Background(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.Code)
	assert.Equal(t, "Only instructors and TAs can recommend study guides", appErr.Message)
	repo.AssertNotCalled(t, "InsertStudyGuideRecommendation", mock.Anything, mock.Anything)
}

// ON CONFLICT DO NOTHING + RETURNING surfaces as sql.ErrNoRows on the
// joined SELECT when the (guide, viewer) row already exists.
func TestService_RecommendStudyGuide_Duplicate_409(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, true), nil)
	repo.EXPECT().InsertStudyGuideRecommendation(mock.Anything, mock.Anything).
		Return(db.InsertStudyGuideRecommendationRow{}, sql.ErrNoRows)

	svc := studyguides.NewService(repo)
	_, err := svc.RecommendStudyGuide(context.Background(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusConflict, appErr.Code)
	assert.Equal(t, "You have already recommended this study guide", appErr.Message)
}

func TestService_RecommendStudyGuide_GateError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).
		Return(db.ViewerCanRecommendForGuideRow{}, errors.New("gate down"))

	svc := studyguides.NewService(repo)
	_, err := svc.RecommendStudyGuide(context.Background(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gate down")
}

func TestService_RecommendStudyGuide_InsertError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, true), nil)
	repo.EXPECT().InsertStudyGuideRecommendation(mock.Anything, mock.Anything).
		Return(db.InsertStudyGuideRecommendationRow{}, errors.New("insert blew up"))

	svc := studyguides.NewService(repo)
	_, err := svc.RecommendStudyGuide(context.Background(), studyguides.RecommendStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert blew up")
}

// ---------------------------------------------------------------------
// RemoveRecommendation (ASK-101)
// ---------------------------------------------------------------------

func TestService_RemoveRecommendation_Success(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, true), nil)
	repo.EXPECT().DeleteStudyGuideRecommendation(mock.Anything, mock.Anything).Return(int64(1), nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveRecommendation(context.Background(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.NoError(t, err)
}

func TestService_RemoveRecommendation_GuideMissing_404_GuideMessage(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(false, false), nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveRecommendation(context.Background(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
	repo.AssertNotCalled(t, "DeleteStudyGuideRecommendation", mock.Anything, mock.Anything)
}

// Former TA who lost the role can't manage their old recommendation.
func TestService_RemoveRecommendation_NotInstructorOrTA_403(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, false), nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveRecommendation(context.Background(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.Code)
	assert.Equal(t, "Only instructors and TAs can manage recommendations", appErr.Message)
	repo.AssertNotCalled(t, "DeleteStudyGuideRecommendation", mock.Anything, mock.Anything)
}

// Guide exists, viewer has role, but viewer never recommended this guide.
func TestService_RemoveRecommendation_NoExistingRecommendation_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, true), nil)
	repo.EXPECT().DeleteStudyGuideRecommendation(mock.Anything, mock.Anything).Return(int64(0), nil)

	svc := studyguides.NewService(repo)
	err := svc.RemoveRecommendation(context.Background(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Recommendation not found", appErr.Message)
}

func TestService_RemoveRecommendation_GateError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).
		Return(db.ViewerCanRecommendForGuideRow{}, errors.New("gate down"))

	svc := studyguides.NewService(repo)
	err := svc.RemoveRecommendation(context.Background(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gate down")
}

func TestService_RemoveRecommendation_DeleteError_500(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().ViewerCanRecommendForGuide(mock.Anything, mock.Anything).Return(gateRow(true, true), nil)
	repo.EXPECT().DeleteStudyGuideRecommendation(mock.Anything, mock.Anything).
		Return(int64(0), errors.New("delete blew up"))

	svc := studyguides.NewService(repo)
	err := svc.RemoveRecommendation(context.Background(), studyguides.RemoveRecommendationParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete blew up")
}

// ---------------------------------------------------------------------
// UpdateStudyGuide (ASK-129)
// ---------------------------------------------------------------------

// strPtr is a tiny convenience used only by the update tests for
// building the optional pointer fields on UpdateStudyGuideParams.
func strPtr(s string) *string { return &s }

// expectUpdateAndRehydrate sets up the lock + creator-check + update +
// re-hydrate fan-out a Service.UpdateStudyGuide call requires when the
// caller is the legitimate creator. Returns a pointer to the captured
// sqlc params so individual tests can assert the COALESCE-narg shape
// the service ended up sending to the DB.
func expectUpdateAndRehydrate(t *testing.T, repo *mock_studyguides.MockRepository, guideID, creatorID uuid.UUID) *db.UpdateStudyGuideParams {
	t.Helper()

	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(studyguides.Repository) error) error {
			return fn(repo)
		})
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(guideID),
			CreatorID: utils.UUID(creatorID),
		}, nil)

	captured := &db.UpdateStudyGuideParams{}
	repo.EXPECT().UpdateStudyGuide(mock.Anything, mock.Anything).
		Run(func(_ context.Context, arg db.UpdateStudyGuideParams) {
			*captured = arg
		}).Return(nil)

	// Re-hydrate stage: GetStudyGuide does the GetStudyGuideDetail +
	// 5-way sibling fan-out. Wire all 6 with empty results; the tests
	// don't need to inspect the response here, just confirm the path
	// completes without error.
	repo.EXPECT().GetStudyGuideDetail(mock.Anything, mock.Anything).
		Return(detailFixture(t, guideID, uuid.New(), creatorID), nil)
	repo.EXPECT().GetUserVoteForGuide(mock.Anything, mock.Anything).
		Return(db.VoteDirection(""), sql.ErrNoRows)
	repo.EXPECT().ListGuideRecommenders(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideQuizzesWithQuestionCount(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideResources(mock.Anything, mock.Anything).Return(nil, nil)
	repo.EXPECT().ListGuideFiles(mock.Anything, mock.Anything).Return(nil, nil)

	return captured
}

func TestService_UpdateStudyGuide_TitleOnly_PassesOtherNargsAsNULL(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, creatorID := uuid.New(), uuid.New()

	captured := expectUpdateAndRehydrate(t, repo, guideID, creatorID)

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
		Title: strPtr("New Title"),
	})
	require.NoError(t, err)
	require.True(t, captured.Title.Valid, "title should be a non-NULL narg")
	assert.Equal(t, "New Title", captured.Title.String)
	assert.False(t, captured.Description.Valid, "description should stay NULL (don't update)")
	assert.False(t, captured.Content.Valid, "content should stay NULL (don't update)")
	assert.Nil(t, captured.Tags, "tags should be nil (don't update)")
}

func TestService_UpdateStudyGuide_TitleTrimmed(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, creatorID := uuid.New(), uuid.New()
	captured := expectUpdateAndRehydrate(t, repo, guideID, creatorID)

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
		Title: strPtr("  Trimmed  "),
	})
	require.NoError(t, err)
	assert.Equal(t, "Trimmed", captured.Title.String)
}

// Tags non-nil but empty must reach SQL as a non-nil empty slice so
// the COALESCE replaces with []. nil tags would mean "don't update".
func TestService_UpdateStudyGuide_TagsClearedWithEmptySlice(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, creatorID := uuid.New(), uuid.New()
	captured := expectUpdateAndRehydrate(t, repo, guideID, creatorID)

	emptyTags := []string{}
	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
		Tags: &emptyTags,
	})
	require.NoError(t, err)
	require.NotNil(t, captured.Tags, "empty slice must reach SQL as non-nil so COALESCE replaces (not skips)")
	assert.Empty(t, captured.Tags)
}

func TestService_UpdateStudyGuide_TagsNormalized(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, creatorID := uuid.New(), uuid.New()
	captured := expectUpdateAndRehydrate(t, repo, guideID, creatorID)

	tags := []string{" Midterm ", "FINAL", "midterm"}
	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
		Tags: &tags,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"midterm", "final"}, captured.Tags)
}

// Description with whitespace gets dropped (no-op on that field) --
// matches CreateStudyGuide. Documented limitation: clearing isn't
// supported via PATCH.
func TestService_UpdateStudyGuide_DescriptionWhitespace_DropsToNoOp(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, creatorID := uuid.New(), uuid.New()
	captured := expectUpdateAndRehydrate(t, repo, guideID, creatorID)

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
		Description: strPtr("   "),
		// At least one other field present so the at-least-one-field
		// rule passes (whitespace description alone normalizes away
		// to nothing, which would trip validateUpdateParams indirectly
		// -- but actually our validator only checks the input pointer,
		// not post-normalization, so a lone whitespace description
		// would also pass validation but be a SQL no-op. Adding a
		// title here keeps the test focused on the normalization
		// behavior, not the at-least-one rule).
		Title: strPtr("T"),
	})
	require.NoError(t, err)
	assert.False(t, captured.Description.Valid, "whitespace-only description should be dropped")
}

func TestService_UpdateStudyGuide_EmptyBody_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "body")
}

func TestService_UpdateStudyGuide_TitleEmptyAfterTrim_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Title: strPtr("   "),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "title")
}

func TestService_UpdateStudyGuide_TitleTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	long := strings.Repeat("a", studyguides.MaxTitleLength+1)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Title: &long,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "title")
}

func TestService_UpdateStudyGuide_DescriptionTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	long := strings.Repeat("a", studyguides.MaxDescriptionLength+1)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Description: &long,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "description")
}

func TestService_UpdateStudyGuide_ContentTooLong_400(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	svc := studyguides.NewService(repo)
	long := strings.Repeat("a", studyguides.MaxContentLength+1)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Content: &long,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "content")
}

// Tag normalization is the gate -- a 21-element input or a 51-char tag
// should fail BEFORE the tx opens (a clean 400 vs a rolled-back tx).
func TestService_UpdateStudyGuide_TooManyTags_400_BeforeTx(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	tags := make([]string, studyguides.MaxTagsCount+1)
	for i := range tags {
		tags[i] = fmt.Sprintf("t%d", i)
	}
	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Tags: &tags,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusBadRequest, appErr.Code)
	assert.Contains(t, appErr.Details, "tags")
	repo.AssertNotCalled(t, "InTx", mock.Anything, mock.Anything)
}

func TestService_UpdateStudyGuide_GuideMissing_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(studyguides.Repository) error) error {
			return fn(repo)
		})
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{}, sql.ErrNoRows)

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Title: strPtr("T"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	assert.Equal(t, "Study guide not found", appErr.Message)
	repo.AssertNotCalled(t, "UpdateStudyGuide", mock.Anything, mock.Anything)
}

func TestService_UpdateStudyGuide_AlreadyDeleted_404(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(studyguides.Repository) error) error {
			return fn(repo)
		})
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(uuid.New()),
			CreatorID: utils.UUID(uuid.New()),
			DeletedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}, nil)

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: uuid.New(),
		Title: strPtr("T"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusNotFound, appErr.Code)
	repo.AssertNotCalled(t, "UpdateStudyGuide", mock.Anything, mock.Anything)
}

func TestService_UpdateStudyGuide_NotCreator_403(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(studyguides.Repository) error) error {
			return fn(repo)
		})
	creatorID := uuid.New()
	otherViewer := uuid.New()
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(uuid.New()),
			CreatorID: utils.UUID(creatorID),
		}, nil)

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: uuid.New(), ViewerID: otherViewer,
		Title: strPtr("T"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusForbidden, appErr.Code)
	repo.AssertNotCalled(t, "UpdateStudyGuide", mock.Anything, mock.Anything)
}

func TestService_UpdateStudyGuide_UpdateError_500_PropagatesAndRollsBack(t *testing.T) {
	repo := mock_studyguides.NewMockRepository(t)
	guideID, creatorID := uuid.New(), uuid.New()

	repo.EXPECT().InTx(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(studyguides.Repository) error) error {
			return fn(repo)
		})
	repo.EXPECT().GetStudyGuideByIDForUpdate(mock.Anything, mock.Anything).
		Return(db.GetStudyGuideByIDForUpdateRow{
			ID:        utils.UUID(guideID),
			CreatorID: utils.UUID(creatorID),
		}, nil)
	repo.EXPECT().UpdateStudyGuide(mock.Anything, mock.Anything).
		Return(errors.New("update blew up"))

	svc := studyguides.NewService(repo)
	_, err := svc.UpdateStudyGuide(context.Background(), studyguides.UpdateStudyGuideParams{
		StudyGuideID: guideID, ViewerID: creatorID,
		Title: strPtr("T"),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update blew up")
}

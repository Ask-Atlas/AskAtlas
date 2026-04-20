package dashboard_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/dashboard"
	mock_dashboard "github.com/Ask-Atlas/AskAtlas/api/internal/dashboard/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Reusable shorthand for "no rows" return on the term-resolver mocks.
func errNoRows() error { return sql.ErrNoRows }

// emptyStats builds a zero-valued GetUserPracticeStatsRow.
func emptyStats() db.GetUserPracticeStatsRow {
	return db.GetUserPracticeStatsRow{
		SessionsCompleted: 0,
		TotalCorrect:      0,
		TotalQuestions:    0,
	}
}

// emptyFileStats builds a zero-valued GetUserFileStatsRow.
func emptyFileStats() db.GetUserFileStatsRow {
	return db.GetUserFileStatsRow{
		TotalCount: 0,
		TotalSize:  0,
	}
}

// stubAllSectionsEmpty wires repo expectations so every section
// returns its empty-state shape. Used by tests that only care
// about the "all empty" case.
func stubAllSectionsEmpty(t *testing.T, repo *mock_dashboard.MockRepository) {
	t.Helper()
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserStudyGuides(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), nil).Maybe()
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserFileStats(mock.Anything, mock.Anything).Return(emptyFileStats(), nil).Maybe()
	repo.EXPECT().ListRecentUserFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
}

// =============================================================================
// Term-resolver waterfall.
// =============================================================================

func TestGetDashboard_CoursesSection_TermWaterfall_ActiveStep(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	// Step 1 returns a term -> sibling waterfall steps must NOT be
	// queried (mockery would fail the test if ResolveCurrentTermLastEnded
	// or ResolveCurrentTermLexLatest were called without an EXPECT).
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("Spring 2026", nil)
	repo.EXPECT().ListEnrolledCoursesForTerm(mock.Anything, mock.MatchedBy(func(p db.ListEnrolledCoursesForTermParams) bool {
		return p.Term == "Spring 2026" && p.PageLimit == dashboard.MaxCourses
	})).Return([]db.ListEnrolledCoursesForTermRow{}, nil)
	stubOtherSectionsEmpty(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	require.NotNil(t, result.Courses.CurrentTerm)
	assert.Equal(t, "Spring 2026", *result.Courses.CurrentTerm)
}

func TestGetDashboard_CoursesSection_TermWaterfall_FallsToLastEnded(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	// Step 1 has no row, step 2 returns the last-ended term, step 3
	// must NOT be invoked.
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("Fall 2025", nil)
	repo.EXPECT().ListEnrolledCoursesForTerm(mock.Anything, mock.MatchedBy(func(p db.ListEnrolledCoursesForTermParams) bool {
		return p.Term == "Fall 2025"
	})).Return([]db.ListEnrolledCoursesForTermRow{}, nil)
	stubOtherSectionsEmpty(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	require.NotNil(t, result.Courses.CurrentTerm)
	assert.Equal(t, "Fall 2025", *result.Courses.CurrentTerm)
}

func TestGetDashboard_CoursesSection_TermWaterfall_FallsToLexLatest(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("Spring 2026", nil)
	repo.EXPECT().ListEnrolledCoursesForTerm(mock.Anything, mock.Anything).Return([]db.ListEnrolledCoursesForTermRow{}, nil)
	stubOtherSectionsEmpty(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	require.NotNil(t, result.Courses.CurrentTerm)
	assert.Equal(t, "Spring 2026", *result.Courses.CurrentTerm)
}

func TestGetDashboard_CoursesSection_NoEnrollments_RendersEmpty(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	// All 3 waterfall steps return ErrNoRows -> term is "" -> the
	// courses section ships empty + current_term=null. The
	// ListEnrolledCoursesForTerm query must NOT run because there's
	// no resolved term to query against.
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows())
	stubOtherSectionsEmpty(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	assert.Equal(t, int32(0), result.Courses.EnrolledCount)
	assert.Nil(t, result.Courses.CurrentTerm)
	require.NotNil(t, result.Courses.Courses) // empty, not nil
	assert.Empty(t, result.Courses.Courses)
}

func TestGetDashboard_CoursesSection_TermResolverDBError_PropagatesAs500(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	boom := errors.New("connection lost")
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", boom)
	// Subsequent steps must NOT run -- a real DB error short-circuits
	// the waterfall (only sql.ErrNoRows falls through).

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

// =============================================================================
// Courses payload assembly.
// =============================================================================

func TestGetDashboard_CoursesSection_AssemblesEnrollmentRows(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	courseID := uuid.New()
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("Spring 2026", nil)
	repo.EXPECT().ListEnrolledCoursesForTerm(mock.Anything, mock.Anything).Return(
		[]db.ListEnrolledCoursesForTermRow{{
			CourseID:         utils.UUID(courseID),
			CourseDepartment: "CPTS",
			CourseNumber:     "322",
			CourseTitle:      "Software Engineering Principles I",
			MemberRole:       db.CourseRoleStudent,
			SectionTerm:      "Spring 2026",
		}}, nil)
	stubOtherSectionsEmpty(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	require.Len(t, result.Courses.Courses, 1)
	c := result.Courses.Courses[0]
	assert.Equal(t, courseID, c.ID)
	assert.Equal(t, "CPTS", c.Department)
	assert.Equal(t, "322", c.Number)
	assert.Equal(t, dashboard.MemberRoleStudent, c.Role)
	assert.Equal(t, "Spring 2026", c.SectionTerm)
	assert.Equal(t, int32(1), result.Courses.EnrolledCount)
}

// =============================================================================
// Study-guides section.
// =============================================================================

func TestGetDashboard_StudyGuidesSection_AssemblesCountAndRecent(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	guideID := uuid.New()
	updatedAt := time.Date(2026, 3, 28, 10, 0, 0, 0, time.UTC)
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(7), nil)
	repo.EXPECT().ListRecentUserStudyGuides(mock.Anything, mock.MatchedBy(func(p db.ListRecentUserStudyGuidesParams) bool {
		return p.PageLimit == dashboard.RecentGuidesLimit
	})).Return([]db.ListRecentUserStudyGuidesRow{{
		StudyGuideID:     utils.UUID(guideID),
		StudyGuideTitle:  "Binary Trees",
		UpdatedAt:        pgtype.Timestamptz{Time: updatedAt, Valid: true},
		CourseDepartment: "CPTS",
		CourseNumber:     "322",
	}}, nil)
	stubOtherSectionsExceptGuides(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	assert.Equal(t, int32(7), result.StudyGuides.CreatedCount)
	require.Len(t, result.StudyGuides.Recent, 1)
	g := result.StudyGuides.Recent[0]
	assert.Equal(t, guideID, g.ID)
	assert.Equal(t, "Binary Trees", g.Title)
	assert.True(t, g.UpdatedAt.Equal(updatedAt))
}

// =============================================================================
// Practice section.
// =============================================================================

func TestGetDashboard_PracticeSection_AccuracyComputedFromStats(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	// 3 sessions, 27 of 36 correct -> 75% rounded.
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(db.GetUserPracticeStatsRow{
		SessionsCompleted: 3,
		TotalCorrect:      27,
		TotalQuestions:    36,
	}, nil)
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(36), nil)
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(nil, nil)
	stubOtherSectionsExceptPractice(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	assert.Equal(t, int32(3), result.Practice.SessionsCompleted)
	assert.Equal(t, int32(36), result.Practice.TotalQuestionsAnswered)
	assert.Equal(t, int32(75), result.Practice.OverallAccuracy)
}

func TestGetDashboard_PracticeSection_NoCompletedSessions_AccuracyIsZero(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	// No completed sessions -> SUMs are 0 -> percentage(0, 0) returns
	// 0 (the divide-by-zero guard, not NaN).
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), nil)
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(0), nil)
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(nil, nil)
	stubOtherSectionsExceptPractice(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	assert.Equal(t, int32(0), result.Practice.SessionsCompleted)
	assert.Equal(t, int32(0), result.Practice.TotalQuestionsAnswered)
	assert.Equal(t, int32(0), result.Practice.OverallAccuracy)
	require.NotNil(t, result.Practice.RecentSessions) // [] not nil
	assert.Empty(t, result.Practice.RecentSessions)
}

func TestGetDashboard_PracticeSection_RecentSessionPercentage(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	completedAt := time.Date(2026, 4, 3, 15, 0, 0, 0, time.UTC)
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), nil)
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(0), nil)
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(
		[]db.ListRecentUserSessionsRow{{
			SessionID:       utils.UUID(uuid.New()),
			CompletedAt:     pgtype.Timestamptz{Time: completedAt, Valid: true},
			CorrectAnswers:  4,
			TotalQuestions:  5,
			QuizTitle:       "Tree Traversal Quiz",
			StudyGuideTitle: "Binary Trees",
		}}, nil)
	stubOtherSectionsExceptPractice(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	require.Len(t, result.Practice.RecentSessions, 1)
	s := result.Practice.RecentSessions[0]
	// 4/5 = 80%
	assert.Equal(t, int32(80), s.ScorePercentage)
	assert.Equal(t, "Tree Traversal Quiz", s.QuizTitle)
	assert.Equal(t, "Binary Trees", s.StudyGuideTitle)
}

// =============================================================================
// Files section.
// =============================================================================

func TestGetDashboard_FilesSection_AssemblesStatsAndRecent(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	fileID := uuid.New()
	updatedAt := time.Date(2026, 4, 4, 10, 0, 0, 0, time.UTC)
	repo.EXPECT().GetUserFileStats(mock.Anything, mock.Anything).Return(db.GetUserFileStatsRow{
		TotalCount: 15,
		TotalSize:  52428800,
	}, nil)
	repo.EXPECT().ListRecentUserFiles(mock.Anything, mock.MatchedBy(func(p db.ListRecentUserFilesParams) bool {
		return p.PageLimit == dashboard.RecentFilesLimit
	})).Return([]db.ListRecentUserFilesRow{{
		FileID:       utils.UUID(fileID),
		FileName:     "Lecture Notes Week 12.pdf",
		FileMimeType: "application/pdf",
		UpdatedAt:    pgtype.Timestamptz{Time: updatedAt, Valid: true},
	}}, nil)
	stubOtherSectionsExceptFiles(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)
	assert.Equal(t, int32(15), result.Files.TotalCount)
	assert.Equal(t, int64(52428800), result.Files.TotalSize)
	require.Len(t, result.Files.Recent, 1)
	f := result.Files.Recent[0]
	assert.Equal(t, fileID, f.ID)
	assert.Equal(t, "Lecture Notes Week 12.pdf", f.Name)
	assert.Equal(t, "application/pdf", f.MimeType)
}

// =============================================================================
// Per-section error propagation.
// =============================================================================

func TestGetDashboard_StudyGuideQueryFails_PropagatesAs500(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows())
	boom := errors.New("connection lost")
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(0), boom)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

func TestGetDashboard_PracticeStatsFails_PropagatesAs500(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)

	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows())
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(0), nil)
	repo.EXPECT().ListRecentUserStudyGuides(mock.Anything, mock.Anything).Return(nil, nil)
	boom := errors.New("connection lost")
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), boom)

	_, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.Error(t, err)
	assert.ErrorIs(t, err, boom)
}

// =============================================================================
// All-empty assembly (full happy-path with no data).
// =============================================================================

func TestGetDashboard_NoData_AllSectionsEmpty(t *testing.T) {
	repo := mock_dashboard.NewMockRepository(t)
	svc := dashboard.NewService(repo)
	stubAllSectionsEmpty(t, repo)

	result, err := svc.GetDashboard(context.Background(), dashboard.GetDashboardParams{ViewerID: uuid.New()})
	require.NoError(t, err)

	// All count fields zero, all list fields non-nil empty.
	assert.Equal(t, int32(0), result.Courses.EnrolledCount)
	assert.Nil(t, result.Courses.CurrentTerm)
	require.NotNil(t, result.Courses.Courses)
	assert.Empty(t, result.Courses.Courses)

	assert.Equal(t, int32(0), result.StudyGuides.CreatedCount)
	require.NotNil(t, result.StudyGuides.Recent)
	assert.Empty(t, result.StudyGuides.Recent)

	assert.Equal(t, int32(0), result.Practice.SessionsCompleted)
	assert.Equal(t, int32(0), result.Practice.TotalQuestionsAnswered)
	assert.Equal(t, int32(0), result.Practice.OverallAccuracy)
	require.NotNil(t, result.Practice.RecentSessions)
	assert.Empty(t, result.Practice.RecentSessions)

	assert.Equal(t, int32(0), result.Files.TotalCount)
	assert.Equal(t, int64(0), result.Files.TotalSize)
	require.NotNil(t, result.Files.Recent)
	assert.Empty(t, result.Files.Recent)
}

// =============================================================================
// Helpers (one per "other sections except X" pattern).
// =============================================================================

func stubOtherSectionsEmpty(t *testing.T, repo *mock_dashboard.MockRepository) {
	t.Helper()
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserStudyGuides(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), nil).Maybe()
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserFileStats(mock.Anything, mock.Anything).Return(emptyFileStats(), nil).Maybe()
	repo.EXPECT().ListRecentUserFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
}

func stubOtherSectionsExceptGuides(t *testing.T, repo *mock_dashboard.MockRepository) {
	t.Helper()
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), nil).Maybe()
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserFileStats(mock.Anything, mock.Anything).Return(emptyFileStats(), nil).Maybe()
	repo.EXPECT().ListRecentUserFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
}

func stubOtherSectionsExceptPractice(t *testing.T, repo *mock_dashboard.MockRepository) {
	t.Helper()
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserStudyGuides(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserFileStats(mock.Anything, mock.Anything).Return(emptyFileStats(), nil).Maybe()
	repo.EXPECT().ListRecentUserFiles(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
}

func stubOtherSectionsExceptFiles(t *testing.T, repo *mock_dashboard.MockRepository) {
	t.Helper()
	repo.EXPECT().ResolveCurrentTermActive(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLastEnded(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().ResolveCurrentTermLexLatest(mock.Anything, mock.Anything).Return("", errNoRows()).Maybe()
	repo.EXPECT().CountUserStudyGuides(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserStudyGuides(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	repo.EXPECT().GetUserPracticeStats(mock.Anything, mock.Anything).Return(emptyStats(), nil).Maybe()
	repo.EXPECT().CountUserAnsweredQuestions(mock.Anything, mock.Anything).Return(int64(0), nil).Maybe()
	repo.EXPECT().ListRecentUserSessions(mock.Anything, mock.Anything).Return(nil, nil).Maybe()
}

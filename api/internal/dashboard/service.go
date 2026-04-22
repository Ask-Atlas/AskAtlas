package dashboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// Service implements the GET /api/me/dashboard business logic.
type Service struct {
	repo Repository
}

// NewService wires a dashboard Service over the given repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetDashboard assembles the full dashboard payload for the viewer
// (ASK-155).
//
// Strategy:
//   - Resolve the current term via the 3-step waterfall (active ->
//     last-ended -> lex-latest). When the user has no enrollments
//     at all every step returns sql.ErrNoRows; the resolved term is
//     nil and the courses section ships as
//     {enrolled_count: 0, current_term: null, courses: []}.
//   - For each of the 4 sections fan out to per-section queries
//     and assemble in-process. Sequential rather than goroutine
//     fan-out: each query is sub-2ms on indexed reads at MVP scale,
//     and the connection-pool churn from 10 concurrent queries
//     would outweigh the parallelism win for what is already a
//     single-client request.
//   - Any per-query error fails the entire request (per spec --
//     never return partial data). Wrapped with section context so
//     a 500 logs identify the failing section.
func (s *Service) GetDashboard(ctx context.Context, p GetDashboardParams) (DashboardData, error) {
	viewerPgxID := utils.UUID(p.ViewerID)

	courses, err := s.buildCoursesSection(ctx, viewerPgxID)
	if err != nil {
		return DashboardData{}, fmt.Errorf("GetDashboard: courses: %w", err)
	}
	guides, err := s.buildStudyGuidesSection(ctx, viewerPgxID)
	if err != nil {
		return DashboardData{}, fmt.Errorf("GetDashboard: study guides: %w", err)
	}
	practice, err := s.buildPracticeSection(ctx, viewerPgxID)
	if err != nil {
		return DashboardData{}, fmt.Errorf("GetDashboard: practice: %w", err)
	}
	files, err := s.buildFilesSection(ctx, viewerPgxID)
	if err != nil {
		return DashboardData{}, fmt.Errorf("GetDashboard: files: %w", err)
	}

	return DashboardData{
		Courses:     courses,
		StudyGuides: guides,
		Practice:    practice,
		Files:       files,
	}, nil
}

// buildCoursesSection runs the term-resolver waterfall, then
// (when a term resolves) fetches the enrollments capped at
// MaxCourses. When the user has no enrollments the section ships
// with current_term=null + enrolled_count=0 + empty list.
func (s *Service) buildCoursesSection(ctx context.Context, viewer pgtype.UUID) (DashboardCoursesSection, error) {
	emptySection := DashboardCoursesSection{
		EnrolledCount: 0,
		CurrentTerm:   nil,
		Courses:       []DashboardCourseSummary{},
	}

	term, err := s.resolveCurrentTerm(ctx, viewer)
	if err != nil {
		return DashboardCoursesSection{}, err
	}
	if term == "" {
		// No enrollments at all -- ship the empty section.
		return emptySection, nil
	}

	rows, err := s.repo.ListEnrolledCoursesForTerm(ctx, db.ListEnrolledCoursesForTermParams{
		ViewerID:  viewer,
		Term:      term,
		PageLimit: MaxCourses,
	})
	if err != nil {
		return DashboardCoursesSection{}, fmt.Errorf("ListEnrolledCoursesForTerm: %w", err)
	}

	out := make([]DashboardCourseSummary, 0, len(rows))
	for _, r := range rows {
		c, err := mapCourse(r)
		if err != nil {
			return DashboardCoursesSection{}, err
		}
		out = append(out, c)
	}

	return DashboardCoursesSection{
		EnrolledCount: int32(len(out)),
		CurrentTerm:   &term,
		Courses:       out,
	}, nil
}

// resolveCurrentTerm runs the 3-step waterfall. Returns "" (and a
// nil error) when every step returns sql.ErrNoRows -- that
// indicates "user has no enrollments anywhere". A real DB error
// at any step short-circuits and propagates.
func (s *Service) resolveCurrentTerm(ctx context.Context, viewer pgtype.UUID) (string, error) {
	if term, err := s.repo.ResolveCurrentTermActive(ctx, viewer); err == nil {
		return term, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("resolveCurrentTerm: active: %w", err)
	}

	if term, err := s.repo.ResolveCurrentTermLastEnded(ctx, viewer); err == nil {
		return term, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("resolveCurrentTerm: last-ended: %w", err)
	}

	if term, err := s.repo.ResolveCurrentTermLexLatest(ctx, viewer); err == nil {
		return term, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("resolveCurrentTerm: lex-latest: %w", err)
	}

	return "", nil
}

// buildStudyGuidesSection fetches the count + recent guides. Both
// queries filter sg.deleted_at IS NULL in SQL.
func (s *Service) buildStudyGuidesSection(ctx context.Context, viewer pgtype.UUID) (DashboardStudyGuidesSection, error) {
	count, err := s.repo.CountUserStudyGuides(ctx, viewer)
	if err != nil {
		return DashboardStudyGuidesSection{}, fmt.Errorf("CountUserStudyGuides: %w", err)
	}

	rows, err := s.repo.ListRecentUserStudyGuides(ctx, db.ListRecentUserStudyGuidesParams{
		ViewerID:  viewer,
		PageLimit: RecentGuidesLimit,
	})
	if err != nil {
		return DashboardStudyGuidesSection{}, fmt.Errorf("ListRecentUserStudyGuides: %w", err)
	}

	out := make([]DashboardStudyGuideSummary, 0, len(rows))
	for _, r := range rows {
		g, err := mapStudyGuide(r)
		if err != nil {
			return DashboardStudyGuidesSection{}, err
		}
		out = append(out, g)
	}

	return DashboardStudyGuidesSection{
		CreatedCount: int32(count),
		Recent:       out,
	}, nil
}

// buildPracticeSection fetches the aggregate stats + answered-
// questions count + recent sessions, computes overall_accuracy
// in Go, and assembles the section.
func (s *Service) buildPracticeSection(ctx context.Context, viewer pgtype.UUID) (DashboardPracticeSection, error) {
	stats, err := s.repo.GetUserPracticeStats(ctx, viewer)
	if err != nil {
		return DashboardPracticeSection{}, fmt.Errorf("GetUserPracticeStats: %w", err)
	}

	answered, err := s.repo.CountUserAnsweredQuestions(ctx, viewer)
	if err != nil {
		return DashboardPracticeSection{}, fmt.Errorf("CountUserAnsweredQuestions: %w", err)
	}

	rows, err := s.repo.ListRecentUserSessions(ctx, db.ListRecentUserSessionsParams{
		ViewerID:  viewer,
		PageLimit: RecentSessionsLimit,
	})
	if err != nil {
		return DashboardPracticeSection{}, fmt.Errorf("ListRecentUserSessions: %w", err)
	}

	out := make([]DashboardSessionSummary, 0, len(rows))
	for _, r := range rows {
		sess, err := mapSession(r)
		if err != nil {
			return DashboardPracticeSection{}, err
		}
		out = append(out, sess)
	}

	return DashboardPracticeSection{
		SessionsCompleted:      int32(stats.SessionsCompleted),
		TotalQuestionsAnswered: int32(answered),
		// percentage() returns 0 when TotalQuestions is 0 -- the
		// SUM in the SQL query is COALESCE(..., 0) so this
		// branch is reachable for users with no completed sessions.
		OverallAccuracy: percentage(stats.TotalCorrect, stats.TotalQuestions),
		RecentSessions:  out,
	}, nil
}

// buildFilesSection fetches the count + total bytes + recent files.
// All queries filter on deletion_status IS NULL AND status='complete'
// in SQL.
func (s *Service) buildFilesSection(ctx context.Context, viewer pgtype.UUID) (DashboardFilesSection, error) {
	stats, err := s.repo.GetUserFileStats(ctx, viewer)
	if err != nil {
		return DashboardFilesSection{}, fmt.Errorf("GetUserFileStats: %w", err)
	}

	rows, err := s.repo.ListRecentUserFiles(ctx, db.ListRecentUserFilesParams{
		ViewerID:  viewer,
		PageLimit: RecentFilesLimit,
	})
	if err != nil {
		return DashboardFilesSection{}, fmt.Errorf("ListRecentUserFiles: %w", err)
	}

	out := make([]DashboardFileSummary, 0, len(rows))
	for _, r := range rows {
		f, err := mapFile(r)
		if err != nil {
			return DashboardFilesSection{}, err
		}
		out = append(out, f)
	}

	return DashboardFilesSection{
		TotalCount: int32(stats.TotalCount),
		TotalSize:  stats.TotalSize,
		Recent:     out,
	}, nil
}

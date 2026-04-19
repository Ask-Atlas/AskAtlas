package studyguides

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// AssertCourseExists is the 404-distinguishing preflight. Handler calls
// it before ListStudyGuides to emit a tailored "Course not found" 404
// when the course_id doesn't resolve. Pulling the check here (rather
// than interpreting an empty result) matches the pattern from
// courses.Service.assertCourseAndSection.
func (s *Service) AssertCourseExists(ctx context.Context, courseID uuid.UUID) error {
	exists, err := s.repo.CourseExistsForGuides(ctx, utils.UUID(courseID))
	if err != nil {
		return fmt.Errorf("AssertCourseExists: %w", err)
	}
	if !exists {
		return apperrors.NewNotFound("Course not found")
	}
	return nil
}

// GetStudyGuide fetches the full study-guide detail including nested
// course + creator + recommenders + quizzes + resources + files + the
// viewer's own vote state.
//
// This is a pure read -- no view-count increment, no last-viewed
// upsert, no mutation. View tracking will live on its own dedicated
// POST (future ticket mirroring ASK-134's POST /files/{id}/view).
//
// Runs 6 queries total: GetStudyGuideDetail (404 candidate, must
// complete before the rest fan out so we don't make 5 speculative
// round trips for a guide that doesn't exist), then the 5 independent
// sibling queries (user_vote, recommenders, quizzes, resources,
// files) issued in parallel via errgroup. The 5-wide fan-out cuts
// end-to-end latency from sum(query_time) to max(query_time) while
// preserving the strict error-propagation semantics -- any sibling
// error short-circuits the rest via ctx cancellation and returns a
// 500 to the handler.
//
// Soft-deleted guides return apperrors.ErrNotFound via the underlying
// query's WHERE sg.deleted_at IS NULL. The handler maps that to a
// 404 'Study guide not found'.
func (s *Service) GetStudyGuide(ctx context.Context, p GetStudyGuideParams) (StudyGuideDetail, error) {
	guidePgxID := utils.UUID(p.StudyGuideID)
	viewerPgxID := utils.UUID(p.ViewerID)

	row, err := s.repo.GetStudyGuideDetail(ctx, guidePgxID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StudyGuideDetail{}, apperrors.NewNotFound("Study guide not found")
		}
		return StudyGuideDetail{}, fmt.Errorf("GetStudyGuide: detail: %w", err)
	}

	detail, err := mapStudyGuideDetail(row)
	if err != nil {
		return StudyGuideDetail{}, fmt.Errorf("GetStudyGuide: map detail: %w", err)
	}

	// The 5 sibling queries are independent of each other -- fan out
	// in parallel via errgroup. Each goroutine owns its slot on
	// `detail` (different field per slot), so there's no data race on
	// the struct.
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		vote, err := s.repo.GetUserVoteForGuide(gctx, db.GetUserVoteForGuideParams{
			StudyGuideID: guidePgxID,
			ViewerID:     viewerPgxID,
		})
		switch {
		case err == nil:
			v := GuideVote(vote)
			detail.UserVote = &v
		case errors.Is(err, sql.ErrNoRows):
			detail.UserVote = nil
		default:
			return fmt.Errorf("user vote: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideRecommenders(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("recommenders: %w", err)
		}
		detail.RecommendedBy = make([]Creator, 0, len(rows))
		for _, r := range rows {
			rec, err := mapRecommender(r)
			if err != nil {
				return fmt.Errorf("map recommender: %w", err)
			}
			detail.RecommendedBy = append(detail.RecommendedBy, rec)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideQuizzesWithQuestionCount(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("quizzes: %w", err)
		}
		detail.Quizzes = make([]Quiz, 0, len(rows))
		for _, q := range rows {
			quiz, err := mapQuiz(q)
			if err != nil {
				return fmt.Errorf("map quiz: %w", err)
			}
			detail.Quizzes = append(detail.Quizzes, quiz)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideResources(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("resources: %w", err)
		}
		detail.Resources = make([]Resource, 0, len(rows))
		for _, r := range rows {
			res, err := mapResource(r)
			if err != nil {
				return fmt.Errorf("map resource: %w", err)
			}
			detail.Resources = append(detail.Resources, res)
		}
		return nil
	})

	g.Go(func() error {
		rows, err := s.repo.ListGuideFiles(gctx, guidePgxID)
		if err != nil {
			return fmt.Errorf("files: %w", err)
		}
		detail.Files = make([]GuideFile, 0, len(rows))
		for _, f := range rows {
			gf, err := mapGuideFile(f)
			if err != nil {
				return fmt.Errorf("map file: %w", err)
			}
			detail.Files = append(detail.Files, gf)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return StudyGuideDetail{}, fmt.Errorf("GetStudyGuide: %w", err)
	}

	return detail, nil
}

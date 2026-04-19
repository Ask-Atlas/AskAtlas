package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service. The
// production impl is sqlc-backed (sqlc_repository.go); tests inject
// a mockery-generated mock. Mirrors the
// quizzes.Repository / studyguides.Repository pattern.
//
// CountQuizQuestions is duplicated from quizzes.Repository (also
// added in ASK-115) -- both packages legitimately need a snapshot
// count, but importing the quizzes Repository would couple the two
// surfaces. Sqlc generates per-package queriers anyway, so the
// duplication is a wash.
type Repository interface {
	CheckQuizLiveForSession(ctx context.Context, quizID pgtype.UUID) (bool, error)
	DeleteStaleIncompleteSessions(ctx context.Context, arg db.DeleteStaleIncompleteSessionsParams) error
	FindIncompleteSession(ctx context.Context, arg db.FindIncompleteSessionParams) (db.FindIncompleteSessionRow, error)
	InsertPracticeSessionIfAbsent(ctx context.Context, arg db.InsertPracticeSessionIfAbsentParams) (db.InsertPracticeSessionIfAbsentRow, error)
	SnapshotQuizQuestions(ctx context.Context, arg db.SnapshotQuizQuestionsParams) error
	ListSessionAnswers(ctx context.Context, sessionID pgtype.UUID) ([]db.ListSessionAnswersRow, error)
	CountQuizQuestions(ctx context.Context, quizID pgtype.UUID) (int64, error)

	// InTx runs fn inside a single Postgres transaction. The
	// Repository passed to fn is scoped to the tx via Queries.WithTx,
	// so any sqlc call made through it participates in the same tx.
	// Used by StartSession for the atomic insert-session +
	// snapshot-questions write.
	InTx(ctx context.Context, fn func(Repository) error) error
}

// Service is the business-logic layer for the sessions feature.
type Service struct {
	repo Repository
}

// NewService creates a new Service backed by the given Repository.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// StartSession starts a new practice session OR resumes the user's
// existing in-progress session for a quiz (ASK-128). The result's
// Created flag tells the handler whether to render 201 (created) or
// 200 (resumed).
//
// Order of operations:
//  1. CheckQuizLiveForSession -- quiz live AND parent guide live.
//     A missing/deleted/parent-deleted quiz all return 404; the
//     caller cannot distinguish (info-leak prevention, same rule as
//     GetQuiz / DeleteQuiz / UpdateQuiz).
//  2. DeleteStaleIncompleteSessions -- hard-delete this user's
//     stale incomplete session for this quiz (started_at > 7 days
//     ago) so the next step sees a clean slate. Idempotent (no-op
//     when nothing is stale). Per spec AC6: a stale session forces
//     a fresh start (201), not a resume (200).
//  3. FindIncompleteSession -- if found, hydrate the existing row
//     plus its answers and return Created=false (200 resume). The
//     partial unique index guarantees AT MOST one match.
//  4. Inside InTx:
//     a. CountQuizQuestions -- snapshot count. The quizzes
//     service caps at 100 questions per quiz, so the int64 ->
//     int32 narrowing for total_questions is safe in practice;
//     we still bound-check defensively.
//     b. InsertPracticeSessionIfAbsent -- ON CONFLICT DO NOTHING
//     against the partial unique index. On race-loss the
//     query returns sql.ErrNoRows; we mark raceLost and exit
//     the tx cleanly.
//     c. SnapshotQuizQuestions -- bulk-insert one
//     practice_session_questions row per quiz_questions row.
//  5. If raceLost: re-run FindIncompleteSession, return as 200
//     resume. The losing request never wrote a snapshot.
//  6. Otherwise: load the (always-empty) answers list and return
//     Created=true (201).
//
// Why not lock the parent quiz row during the tx: the snapshot
// captures questions at commit time, which is fine -- a question
// added a microsecond later is "after" by the spec and is excluded
// (AC3); a deletion mid-snapshot is observed in our snapshot's
// COMMITTED isolation level (questions visible at tx start are
// snapshotted, deletions concurrent with snapshot use ON DELETE
// SET NULL on the practice_session_questions side post-commit).
func (s *Service) StartSession(ctx context.Context, p StartSessionParams) (StartSessionResult, error) {
	quizPgxID := utils.UUID(p.QuizID)
	userPgxID := utils.UUID(p.UserID)

	live, err := s.repo.CheckQuizLiveForSession(ctx, quizPgxID)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: live check: %w", err)
	}
	if !live {
		return StartSessionResult{}, apperrors.NewNotFound("Quiz not found")
	}

	if err := s.repo.DeleteStaleIncompleteSessions(ctx, db.DeleteStaleIncompleteSessionsParams{
		UserID: userPgxID,
		QuizID: quizPgxID,
	}); err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: stale cleanup: %w", err)
	}

	existing, err := s.repo.FindIncompleteSession(ctx, db.FindIncompleteSessionParams{
		UserID: userPgxID,
		QuizID: quizPgxID,
	})
	if err == nil {
		return s.hydrateExisting(ctx, existing)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return StartSessionResult{}, fmt.Errorf("StartSession: resume probe: %w", err)
	}

	var inserted db.InsertPracticeSessionIfAbsentRow
	raceLost := false
	if err := s.repo.InTx(ctx, func(tx Repository) error {
		count, err := tx.CountQuizQuestions(ctx, quizPgxID)
		if err != nil {
			return fmt.Errorf("count questions: %w", err)
		}
		if count < 0 || count > math.MaxInt32 {
			return fmt.Errorf("count out of range: %d", count)
		}

		ins, err := tx.InsertPracticeSessionIfAbsent(ctx, db.InsertPracticeSessionIfAbsentParams{
			UserID:         userPgxID,
			QuizID:         quizPgxID,
			TotalQuestions: int32(count),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Race-lost: a concurrent start already inserted the
				// incomplete session. Bail cleanly so the caller can
				// re-fetch and resume.
				raceLost = true
				return nil
			}
			return fmt.Errorf("insert: %w", err)
		}
		inserted = ins

		if err := tx.SnapshotQuizQuestions(ctx, db.SnapshotQuizQuestionsParams{
			SessionID: ins.ID,
			QuizID:    quizPgxID,
		}); err != nil {
			return fmt.Errorf("snapshot: %w", err)
		}
		return nil
	}); err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: tx: %w", err)
	}

	if raceLost {
		row, err := s.repo.FindIncompleteSession(ctx, db.FindIncompleteSessionParams{
			UserID: userPgxID,
			QuizID: quizPgxID,
		})
		if err != nil {
			// Edge case: race-lost AND no incomplete on re-fetch (the
			// winner already completed it in the microseconds between
			// our race-loss and the re-fetch). Surface as 500 rather
			// than guess at the right semantics -- this should be
			// vanishingly rare.
			return StartSessionResult{}, fmt.Errorf("StartSession: race re-fetch: %w", err)
		}
		return s.hydrateExisting(ctx, row)
	}

	answers, err := s.repo.ListSessionAnswers(ctx, inserted.ID)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: list answers: %w", err)
	}
	detail, err := mapInsertedSession(inserted, answers)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("StartSession: map: %w", err)
	}
	return StartSessionResult{Session: detail, Created: true}, nil
}

// hydrateExisting loads a found-incomplete-session row's answers
// and projects it onto a Created=false StartSessionResult. Shared
// between the natural resume path (step 3 above) and the
// race-lost fallback (step 5).
func (s *Service) hydrateExisting(ctx context.Context, row db.FindIncompleteSessionRow) (StartSessionResult, error) {
	answers, err := s.repo.ListSessionAnswers(ctx, row.ID)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("hydrateExisting: list answers: %w", err)
	}
	detail, err := mapFoundSession(row, answers)
	if err != nil {
		return StartSessionResult{}, fmt.Errorf("hydrateExisting: map: %w", err)
	}
	return StartSessionResult{Session: detail, Created: false}, nil
}

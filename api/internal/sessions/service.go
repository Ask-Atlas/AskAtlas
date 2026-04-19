package sessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by Service. The
// production impl is sqlc-backed (sqlc_repository.go); tests inject
// a mockery-generated mock. Mirrors the
// quizzes.Repository / studyguides.Repository pattern.
type Repository interface {
	CheckQuizLiveForSession(ctx context.Context, quizID pgtype.UUID) (bool, error)
	DeleteStaleIncompleteSessions(ctx context.Context, arg db.DeleteStaleIncompleteSessionsParams) error
	FindIncompleteSession(ctx context.Context, arg db.FindIncompleteSessionParams) (db.FindIncompleteSessionRow, error)
	InsertPracticeSessionIfAbsent(ctx context.Context, arg db.InsertPracticeSessionIfAbsentParams) (db.InsertPracticeSessionIfAbsentRow, error)
	SnapshotQuizQuestionsAndUpdateCount(ctx context.Context, arg db.SnapshotQuizQuestionsAndUpdateCountParams) (int32, error)
	ListSessionAnswers(ctx context.Context, sessionID pgtype.UUID) ([]db.ListSessionAnswersRow, error)

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
//     incomplete session for this quiz when started_at is older
//     than 7 days, so the next step sees a clean slate. Idempotent
//     (no-op when nothing is stale). Per spec AC6: a stale session
//     forces a fresh start (201), not a resume (200).
//  3. FindIncompleteSession -- if found, hydrate the existing row
//     plus its answers and return Created=false (200 resume). The
//     partial unique index guarantees AT MOST one match.
//  4. Inside InTx:
//     a. InsertPracticeSessionIfAbsent -- inserts the session row
//     with total_questions = 0 (column default). ON CONFLICT DO
//     NOTHING against the partial unique index. On race-loss
//     the query returns sql.ErrNoRows; we mark raceLost and
//     exit the tx cleanly.
//     b. SnapshotQuizQuestionsAndUpdateCount -- single CTE
//     statement that bulk-inserts the snapshot rows AND updates
//     the session's total_questions to the actual snapshot
//     count. Single statement = single Postgres snapshot, so
//     count and snapshot are guaranteed consistent even under
//     concurrent quiz edits at READ COMMITTED isolation
//     (gemini + copilot PR feedback). The returned int32 is
//     the new total_questions; we sync it onto inserted in
//     memory so the response carries the correct value
//     without a round-trip.
//  5. If raceLost: re-run FindIncompleteSession, return as 200
//     resume. The losing request never wrote a snapshot.
//  6. Otherwise: return Created=true (201) with answers=[] (we
//     skip ListSessionAnswers because a fresh session never has
//     answers -- gemini PR feedback).
//
// Why not lock the parent quiz row during the tx: the snapshot
// captures questions in a single statement, so a concurrent
// edit either lands entirely before our statement (and is in the
// snapshot) or entirely after (and is not). The CTE in
// SnapshotQuizQuestionsAndUpdateCount makes this race-free.
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
		ins, err := tx.InsertPracticeSessionIfAbsent(ctx, db.InsertPracticeSessionIfAbsentParams{
			UserID: userPgxID,
			QuizID: quizPgxID,
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

		total, err := tx.SnapshotQuizQuestionsAndUpdateCount(ctx, db.SnapshotQuizQuestionsAndUpdateCountParams{
			SessionID: ins.ID,
			QuizID:    quizPgxID,
		})
		if err != nil {
			return fmt.Errorf("snapshot: %w", err)
		}
		// Sync the in-memory session's TotalQuestions to the
		// authoritative value derived from the actual snapshot row
		// count -- the inserted row's TotalQuestions was 0 (column
		// default) before the CTE updated it.
		inserted.TotalQuestions = total
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

	// Fresh session -- no answers exist yet, so skip the
	// ListSessionAnswers round-trip and pass an empty slice
	// (gemini PR feedback).
	detail, err := mapInsertedSession(inserted, nil)
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

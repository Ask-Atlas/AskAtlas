package sessions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sqlcRepository is the production Repository implementation backed
// by sqlc-generated queries. Each method is a thin pass-through;
// service logic (404 dispatch, stale-cleanup, race handling)
// lives in service.go. Holds both pool + queries: pool is needed
// to begin transactions for InTx; queries is the non-tx default.
// Same shape as quizzes.sqlcRepository.
type sqlcRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated
// Postgres queries. Takes the pgxpool.Pool alongside the Queries
// instance so InTx can begin transactions.
func NewSQLCRepository(pool *pgxpool.Pool, queries *db.Queries) Repository {
	return &sqlcRepository{pool: pool, queries: queries}
}

// InTx runs fn inside a single Postgres transaction. The Repository
// passed to fn is scoped to the tx via Queries.WithTx, so any sqlc
// call made through it participates in the same tx. Commits on a
// nil return; rolls back on any error (including panics, via the
// deferred Rollback call). Mirrors quizzes.sqlcRepository.InTx.
func (r *sqlcRepository) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("InTx: begin tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Error("sessions: failed to rollback transaction", "error", rollbackErr)
		}
	}()

	txRepo := &sqlcRepository{pool: r.pool, queries: r.queries.WithTx(tx)}
	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("InTx: commit: %w", err)
	}
	return nil
}

func (r *sqlcRepository) CheckQuizLiveForSession(ctx context.Context, quizID pgtype.UUID) (bool, error) {
	return r.queries.CheckQuizLiveForSession(ctx, quizID)
}

func (r *sqlcRepository) DeleteStaleIncompleteSessions(ctx context.Context, arg db.DeleteStaleIncompleteSessionsParams) error {
	return r.queries.DeleteStaleIncompleteSessions(ctx, arg)
}

func (r *sqlcRepository) FindIncompleteSession(ctx context.Context, arg db.FindIncompleteSessionParams) (db.FindIncompleteSessionRow, error) {
	return r.queries.FindIncompleteSession(ctx, arg)
}

func (r *sqlcRepository) InsertPracticeSessionIfAbsent(ctx context.Context, arg db.InsertPracticeSessionIfAbsentParams) (db.InsertPracticeSessionIfAbsentRow, error) {
	return r.queries.InsertPracticeSessionIfAbsent(ctx, arg)
}

func (r *sqlcRepository) SnapshotQuizQuestionsAndUpdateCount(ctx context.Context, arg db.SnapshotQuizQuestionsAndUpdateCountParams) (int32, error) {
	return r.queries.SnapshotQuizQuestionsAndUpdateCount(ctx, arg)
}

func (r *sqlcRepository) ListSessionAnswers(ctx context.Context, sessionID pgtype.UUID) ([]db.ListSessionAnswersRow, error) {
	return r.queries.ListSessionAnswers(ctx, sessionID)
}

func (r *sqlcRepository) GetSessionForAnswerSubmission(ctx context.Context, id pgtype.UUID) (db.GetSessionForAnswerSubmissionRow, error) {
	return r.queries.GetSessionForAnswerSubmission(ctx, id)
}

func (r *sqlcRepository) CheckQuestionInSessionSnapshot(ctx context.Context, arg db.CheckQuestionInSessionSnapshotParams) (bool, error) {
	return r.queries.CheckQuestionInSessionSnapshot(ctx, arg)
}

func (r *sqlcRepository) GetQuizQuestionByID(ctx context.Context, id pgtype.UUID) (db.GetQuizQuestionByIDRow, error) {
	return r.queries.GetQuizQuestionByID(ctx, id)
}

func (r *sqlcRepository) GetCorrectOptionText(ctx context.Context, questionID pgtype.UUID) (string, error) {
	return r.queries.GetCorrectOptionText(ctx, questionID)
}

func (r *sqlcRepository) InsertPracticeAnswer(ctx context.Context, arg db.InsertPracticeAnswerParams) (db.InsertPracticeAnswerRow, error) {
	return r.queries.InsertPracticeAnswer(ctx, arg)
}

func (r *sqlcRepository) IncrementSessionCorrectAnswers(ctx context.Context, id pgtype.UUID) error {
	return r.queries.IncrementSessionCorrectAnswers(ctx, id)
}

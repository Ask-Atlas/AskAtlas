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

func (r *sqlcRepository) SnapshotQuizQuestions(ctx context.Context, arg db.SnapshotQuizQuestionsParams) error {
	return r.queries.SnapshotQuizQuestions(ctx, arg)
}

func (r *sqlcRepository) ListSessionAnswers(ctx context.Context, sessionID pgtype.UUID) ([]db.ListSessionAnswersRow, error) {
	return r.queries.ListSessionAnswers(ctx, sessionID)
}

func (r *sqlcRepository) CountQuizQuestions(ctx context.Context, quizID pgtype.UUID) (int64, error) {
	return r.queries.CountQuizQuestions(ctx, quizID)
}

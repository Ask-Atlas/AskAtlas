package quizzes

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
// service logic (validation, per-type branching, transaction
// orchestration) lives in service.go.
//
// Holds both pool + queries: pool is needed to begin transactions
// for InTx; queries is the non-tx default. Same shape as
// studyguides.sqlcRepository.
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
// deferred Rollback call). Mirrors studyguides.sqlcRepository.InTx.
func (r *sqlcRepository) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("InTx: begin tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Error("quizzes: failed to rollback transaction", "error", rollbackErr)
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

func (r *sqlcRepository) GuideExistsAndLiveForQuizzes(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.GuideExistsAndLiveForQuizzes(ctx, id)
}

func (r *sqlcRepository) InsertQuiz(ctx context.Context, arg db.InsertQuizParams) (db.InsertQuizRow, error) {
	return r.queries.InsertQuiz(ctx, arg)
}

func (r *sqlcRepository) InsertQuizQuestion(ctx context.Context, arg db.InsertQuizQuestionParams) (pgtype.UUID, error) {
	return r.queries.InsertQuizQuestion(ctx, arg)
}

func (r *sqlcRepository) InsertQuizAnswerOption(ctx context.Context, arg db.InsertQuizAnswerOptionParams) error {
	return r.queries.InsertQuizAnswerOption(ctx, arg)
}

func (r *sqlcRepository) GetQuizDetail(ctx context.Context, id pgtype.UUID) (db.GetQuizDetailRow, error) {
	return r.queries.GetQuizDetail(ctx, id)
}

func (r *sqlcRepository) ListQuizQuestionsByQuiz(ctx context.Context, quizID pgtype.UUID) ([]db.ListQuizQuestionsByQuizRow, error) {
	return r.queries.ListQuizQuestionsByQuiz(ctx, quizID)
}

func (r *sqlcRepository) ListQuizAnswerOptionsByQuiz(ctx context.Context, quizID pgtype.UUID) ([]db.QuizAnswerOption, error) {
	return r.queries.ListQuizAnswerOptionsByQuiz(ctx, quizID)
}

func (r *sqlcRepository) ListQuizzesByStudyGuide(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListQuizzesByStudyGuideRow, error) {
	return r.queries.ListQuizzesByStudyGuide(ctx, studyGuideID)
}

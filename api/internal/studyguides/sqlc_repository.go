package studyguides

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

// sqlcRepository is the production Repository implementation backed by
// sqlc-generated queries. Each method is a thin pass-through; service
// logic (validation, sort dispatch, cursor encoding) lives in service.go.
//
// Holds both pool + queries: pool is needed to begin transactions for
// InTx (see DeleteStudyGuide); queries is the non-tx default.
type sqlcRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated
// Postgres queries. Takes the pgxpool.Pool alongside the Queries
// instance so InTx can begin transactions; same shape as
// files.NewSQLCRepository.
func NewSQLCRepository(pool *pgxpool.Pool, queries *db.Queries) Repository {
	return &sqlcRepository{pool: pool, queries: queries}
}

// InTx runs fn inside a single Postgres transaction. The Repository
// passed to fn is scoped to the tx via Queries.WithTx, so any sqlc
// call made through it participates in the same tx. Commits on a nil
// return; rolls back on any error (including panics, via the deferred
// Rollback call). Mirrors files.sqlcRepository.InTx -- the pattern is
// already battle-tested in the files surface (DeleteFile + grants).
func (r *sqlcRepository) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("InTx: begin tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Error("studyguides: failed to rollback transaction", "error", rollbackErr)
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

func (r *sqlcRepository) ListStudyGuidesScoreDesc(ctx context.Context, arg db.ListStudyGuidesScoreDescParams) ([]db.ListStudyGuidesScoreDescRow, error) {
	return r.queries.ListStudyGuidesScoreDesc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesScoreAsc(ctx context.Context, arg db.ListStudyGuidesScoreAscParams) ([]db.ListStudyGuidesScoreAscRow, error) {
	return r.queries.ListStudyGuidesScoreAsc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesViewsDesc(ctx context.Context, arg db.ListStudyGuidesViewsDescParams) ([]db.ListStudyGuidesViewsDescRow, error) {
	return r.queries.ListStudyGuidesViewsDesc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesViewsAsc(ctx context.Context, arg db.ListStudyGuidesViewsAscParams) ([]db.ListStudyGuidesViewsAscRow, error) {
	return r.queries.ListStudyGuidesViewsAsc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesNewestDesc(ctx context.Context, arg db.ListStudyGuidesNewestDescParams) ([]db.ListStudyGuidesNewestDescRow, error) {
	return r.queries.ListStudyGuidesNewestDesc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesNewestAsc(ctx context.Context, arg db.ListStudyGuidesNewestAscParams) ([]db.ListStudyGuidesNewestAscRow, error) {
	return r.queries.ListStudyGuidesNewestAsc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesUpdatedDesc(ctx context.Context, arg db.ListStudyGuidesUpdatedDescParams) ([]db.ListStudyGuidesUpdatedDescRow, error) {
	return r.queries.ListStudyGuidesUpdatedDesc(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuidesUpdatedAsc(ctx context.Context, arg db.ListStudyGuidesUpdatedAscParams) ([]db.ListStudyGuidesUpdatedAscRow, error) {
	return r.queries.ListStudyGuidesUpdatedAsc(ctx, arg)
}

func (r *sqlcRepository) CourseExistsForGuides(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.CourseExistsForGuides(ctx, id)
}

func (r *sqlcRepository) GetStudyGuideDetail(ctx context.Context, id pgtype.UUID) (db.GetStudyGuideDetailRow, error) {
	return r.queries.GetStudyGuideDetail(ctx, id)
}

func (r *sqlcRepository) GetUserVoteForGuide(ctx context.Context, arg db.GetUserVoteForGuideParams) (db.VoteDirection, error) {
	return r.queries.GetUserVoteForGuide(ctx, arg)
}

func (r *sqlcRepository) ListGuideRecommenders(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideRecommendersRow, error) {
	return r.queries.ListGuideRecommenders(ctx, studyGuideID)
}

func (r *sqlcRepository) ListGuideQuizzesWithQuestionCount(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideQuizzesWithQuestionCountRow, error) {
	return r.queries.ListGuideQuizzesWithQuestionCount(ctx, studyGuideID)
}

func (r *sqlcRepository) ListGuideResources(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideResourcesRow, error) {
	return r.queries.ListGuideResources(ctx, studyGuideID)
}

func (r *sqlcRepository) ListGuideFiles(ctx context.Context, studyGuideID pgtype.UUID) ([]db.ListGuideFilesRow, error) {
	return r.queries.ListGuideFiles(ctx, studyGuideID)
}

func (r *sqlcRepository) InsertStudyGuide(ctx context.Context, arg db.InsertStudyGuideParams) (db.InsertStudyGuideRow, error) {
	return r.queries.InsertStudyGuide(ctx, arg)
}

func (r *sqlcRepository) GetStudyGuideByIDForUpdate(ctx context.Context, id pgtype.UUID) (db.GetStudyGuideByIDForUpdateRow, error) {
	return r.queries.GetStudyGuideByIDForUpdate(ctx, id)
}

func (r *sqlcRepository) SoftDeleteStudyGuide(ctx context.Context, id pgtype.UUID) error {
	return r.queries.SoftDeleteStudyGuide(ctx, id)
}

func (r *sqlcRepository) SoftDeleteQuizzesForGuide(ctx context.Context, studyGuideID pgtype.UUID) error {
	return r.queries.SoftDeleteQuizzesForGuide(ctx, studyGuideID)
}

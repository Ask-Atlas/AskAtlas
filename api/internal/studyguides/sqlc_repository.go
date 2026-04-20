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

func (r *sqlcRepository) ListMyStudyGuidesUpdated(ctx context.Context, arg db.ListMyStudyGuidesUpdatedParams) ([]db.ListMyStudyGuidesUpdatedRow, error) {
	return r.queries.ListMyStudyGuidesUpdated(ctx, arg)
}

func (r *sqlcRepository) ListMyStudyGuidesNewest(ctx context.Context, arg db.ListMyStudyGuidesNewestParams) ([]db.ListMyStudyGuidesNewestRow, error) {
	return r.queries.ListMyStudyGuidesNewest(ctx, arg)
}

func (r *sqlcRepository) ListMyStudyGuidesTitle(ctx context.Context, arg db.ListMyStudyGuidesTitleParams) ([]db.ListMyStudyGuidesTitleRow, error) {
	return r.queries.ListMyStudyGuidesTitle(ctx, arg)
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

func (r *sqlcRepository) UpdateStudyGuide(ctx context.Context, arg db.UpdateStudyGuideParams) error {
	return r.queries.UpdateStudyGuide(ctx, arg)
}

func (r *sqlcRepository) GuideExistsAndLive(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.GuideExistsAndLive(ctx, id)
}

func (r *sqlcRepository) UpsertStudyGuideVote(ctx context.Context, arg db.UpsertStudyGuideVoteParams) error {
	return r.queries.UpsertStudyGuideVote(ctx, arg)
}

func (r *sqlcRepository) ComputeGuideVoteScore(ctx context.Context, studyGuideID pgtype.UUID) (int64, error) {
	return r.queries.ComputeGuideVoteScore(ctx, studyGuideID)
}

func (r *sqlcRepository) DeleteStudyGuideVote(ctx context.Context, arg db.DeleteStudyGuideVoteParams) (int64, error) {
	return r.queries.DeleteStudyGuideVote(ctx, arg)
}

func (r *sqlcRepository) ViewerCanRecommendForGuide(ctx context.Context, arg db.ViewerCanRecommendForGuideParams) (db.ViewerCanRecommendForGuideRow, error) {
	return r.queries.ViewerCanRecommendForGuide(ctx, arg)
}

func (r *sqlcRepository) InsertStudyGuideRecommendation(ctx context.Context, arg db.InsertStudyGuideRecommendationParams) (db.InsertStudyGuideRecommendationRow, error) {
	return r.queries.InsertStudyGuideRecommendation(ctx, arg)
}

func (r *sqlcRepository) DeleteStudyGuideRecommendation(ctx context.Context, arg db.DeleteStudyGuideRecommendationParams) (int64, error) {
	return r.queries.DeleteStudyGuideRecommendation(ctx, arg)
}

func (r *sqlcRepository) URLAlreadyAttachedToGuide(ctx context.Context, arg db.URLAlreadyAttachedToGuideParams) (bool, error) {
	return r.queries.URLAlreadyAttachedToGuide(ctx, arg)
}

func (r *sqlcRepository) UpsertResource(ctx context.Context, arg db.UpsertResourceParams) error {
	return r.queries.UpsertResource(ctx, arg)
}

func (r *sqlcRepository) GetResourceByCreatorURL(ctx context.Context, arg db.GetResourceByCreatorURLParams) (db.GetResourceByCreatorURLRow, error) {
	return r.queries.GetResourceByCreatorURL(ctx, arg)
}

func (r *sqlcRepository) InsertGuideResource(ctx context.Context, arg db.InsertGuideResourceParams) error {
	return r.queries.InsertGuideResource(ctx, arg)
}

func (r *sqlcRepository) GetGuideResourceAttacher(ctx context.Context, arg db.GetGuideResourceAttacherParams) (pgtype.UUID, error) {
	return r.queries.GetGuideResourceAttacher(ctx, arg)
}

func (r *sqlcRepository) DeleteGuideResource(ctx context.Context, arg db.DeleteGuideResourceParams) (int64, error) {
	return r.queries.DeleteGuideResource(ctx, arg)
}

func (r *sqlcRepository) GetFileForAttach(ctx context.Context, id pgtype.UUID) (db.GetFileForAttachRow, error) {
	return r.queries.GetFileForAttach(ctx, id)
}

func (r *sqlcRepository) InsertGuideFile(ctx context.Context, arg db.InsertGuideFileParams) (pgtype.Timestamptz, error) {
	return r.queries.InsertGuideFile(ctx, arg)
}

func (r *sqlcRepository) GuideFileAttached(ctx context.Context, arg db.GuideFileAttachedParams) (bool, error) {
	return r.queries.GuideFileAttached(ctx, arg)
}

func (r *sqlcRepository) DeleteGuideFile(ctx context.Context, arg db.DeleteGuideFileParams) (int64, error) {
	return r.queries.DeleteGuideFile(ctx, arg)
}

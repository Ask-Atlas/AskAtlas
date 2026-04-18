package studyguides

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// sqlcRepository is the production Repository implementation backed by
// sqlc-generated queries. Each method is a thin pass-through; service
// logic (validation, sort dispatch, cursor encoding) lives in service.go.
type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated
// Postgres queries.
func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
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

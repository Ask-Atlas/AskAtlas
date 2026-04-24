package refs

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated Postgres queries.
func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) ListStudyGuideRefSummaries(ctx context.Context, arg db.ListStudyGuideRefSummariesParams) ([]db.ListStudyGuideRefSummariesRow, error) {
	return r.queries.ListStudyGuideRefSummaries(ctx, arg)
}

func (r *sqlcRepository) ListQuizRefSummaries(ctx context.Context, ids []pgtype.UUID) ([]db.ListQuizRefSummariesRow, error) {
	return r.queries.ListQuizRefSummaries(ctx, ids)
}

func (r *sqlcRepository) ListFileRefSummaries(ctx context.Context, arg db.ListFileRefSummariesParams) ([]db.ListFileRefSummariesRow, error) {
	return r.queries.ListFileRefSummaries(ctx, arg)
}

func (r *sqlcRepository) ListCourseRefSummaries(ctx context.Context, ids []pgtype.UUID) ([]db.ListCourseRefSummariesRow, error) {
	return r.queries.ListCourseRefSummaries(ctx, ids)
}

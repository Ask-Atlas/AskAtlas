package favorites

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
)

// Repository is the data-access surface required by the favorites
// service. Three independent list calls keyed off the viewer; the
// service merges them in-process and applies offset/limit windowing.
// Defined here (where it is used) rather than alongside db.Queries
// to keep the surface small and to allow mockery-generated mocks
// for service tests.
type Repository interface {
	ListFileFavorites(ctx context.Context, arg db.ListFileFavoritesParams) ([]db.ListFileFavoritesRow, error)
	ListStudyGuideFavorites(ctx context.Context, arg db.ListStudyGuideFavoritesParams) ([]db.ListStudyGuideFavoritesRow, error)
	ListCourseFavorites(ctx context.Context, arg db.ListCourseFavoritesParams) ([]db.ListCourseFavoritesRow, error)
}

type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated Postgres queries.
func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) ListFileFavorites(ctx context.Context, arg db.ListFileFavoritesParams) ([]db.ListFileFavoritesRow, error) {
	return r.queries.ListFileFavorites(ctx, arg)
}

func (r *sqlcRepository) ListStudyGuideFavorites(ctx context.Context, arg db.ListStudyGuideFavoritesParams) ([]db.ListStudyGuideFavoritesRow, error) {
	return r.queries.ListStudyGuideFavorites(ctx, arg)
}

func (r *sqlcRepository) ListCourseFavorites(ctx context.Context, arg db.ListCourseFavoritesParams) ([]db.ListCourseFavoritesRow, error) {
	return r.queries.ListCourseFavorites(ctx, arg)
}

package schools

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
)

type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated Postgres queries.
func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) ListSchools(ctx context.Context, arg db.ListSchoolsParams) ([]db.School, error) {
	return r.queries.ListSchools(ctx, arg)
}

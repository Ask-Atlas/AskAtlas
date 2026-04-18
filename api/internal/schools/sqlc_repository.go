package schools

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *sqlcRepository) GetSchool(ctx context.Context, id pgtype.UUID) (db.School, error) {
	row, err := r.queries.GetSchool(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.School{}, fmt.Errorf("GetSchool: %w", apperrors.ErrNotFound)
		}
		return db.School{}, fmt.Errorf("GetSchool: %w", err)
	}
	return row, nil
}

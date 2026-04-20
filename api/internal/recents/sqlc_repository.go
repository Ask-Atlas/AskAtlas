package recents

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
)

// Repository is the data-access surface required by the recents
// service. Three independent list calls keyed off the viewer; the
// service merges them in-process. Defined here (where it is used)
// rather than alongside db.Queries to keep the surface small and
// to allow mockery-generated mocks for service tests.
type Repository interface {
	ListRecentFiles(ctx context.Context, arg db.ListRecentFilesParams) ([]db.ListRecentFilesRow, error)
	ListRecentStudyGuides(ctx context.Context, arg db.ListRecentStudyGuidesParams) ([]db.ListRecentStudyGuidesRow, error)
	ListRecentCourses(ctx context.Context, arg db.ListRecentCoursesParams) ([]db.ListRecentCoursesRow, error)
}

type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository returns a Repository backed by sqlc-generated Postgres queries.
func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) ListRecentFiles(ctx context.Context, arg db.ListRecentFilesParams) ([]db.ListRecentFilesRow, error) {
	return r.queries.ListRecentFiles(ctx, arg)
}

func (r *sqlcRepository) ListRecentStudyGuides(ctx context.Context, arg db.ListRecentStudyGuidesParams) ([]db.ListRecentStudyGuidesRow, error) {
	return r.queries.ListRecentStudyGuides(ctx, arg)
}

func (r *sqlcRepository) ListRecentCourses(ctx context.Context, arg db.ListRecentCoursesParams) ([]db.ListRecentCoursesRow, error) {
	return r.queries.ListRecentCourses(ctx, arg)
}

package courses

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

func (r *sqlcRepository) ListCoursesDepartmentAsc(ctx context.Context, arg db.ListCoursesDepartmentAscParams) ([]db.ListCoursesDepartmentAscRow, error) {
	return r.queries.ListCoursesDepartmentAsc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesDepartmentDesc(ctx context.Context, arg db.ListCoursesDepartmentDescParams) ([]db.ListCoursesDepartmentDescRow, error) {
	return r.queries.ListCoursesDepartmentDesc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesNumberAsc(ctx context.Context, arg db.ListCoursesNumberAscParams) ([]db.ListCoursesNumberAscRow, error) {
	return r.queries.ListCoursesNumberAsc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesNumberDesc(ctx context.Context, arg db.ListCoursesNumberDescParams) ([]db.ListCoursesNumberDescRow, error) {
	return r.queries.ListCoursesNumberDesc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesTitleAsc(ctx context.Context, arg db.ListCoursesTitleAscParams) ([]db.ListCoursesTitleAscRow, error) {
	return r.queries.ListCoursesTitleAsc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesTitleDesc(ctx context.Context, arg db.ListCoursesTitleDescParams) ([]db.ListCoursesTitleDescRow, error) {
	return r.queries.ListCoursesTitleDesc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesCreatedAtAsc(ctx context.Context, arg db.ListCoursesCreatedAtAscParams) ([]db.ListCoursesCreatedAtAscRow, error) {
	return r.queries.ListCoursesCreatedAtAsc(ctx, arg)
}

func (r *sqlcRepository) ListCoursesCreatedAtDesc(ctx context.Context, arg db.ListCoursesCreatedAtDescParams) ([]db.ListCoursesCreatedAtDescRow, error) {
	return r.queries.ListCoursesCreatedAtDesc(ctx, arg)
}

// GetCourse maps pgx/sql.ErrNoRows to apperrors.ErrNotFound, matching the
// pattern in internal/files/sqlc_repository.go and internal/schools.
func (r *sqlcRepository) GetCourse(ctx context.Context, id pgtype.UUID) (db.GetCourseRow, error) {
	row, err := r.queries.GetCourse(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetCourseRow{}, fmt.Errorf("GetCourse: %w", apperrors.ErrNotFound)
		}
		return db.GetCourseRow{}, fmt.Errorf("GetCourse: %w", err)
	}
	return row, nil
}

func (r *sqlcRepository) ListCourseSections(ctx context.Context, courseID pgtype.UUID) ([]db.ListCourseSectionsRow, error) {
	return r.queries.ListCourseSections(ctx, courseID)
}

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

func (r *sqlcRepository) CourseExists(ctx context.Context, id pgtype.UUID) (bool, error) {
	return r.queries.CourseExists(ctx, id)
}

func (r *sqlcRepository) SectionInCourseExists(ctx context.Context, arg db.SectionInCourseExistsParams) (bool, error) {
	return r.queries.SectionInCourseExists(ctx, arg)
}

// JoinSection forwards to the sqlc-generated INSERT … ON CONFLICT DO NOTHING
// RETURNING. The query returns sql.ErrNoRows when the row already exists;
// the service treats that as the 409 already-a-member case, so this method
// passes the error through unchanged rather than mapping it to ErrConflict.
// The two callers (Service.JoinSection and tests) get to decide.
func (r *sqlcRepository) JoinSection(ctx context.Context, arg db.JoinSectionParams) (db.CourseMember, error) {
	return r.queries.JoinSection(ctx, arg)
}

// LeaveSection forwards to the sqlc-generated DELETE … RETURNING. The query
// returns sql.ErrNoRows when the membership row does not exist; the service
// translates that to the 404 not-a-member case.
func (r *sqlcRepository) LeaveSection(ctx context.Context, arg db.LeaveSectionParams) (pgtype.UUID, error) {
	return r.queries.LeaveSection(ctx, arg)
}

func (r *sqlcRepository) ListMyEnrollments(ctx context.Context, arg db.ListMyEnrollmentsParams) ([]db.ListMyEnrollmentsRow, error) {
	return r.queries.ListMyEnrollments(ctx, arg)
}

// GetMembership forwards to the sqlc-generated lookup. Returns sql.ErrNoRows
// when the user is not a member; the service maps that to {enrolled: false}
// rather than treating it as an error.
func (r *sqlcRepository) GetMembership(ctx context.Context, arg db.GetMembershipParams) (db.GetMembershipRow, error) {
	return r.queries.GetMembership(ctx, arg)
}

func (r *sqlcRepository) ListSectionMembers(ctx context.Context, arg db.ListSectionMembersParams) ([]db.ListSectionMembersRow, error) {
	return r.queries.ListSectionMembers(ctx, arg)
}

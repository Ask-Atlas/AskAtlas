package favorites

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository is the data-access surface required by the favorites
// service. The three List* calls power GET /api/me/favorites; the
// six new Check* + Toggle* calls power the per-entity favorite-toggle
// endpoints (ASK-130 / ASK-156 / ASK-157).
//
// Defined here (where it is used) rather than alongside db.Queries
// to keep the surface small and to allow mockery-generated mocks
// for service tests.
type Repository interface {
	ListFileFavorites(ctx context.Context, arg db.ListFileFavoritesParams) ([]db.ListFileFavoritesRow, error)
	ListStudyGuideFavorites(ctx context.Context, arg db.ListStudyGuideFavoritesParams) ([]db.ListStudyGuideFavoritesRow, error)
	ListCourseFavorites(ctx context.Context, arg db.ListCourseFavoritesParams) ([]db.ListCourseFavoritesRow, error)

	// CheckFileExists / CheckStudyGuideExists / CheckCourseExists
	// return apperrors.ErrNotFound when the parent entity is missing
	// or in a deletion lifecycle (per the spec, both map to 404).
	CheckFileExists(ctx context.Context, fileID pgtype.UUID) error
	CheckStudyGuideExists(ctx context.Context, studyGuideID pgtype.UUID) error
	CheckCourseExists(ctx context.Context, courseID pgtype.UUID) error

	// Toggle* atomically inserts when missing / deletes when present
	// in a single CTE round trip. The caller is expected to gate
	// existence first via the matching Check* probe.
	ToggleFileFavorite(ctx context.Context, arg db.ToggleFileFavoriteParams) (db.ToggleFileFavoriteRow, error)
	ToggleStudyGuideFavorite(ctx context.Context, arg db.ToggleStudyGuideFavoriteParams) (db.ToggleStudyGuideFavoriteRow, error)
	ToggleCourseFavorite(ctx context.Context, arg db.ToggleCourseFavoriteParams) (db.ToggleCourseFavoriteRow, error)
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

// checkExists is the shared sql.ErrNoRows -> apperrors.ErrNotFound
// translation used by all three Check* probes. The integer column the
// query returns ("SELECT 1") is discarded; we only care about the
// no-rows signal.
func checkExists(_ int32, err error, label string) error {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", label, apperrors.ErrNotFound)
		}
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}

func (r *sqlcRepository) CheckFileExists(ctx context.Context, fileID pgtype.UUID) error {
	val, err := r.queries.CheckFileExists(ctx, fileID)
	return checkExists(val, err, "CheckFileExists")
}

func (r *sqlcRepository) CheckStudyGuideExists(ctx context.Context, studyGuideID pgtype.UUID) error {
	val, err := r.queries.CheckStudyGuideExists(ctx, studyGuideID)
	return checkExists(val, err, "CheckStudyGuideExists")
}

func (r *sqlcRepository) CheckCourseExists(ctx context.Context, courseID pgtype.UUID) error {
	val, err := r.queries.CheckCourseExists(ctx, courseID)
	return checkExists(val, err, "CheckCourseExists")
}

func (r *sqlcRepository) ToggleFileFavorite(ctx context.Context, arg db.ToggleFileFavoriteParams) (db.ToggleFileFavoriteRow, error) {
	row, err := r.queries.ToggleFileFavorite(ctx, arg)
	if err != nil {
		return db.ToggleFileFavoriteRow{}, fmt.Errorf("ToggleFileFavorite: %w", err)
	}
	return row, nil
}

func (r *sqlcRepository) ToggleStudyGuideFavorite(ctx context.Context, arg db.ToggleStudyGuideFavoriteParams) (db.ToggleStudyGuideFavoriteRow, error) {
	row, err := r.queries.ToggleStudyGuideFavorite(ctx, arg)
	if err != nil {
		return db.ToggleStudyGuideFavoriteRow{}, fmt.Errorf("ToggleStudyGuideFavorite: %w", err)
	}
	return row, nil
}

func (r *sqlcRepository) ToggleCourseFavorite(ctx context.Context, arg db.ToggleCourseFavoriteParams) (db.ToggleCourseFavoriteRow, error) {
	row, err := r.queries.ToggleCourseFavorite(ctx, arg)
	if err != nil {
		return db.ToggleCourseFavoriteRow{}, fmt.Errorf("ToggleCourseFavorite: %w", err)
	}
	return row, nil
}

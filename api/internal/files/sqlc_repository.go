package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

type sqlcRepository struct {
	queries *db.Queries
}

func NewSQLCRepository(queries *db.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) GetFileIfViewable(ctx context.Context, arg db.GetFileIfViewableParams) (db.File, error) {
	slog.Debug("getting file if viewable", "file_id", arg.FileID, "viewer_id", arg.ViewerID)

	file, err := r.queries.GetFileIfViewable(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.File{}, fmt.Errorf("GetFileIfViewable: %w", apperrors.ErrNotFound)
		}
		return db.File{}, fmt.Errorf("GetFileIfViewable: %w", err)
	}

	return file, nil
}

func (r *sqlcRepository) ListOwnedFilesUpdatedDesc(ctx context.Context, arg db.ListOwnedFilesUpdatedDescParams) ([]db.ListOwnedFilesUpdatedDescRow, error) {
	slog.Debug("listing owned files updated desc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesUpdatedDesc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesUpdatedDesc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesUpdatedAsc(ctx context.Context, arg db.ListOwnedFilesUpdatedAscParams) ([]db.ListOwnedFilesUpdatedAscRow, error) {
	slog.Debug("listing owned files updated asc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesUpdatedAsc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesUpdatedAsc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesCreatedDesc(ctx context.Context, arg db.ListOwnedFilesCreatedDescParams) ([]db.ListOwnedFilesCreatedDescRow, error) {
	slog.Debug("listing owned files created desc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesCreatedDesc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesCreatedDesc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesCreatedAsc(ctx context.Context, arg db.ListOwnedFilesCreatedAscParams) ([]db.ListOwnedFilesCreatedAscRow, error) {
	slog.Debug("listing owned files created asc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesCreatedAsc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesCreatedAsc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesNameAsc(ctx context.Context, arg db.ListOwnedFilesNameAscParams) ([]db.ListOwnedFilesNameAscRow, error) {
	slog.Debug("listing owned files name asc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesNameAsc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesNameAsc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesNameDesc(ctx context.Context, arg db.ListOwnedFilesNameDescParams) ([]db.ListOwnedFilesNameDescRow, error) {
	slog.Debug("listing owned files name desc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesNameDesc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesNameDesc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesSizeAsc(ctx context.Context, arg db.ListOwnedFilesSizeAscParams) ([]db.ListOwnedFilesSizeAscRow, error) {
	slog.Debug("listing owned files size asc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesSizeAsc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesSizeAsc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesSizeDesc(ctx context.Context, arg db.ListOwnedFilesSizeDescParams) ([]db.ListOwnedFilesSizeDescRow, error) {
	slog.Debug("listing owned files size desc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesSizeDesc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesSizeDesc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesStatusAsc(ctx context.Context, arg db.ListOwnedFilesStatusAscParams) ([]db.ListOwnedFilesStatusAscRow, error) {
	slog.Debug("listing owned files status asc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesStatusAsc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesStatusAsc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesStatusDesc(ctx context.Context, arg db.ListOwnedFilesStatusDescParams) ([]db.ListOwnedFilesStatusDescRow, error) {
	slog.Debug("listing owned files status desc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesStatusDesc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesStatusDesc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesMimeAsc(ctx context.Context, arg db.ListOwnedFilesMimeAscParams) ([]db.ListOwnedFilesMimeAscRow, error) {
	slog.Debug("listing owned files mime asc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesMimeAsc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesMimeAsc: %w", err)
	}

	return files, nil
}

func (r *sqlcRepository) ListOwnedFilesMimeDesc(ctx context.Context, arg db.ListOwnedFilesMimeDescParams) ([]db.ListOwnedFilesMimeDescRow, error) {
	slog.Debug("listing owned files mime desc", "owner_id", arg.OwnerID)

	files, err := r.queries.ListOwnedFilesMimeDesc(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("ListOwnedFilesMimeDesc: %w", err)
	}

	return files, nil
}

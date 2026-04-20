package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sqlcRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewSQLCRepository creates a postgres-backed db Repository instance.
func NewSQLCRepository(pool *pgxpool.Pool, queries *db.Queries) Repository {
	return &sqlcRepository{pool: pool, queries: queries}
}

func (r *sqlcRepository) InTx(ctx context.Context, fn func(Repository) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("InTx: begin tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", rollbackErr)
		}
	}()

	txRepo := &sqlcRepository{pool: r.pool, queries: r.queries.WithTx(tx)}
	if err := fn(txRepo); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("InTx: commit: %w", err)
	}

	return nil
}

func (r *sqlcRepository) InsertFile(ctx context.Context, arg db.InsertFileParams) (db.File, error) {
	slog.Debug("inserting file", "user_id", arg.UserID, "name", arg.Name)
	file, err := r.queries.InsertFile(ctx, arg)
	if err != nil {
		return db.File{}, fmt.Errorf("InsertFile: %w", err)
	}
	return file, nil
}

func (r *sqlcRepository) UpdateFileStatus(ctx context.Context, arg db.UpdateFileStatusParams) error {
	slog.Debug("updating file status", "file_id", arg.FileID, "status", arg.Status)
	if err := r.queries.UpdateFileStatus(ctx, arg); err != nil {
		return fmt.Errorf("UpdateFileStatus: %w", err)
	}
	return nil
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

func (r *sqlcRepository) UpsertFileGrant(ctx context.Context, arg db.UpsertFileGrantParams) (db.FileGrant, error) {
	slog.Debug("upserting file grant", "file_id", arg.FileID, "grantee_type", arg.GranteeType, "grantee_id", arg.GranteeID, "permission", arg.Permission)
	row, err := r.queries.UpsertFileGrant(ctx, arg)
	if err != nil {
		return db.FileGrant{}, fmt.Errorf("UpsertFileGrant: %w", err)
	}
	return row, nil
}

func (r *sqlcRepository) RevokeFileGrant(ctx context.Context, arg db.RevokeFileGrantParams) error {
	slog.Debug("revoking file grant", "file_id", arg.FileID, "grantee_type", arg.GranteeType, "grantee_id", arg.GranteeID, "permission", arg.Permission)
	if err := r.queries.RevokeFileGrant(ctx, arg); err != nil {
		return fmt.Errorf("RevokeFileGrant: %w", err)
	}
	return nil
}

func (r *sqlcRepository) GetFileByOwner(ctx context.Context, arg db.GetFileByOwnerParams) (db.GetFileByOwnerRow, error) {
	slog.Debug("getting file by owner", "file_id", arg.FileID, "owner_id", arg.OwnerID)
	file, err := r.queries.GetFileByOwner(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetFileByOwnerRow{}, fmt.Errorf("GetFileByOwner: %w", apperrors.ErrNotFound)
		}
		return db.GetFileByOwnerRow{}, fmt.Errorf("GetFileByOwner: %w", err)
	}
	return file, nil
}

func (r *sqlcRepository) SoftDeleteFile(ctx context.Context, arg db.SoftDeleteFileParams) (int64, error) {
	slog.Debug("soft deleting file", "file_id", arg.FileID, "owner_id", arg.OwnerID)
	rows, err := r.queries.SoftDeleteFile(ctx, arg)
	if err != nil {
		return 0, fmt.Errorf("SoftDeleteFile: %w", err)
	}
	return rows, nil
}

func (r *sqlcRepository) SetFileDeletionJobID(ctx context.Context, arg db.SetFileDeletionJobIDParams) error {
	slog.Debug("setting deletion job id", "file_id", arg.FileID)
	if err := r.queries.SetFileDeletionJobID(ctx, arg); err != nil {
		return fmt.Errorf("SetFileDeletionJobID: %w", err)
	}
	return nil
}

func (r *sqlcRepository) GetFileForUpdate(ctx context.Context, fileID pgtype.UUID) (db.GetFileForUpdateRow, error) {
	slog.Debug("getting file for update", "file_id", fileID)
	row, err := r.queries.GetFileForUpdate(ctx, fileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetFileForUpdateRow{}, fmt.Errorf("GetFileForUpdate: %w", apperrors.ErrNotFound)
		}
		return db.GetFileForUpdateRow{}, fmt.Errorf("GetFileForUpdate: %w", err)
	}
	return row, nil
}

func (r *sqlcRepository) PatchFile(ctx context.Context, arg db.PatchFileParams) (db.PatchFileRow, error) {
	slog.Debug("patching file", "file_id", arg.FileID, "owner_id", arg.OwnerID)
	row, err := r.queries.PatchFile(ctx, arg)
	if err != nil {
		// A concurrent DELETE between the GetFileForUpdate probe and
		// this UPDATE drops the WHERE clause to zero rows, so the CTE
		// scan returns sql.ErrNoRows -- map to 404 to match the spec.
		if errors.Is(err, sql.ErrNoRows) {
			return db.PatchFileRow{}, fmt.Errorf("PatchFile: %w", apperrors.ErrNotFound)
		}
		return db.PatchFileRow{}, fmt.Errorf("PatchFile: %w", err)
	}
	return row, nil
}

func (r *sqlcRepository) InsertFileView(ctx context.Context, arg db.InsertFileViewParams) error {
	if err := r.queries.InsertFileView(ctx, arg); err != nil {
		return fmt.Errorf("InsertFileView: %w", err)
	}
	return nil
}

func (r *sqlcRepository) UpsertFileLastViewed(ctx context.Context, arg db.UpsertFileLastViewedParams) error {
	if err := r.queries.UpsertFileLastViewed(ctx, arg); err != nil {
		return fmt.Errorf("UpsertFileLastViewed: %w", err)
	}
	return nil
}

func (r *sqlcRepository) MarkFileDeleted(ctx context.Context, fileID pgtype.UUID) error {
	slog.Debug("marking file deleted", "file_id", fileID)
	if err := r.queries.MarkFileDeleted(ctx, fileID); err != nil {
		return fmt.Errorf("MarkFileDeleted: %w", err)
	}
	return nil
}

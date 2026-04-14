// Package files contains the business logic, models, and data access layer for managing uploaded files.
package files

import (
	"fmt"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

// sharedRow holds the common fields present in every sqlc-generated
// ListOwnedFiles*Row type.
type sharedRow struct {
	ID           pgtype.UUID
	UserID       pgtype.UUID
	Name         string
	Size         int64
	MimeType     db.MimeType
	Status       db.UploadStatus
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	FavoritedAt  pgtype.Timestamptz
	LastViewedAt pgtype.Timestamptz
}

// mapListRow converts a database sharedRow into the domain File standard model.
func mapListRow(r sharedRow) (File, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return File{}, fmt.Errorf("mapListRow: ID: %w", err)
	}
	userID, err := utils.PgxToGoogleUUID(r.UserID)
	if err != nil {
		return File{}, fmt.Errorf("mapListRow: UserID: %w", err)
	}

	return File{
		ID:           id,
		UserID:       userID,
		Name:         r.Name,
		Size:         r.Size,
		MimeType:     string(r.MimeType),
		Status:       string(r.Status),
		CreatedAt:    r.CreatedAt.Time,
		UpdatedAt:    r.UpdatedAt.Time,
		FavoritedAt:  utils.TimestamptzPtr(r.FavoritedAt),
		LastViewedAt: utils.TimestamptzPtr(r.LastViewedAt),
	}, nil
}

// mapListRows converts a slice of database sharedRows into a slice of domain File models.
func mapListRows(rows []sharedRow) ([]File, error) {
	out := make([]File, 0, len(rows))
	for _, r := range rows {
		f, err := mapListRow(r)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

// toShared converts any of the twelve sqlc ListOwnedFiles*Row types to sharedRow.
// All variants have identical fields — only the type name differs.
func toShared[R any](
	rows []R,
	fn func(R) sharedRow,
) []sharedRow {
	out := make([]sharedRow, len(rows))
	for i, r := range rows {
		out[i] = fn(r)
	}
	return out
}

func sharedFromUpdatedDesc(r db.ListOwnedFilesUpdatedDescRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromUpdatedAsc(r db.ListOwnedFilesUpdatedAscRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromCreatedDesc(r db.ListOwnedFilesCreatedDescRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromCreatedAsc(r db.ListOwnedFilesCreatedAscRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromNameAsc(r db.ListOwnedFilesNameAscRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromNameDesc(r db.ListOwnedFilesNameDescRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromSizeAsc(r db.ListOwnedFilesSizeAscRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromSizeDesc(r db.ListOwnedFilesSizeDescRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromStatusAsc(r db.ListOwnedFilesStatusAscRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromStatusDesc(r db.ListOwnedFilesStatusDescRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromMimeAsc(r db.ListOwnedFilesMimeAscRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}
func sharedFromMimeDesc(r db.ListOwnedFilesMimeDescRow) sharedRow {
	return sharedRow{r.ID, r.UserID, r.Name, r.Size, r.MimeType, r.Status, r.CreatedAt, r.UpdatedAt, r.FavoritedAt, r.LastViewedAt}
}

// mapGrantRow converts a database UpsertFileGrantRow into the domain Grant model.
func mapGrantRow(r db.UpsertFileGrantRow) (Grant, error) {
	id, err := utils.PgxToGoogleUUID(r.ID)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrantRow: ID: %w", err)
	}
	fileID, err := utils.PgxToGoogleUUID(r.FileID)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrantRow: FileID: %w", err)
	}
	granteeID, err := utils.PgxToGoogleUUID(r.GranteeID)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrantRow: GranteeID: %w", err)
	}
	grantedBy, err := utils.PgxToGoogleUUID(r.GrantedBy)
	if err != nil {
		return Grant{}, fmt.Errorf("mapGrantRow: GrantedBy: %w", err)
	}

	return Grant{
		ID:          id,
		FileID:      fileID,
		GranteeType: string(r.GranteeType),
		GranteeID:   granteeID,
		Permission:  string(r.Permission),
		GrantedBy:   grantedBy,
		CreatedAt:   r.CreatedAt.Time,
	}, nil
}

// mapDBFile converts a database File model into the domain File standard model.
func mapDBFile(f db.File) (File, error) {
	id, err := utils.PgxToGoogleUUID(f.ID)
	if err != nil {
		return File{}, fmt.Errorf("mapDBFile: ID: %w", err)
	}
	userID, err := utils.PgxToGoogleUUID(f.UserID)
	if err != nil {
		return File{}, fmt.Errorf("mapDBFile: UserID: %w", err)
	}

	return File{
		ID:        id,
		UserID:    userID,
		Name:      f.Name,
		Size:      f.Size,
		MimeType:  string(f.MimeType),
		Status:    string(f.Status),
		CreatedAt: f.CreatedAt.Time,
		UpdatedAt: f.UpdatedAt.Time,
	}, nil
}

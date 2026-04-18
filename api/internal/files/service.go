package files

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository defines the data-access interface required by the files Service.
type Repository interface {
	InTx(ctx context.Context, fn func(Repository) error) error

	GetFileIfViewable(ctx context.Context, arg db.GetFileIfViewableParams) (db.File, error)
	GetFileByOwner(ctx context.Context, arg db.GetFileByOwnerParams) (db.GetFileByOwnerRow, error)
	SoftDeleteFile(ctx context.Context, arg db.SoftDeleteFileParams) (int64, error)
	SetFileDeletionJobID(ctx context.Context, arg db.SetFileDeletionJobIDParams) error
	MarkFileDeleted(ctx context.Context, fileID pgtype.UUID) error
	UpdateFile(ctx context.Context, arg db.UpdateFileParams) (db.UpdateFileRow, error)

	ListOwnedFilesUpdatedDesc(ctx context.Context, arg db.ListOwnedFilesUpdatedDescParams) ([]db.ListOwnedFilesUpdatedDescRow, error)
	ListOwnedFilesUpdatedAsc(ctx context.Context, arg db.ListOwnedFilesUpdatedAscParams) ([]db.ListOwnedFilesUpdatedAscRow, error)
	ListOwnedFilesCreatedDesc(ctx context.Context, arg db.ListOwnedFilesCreatedDescParams) ([]db.ListOwnedFilesCreatedDescRow, error)
	ListOwnedFilesCreatedAsc(ctx context.Context, arg db.ListOwnedFilesCreatedAscParams) ([]db.ListOwnedFilesCreatedAscRow, error)
	ListOwnedFilesNameAsc(ctx context.Context, arg db.ListOwnedFilesNameAscParams) ([]db.ListOwnedFilesNameAscRow, error)
	ListOwnedFilesNameDesc(ctx context.Context, arg db.ListOwnedFilesNameDescParams) ([]db.ListOwnedFilesNameDescRow, error)
	ListOwnedFilesSizeAsc(ctx context.Context, arg db.ListOwnedFilesSizeAscParams) ([]db.ListOwnedFilesSizeAscRow, error)
	ListOwnedFilesSizeDesc(ctx context.Context, arg db.ListOwnedFilesSizeDescParams) ([]db.ListOwnedFilesSizeDescRow, error)
	ListOwnedFilesStatusAsc(ctx context.Context, arg db.ListOwnedFilesStatusAscParams) ([]db.ListOwnedFilesStatusAscRow, error)
	ListOwnedFilesStatusDesc(ctx context.Context, arg db.ListOwnedFilesStatusDescParams) ([]db.ListOwnedFilesStatusDescRow, error)
	ListOwnedFilesMimeAsc(ctx context.Context, arg db.ListOwnedFilesMimeAscParams) ([]db.ListOwnedFilesMimeAscRow, error)
	ListOwnedFilesMimeDesc(ctx context.Context, arg db.ListOwnedFilesMimeDescParams) ([]db.ListOwnedFilesMimeDescRow, error)
}

// Service is the business-logic layer for the files feature.
type Service struct {
	repo       Repository
	queryTable map[sortKey]queryFn
}

// NewService creates a new Service instance configured with the given repository.
func NewService(repo Repository) *Service {
	s := &Service{repo: repo}
	s.queryTable = map[sortKey]queryFn{
		{SortFieldUpdatedAt, SortDirDesc}: s.queryUpdatedDesc,
		{SortFieldUpdatedAt, SortDirAsc}:  s.queryUpdatedAsc,
		{SortFieldCreatedAt, SortDirDesc}: s.queryCreatedDesc,
		{SortFieldCreatedAt, SortDirAsc}:  s.queryCreatedAsc,
		{SortFieldName, SortDirAsc}:       s.queryNameAsc,
		{SortFieldName, SortDirDesc}:      s.queryNameDesc,
		{SortFieldSize, SortDirAsc}:       s.querySizeAsc,
		{SortFieldSize, SortDirDesc}:      s.querySizeDesc,
		{SortFieldStatus, SortDirAsc}:     s.queryStatusAsc,
		{SortFieldStatus, SortDirDesc}:    s.queryStatusDesc,
		{SortFieldMimeType, SortDirAsc}:   s.queryMimeAsc,
		{SortFieldMimeType, SortDirDesc}:  s.queryMimeDesc,
	}
	return s
}

// GetFile retrieves a single file, verifying that the requesting user has access to it.
func (s *Service) GetFile(ctx context.Context, p GetFileParams) (File, error) {
	row, err := s.repo.GetFileIfViewable(ctx, db.GetFileIfViewableParams{
		FileID:        utils.UUID(p.FileID),
		ViewerID:      utils.UUID(p.ViewerID),
		CourseIds:     uuidsToPgtype(p.CourseIDs),
		StudyGuideIds: uuidsToPgtype(p.StudyGuideIDs),
	})
	if err != nil {
		return File{}, err
	}
	return mapDBFile(row)
}

// ListFiles queries the repository for a paginated list of files matching the given parameters.
func (s *Service) ListFiles(ctx context.Context, p ListFilesParams) ([]File, *string, error) {
	if p.Scope != ScopeOwned {
		return nil, nil, fmt.Errorf("ListFiles: scope %q is not yet implemented", p.Scope)
	}

	limit := int32(p.PageLimit + 1) // +1 to detect next page without a COUNT

	files, err := s.dispatchListQuery(ctx, p, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("ListFiles: %w", err)
	}

	hasMore := len(files) > p.PageLimit
	if hasMore {
		files = files[:p.PageLimit]
	}

	var nextCursor *string
	if hasMore && len(files) > 0 {
		encoded, err := EncodeCursor(buildCursor(files[len(files)-1], p.SortField))
		if err != nil {
			return nil, nil, fmt.Errorf("ListFiles: encode cursor: %w", err)
		}
		nextCursor = &encoded
	}

	return files, nextCursor, nil
}

// UpdateFile renames a file after validating the new name. The name is trimmed
// of leading/trailing whitespace before validation. Returns apperrors.ErrNotFound
// if the file does not belong to the caller or is in a deletion state.
func (s *Service) UpdateFile(ctx context.Context, p UpdateFileParams) (File, error) {
	name := strings.TrimSpace(p.Name)
	if err := validateFileName(name); err != nil {
		return File{}, err
	}

	row, err := s.repo.UpdateFile(ctx, db.UpdateFileParams{
		FileID:  utils.UUID(p.FileID),
		OwnerID: utils.UUID(p.OwnerID),
		Name:    name,
	})
	if err != nil {
		return File{}, err
	}
	return mapUpdateFileRow(row)
}

const maxFileNameLength = 255

// validateFileName checks that a (already-trimmed) file name is non-empty,
// within length limits, and free of dangerous characters.
func validateFileName(name string) *apperrors.AppError {
	details := make(map[string]string)

	if name == "" {
		details["name"] = "must not be empty"
		return apperrors.NewBadRequest("Invalid file name", details)
	}

	if utf8.RuneCountInString(name) > maxFileNameLength {
		details["name"] = fmt.Sprintf("must not exceed %d characters", maxFileNameLength)
		return apperrors.NewBadRequest("Invalid file name", details)
	}

	var invalid []string
	seen := make(map[string]bool)
	for _, r := range name {
		var ch string
		switch {
		case r == '/':
			ch = "/"
		case r == '\\':
			ch = "\\"
		case r == 0:
			ch = "null byte"
		case unicode.IsControl(r):
			ch = "control character"
		default:
			continue
		}
		if !seen[ch] {
			seen[ch] = true
			invalid = append(invalid, ch)
		}
	}
	if len(invalid) > 0 {
		details["name"] = "contains invalid characters: " + strings.Join(invalid, ", ")
		return apperrors.NewBadRequest("Invalid file name", details)
	}

	return nil
}

// DeleteFileParams holds the inputs required to initiate file deletion.
type DeleteFileParams struct {
	FileID  uuid.UUID
	OwnerID uuid.UUID
}

// QStashPublisher is the interface the service uses to publish async jobs.
// Allows the concrete qstashclient.Client to be swapped for a test double.
type QStashPublisher interface {
	PublishDeleteFile(ctx context.Context, msg qstashclient.DeleteFileMessage) (string, error)
}

// DeleteFile soft-deletes the file within a transaction, then publishes an async
// S3 cleanup job via QStash. Returns apperrors.ErrNotFound if the file does not
// belong to the caller or is already in a deletion state.
func (s *Service) DeleteFile(ctx context.Context, p DeleteFileParams, publisher QStashPublisher) error {
	var file db.GetFileByOwnerRow
	if err := s.repo.InTx(ctx, func(txRepo Repository) error {
		var err error
		file, err = txRepo.GetFileByOwner(ctx, db.GetFileByOwnerParams{
			FileID:  utils.UUID(p.FileID),
			OwnerID: utils.UUID(p.OwnerID),
		})
		if err != nil {
			return err
		}

		rows, err := txRepo.SoftDeleteFile(ctx, db.SoftDeleteFileParams{
			FileID:  utils.UUID(p.FileID),
			OwnerID: utils.UUID(p.OwnerID),
		})
		if err != nil {
			return fmt.Errorf("DeleteFile: soft delete: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("DeleteFile: %w", apperrors.ErrNotFound)
		}

		return nil
	}); err != nil {
		return err
	}

	jobID, err := publisher.PublishDeleteFile(ctx, qstashclient.DeleteFileMessage{
		FileID:      p.FileID.String(),
		S3Key:       file.S3Key,
		UserID:      p.OwnerID.String(),
		RequestedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("DeleteFile: publish delete job: %w", err)
	}

	if err := s.repo.SetFileDeletionJobID(ctx, db.SetFileDeletionJobIDParams{
		FileID: utils.UUID(p.FileID),
		JobID:  utils.Text(&jobID),
	}); err != nil {
		slog.Error("DeleteFile: failed to set deletion_job_id", "file_id", p.FileID, "error", err)
	}

	return nil
}

type sortKey struct {
	Field SortField
	Dir   SortDir
}

type queryFn func(ctx context.Context, f dbFilters, limit int32) ([]File, error)

func (s *Service) dispatchListQuery(ctx context.Context, p ListFilesParams, limit int32) ([]File, error) {
	fn, ok := s.queryTable[sortKey{p.SortField, p.SortDir}]
	if !ok {
		return nil, fmt.Errorf("dispatchListQuery: %w: unsupported sort %s/%s",
			apperrors.ErrInvalidInput, p.SortField, p.SortDir)
	}
	return fn(ctx, toDBFilters(p), limit)
}

func (s *Service) queryUpdatedDesc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesUpdatedDesc(ctx, db.ListOwnedFilesUpdatedDescParams{
		ViewerID:        f.ViewerID,
		OwnerID:         f.OwnerID,
		Status:          f.Status,
		MimeType:        f.MimeType,
		MinSize:         f.MinSize,
		MaxSize:         f.MaxSize,
		CreatedFrom:     f.CreatedFrom,
		CreatedTo:       f.CreatedTo,
		UpdatedFrom:     f.UpdatedFrom,
		UpdatedTo:       f.UpdatedTo,
		Q:               f.Q,
		PageLimit:       limit,
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromUpdatedDesc))
}

func (s *Service) queryUpdatedAsc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesUpdatedAsc(ctx, db.ListOwnedFilesUpdatedAscParams{
		ViewerID:        f.ViewerID,
		OwnerID:         f.OwnerID,
		Status:          f.Status,
		MimeType:        f.MimeType,
		MinSize:         f.MinSize,
		MaxSize:         f.MaxSize,
		CreatedFrom:     f.CreatedFrom,
		CreatedTo:       f.CreatedTo,
		UpdatedFrom:     f.UpdatedFrom,
		UpdatedTo:       f.UpdatedTo,
		Q:               f.Q,
		PageLimit:       limit,
		CursorUpdatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.UpdatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromUpdatedAsc))
}

func (s *Service) queryCreatedDesc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesCreatedDesc(ctx, db.ListOwnedFilesCreatedDescParams{
		ViewerID:        f.ViewerID,
		OwnerID:         f.OwnerID,
		Status:          f.Status,
		MimeType:        f.MimeType,
		MinSize:         f.MinSize,
		MaxSize:         f.MaxSize,
		CreatedFrom:     f.CreatedFrom,
		CreatedTo:       f.CreatedTo,
		UpdatedFrom:     f.UpdatedFrom,
		UpdatedTo:       f.UpdatedTo,
		Q:               f.Q,
		PageLimit:       limit,
		CursorCreatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromCreatedDesc))
}

func (s *Service) queryCreatedAsc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesCreatedAsc(ctx, db.ListOwnedFilesCreatedAscParams{
		ViewerID:        f.ViewerID,
		OwnerID:         f.OwnerID,
		Status:          f.Status,
		MimeType:        f.MimeType,
		MinSize:         f.MinSize,
		MaxSize:         f.MaxSize,
		CreatedFrom:     f.CreatedFrom,
		CreatedTo:       f.CreatedTo,
		UpdatedFrom:     f.UpdatedFrom,
		UpdatedTo:       f.UpdatedTo,
		Q:               f.Q,
		PageLimit:       limit,
		CursorCreatedAt: utils.CursorTimestamptz(f.Cursor, func(c *Cursor) *time.Time { return c.CreatedAt }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromCreatedAsc))
}

func (s *Service) queryNameAsc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesNameAsc(ctx, db.ListOwnedFilesNameAscParams{
		ViewerID:        f.ViewerID,
		OwnerID:         f.OwnerID,
		Status:          f.Status,
		MimeType:        f.MimeType,
		MinSize:         f.MinSize,
		MaxSize:         f.MaxSize,
		CreatedFrom:     f.CreatedFrom,
		CreatedTo:       f.CreatedTo,
		UpdatedFrom:     f.UpdatedFrom,
		UpdatedTo:       f.UpdatedTo,
		Q:               f.Q,
		PageLimit:       limit,
		CursorNameLower: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.NameLower }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromNameAsc))
}

func (s *Service) queryNameDesc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesNameDesc(ctx, db.ListOwnedFilesNameDescParams{
		ViewerID:        f.ViewerID,
		OwnerID:         f.OwnerID,
		Status:          f.Status,
		MimeType:        f.MimeType,
		MinSize:         f.MinSize,
		MaxSize:         f.MaxSize,
		CreatedFrom:     f.CreatedFrom,
		CreatedTo:       f.CreatedTo,
		UpdatedFrom:     f.UpdatedFrom,
		UpdatedTo:       f.UpdatedTo,
		Q:               f.Q,
		PageLimit:       limit,
		CursorNameLower: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.NameLower }),
		CursorID:        utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromNameDesc))
}

func (s *Service) querySizeAsc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesSizeAsc(ctx, db.ListOwnedFilesSizeAscParams{
		ViewerID:    f.ViewerID,
		OwnerID:     f.OwnerID,
		Status:      f.Status,
		MimeType:    f.MimeType,
		MinSize:     f.MinSize,
		MaxSize:     f.MaxSize,
		CreatedFrom: f.CreatedFrom,
		CreatedTo:   f.CreatedTo,
		UpdatedFrom: f.UpdatedFrom,
		UpdatedTo:   f.UpdatedTo,
		Q:           f.Q,
		PageLimit:   limit,
		CursorSize:  utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.Size }),
		CursorID:    utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromSizeAsc))
}

func (s *Service) querySizeDesc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesSizeDesc(ctx, db.ListOwnedFilesSizeDescParams{
		ViewerID:    f.ViewerID,
		OwnerID:     f.OwnerID,
		Status:      f.Status,
		MimeType:    f.MimeType,
		MinSize:     f.MinSize,
		MaxSize:     f.MaxSize,
		CreatedFrom: f.CreatedFrom,
		CreatedTo:   f.CreatedTo,
		UpdatedFrom: f.UpdatedFrom,
		UpdatedTo:   f.UpdatedTo,
		Q:           f.Q,
		PageLimit:   limit,
		CursorSize:  utils.CursorInt8(f.Cursor, func(c *Cursor) *int64 { return c.Size }),
		CursorID:    utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromSizeDesc))
}

func (s *Service) queryStatusAsc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesStatusAsc(ctx, db.ListOwnedFilesStatusAscParams{
		ViewerID:     f.ViewerID,
		OwnerID:      f.OwnerID,
		Status:       f.Status,
		MimeType:     f.MimeType,
		MinSize:      f.MinSize,
		MaxSize:      f.MaxSize,
		CreatedFrom:  f.CreatedFrom,
		CreatedTo:    f.CreatedTo,
		UpdatedFrom:  f.UpdatedFrom,
		UpdatedTo:    f.UpdatedTo,
		Q:            f.Q,
		PageLimit:    limit,
		CursorStatus: utils.CursorNullUploadStatus(f.Cursor, func(c *Cursor) *string { return c.Status }),
		CursorID:     utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromStatusAsc))
}

func (s *Service) queryStatusDesc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesStatusDesc(ctx, db.ListOwnedFilesStatusDescParams{
		ViewerID:     f.ViewerID,
		OwnerID:      f.OwnerID,
		Status:       f.Status,
		MimeType:     f.MimeType,
		MinSize:      f.MinSize,
		MaxSize:      f.MaxSize,
		CreatedFrom:  f.CreatedFrom,
		CreatedTo:    f.CreatedTo,
		UpdatedFrom:  f.UpdatedFrom,
		UpdatedTo:    f.UpdatedTo,
		Q:            f.Q,
		PageLimit:    limit,
		CursorStatus: utils.CursorNullUploadStatus(f.Cursor, func(c *Cursor) *string { return c.Status }),
		CursorID:     utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromStatusDesc))
}

func (s *Service) queryMimeAsc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesMimeAsc(ctx, db.ListOwnedFilesMimeAscParams{
		ViewerID:       f.ViewerID,
		OwnerID:        f.OwnerID,
		Status:         f.Status,
		MimeType:       f.MimeType,
		MinSize:        f.MinSize,
		MaxSize:        f.MaxSize,
		CreatedFrom:    f.CreatedFrom,
		CreatedTo:      f.CreatedTo,
		UpdatedFrom:    f.UpdatedFrom,
		UpdatedTo:      f.UpdatedTo,
		Q:              f.Q,
		PageLimit:      limit,
		CursorMimeType: utils.CursorNullMimeType(f.Cursor, func(c *Cursor) *string { return c.MimeType }),
		CursorID:       utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromMimeAsc))
}

func (s *Service) queryMimeDesc(ctx context.Context, f dbFilters, limit int32) ([]File, error) {
	rows, err := s.repo.ListOwnedFilesMimeDesc(ctx, db.ListOwnedFilesMimeDescParams{
		ViewerID:       f.ViewerID,
		OwnerID:        f.OwnerID,
		Status:         f.Status,
		MimeType:       f.MimeType,
		MinSize:        f.MinSize,
		MaxSize:        f.MaxSize,
		CreatedFrom:    f.CreatedFrom,
		CreatedTo:      f.CreatedTo,
		UpdatedFrom:    f.UpdatedFrom,
		UpdatedTo:      f.UpdatedTo,
		Q:              f.Q,
		PageLimit:      limit,
		CursorMimeType: utils.CursorNullMimeType(f.Cursor, func(c *Cursor) *string { return c.MimeType }),
		CursorID:       utils.CursorUUID(f.Cursor, func(c *Cursor) [16]byte { return c.ID }),
	})
	if err != nil {
		return nil, err
	}
	return mapListRows(toShared(rows, sharedFromMimeDesc))
}

// dbFilters holds the resolved pgtype values shared across all list queries.
type dbFilters struct {
	ViewerID pgtype.UUID
	OwnerID  pgtype.UUID
	Cursor   *Cursor

	Status      db.NullUploadStatus
	MimeType    db.NullMimeType
	MinSize     pgtype.Int8
	MaxSize     pgtype.Int8
	CreatedFrom pgtype.Timestamptz
	CreatedTo   pgtype.Timestamptz
	UpdatedFrom pgtype.Timestamptz
	UpdatedTo   pgtype.Timestamptz
	Q           pgtype.Text
}

func toDBFilters(p ListFilesParams) dbFilters {
	return dbFilters{
		ViewerID:    utils.UUID(p.ViewerID),
		OwnerID:     utils.UUID(p.OwnerID),
		Cursor:      p.Cursor,
		Status:      utils.NullUploadStatus(p.Status),
		MimeType:    utils.NullMimeType(p.MimeType),
		MinSize:     utils.Int8(p.MinSize),
		MaxSize:     utils.Int8(p.MaxSize),
		CreatedFrom: utils.Timestamptz(p.CreatedFrom),
		CreatedTo:   utils.Timestamptz(p.CreatedTo),
		UpdatedFrom: utils.Timestamptz(p.UpdatedFrom),
		UpdatedTo:   utils.Timestamptz(p.UpdatedTo),
		Q:           utils.Text(p.Q),
	}
}

func buildCursor(f File, field SortField) Cursor {
	c := Cursor{ID: f.ID}
	switch field {
	case SortFieldUpdatedAt:
		t := f.UpdatedAt
		c.UpdatedAt = &t
	case SortFieldCreatedAt:
		t := f.CreatedAt
		c.CreatedAt = &t
	case SortFieldName:
		lower := strings.ToLower(f.Name)
		c.NameLower = &lower
	case SortFieldSize:
		s := f.Size
		c.Size = &s
	case SortFieldStatus:
		st := f.Status
		c.Status = &st
	case SortFieldMimeType:
		m := f.MimeType
		c.MimeType = &m
	}
	return c
}

// uuidsToPgtype converts []uuid.UUID to []pgtype.UUID for pgx array params.
func uuidsToPgtype(ids []uuid.UUID) []pgtype.UUID {
	out := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		out[i] = utils.UUID(id)
	}
	return out
}

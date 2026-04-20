package files

import (
	"context"
	"errors"
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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// Repository defines the data-access interface required by the files Service.
type Repository interface {
	InTx(ctx context.Context, fn func(Repository) error) error

	InsertFile(ctx context.Context, arg db.InsertFileParams) (db.File, error)
	UpdateFileStatus(ctx context.Context, arg db.UpdateFileStatusParams) error
	GetFileIfViewable(ctx context.Context, arg db.GetFileIfViewableParams) (db.File, error)
	GetFileByOwner(ctx context.Context, arg db.GetFileByOwnerParams) (db.GetFileByOwnerRow, error)
	SoftDeleteFile(ctx context.Context, arg db.SoftDeleteFileParams) (int64, error)
	SetFileDeletionJobID(ctx context.Context, arg db.SetFileDeletionJobIDParams) error
	MarkFileDeleted(ctx context.Context, fileID pgtype.UUID) error
	GetFileForUpdate(ctx context.Context, fileID pgtype.UUID) (db.GetFileForUpdateRow, error)
	PatchFile(ctx context.Context, arg db.PatchFileParams) (db.PatchFileRow, error)
	InsertFileView(ctx context.Context, arg db.InsertFileViewParams) error
	UpsertFileLastViewed(ctx context.Context, arg db.UpsertFileLastViewedParams) error

	InsertFileGrant(ctx context.Context, arg db.InsertFileGrantParams) (db.FileGrant, error)
	RevokeFileGrant(ctx context.Context, arg db.RevokeFileGrantParams) (int64, error)
	CheckUserExists(ctx context.Context, userID pgtype.UUID) error
	CheckCourseExists(ctx context.Context, courseID pgtype.UUID) error
	CheckStudyGuideExists(ctx context.Context, studyGuideID pgtype.UUID) error

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
//
// Note: as of ASK-105 the service no longer presigns S3 URLs. The
// caller (Next.js server) generates the s3_key + presigns the
// upload separately; the Go API only manages metadata records.
// File deletions still go through QStash + the jobs handler,
// which holds its own S3 client -- the files Service itself is
// pure metadata.
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

// CreateFile inserts a `pending` file metadata record (ASK-105).
// The caller (typically the Next.js server) provides the S3 key
// it generated and presigned separately; this method NEVER touches
// S3. The handler is responsible for validating the request body
// (name, mime_type, size, s3_key) before calling -- this method
// trusts the params struct.
//
// Server-generated UUID for the file ID; user_id always comes from
// the JWT (handler-resolved), never from the request body.
func (s *Service) CreateFile(ctx context.Context, p CreateFileParams) (File, error) {
	row, err := s.repo.InsertFile(ctx, db.InsertFileParams{
		ID:       utils.UUID(uuid.New()),
		UserID:   utils.UUID(p.UserID),
		S3Key:    p.S3Key,
		Name:     p.Name,
		MimeType: p.MimeType,
		Size:     p.Size,
	})
	if err != nil {
		return File{}, fmt.Errorf("CreateFile: insert: %w", err)
	}

	file, err := mapDBFile(row)
	if err != nil {
		return File{}, fmt.Errorf("CreateFile: map: %w", err)
	}
	return file, nil
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

// UpdateFile applies a partial update (PATCH /api/files/{file_id},
// ASK-113). Both Name and Status are optional but at least one must be
// non-nil; the handler is expected to enforce the at-least-one rule but
// this method re-checks defensively.
//
// Validation order (matters for the error the caller sees):
//  1. At-least-one-field    -> 400 "At least one field must be provided"
//  2. Per-field validation  -> 400 with a `details` map for both fields
//  3. 404 / 403 (existence + ownership via GetFileForUpdate)
//  4. Status transition     -> 400 INVALID_TRANSITION
//  5. Apply PatchFile and return the joined post-update row
//
// Status transitions: the only allowed targets are `complete` and
// `failed`, both reachable only from `pending`. Every other current
// status is terminal -- this matches the spec and prevents resurrecting
// a `failed` upload via PATCH (the user must POST a fresh file).
func (s *Service) UpdateFile(ctx context.Context, p UpdateFileParams) (File, error) {
	if p.Name == nil && p.Status == nil {
		return File{}, apperrors.NewBadRequest("At least one field must be provided", nil)
	}

	// Per-field validation -- accumulate so the caller sees both
	// problems at once (matches the spec's "details with both fields"
	// edge case).
	details := make(map[string]string)
	var trimmedName *string
	if p.Name != nil {
		t := strings.TrimSpace(*p.Name)
		if t == "" {
			details["name"] = "name cannot be empty"
		} else if utf8.RuneCountInString(t) > maxFileNameLength {
			details["name"] = fmt.Sprintf("must be %d characters or fewer", maxFileNameLength)
		} else if invalid := invalidNameChars(t); len(invalid) > 0 {
			details["name"] = "contains invalid characters: " + strings.Join(invalid, ", ")
		} else {
			trimmedName = &t
		}
	}
	if p.Status != nil {
		switch *p.Status {
		case "complete", "failed":
			// valid target
		default:
			details["status"] = "must be 'complete' or 'failed'"
		}
	}
	if len(details) > 0 {
		return File{}, apperrors.NewBadRequest("Invalid request body", details)
	}

	// Existence + ownership probe. Soft-deleted rows are filtered out
	// at the SQL level so they always map to 404 here, matching the
	// spec's "Resource not found" rule for deleted files.
	current, err := s.repo.GetFileForUpdate(ctx, utils.UUID(p.FileID))
	if err != nil {
		return File{}, err // sql.ErrNoRows -> apperrors.ErrNotFound via ToHTTPError
	}
	currentOwner, err := utils.PgxToGoogleUUID(current.UserID)
	if err != nil {
		return File{}, fmt.Errorf("UpdateFile: decode owner: %w", err)
	}
	if currentOwner != p.OwnerID {
		return File{}, apperrors.NewForbidden()
	}

	// Status transition validation. We only enforce this when the
	// caller asked to change status; pure renames pass through
	// regardless of current status (a user can rename a `complete`
	// file).
	if p.Status != nil && string(current.Status) != "pending" {
		return File{}, apperrors.NewBadRequest("Invalid status transition", map[string]string{
			"status": fmt.Sprintf("cannot transition from '%s' to '%s'", current.Status, *p.Status),
		})
	}

	row, err := s.repo.PatchFile(ctx, db.PatchFileParams{
		FileID:   utils.UUID(p.FileID),
		OwnerID:  utils.UUID(p.OwnerID),
		ViewerID: utils.UUID(p.ViewerID),
		Name:     utils.Text(trimmedName),
		Status:   utils.NullUploadStatus(p.Status),
	})
	if err != nil {
		// A concurrent DELETE between GetFileForUpdate and PatchFile
		// drops us into sql.ErrNoRows here -- map to 404 to match the
		// spec's "If DELETE wins, PATCH returns 404" rule.
		return File{}, err
	}
	return mapPatchFileRow(row)
}

// invalidNameChars returns a deduped list of disallowed characters
// found in name (path separators, NULs, control chars). Lifted out of
// validateFileName so PATCH can run the same check while building a
// details map without throwing away the helper's other branches.
func invalidNameChars(name string) []string {
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
	return invalid
}

const maxFileNameLength = 255

// RecordFileView logs a view event for POST /api/files/{file_id}/view
// (ASK-134). Two writes per call:
//   - file_views: append-only analytics row.
//   - file_last_viewed: per-(viewer, file) most-recent-timestamp row,
//     upserted so the recents sidebar always reads the latest view.
//
// Existence is gated by GetFileForUpdate so missing or soft-deleted
// files map to apperrors.ErrNotFound -> 404 before either write fires;
// otherwise dangling view rows could accumulate.
//
// Per the spec the two writes are NOT wrapped in a transaction --
// partial failure (view logged but last_viewed stale) is acceptable
// at MVP scale and an UpsertFileLastViewed retry on the next view
// repairs the timestamp. If file_views fails the call returns 500
// without touching last_viewed; if last_viewed fails the analytics
// row already landed so the call still returns 500 but the user
// retries naturally on their next file open.
func (s *Service) RecordFileView(ctx context.Context, viewerID, fileID uuid.UUID) error {
	if _, err := s.repo.GetFileForUpdate(ctx, utils.UUID(fileID)); err != nil {
		return err // sql.ErrNoRows -> apperrors.ErrNotFound via ToHTTPError
	}
	if err := s.repo.InsertFileView(ctx, db.InsertFileViewParams{
		FileID: utils.UUID(fileID),
		UserID: utils.UUID(viewerID),
	}); err != nil {
		return fmt.Errorf("RecordFileView: insert view: %w", err)
	}
	if err := s.repo.UpsertFileLastViewed(ctx, db.UpsertFileLastViewedParams{
		UserID: utils.UUID(viewerID),
		FileID: utils.UUID(fileID),
	}); err != nil {
		return fmt.Errorf("RecordFileView: upsert last viewed: %w", err)
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

// publicSentinelUUID represents "public access" when used as a
// grantee_id with grantee_type=user. ASK-122 carves this UUID out of
// the users-table existence check; every other grantee_id (including
// for course / study_guide) is validated against the target table.
var publicSentinelUUID = uuid.UUID{}

// validGranteeTypes / validPermissions guard the service from junk
// strings reaching the DB. The openapi wrapper enforces these at the
// HTTP boundary; the service re-validates so internal Go callers
// can't bypass it.
var (
	validGranteeTypes = map[string]struct{}{"user": {}, "course": {}, "study_guide": {}}
	validPermissions  = map[string]struct{}{"view": {}, "share": {}, "delete": {}}
)

// CreateGrant creates a file_grants row for ASK-122. Validation order
// matches the spec's error precedence:
//  1. grantee_type / permission enum  -> 400 with details
//  2. file existence (deleted -> 404) -> 404
//  3. file ownership                  -> 403
//  4. grantee existence (per type)    -> 400 with grantee_id detail
//  5. INSERT; unique violation        -> 409
//
// The public sentinel UUID is exempt from step 4 only when
// grantee_type=user; for course / study_guide the sentinel falls into
// the normal not-found branch.
func (s *Service) CreateGrant(ctx context.Context, p CreateGrantParams) (Grant, error) {
	if appErr := validateGranteeFields(p.GranteeType, p.Permission); appErr != nil {
		return Grant{}, appErr
	}

	current, err := s.repo.GetFileForUpdate(ctx, utils.UUID(p.FileID))
	if err != nil {
		return Grant{}, err // sql.ErrNoRows -> ErrNotFound -> 404
	}
	currentOwner, err := utils.PgxToGoogleUUID(current.UserID)
	if err != nil {
		return Grant{}, fmt.Errorf("CreateGrant: decode owner: %w", err)
	}
	if currentOwner != p.OwnerID {
		return Grant{}, apperrors.NewForbidden()
	}

	if appErr := s.validateGranteeExists(ctx, p.GranteeType, p.GranteeID); appErr != nil {
		return Grant{}, appErr
	}

	row, err := s.repo.InsertFileGrant(ctx, db.InsertFileGrantParams{
		FileID:      utils.UUID(p.FileID),
		GranteeType: db.GranteeType(p.GranteeType),
		GranteeID:   utils.UUID(p.GranteeID),
		Permission:  db.Permission(p.Permission),
		GrantedBy:   utils.UUID(p.OwnerID),
	})
	if err != nil {
		// PostgreSQL unique-violation -> 409 Conflict. The
		// file_grants UNIQUE (file_id, grantee_type, grantee_id,
		// permission) constraint maps the duplicate-grant case to
		// sqlstate 23505.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Grant{}, apperrors.ErrConflict
		}
		return Grant{}, fmt.Errorf("CreateGrant: %w", err)
	}

	return mapGrantRow(row)
}

// RevokeGrant deletes a file_grants row for ASK-125. Same 404 / 403
// gating as CreateGrant; the body fields are validated but no
// grantee-existence check (the grant either exists in file_grants or
// it doesn't). RevokeFileGrant returning 0 rows means the grant was
// missing -- spec says 404 (not idempotent no-op).
func (s *Service) RevokeGrant(ctx context.Context, p RevokeGrantParams) error {
	if appErr := validateGranteeFields(p.GranteeType, p.Permission); appErr != nil {
		return appErr
	}

	current, err := s.repo.GetFileForUpdate(ctx, utils.UUID(p.FileID))
	if err != nil {
		return err
	}
	currentOwner, err := utils.PgxToGoogleUUID(current.UserID)
	if err != nil {
		return fmt.Errorf("RevokeGrant: decode owner: %w", err)
	}
	if currentOwner != p.OwnerID {
		return apperrors.NewForbidden()
	}

	rows, err := s.repo.RevokeFileGrant(ctx, db.RevokeFileGrantParams{
		FileID:      utils.UUID(p.FileID),
		GranteeType: db.GranteeType(p.GranteeType),
		GranteeID:   utils.UUID(p.GranteeID),
		Permission:  db.Permission(p.Permission),
	})
	if err != nil {
		return fmt.Errorf("RevokeGrant: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("RevokeGrant: %w", apperrors.ErrNotFound)
	}
	return nil
}

// validateGranteeFields rejects unknown grantee_type / permission
// values up-front so the error response carries both detail keys
// when both fields are bad (matches the spec's "details for both
// fields" edge case).
func validateGranteeFields(granteeType, permission string) *apperrors.AppError {
	details := make(map[string]string)
	if _, ok := validGranteeTypes[granteeType]; !ok {
		details["grantee_type"] = "must be 'user', 'course', or 'study_guide'"
	}
	if _, ok := validPermissions[permission]; !ok {
		details["permission"] = "must be 'view', 'share', or 'delete'"
	}
	if len(details) > 0 {
		return apperrors.NewBadRequest("Invalid request body", details)
	}
	return nil
}

// validateGranteeExists looks the grantee_id up in the table that
// matches grantee_type. For user grantees the public sentinel UUID
// is exempt -- it represents "public access" and does not correspond
// to a real users row. ErrNoRows from the probe maps to a 400
// VALIDATION_ERROR per the spec, NOT a 404 (the missing entity is
// the grantee, not the file).
func (s *Service) validateGranteeExists(ctx context.Context, granteeType string, granteeID uuid.UUID) *apperrors.AppError {
	pgID := utils.UUID(granteeID)
	var probeErr error
	switch granteeType {
	case "user":
		if granteeID == publicSentinelUUID {
			return nil // sentinel exempt from users lookup
		}
		probeErr = s.repo.CheckUserExists(ctx, pgID)
	case "course":
		probeErr = s.repo.CheckCourseExists(ctx, pgID)
	case "study_guide":
		probeErr = s.repo.CheckStudyGuideExists(ctx, pgID)
	default:
		// Unreachable: validateGranteeFields already gated this.
		return apperrors.NewBadRequest("Invalid request body", map[string]string{
			"grantee_type": "must be 'user', 'course', or 'study_guide'",
		})
	}
	if probeErr != nil {
		if errors.Is(probeErr, apperrors.ErrNotFound) {
			return apperrors.NewBadRequest("Grantee not found", map[string]string{
				"grantee_id": fmt.Sprintf("no %s with this ID", granteeType),
			})
		}
		// Real DB error -- propagate as 500 via the wrapping
		// AppError converter in the handler.
		return &apperrors.AppError{
			Code:    500,
			Status:  "Internal Server Error",
			Message: "Something went wrong",
			Cause:   probeErr,
		}
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
		CursorMimeType: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.MimeType }),
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
		CursorMimeType: utils.CursorText(f.Cursor, func(c *Cursor) *string { return c.MimeType }),
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
	MimeType    pgtype.Text
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
		MimeType:    utils.Text(p.MimeType),
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

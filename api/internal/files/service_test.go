package files_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	mock_files "github.com/Ask-Atlas/AskAtlas/api/internal/files/mocks"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ListFiles_Scope(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	params := files.ListFilesParams{
		Scope: files.ScopeCourse, // Unsupported
	}

	_, _, err := svc.ListFiles(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("scope %q is not yet implemented", files.ScopeCourse))
}

func TestService_ListFiles_Pagination_HasNextPage(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	viewerID := uuid.New()
	params := files.ListFilesParams{
		ViewerID:  viewerID,
		OwnerID:   viewerID,
		Scope:     files.ScopeOwned,
		SortField: files.SortFieldUpdatedAt,
		SortDir:   files.SortDirDesc,
		PageLimit: 2,
	}

	// Prepare 3 rows (Limit + 1) to simulate "HasNextPage"
	now := time.Now()
	rows := []db.ListOwnedFilesUpdatedDescRow{
		{ID: utils.UUID(uuid.New()), UserID: utils.UUID(viewerID), Name: "f1", UpdatedAt: utils.Timestamptz(&now)},
		{ID: utils.UUID(uuid.New()), UserID: utils.UUID(viewerID), Name: "f2", UpdatedAt: utils.Timestamptz(&now)},
		{ID: utils.UUID(uuid.New()), UserID: utils.UUID(viewerID), Name: "f3", UpdatedAt: utils.Timestamptz(&now)},
	}

	repo.EXPECT().
		ListOwnedFilesUpdatedDesc(mock.Anything, mock.MatchedBy(func(arg db.ListOwnedFilesUpdatedDescParams) bool {
			return arg.PageLimit == 3 // Expect PageLimit + 1
		})).
		Return(rows, nil)

	results, nextCursor, err := svc.ListFiles(context.Background(), params)
	require.NoError(t, err)

	assert.Len(t, results, 2, "expected results to be trimmed to PageLimit")
	assert.NotNil(t, nextCursor, "expected nextCursor to be present")
}

func TestService_ListFiles_Pagination_NoNextPage(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	viewerID := uuid.New()
	params := files.ListFilesParams{
		ViewerID:  viewerID,
		OwnerID:   viewerID,
		Scope:     files.ScopeOwned,
		SortField: files.SortFieldUpdatedAt,
		SortDir:   files.SortDirDesc,
		PageLimit: 5,
	}

	// Prepare 2 rows (under limit)
	now := time.Now()
	rows := []db.ListOwnedFilesUpdatedDescRow{
		{ID: utils.UUID(uuid.New()), UserID: utils.UUID(viewerID), Name: "f1", UpdatedAt: utils.Timestamptz(&now)},
		{ID: utils.UUID(uuid.New()), UserID: utils.UUID(viewerID), Name: "f2", UpdatedAt: utils.Timestamptz(&now)},
	}

	repo.EXPECT().
		ListOwnedFilesUpdatedDesc(mock.Anything, mock.MatchedBy(func(arg db.ListOwnedFilesUpdatedDescParams) bool {
			return arg.PageLimit == 6 // Expect PageLimit + 1
		})).
		Return(rows, nil)

	results, nextCursor, err := svc.ListFiles(context.Background(), params)
	require.NoError(t, err)

	assert.Len(t, results, 2)
	assert.Nil(t, nextCursor, "expected nextCursor to be nil")
}

// As of ASK-105 the service no longer presigns S3 URLs and no longer
// validates / sanitizes the file name -- the caller (Next.js server)
// generates the s3_key and the handler trims/length-checks the name.
// The service only persists what it's given.
func TestService_CreateFile_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	userID := uuid.New()
	now := time.Now()
	callerS3Key := "uploads/abc123/lecture-notes.pdf"

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "lecture-notes.pdf",
		MimeType: "application/pdf",
		Size:     1048576,
		S3Key:    callerS3Key,
	}

	var capturedID pgtype.UUID

	repo.EXPECT().
		InsertFile(mock.Anything, mock.MatchedBy(func(arg db.InsertFileParams) bool {
			capturedID = arg.ID
			return arg.ID.Valid &&
				arg.UserID == utils.UUID(userID) &&
				arg.Name == "lecture-notes.pdf" &&
				arg.MimeType == "application/pdf" &&
				arg.Size == int64(1048576) &&
				// The service must store the caller-supplied key as-is,
				// NOT regenerate one server-side.
				arg.S3Key == callerS3Key
		})).
		RunAndReturn(func(_ context.Context, arg db.InsertFileParams) (db.File, error) {
			return db.File{
				ID:        arg.ID,
				UserID:    arg.UserID,
				S3Key:     arg.S3Key,
				Name:      arg.Name,
				MimeType:  arg.MimeType,
				Size:      arg.Size,
				Status:    db.UploadStatusPending,
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			}, nil
		})

	file, err := svc.CreateFile(context.Background(), params)
	require.NoError(t, err)

	wantID, err := utils.PgxToGoogleUUID(capturedID)
	require.NoError(t, err)
	assert.Equal(t, wantID, file.ID)
	assert.Equal(t, "lecture-notes.pdf", file.Name)
	assert.Equal(t, int64(1048576), file.Size)
	assert.Equal(t, "application/pdf", file.MimeType)
	assert.Equal(t, "pending", file.Status)
}

func TestService_CreateFile_InsertError(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	params := files.CreateFileParams{
		UserID:   uuid.New(),
		Name:     "file.pdf",
		MimeType: "application/pdf",
		Size:     100,
		S3Key:    "uploads/abc/file.pdf",
	}

	repo.EXPECT().
		InsertFile(mock.Anything, mock.Anything).
		Return(db.File{}, fmt.Errorf("db error"))

	_, err := svc.CreateFile(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreateFile: insert")
}

// Service stores the S3 key as-is. Verifies that even an unusual
// caller-supplied key (e.g. one containing `..`) is persisted
// unchanged -- the Go API trusts the caller (Next.js) to generate
// safe keys, and the s3_key is opaque to the API surface (it is
// never returned in any response).
func TestService_CreateFile_StoresS3KeyVerbatim(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	userID := uuid.New()
	now := time.Now()
	weirdKey := "uploads/../../weird/path with spaces.pdf"

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "doc.pdf",
		MimeType: "application/pdf",
		Size:     100,
		S3Key:    weirdKey,
	}

	var captured string
	repo.EXPECT().
		InsertFile(mock.Anything, mock.MatchedBy(func(arg db.InsertFileParams) bool {
			captured = arg.S3Key
			return true
		})).
		RunAndReturn(func(_ context.Context, arg db.InsertFileParams) (db.File, error) {
			return db.File{
				ID:        arg.ID,
				UserID:    arg.UserID,
				S3Key:     arg.S3Key,
				Name:      arg.Name,
				MimeType:  arg.MimeType,
				Size:      arg.Size,
				Status:    db.UploadStatusPending,
				CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			}, nil
		})

	_, err := svc.CreateFile(context.Background(), params)
	require.NoError(t, err)
	assert.Equal(t, weirdKey, captured, "service must store the caller-supplied s3_key verbatim, with no sanitization")
}

// ----------------------------------------------------------------------
// PATCH /api/files/{file_id} -- ASK-113.
//
// Coverage matrix (mapped to the 10 acceptance criteria + edge cases
// in the ticket):
//
//   AC1  pending -> complete                        StatusComplete_Success
//   AC2  pending -> failed                          StatusFailed_Success
//   AC3  complete -> failed                         InvalidTransition_*
//   AC4  failed -> complete                         InvalidTransition_*
//   AC5  rename only on a complete file             RenameOnly_OnCompleteFile
//   AC6  rename + status atomically                 NameAndStatus_Atomic
//   AC7  not the owner                              NotOwner
//   AC8  non-existent file                          NotFound
//   AC9  soft-deleted file                          NotFound (same path -- SQL filter)
//   AC10 empty body                                 EmptyBody
//
// Edge cases:
//   - status enum violation                         InvalidStatusValue
//   - both name + status invalid                    BothFieldsInvalid_DetailsForBoth
//   - name 255 chars (boundary in)                  Name255Chars_Accepted
//   - name 256 chars (boundary out)                 NameTooLong
//   - name "" / "   "                               EmptyName_AfterTrim
//   - name with /, \, control chars                 DangerousChars
//   - DB error mid-PATCH                            PatchFile_Error
// ----------------------------------------------------------------------

// patchFileTestSetup primes a successful GetFileForUpdate response so a
// test can focus on whatever it's actually exercising. Returns the
// fresh repo + service + file/owner ids so callers can compose
// further EXPECT() lines.
func patchFileTestSetup(t *testing.T, currentStatus db.UploadStatus) (*mock_files.MockRepository, *files.Service, uuid.UUID, uuid.UUID) {
	t.Helper()
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)
	fid := uuid.New()
	oid := uuid.New()

	repo.EXPECT().
		GetFileForUpdate(mock.Anything, utils.UUID(fid)).
		Return(db.GetFileForUpdateRow{
			ID:     utils.UUID(fid),
			UserID: utils.UUID(oid),
			Status: currentStatus,
		}, nil)

	return repo, svc, fid, oid
}

// expectPatchFile registers a PatchFile expectation that returns a
// canned post-update row. Tests assert on the matched arg.
func expectPatchFile(repo *mock_files.MockRepository, fid, oid uuid.UUID, match func(arg db.PatchFileParams) bool, returnedName string, returnedStatus db.UploadStatus) {
	now := time.Now()
	repo.EXPECT().
		PatchFile(mock.Anything, mock.MatchedBy(match)).
		Return(db.PatchFileRow{
			ID:        utils.UUID(fid),
			UserID:    utils.UUID(oid),
			Name:      returnedName,
			Size:      1024,
			MimeType:  "application/pdf",
			Status:    returnedStatus,
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)
}

func strPtr(s string) *string { return &s }

func TestService_UpdateFile_StatusComplete_Success(t *testing.T) {
	repo, svc, fid, oid := patchFileTestSetup(t, "pending")
	expectPatchFile(repo, fid, oid, func(arg db.PatchFileParams) bool {
		return arg.FileID == utils.UUID(fid) &&
			arg.OwnerID == utils.UUID(oid) &&
			arg.ViewerID == utils.UUID(oid) &&
			!arg.Name.Valid &&
			arg.Status.Valid && arg.Status.UploadStatus == "complete"
	}, "notes.pdf", "complete")

	f, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:   fid,
		OwnerID:  oid,
		ViewerID: oid,
		Status:   strPtr("complete"),
	})
	require.NoError(t, err)
	assert.Equal(t, "complete", f.Status)
}

func TestService_UpdateFile_StatusFailed_Success(t *testing.T) {
	repo, svc, fid, oid := patchFileTestSetup(t, "pending")
	expectPatchFile(repo, fid, oid, func(arg db.PatchFileParams) bool {
		return arg.Status.Valid && arg.Status.UploadStatus == "failed"
	}, "notes.pdf", "failed")

	f, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:   fid,
		OwnerID:  oid,
		ViewerID: oid,
		Status:   strPtr("failed"),
	})
	require.NoError(t, err)
	assert.Equal(t, "failed", f.Status)
}

func TestService_UpdateFile_InvalidTransition_CompleteToFailed(t *testing.T) {
	_, svc, fid, oid := patchFileTestSetup(t, "complete")

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Status: strPtr("failed"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["status"], "cannot transition from 'complete' to 'failed'")
}

func TestService_UpdateFile_InvalidTransition_FailedToComplete(t *testing.T) {
	_, svc, fid, oid := patchFileTestSetup(t, "failed")

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Status: strPtr("complete"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Contains(t, appErr.Details["status"], "cannot transition from 'failed' to 'complete'")
}

func TestService_UpdateFile_RenameOnly_OnCompleteFile(t *testing.T) {
	// AC5: A file already in `complete` can still be renamed --
	// status validation only fires when the caller asked to change it.
	repo, svc, fid, oid := patchFileTestSetup(t, "complete")
	expectPatchFile(repo, fid, oid, func(arg db.PatchFileParams) bool {
		return arg.Name.Valid && arg.Name.String == "renamed.pdf" && !arg.Status.Valid
	}, "renamed.pdf", "complete")

	f, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Name: strPtr("renamed.pdf"),
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed.pdf", f.Name)
	assert.Equal(t, "complete", f.Status)
}

func TestService_UpdateFile_NameAndStatus_Atomic(t *testing.T) {
	// AC6: pending file, both name + status are sent in the same call.
	repo, svc, fid, oid := patchFileTestSetup(t, "pending")
	expectPatchFile(repo, fid, oid, func(arg db.PatchFileParams) bool {
		return arg.Name.Valid && arg.Name.String == "new-name.pdf" &&
			arg.Status.Valid && arg.Status.UploadStatus == "complete"
	}, "new-name.pdf", "complete")

	f, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Name:   strPtr("new-name.pdf"),
		Status: strPtr("complete"),
	})
	require.NoError(t, err)
	assert.Equal(t, "new-name.pdf", f.Name)
	assert.Equal(t, "complete", f.Status)
}

func TestService_UpdateFile_NotOwner(t *testing.T) {
	// AC7: caller is not the file owner -> 403, no PatchFile call.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)
	fid := uuid.New()
	owner := uuid.New()
	caller := uuid.New() // different user

	repo.EXPECT().
		GetFileForUpdate(mock.Anything, utils.UUID(fid)).
		Return(db.GetFileForUpdateRow{
			ID:     utils.UUID(fid),
			UserID: utils.UUID(owner),
			Status: "pending",
		}, nil)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: caller, ViewerID: caller,
		Name: strPtr("renamed.pdf"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

func TestService_UpdateFile_NotFound(t *testing.T) {
	// AC8 + AC9: missing or soft-deleted -> 404. Both reach service
	// as sql.ErrNoRows from the repo layer.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().
		GetFileForUpdate(mock.Anything, mock.Anything).
		Return(db.GetFileForUpdateRow{}, fmt.Errorf("GetFileForUpdate: %w", apperrors.ErrNotFound))

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
		Name: strPtr("renamed.pdf"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_UpdateFile_EmptyBody(t *testing.T) {
	// AC10: both fields nil -> 400 before we touch the DB.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Equal(t, "At least one field must be provided", appErr.Message)
}

func TestService_UpdateFile_InvalidStatusValue(t *testing.T) {
	// "pending" is not a valid target (neither is "deleted" or any
	// other arbitrary string). Service rejects before DB.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
		Status: strPtr("pending"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "must be 'complete' or 'failed'", appErr.Details["status"])
}

func TestService_UpdateFile_BothFieldsInvalid_DetailsForBoth(t *testing.T) {
	// Edge case: both name and status are invalid -> details map
	// surfaces both.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
		Name:   strPtr("   "), // empty after trim
		Status: strPtr("bogus"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.NotEmpty(t, appErr.Details["name"])
	assert.NotEmpty(t, appErr.Details["status"])
}

func TestService_UpdateFile_Name255Chars_Accepted(t *testing.T) {
	// Boundary: exactly 255 ASCII characters is accepted.
	name := strings.Repeat("a", 255)
	repo, svc, fid, oid := patchFileTestSetup(t, "complete")
	expectPatchFile(repo, fid, oid, func(arg db.PatchFileParams) bool {
		return arg.Name.Valid && arg.Name.String == name
	}, name, "complete")

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Name: &name,
	})
	require.NoError(t, err)
}

func TestService_UpdateFile_NameTooLong(t *testing.T) {
	// Boundary: 256 chars is over.
	name := strings.Repeat("a", 256)
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
		Name: &name,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Contains(t, appErr.Details["name"], "255")
}

func TestService_UpdateFile_EmptyName_AfterTrim(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
		Name: strPtr("   "),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "name cannot be empty", appErr.Details["name"])
}

func TestService_UpdateFile_DangerousChars(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: uuid.New(), OwnerID: uuid.New(), ViewerID: uuid.New(),
		Name: strPtr("my/file\\name.pdf"),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Contains(t, appErr.Details["name"], "/")
	assert.Contains(t, appErr.Details["name"], "\\")
}

func TestService_UpdateFile_WhitespaceTrimmed(t *testing.T) {
	// Service trims before sending to SQL -- the repo sees the trimmed
	// form, not the padded original.
	repo, svc, fid, oid := patchFileTestSetup(t, "complete")
	expectPatchFile(repo, fid, oid, func(arg db.PatchFileParams) bool {
		return arg.Name.Valid && arg.Name.String == "trimmed.pdf"
	}, "trimmed.pdf", "complete")

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Name: strPtr("  trimmed.pdf  "),
	})
	require.NoError(t, err)
}

func TestService_UpdateFile_PatchFile_Error(t *testing.T) {
	// DB error after probe succeeds -> error propagated unchanged.
	// Models the "concurrent DELETE wins" race in the spec.
	repo, svc, fid, oid := patchFileTestSetup(t, "pending")
	repo.EXPECT().
		PatchFile(mock.Anything, mock.Anything).
		Return(db.PatchFileRow{}, fmt.Errorf("PatchFile: %w", apperrors.ErrNotFound))

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID: fid, OwnerID: oid, ViewerID: oid,
		Status: strPtr("complete"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_GetFile(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fid := uuid.New()
	vid := uuid.New()
	params := files.GetFileParams{
		FileID:   fid,
		ViewerID: vid,
	}

	row := db.File{
		ID:        utils.UUID(fid),
		UserID:    utils.UUID(vid),
		Name:      "test.txt",
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	repo.EXPECT().
		GetFileIfViewable(mock.Anything, mock.MatchedBy(func(arg db.GetFileIfViewableParams) bool {
			return arg.FileID == row.ID && arg.ViewerID == row.UserID
		})).
		Return(row, nil)

	f, err := svc.GetFile(context.Background(), params)
	require.NoError(t, err)
	assert.Equal(t, fid, f.ID)
	assert.Equal(t, "test.txt", f.Name)
}

// ----------------------------------------------------------------------
// POST /api/files/{file_id}/view (ASK-134).
//
// Coverage:
//   AC1: existence check passes -> both writes fire, no error
//   AC3+4: existence check returns ErrNotFound -> neither write runs
//   InsertFileView fails -> 500-class error, UpsertFileLastViewed not called
//   UpsertFileLastViewed fails -> 500-class error after view already logged
// ----------------------------------------------------------------------

func TestService_RecordFileView_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	viewerID := uuid.New()
	fileID := uuid.New()

	repo.EXPECT().
		GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(db.GetFileForUpdateRow{
			ID:     utils.UUID(fileID),
			UserID: utils.UUID(uuid.New()), // arbitrary owner -- viewing is permission-less
			Status: "complete",
		}, nil)
	repo.EXPECT().
		InsertFileView(mock.Anything, mock.MatchedBy(func(arg db.InsertFileViewParams) bool {
			return arg.FileID == utils.UUID(fileID) && arg.UserID == utils.UUID(viewerID)
		})).
		Return(nil)
	repo.EXPECT().
		UpsertFileLastViewed(mock.Anything, mock.MatchedBy(func(arg db.UpsertFileLastViewedParams) bool {
			return arg.FileID == utils.UUID(fileID) && arg.UserID == utils.UUID(viewerID)
		})).
		Return(nil)

	err := svc.RecordFileView(context.Background(), viewerID, fileID)
	require.NoError(t, err)
}

func TestService_RecordFileView_FileNotFound(t *testing.T) {
	// Existence probe returns ErrNotFound -- neither view write
	// fires (mockery would fail the test on any unexpected call).
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().
		GetFileForUpdate(mock.Anything, mock.Anything).
		Return(db.GetFileForUpdateRow{}, fmt.Errorf("GetFileForUpdate: %w", apperrors.ErrNotFound))

	err := svc.RecordFileView(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_RecordFileView_InsertViewFails(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(db.GetFileForUpdateRow{Status: "complete"}, nil)
	repo.EXPECT().InsertFileView(mock.Anything, mock.Anything).
		Return(errors.New("connection lost"))
	// UpsertFileLastViewed must NOT be called when the view insert fails.

	err := svc.RecordFileView(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection lost")
}

func TestService_RecordFileView_UpsertLastViewedFails(t *testing.T) {
	// View insert succeeded but last-viewed upsert failed. The spec
	// accepts partial failure; we still surface the error so the
	// client can retry. The dangling file_views row is harmless
	// (idempotent analytics).
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(db.GetFileForUpdateRow{Status: "complete"}, nil)
	repo.EXPECT().InsertFileView(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().UpsertFileLastViewed(mock.Anything, mock.Anything).
		Return(errors.New("deadlock detected"))

	err := svc.RecordFileView(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deadlock detected")
}

// GetDownloadURL (ASK-205)

func TestService_GetDownloadURL_HappyPath(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	gen := mock_files.NewMockDownloadURLGenerator(t)
	svc := files.NewService(repo, files.WithDownloadURLGenerator(gen))

	viewerID := uuid.New()
	fileID := uuid.New()
	s3Key := "uploads/abc/doc.pdf"
	want := "https://s3.example/b/" + s3Key + "?sig=xyz"

	repo.EXPECT().
		GetFileIfViewable(mock.Anything, mock.MatchedBy(func(p db.GetFileIfViewableParams) bool {
			return p.FileID == utils.UUID(fileID) && p.ViewerID == utils.UUID(viewerID)
		})).
		Return(db.File{
			S3Key:  s3Key,
			Status: "complete",
		}, nil)

	gen.EXPECT().
		GeneratePresignedGetURL(mock.Anything, s3Key).
		Return(want, nil)

	got, err := svc.GetDownloadURL(context.Background(), viewerID, fileID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestService_GetDownloadURL_NoGeneratorConfigured(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.GetDownloadURL(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download URL generator not configured")
}

func TestService_GetDownloadURL_ForwardsRepoNotFound(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	gen := mock_files.NewMockDownloadURLGenerator(t)
	svc := files.NewService(repo, files.WithDownloadURLGenerator(gen))

	repo.EXPECT().GetFileIfViewable(mock.Anything, mock.Anything).
		Return(db.File{}, apperrors.ErrNotFound)

	_, err := svc.GetDownloadURL(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_GetDownloadURL_NonCompleteStatusIs404(t *testing.T) {
	cases := []string{"pending", "failed"}

	for _, status := range cases {
		t.Run(status, func(t *testing.T) {
			repo := mock_files.NewMockRepository(t)
			gen := mock_files.NewMockDownloadURLGenerator(t)
			svc := files.NewService(repo, files.WithDownloadURLGenerator(gen))

			repo.EXPECT().GetFileIfViewable(mock.Anything, mock.Anything).
				Return(db.File{S3Key: "uploads/x", Status: db.UploadStatus(status)}, nil)

			_, err := svc.GetDownloadURL(context.Background(), uuid.New(), uuid.New())
			require.Error(t, err)
			assert.ErrorIs(t, err, apperrors.ErrNotFound)
		})
	}
}

func TestService_GetDownloadURL_PresignerErrorBubblesUp(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	gen := mock_files.NewMockDownloadURLGenerator(t)
	svc := files.NewService(repo, files.WithDownloadURLGenerator(gen))

	repo.EXPECT().GetFileIfViewable(mock.Anything, mock.Anything).
		Return(db.File{S3Key: "uploads/x", Status: "complete"}, nil)
	gen.EXPECT().GeneratePresignedGetURL(mock.Anything, "uploads/x").
		Return("", errors.New("aws creds missing"))

	_, err := svc.GetDownloadURL(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "presign")
	assert.Contains(t, err.Error(), "aws creds missing")
}

func TestService_EnqueueExtractJob_PublishesWithS3KeyAndMime(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	pub := mock_files.NewMockQStashPublisher(t)
	svc := files.NewService(repo)

	fileID, ownerID := uuid.New(), uuid.New()
	repo.EXPECT().GetFileByOwner(mock.Anything, db.GetFileByOwnerParams{
		FileID:  utils.UUID(fileID),
		OwnerID: utils.UUID(ownerID),
	}).Return(db.GetFileByOwnerRow{
		ID:       utils.UUID(fileID),
		UserID:   utils.UUID(ownerID),
		S3Key:    "uploads/abc.pdf",
		MimeType: "application/pdf",
		Status:   "complete",
	}, nil)

	pub.EXPECT().PublishExtractFile(mock.Anything, mock.MatchedBy(func(msg qstashclient.ExtractFileMessage) bool {
		return msg.FileID == fileID.String() &&
			msg.S3Key == "uploads/abc.pdf" &&
			msg.MimeType == "application/pdf" &&
			msg.UserID == ownerID.String() &&
			msg.RequestedAt != ""
	})).Return("qstash-msg-id-123", nil)

	require.NoError(t, svc.EnqueueExtractJob(context.Background(), fileID, ownerID, pub))
}

func TestService_EnqueueExtractJob_NotFoundIsErrNotFound(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	pub := mock_files.NewMockQStashPublisher(t)
	svc := files.NewService(repo)

	fileID, ownerID := uuid.New(), uuid.New()
	repo.EXPECT().GetFileByOwner(mock.Anything, mock.Anything).
		Return(db.GetFileByOwnerRow{}, apperrors.ErrNotFound)

	err := svc.EnqueueExtractJob(context.Background(), fileID, ownerID, pub)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
	// Publisher must not be called when lookup fails.
	pub.AssertNotCalled(t, "PublishExtractFile", mock.Anything, mock.Anything)
}

func TestService_EnqueueExtractJob_PublishErrorPropagates(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	pub := mock_files.NewMockQStashPublisher(t)
	svc := files.NewService(repo)

	fileID, ownerID := uuid.New(), uuid.New()
	repo.EXPECT().GetFileByOwner(mock.Anything, mock.Anything).
		Return(db.GetFileByOwnerRow{
			ID: utils.UUID(fileID), UserID: utils.UUID(ownerID),
			S3Key: "k", MimeType: "text/plain", Status: "complete",
		}, nil)
	pub.EXPECT().PublishExtractFile(mock.Anything, mock.Anything).
		Return("", errors.New("qstash 503"))

	err := svc.EnqueueExtractJob(context.Background(), fileID, ownerID, pub)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "publish")
	assert.Contains(t, err.Error(), "qstash 503")
}

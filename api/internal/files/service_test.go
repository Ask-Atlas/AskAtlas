package files_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	mock_files "github.com/Ask-Atlas/AskAtlas/api/internal/files/mocks"
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

func TestService_UpdateFile_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fid := uuid.New()
	oid := uuid.New()
	now := time.Now()

	repo.EXPECT().
		UpdateFile(mock.Anything, mock.MatchedBy(func(arg db.UpdateFileParams) bool {
			return arg.FileID == utils.UUID(fid) &&
				arg.OwnerID == utils.UUID(oid) &&
				arg.Name == "renamed.pdf"
		})).
		Return(db.UpdateFileRow{
			ID:        utils.UUID(fid),
			UserID:    utils.UUID(oid),
			Name:      "renamed.pdf",
			Size:      1024,
			MimeType:  "application/pdf",
			Status:    "complete",
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

	f, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:  fid,
		OwnerID: oid,
		Name:    "renamed.pdf",
	})
	require.NoError(t, err)
	assert.Equal(t, fid, f.ID)
	assert.Equal(t, "renamed.pdf", f.Name)
	assert.Equal(t, int64(1024), f.Size)
}

func TestService_UpdateFile_NotFound(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().
		UpdateFile(mock.Anything, mock.Anything).
		Return(db.UpdateFileRow{}, fmt.Errorf("UpdateFile: %w", apperrors.ErrNotFound))

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:  uuid.New(),
		OwnerID: uuid.New(),
		Name:    "valid-name.pdf",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_UpdateFile_EmptyNameAfterTrim(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:  uuid.New(),
		OwnerID: uuid.New(),
		Name:    "   ",
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["name"], "must not be empty")
}

func TestService_UpdateFile_NameTooLong(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	longName := string(make([]byte, 256))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:  uuid.New(),
		OwnerID: uuid.New(),
		Name:    longName,
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["name"], "255")
}

func TestService_UpdateFile_DangerousChars(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:  uuid.New(),
		OwnerID: uuid.New(),
		Name:    "my/file\\name.pdf",
	})
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["name"], "/")
	assert.Contains(t, appErr.Details["name"], "\\")
}

func TestService_UpdateFile_WhitespaceTrimmed(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fid := uuid.New()
	oid := uuid.New()
	now := time.Now()

	repo.EXPECT().
		UpdateFile(mock.Anything, mock.MatchedBy(func(arg db.UpdateFileParams) bool {
			return arg.Name == "trimmed.pdf" // Verify whitespace was stripped
		})).
		Return(db.UpdateFileRow{
			ID:        utils.UUID(fid),
			UserID:    utils.UUID(oid),
			Name:      "trimmed.pdf",
			Size:      512,
			MimeType:  "application/pdf",
			Status:    "complete",
			CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

	f, err := svc.UpdateFile(context.Background(), files.UpdateFileParams{
		FileID:  fid,
		OwnerID: oid,
		Name:    "  trimmed.pdf  ",
	})
	require.NoError(t, err)
	assert.Equal(t, "trimmed.pdf", f.Name)
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

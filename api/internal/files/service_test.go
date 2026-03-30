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

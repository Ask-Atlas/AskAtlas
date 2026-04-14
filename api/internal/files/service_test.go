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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_ListFiles_Scope(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo, nil)

	params := files.ListFilesParams{
		Scope: files.ScopeCourse, // Unsupported
	}

	_, _, err := svc.ListFiles(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("scope %q is not yet implemented", files.ScopeCourse))
}

func TestService_ListFiles_Pagination_HasNextPage(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo, nil)

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
	svc := files.NewService(repo, nil)

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

func TestService_CreateFile_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	s3Mock := mock_files.NewMockS3Uploader(t)
	svc := files.NewService(repo, s3Mock)

	userID := uuid.New()
	fileID := uuid.New()
	now := time.Now()

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "lecture-notes.pdf",
		MimeType: "application/pdf",
		Size:     1048576,
	}

	insertedFile := db.File{
		ID:        utils.UUID(fileID),
		UserID:    utils.UUID(userID),
		S3Key:     "users/" + userID.String() + "/files",
		Name:      "lecture-notes.pdf",
		MimeType:  db.MimeTypeApplicationPdf,
		Size:      1048576,
		Status:    db.UploadStatusPending,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	repo.EXPECT().
		InsertFile(mock.Anything, mock.MatchedBy(func(arg db.InsertFileParams) bool {
			return arg.UserID == utils.UUID(userID) &&
				arg.Name == "lecture-notes.pdf" &&
				arg.MimeType == db.MimeTypeApplicationPdf &&
				arg.Size == int64(1048576)
		})).
		Return(insertedFile, nil)

	expectedKey := fmt.Sprintf("users/%s/files/%s/lecture-notes.pdf", userID.String(), fileID.String())
	s3Mock.EXPECT().
		GeneratePresignedPutURL(mock.Anything, expectedKey, "application/pdf", int64(1048576)).
		Return("https://s3.example.com/presigned-url", nil)

	result, err := svc.CreateFile(context.Background(), params)
	require.NoError(t, err)
	assert.Equal(t, fileID, result.File.ID)
	assert.Equal(t, "lecture-notes.pdf", result.File.Name)
	assert.Equal(t, int64(1048576), result.File.Size)
	assert.Equal(t, "application/pdf", result.File.MimeType)
	assert.Equal(t, "pending", result.File.Status)
	assert.Equal(t, "https://s3.example.com/presigned-url", result.UploadURL)
}

func TestService_CreateFile_InsertError(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	s3Mock := mock_files.NewMockS3Uploader(t)
	svc := files.NewService(repo, s3Mock)

	params := files.CreateFileParams{
		UserID:   uuid.New(),
		Name:     "file.pdf",
		MimeType: "application/pdf",
		Size:     100,
	}

	repo.EXPECT().
		InsertFile(mock.Anything, mock.Anything).
		Return(db.File{}, fmt.Errorf("db error"))

	_, err := svc.CreateFile(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreateFile: insert")
}

func TestService_CreateFile_PresignError(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	s3Mock := mock_files.NewMockS3Uploader(t)
	svc := files.NewService(repo, s3Mock)

	userID := uuid.New()
	fileID := uuid.New()
	now := time.Now()

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "file.pdf",
		MimeType: "application/pdf",
		Size:     100,
	}

	insertedFile := db.File{
		ID:        utils.UUID(fileID),
		UserID:    utils.UUID(userID),
		S3Key:     "users/" + userID.String() + "/files",
		Name:      "file.pdf",
		MimeType:  db.MimeTypeApplicationPdf,
		Size:      100,
		Status:    db.UploadStatusPending,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	repo.EXPECT().
		InsertFile(mock.Anything, mock.Anything).
		Return(insertedFile, nil)

	s3Mock.EXPECT().
		GeneratePresignedPutURL(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", fmt.Errorf("s3 error"))

	_, err := svc.CreateFile(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreateFile: presign")
}

func TestService_GetFile(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo, nil)

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

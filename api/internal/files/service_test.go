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
	now := time.Now()

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "lecture-notes.pdf",
		MimeType: "application/pdf",
		Size:     1048576,
	}

	// Capture the generated file ID and s3_key from the insert call.
	var capturedID pgtype.UUID
	var capturedS3Key string

	repo.EXPECT().
		InsertFile(mock.Anything, mock.MatchedBy(func(arg db.InsertFileParams) bool {
			capturedID = arg.ID
			capturedS3Key = arg.S3Key
			return arg.ID.Valid &&
				arg.UserID == utils.UUID(userID) &&
				arg.Name == "lecture-notes.pdf" &&
				arg.MimeType == db.MimeTypeApplicationPdf &&
				arg.Size == int64(1048576) &&
				arg.S3Key != ""
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

	s3Mock.EXPECT().
		GeneratePresignedPutURL(mock.Anything, mock.AnythingOfType("string"), "application/pdf", int64(1048576)).
		Return("https://s3.example.com/presigned-url", nil)

	result, err := svc.CreateFile(context.Background(), params)
	require.NoError(t, err)

	fileID, err := utils.PgxToGoogleUUID(capturedID)
	require.NoError(t, err)

	expectedKey := fmt.Sprintf("users/%s/files/%s/lecture-notes.pdf", userID.String(), fileID.String())
	assert.Equal(t, expectedKey, capturedS3Key, "S3 key stored in DB must include file ID and name")

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
	now := time.Now()

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "file.pdf",
		MimeType: "application/pdf",
		Size:     100,
	}

	repo.EXPECT().
		InsertFile(mock.Anything, mock.Anything).
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

	s3Mock.EXPECT().
		GeneratePresignedPutURL(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("", fmt.Errorf("s3 error"))

	_, err := svc.CreateFile(context.Background(), params)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CreateFile: presign")
}

func TestService_CreateFile_PathTraversal_Rejected(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{"empty name", ""},
		{"dot only", "."},
		{"slash only", "/"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mock_files.NewMockRepository(t)
			s3Mock := mock_files.NewMockS3Uploader(t)
			svc := files.NewService(repo, s3Mock)

			params := files.CreateFileParams{
				UserID:   uuid.New(),
				Name:     tc.fileName,
				MimeType: "application/pdf",
				Size:     100,
			}

			_, err := svc.CreateFile(context.Background(), params)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid input")
		})
	}
}

func TestService_CreateFile_PathTraversal_Sanitized(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	s3Mock := mock_files.NewMockS3Uploader(t)
	svc := files.NewService(repo, s3Mock)

	userID := uuid.New()
	now := time.Now()

	params := files.CreateFileParams{
		UserID:   userID,
		Name:     "../../admin/secrets.pdf",
		MimeType: "application/pdf",
		Size:     100,
	}

	var capturedS3Key string

	repo.EXPECT().
		InsertFile(mock.Anything, mock.MatchedBy(func(arg db.InsertFileParams) bool {
			capturedS3Key = arg.S3Key
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

	s3Mock.EXPECT().
		GeneratePresignedPutURL(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return("https://s3.example.com/presigned-url", nil)

	_, err := svc.CreateFile(context.Background(), params)
	require.NoError(t, err)
	assert.Contains(t, capturedS3Key, "/secrets.pdf", "path traversal should be stripped to base name")
	assert.NotContains(t, capturedS3Key, "..", "S3 key must not contain directory traversal")
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

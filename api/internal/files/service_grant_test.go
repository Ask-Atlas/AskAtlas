package files_test

import (
	"context"
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

func TestService_CreateGrant_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	granteeID := uuid.New()
	grantID := uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().
		GetFileByOwner(mock.Anything, mock.MatchedBy(func(arg db.GetFileByOwnerParams) bool {
			return arg.FileID == utils.UUID(fileID) && arg.OwnerID == utils.UUID(ownerID)
		})).
		Return(db.GetFileByOwnerRow{ID: utils.UUID(fileID)}, nil)

	repo.EXPECT().
		UpsertFileGrant(mock.Anything, mock.MatchedBy(func(arg db.UpsertFileGrantParams) bool {
			return arg.FileID == utils.UUID(fileID) &&
				arg.GranteeType == db.GranteeTypeUser &&
				arg.GranteeID == utils.UUID(granteeID) &&
				arg.Permission == db.PermissionView &&
				arg.GrantedBy == utils.UUID(ownerID)
		})).
		Return(db.FileGrant{
			ID:          utils.UUID(grantID),
			FileID:      utils.UUID(fileID),
			GranteeType: db.GranteeTypeUser,
			GranteeID:   utils.UUID(granteeID),
			Permission:  db.PermissionView,
			GrantedBy:   utils.UUID(ownerID),
			CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		}, nil)

	grant, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   granteeID,
		Permission:  "view",
	})
	require.NoError(t, err)

	assert.Equal(t, grantID, grant.ID)
	assert.Equal(t, fileID, grant.FileID)
	assert.Equal(t, "user", grant.GranteeType)
	assert.Equal(t, granteeID, grant.GranteeID)
	assert.Equal(t, "view", grant.Permission)
	assert.Equal(t, ownerID, grant.GrantedBy)
}

func TestService_CreateGrant_FileNotOwned(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().
		GetFileByOwner(mock.Anything, mock.Anything).
		Return(db.GetFileByOwnerRow{}, apperrors.ErrNotFound)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_CreateGrant_UpsertError(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()

	repo.EXPECT().
		GetFileByOwner(mock.Anything, mock.Anything).
		Return(db.GetFileByOwnerRow{ID: utils.UUID(fileID)}, nil)

	repo.EXPECT().
		UpsertFileGrant(mock.Anything, mock.Anything).
		Return(db.FileGrant{}, assert.AnError)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestService_RevokeGrant_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	granteeID := uuid.New()

	repo.EXPECT().
		GetFileByOwner(mock.Anything, mock.MatchedBy(func(arg db.GetFileByOwnerParams) bool {
			return arg.FileID == utils.UUID(fileID) && arg.OwnerID == utils.UUID(ownerID)
		})).
		Return(db.GetFileByOwnerRow{ID: utils.UUID(fileID)}, nil)

	repo.EXPECT().
		RevokeFileGrant(mock.Anything, mock.MatchedBy(func(arg db.RevokeFileGrantParams) bool {
			return arg.FileID == utils.UUID(fileID) &&
				arg.GranteeType == db.GranteeTypeCourse &&
				arg.GranteeID == utils.UUID(granteeID) &&
				arg.Permission == db.PermissionShare
		})).
		Return(nil)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "course",
		GranteeID:   granteeID,
		Permission:  "share",
	})
	assert.NoError(t, err)
}

func TestService_RevokeGrant_FileNotOwned(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().
		GetFileByOwner(mock.Anything, mock.Anything).
		Return(db.GetFileByOwnerRow{}, apperrors.ErrNotFound)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_RevokeGrant_RepoError(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()

	repo.EXPECT().
		GetFileByOwner(mock.Anything, mock.Anything).
		Return(db.GetFileByOwnerRow{ID: utils.UUID(fileID)}, nil)

	repo.EXPECT().
		RevokeFileGrant(mock.Anything, mock.Anything).
		Return(assert.AnError)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "delete",
	})
	assert.Error(t, err)
}

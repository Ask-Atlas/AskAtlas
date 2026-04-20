package files_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	mock_files "github.com/Ask-Atlas/AskAtlas/api/internal/files/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------
// CreateGrant -- ASK-122 contract.
//
// Validation order under test:
//   1. enum validation         -> 400 with details for both fields
//   2. file existence (probe)  -> 404 (missing or soft-deleted)
//   3. file ownership          -> 403
//   4. grantee existence       -> 400 with grantee_id detail
//   5. INSERT + 23505 dup      -> 409
// ----------------------------------------------------------------------

// fileForUpdate returns a canned GetFileForUpdate row owned by the
// given user. Used to set up the existence/ownership probe in the
// happy-path tests.
func fileForUpdate(fileID, ownerID uuid.UUID) db.GetFileForUpdateRow {
	return db.GetFileForUpdateRow{
		ID:     utils.UUID(fileID),
		UserID: utils.UUID(ownerID),
		Status: "complete",
	}
}

func TestService_CreateGrant_UserGrantee_Success(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	granteeID := uuid.New()
	grantID := uuid.New()
	now := time.Now().UTC()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().CheckUserExists(mock.Anything, utils.UUID(granteeID)).Return(nil)
	repo.EXPECT().
		InsertFileGrant(mock.Anything, mock.MatchedBy(func(arg db.InsertFileGrantParams) bool {
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
	assert.Equal(t, "user", grant.GranteeType)
	assert.Equal(t, "view", grant.Permission)
}

func TestService_CreateGrant_PublicSentinel_SkipsUserLookup(t *testing.T) {
	// AC2: grantee_id = NIL_UUID + grantee_type=user must NOT call
	// CheckUserExists. Mockery would fail the test on any unexpected
	// call.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().InsertFileGrant(mock.Anything, mock.Anything).
		Return(db.FileGrant{
			ID:          utils.UUID(uuid.New()),
			FileID:      utils.UUID(fileID),
			GranteeType: db.GranteeTypeUser,
			GranteeID:   utils.UUID(uuid.UUID{}), // sentinel
			Permission:  db.PermissionView,
			GrantedBy:   utils.UUID(ownerID),
			CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   uuid.UUID{}, // NIL UUID = public sentinel
		Permission:  "view",
	})
	require.NoError(t, err)
}

func TestService_CreateGrant_PublicSentinel_OnlyExemptsUserType(t *testing.T) {
	// AC2 corollary: the sentinel exemption applies ONLY when
	// grantee_type=user. With grantee_type=course, the NIL UUID
	// flows into CheckCourseExists which returns ErrNotFound -> 400.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().CheckCourseExists(mock.Anything, utils.UUID(uuid.UUID{})).
		Return(fmt.Errorf("CheckCourseExists: %w", apperrors.ErrNotFound))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "course",
		GranteeID:   uuid.UUID{},
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["grantee_id"], "no course with this ID")
}

func TestService_CreateGrant_CourseGrantee_NotFound_400(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	courseID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().CheckCourseExists(mock.Anything, utils.UUID(courseID)).
		Return(fmt.Errorf("CheckCourseExists: %w", apperrors.ErrNotFound))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "course",
		GranteeID:   courseID,
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Equal(t, "no course with this ID", appErr.Details["grantee_id"])
}

func TestService_CreateGrant_StudyGuideGrantee_SoftDeleted_400(t *testing.T) {
	// AC5: a soft-deleted study guide must surface as 400 with the
	// grantee_id detail (CheckStudyGuideExists filters deleted_at IS
	// NULL, so soft-deleted guides yield ErrNoRows).
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	guideID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().CheckStudyGuideExists(mock.Anything, utils.UUID(guideID)).
		Return(fmt.Errorf("CheckStudyGuideExists: %w", apperrors.ErrNotFound))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "study_guide",
		GranteeID:   guideID,
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Equal(t, "no study_guide with this ID", appErr.Details["grantee_id"])
}

func TestService_CreateGrant_DuplicateGrant_409(t *testing.T) {
	// AC6: existing grant -> 409. The InsertFileGrant query fails
	// with a pgx PgError carrying sqlstate 23505.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	granteeID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().CheckUserExists(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().InsertFileGrant(mock.Anything, mock.Anything).
		Return(db.FileGrant{}, fmt.Errorf("InsertFileGrant: %w", &pgconn.PgError{Code: "23505"}))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   granteeID,
		Permission:  "view",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestService_CreateGrant_NotOwner_403(t *testing.T) {
	// AC7: caller is not the file owner. GetFileForUpdate succeeds
	// but the returned user_id mismatches; service returns 403
	// without ever touching CheckUserExists or InsertFileGrant.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	owner := uuid.New()
	caller := uuid.New() // different

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, owner), nil)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     caller,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

func TestService_CreateGrant_FileNotFound_404(t *testing.T) {
	// AC8: file missing or soft-deleted. GetFileForUpdate filters
	// soft-deleted at the SQL layer.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(db.GetFileForUpdateRow{}, fmt.Errorf("GetFileForUpdate: %w", apperrors.ErrNotFound))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_CreateGrant_BadGranteeType_400(t *testing.T) {
	// Enum validation runs BEFORE the file probe -- the mock has no
	// expectations so any DB call would fail the test.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "organization",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Contains(t, appErr.Details["grantee_type"], "user")
}

func TestService_CreateGrant_BadPermission_400(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "edit",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Contains(t, appErr.Details["permission"], "view")
}

func TestService_CreateGrant_BothEnumsBad_DetailsForBoth(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "team",
		GranteeID:   uuid.New(),
		Permission:  "admin",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.NotEmpty(t, appErr.Details["grantee_type"])
	assert.NotEmpty(t, appErr.Details["permission"])
}

func TestService_CreateGrant_UserGranteeMissing_400(t *testing.T) {
	// AC9: grantee_type=user, grantee_id is a real UUID but absent
	// from the users table (and is NOT the public sentinel).
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	granteeID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().CheckUserExists(mock.Anything, utils.UUID(granteeID)).
		Return(fmt.Errorf("CheckUserExists: %w", apperrors.ErrNotFound))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   granteeID,
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.Equal(t, "no user with this ID", appErr.Details["grantee_id"])
}

func TestService_CreateGrant_InsertReal500(t *testing.T) {
	// Non-23505 DB error from InsertFileGrant must propagate
	// (handler maps to 500). Owner == caller so we reach the insert.
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	caller := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, caller), nil)
	repo.EXPECT().CheckUserExists(mock.Anything, mock.Anything).Return(nil)
	repo.EXPECT().InsertFileGrant(mock.Anything, mock.Anything).
		Return(db.FileGrant{}, errors.New("connection lost"))

	_, err := svc.CreateGrant(context.Background(), files.CreateGrantParams{
		FileID:      fileID,
		OwnerID:     caller,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, apperrors.ErrConflict)
}

// ----------------------------------------------------------------------
// RevokeGrant -- ASK-125 contract.
// ----------------------------------------------------------------------

func TestService_RevokeGrant_Success_204(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()
	granteeID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, utils.UUID(fileID)).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().
		RevokeFileGrant(mock.Anything, mock.MatchedBy(func(arg db.RevokeFileGrantParams) bool {
			return arg.FileID == utils.UUID(fileID) &&
				arg.GranteeType == db.GranteeTypeCourse &&
				arg.GranteeID == utils.UUID(granteeID) &&
				arg.Permission == db.PermissionShare
		})).
		Return(int64(1), nil)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "course",
		GranteeID:   granteeID,
		Permission:  "share",
	})
	require.NoError(t, err)
}

func TestService_RevokeGrant_GrantMissing_404(t *testing.T) {
	// AC2: no matching grant -> 404 (NOT idempotent no-op).
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().RevokeFileGrant(mock.Anything, mock.Anything).Return(int64(0), nil)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_RevokeGrant_NotOwner_403(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	owner := uuid.New()
	caller := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(fileForUpdate(fileID, owner), nil)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      fileID,
		OwnerID:     caller,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 403, appErr.Code)
}

func TestService_RevokeGrant_FileNotFound_404(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(db.GetFileForUpdateRow{}, fmt.Errorf("GetFileForUpdate: %w", apperrors.ErrNotFound))

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "view",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestService_RevokeGrant_BadEnums_400(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      uuid.New(),
		OwnerID:     uuid.New(),
		GranteeType: "team",
		GranteeID:   uuid.New(),
		Permission:  "admin",
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)
	assert.NotEmpty(t, appErr.Details["grantee_type"])
	assert.NotEmpty(t, appErr.Details["permission"])
}

func TestService_RevokeGrant_RepoError_500(t *testing.T) {
	repo := mock_files.NewMockRepository(t)
	svc := files.NewService(repo)

	fileID := uuid.New()
	ownerID := uuid.New()

	repo.EXPECT().GetFileForUpdate(mock.Anything, mock.Anything).
		Return(fileForUpdate(fileID, ownerID), nil)
	repo.EXPECT().RevokeFileGrant(mock.Anything, mock.Anything).
		Return(int64(0), errors.New("connection lost"))

	err := svc.RevokeGrant(context.Background(), files.RevokeGrantParams{
		FileID:      fileID,
		OwnerID:     ownerID,
		GranteeType: "user",
		GranteeID:   uuid.New(),
		Permission:  "delete",
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, apperrors.ErrNotFound)
}

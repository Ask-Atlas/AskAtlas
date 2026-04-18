package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newTestRouter builds a chi router with the composite handler wired up so
// the oapi-codegen path parameters are resolved correctly.
func newTestRouter(t *testing.T, fh *handlers.FileHandler, gh *handlers.GrantHandler) chi.Router {
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestGrantHandler_CreateGrant_Success(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	userID := uuid.New()
	fileID := uuid.New()
	granteeID := uuid.New()
	grantID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	body := api.CreateGrantRequest{
		GranteeType: api.CreateGrantRequestGranteeTypeUser,
		GranteeId:   granteeID,
		Permission:  api.CreateGrantRequestPermissionView,
	}
	bodyBytes, _ := json.Marshal(body)

	mockGrantSvc.EXPECT().
		CreateGrant(mock.Anything, mock.MatchedBy(func(p files.CreateGrantParams) bool {
			return p.FileID == fileID &&
				p.OwnerID == userID &&
				p.GranteeType == "user" &&
				p.GranteeID == granteeID &&
				p.Permission == "view"
		})).
		Return(files.Grant{
			ID:          grantID,
			FileID:      fileID,
			GranteeType: "user",
			GranteeID:   granteeID,
			Permission:  "view",
			GrantedBy:   userID,
			CreatedAt:   now,
		}, nil)

	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp api.GrantResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, grantID, uuid.UUID(resp.Id))
	assert.Equal(t, fileID, uuid.UUID(resp.FileId))
	assert.Equal(t, "user", resp.GranteeType)
	assert.Equal(t, granteeID, uuid.UUID(resp.GranteeId))
	assert.Equal(t, "view", resp.Permission)
	assert.Equal(t, userID, uuid.UUID(resp.GrantedBy))
}

func TestGrantHandler_CreateGrant_Unauthorized(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	fileID := uuid.New()
	body := api.CreateGrantRequest{
		GranteeType: api.CreateGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.CreateGrantRequestPermissionView,
	}
	bodyBytes, _ := json.Marshal(body)

	// No user in context
	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGrantHandler_CreateGrant_FileNotFound(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	userID := uuid.New()
	fileID := uuid.New()

	body := api.CreateGrantRequest{
		GranteeType: api.CreateGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.CreateGrantRequestPermissionView,
	}
	bodyBytes, _ := json.Marshal(body)

	mockGrantSvc.EXPECT().
		CreateGrant(mock.Anything, mock.Anything).
		Return(files.Grant{}, apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGrantHandler_CreateGrant_ServiceError(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	userID := uuid.New()
	fileID := uuid.New()

	body := api.CreateGrantRequest{
		GranteeType: api.CreateGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.CreateGrantRequestPermissionView,
	}
	bodyBytes, _ := json.Marshal(body)

	mockGrantSvc.EXPECT().
		CreateGrant(mock.Anything, mock.Anything).
		Return(files.Grant{}, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGrantHandler_RevokeGrant_Success(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	userID := uuid.New()
	fileID := uuid.New()
	granteeID := uuid.New()

	body := api.RevokeGrantRequest{
		GranteeType: api.RevokeGrantRequestGranteeTypeCourse,
		GranteeId:   granteeID,
		Permission:  api.RevokeGrantRequestPermissionShare,
	}
	bodyBytes, _ := json.Marshal(body)

	mockGrantSvc.EXPECT().
		RevokeGrant(mock.Anything, mock.MatchedBy(func(p files.RevokeGrantParams) bool {
			return p.FileID == fileID &&
				p.OwnerID == userID &&
				p.GranteeType == "course" &&
				p.GranteeID == granteeID &&
				p.Permission == "share"
		})).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGrantHandler_RevokeGrant_Unauthorized(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	fileID := uuid.New()
	body := api.RevokeGrantRequest{
		GranteeType: api.RevokeGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.RevokeGrantRequestPermissionView,
	}
	bodyBytes, _ := json.Marshal(body)

	// No user in context
	req := httptest.NewRequest(http.MethodDelete, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGrantHandler_RevokeGrant_FileNotFound(t *testing.T) {
	mockFileSvc := mock_handlers.NewMockFileService(t)
	mockGrantSvc := mock_handlers.NewMockGrantService(t)
	fh := handlers.NewFileHandler(mockFileSvc, nil)
	gh := handlers.NewGrantHandler(mockGrantSvc)

	userID := uuid.New()
	fileID := uuid.New()

	body := api.RevokeGrantRequest{
		GranteeType: api.RevokeGrantRequestGranteeTypeUser,
		GranteeId:   uuid.New(),
		Permission:  api.RevokeGrantRequestPermissionView,
	}
	bodyBytes, _ := json.Marshal(body)

	mockGrantSvc.EXPECT().
		RevokeGrant(mock.Anything, mock.Anything).
		Return(apperrors.ErrNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/files/"+fileID.String()+"/grants", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := newTestRouter(t, fh, gh)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

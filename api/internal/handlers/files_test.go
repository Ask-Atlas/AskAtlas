package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
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

func TestFileHandler_ListFiles_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/me/files?page_limit=10", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})

	returnedFiles := []files.File{
		{ID: uuid.New(), Name: "file1.txt", Status: "complete"},
		{ID: uuid.New(), Name: "file2.txt", Status: "pending"},
	}
	nextCursor := "some-cursor"

	mockSvc.EXPECT().ListFiles(mock.Anything, mock.MatchedBy(func(p files.ListFilesParams) bool {
		return p.PageLimit == 10 && p.ViewerID == userID
	})).Return(returnedFiles, &nextCursor, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp api.ListFilesResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Len(t, resp.Files, 2)
	assert.True(t, resp.HasMore)
	assert.Equal(t, &nextCursor, resp.NextCursor)
}

func TestFileHandler_ListFiles_InvalidParams(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/me/files?page_limit=abc", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_ListFiles_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/me/files", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().
		ListFiles(mock.Anything, mock.Anything).
		Return(nil, nil, errors.New("svc error"))

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFileHandler_UpdateFile_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	fileID := uuid.New()
	now := time.Now()

	body := `{"name": "renamed.pdf"}`
	req := httptest.NewRequest(http.MethodPatch, "/files/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().UpdateFile(mock.Anything, mock.MatchedBy(func(p files.UpdateFileParams) bool {
		return p.FileID == fileID && p.OwnerID == userID && p.Name == "renamed.pdf"
	})).Return(files.File{
		ID:        fileID,
		UserID:    userID,
		Name:      "renamed.pdf",
		Size:      1024,
		MimeType:  "application/pdf",
		Status:    "complete",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil)

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp api.FileResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "renamed.pdf", resp.Name)
	assert.Equal(t, int64(1024), resp.Size)
}

func TestFileHandler_UpdateFile_NotFound(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	fileID := uuid.New()

	body := `{"name": "renamed.pdf"}`
	req := httptest.NewRequest(http.MethodPatch, "/files/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().UpdateFile(mock.Anything, mock.Anything).
		Return(files.File{}, apperrors.ErrNotFound)

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestFileHandler_UpdateFile_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	fileID := uuid.New()

	body := `{"name": "renamed.pdf"}`
	req := httptest.NewRequest(http.MethodPatch, "/files/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().UpdateFile(mock.Anything, mock.Anything).
		Return(files.File{}, errors.New("db connection failed"))

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestFileHandler_UpdateFile_ValidationError(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	fileID := uuid.New()

	body := `{"name": "renamed.pdf"}`
	req := httptest.NewRequest(http.MethodPatch, "/files/"+fileID.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().UpdateFile(mock.Anything, mock.Anything).
		Return(files.File{}, apperrors.NewBadRequest("Invalid file name", map[string]string{
			"name": "contains invalid characters: /",
		}))

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp apperrors.AppError
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Details["name"], "/")
}

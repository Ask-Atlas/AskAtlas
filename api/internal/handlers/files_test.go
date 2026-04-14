package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFileHandler_CreateFile_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	fileID := uuid.New()

	body := `{"name":"lecture-notes.pdf","mime_type":"application/pdf","size":1048576}`
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().CreateFile(mock.Anything, mock.MatchedBy(func(p files.CreateFileParams) bool {
		return p.UserID == userID &&
			p.Name == "lecture-notes.pdf" &&
			p.MimeType == "application/pdf" &&
			p.Size == 1048576
	})).Return(files.CreateFileResult{
		File: files.File{
			ID:       fileID,
			UserID:   userID,
			Name:     "lecture-notes.pdf",
			Size:     1048576,
			MimeType: "application/pdf",
			Status:   "pending",
		},
		UploadURL: "https://s3.example.com/presigned-url",
	}, nil)

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp api.CreateFileResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "lecture-notes.pdf", resp.File.Name)
	assert.Equal(t, int64(1048576), resp.File.Size)
	assert.Equal(t, "application/pdf", resp.File.MimeType)
	assert.Equal(t, "pending", resp.File.Status)
	assert.Equal(t, "https://s3.example.com/presigned-url", resp.UploadUrl)
}

func TestFileHandler_CreateFile_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	body := `{"name":"file.pdf","mime_type":"application/pdf","size":100}`
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// no auth context
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFileHandler_CreateFile_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	body := `{"name":"file.pdf","mime_type":"application/pdf","size":100}`
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().
		CreateFile(mock.Anything, mock.Anything).
		Return(files.CreateFileResult{}, errors.New("internal error"))

	r := chi.NewRouter()
	api.HandlerWithOptions(h, api.ChiServerOptions{BaseRouter: r})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

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

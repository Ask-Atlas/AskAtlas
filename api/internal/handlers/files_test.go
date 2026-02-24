package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFileHandler_ListFiles_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/files?page_limit=10", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	returnedFiles := []files.File{
		{ID: uuid.New(), Name: "file1.txt", Status: "complete"},
		{ID: uuid.New(), Name: "file2.txt", Status: "pending"},
	}
	nextCursor := "some-cursor"

	mockSvc.EXPECT().ListFiles(mock.Anything, mock.MatchedBy(func(p files.ListFilesParams) bool {
		return p.PageLimit == 10 && p.ViewerID == userID
	})).Return(returnedFiles, &nextCursor, nil)

	h.ListFiles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp files.ListFilesResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Len(t, resp.Files, 2)
	assert.True(t, resp.HasMore)
	assert.Equal(t, &nextCursor, resp.NextCursor)
}

func TestFileHandler_ListFiles_InvalidParams(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/files?page_limit=0", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.ListFiles(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_ListFiles_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().
		ListFiles(mock.Anything, mock.Anything).
		Return(nil, nil, errors.New("svc error"))

	h.ListFiles(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

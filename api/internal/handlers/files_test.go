package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func fileTestRouter(t *testing.T, fh *handlers.FileHandler) chi.Router {
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	composite := handlers.NewCompositeHandler(fh, gh)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

func TestFileHandler_ListFiles_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/me/files?page_limit=10", nil)
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	r := fileTestRouter(t, h)

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

	r := fileTestRouter(t, h)
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

	r := fileTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

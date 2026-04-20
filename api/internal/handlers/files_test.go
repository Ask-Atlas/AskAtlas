package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func fileTestRouter(t *testing.T, fh *handlers.FileHandler) chi.Router {
	gh := handlers.NewGrantHandler(mock_handlers.NewMockGrantService(t))
	sh := handlers.NewSchoolsHandler(mock_handlers.NewMockSchoolService(t))
	ch := handlers.NewCoursesHandler(mock_handlers.NewMockCourseService(t))
	sgh := handlers.NewStudyGuideHandler(mock_handlers.NewMockStudyGuideService(t))
	qh := handlers.NewQuizzesHandler(mock_handlers.NewMockQuizService(t))
	ssh := handlers.NewSessionsHandler(mock_handlers.NewMockSessionService(t))
	composite := handlers.NewCompositeHandler(fh, gh, sh, ch, sgh, qh, ssh, nil, nil, nil)
	r := chi.NewRouter()
	api.HandlerWithOptions(composite, api.ChiServerOptions{BaseRouter: r})
	return r
}

// Helper: a valid POST /files body with all 4 required fields per
// the new (ASK-105) contract. Tests that need to flip ONE field to
// invalid use this as the baseline + override the bad field.
const validCreateFileBody = `{"name":"lecture-notes.pdf","mime_type":"application/pdf","size":1048576,"s3_key":"uploads/abc123/lecture-notes.pdf"}`

func TestFileHandler_CreateFile_Success(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	fileID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(validCreateFileBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().CreateFile(mock.Anything, mock.MatchedBy(func(p files.CreateFileParams) bool {
		// Verifies the handler passes the caller-supplied s3_key
		// straight through (no rewrite, no server-gen).
		return p.UserID == userID &&
			p.Name == "lecture-notes.pdf" &&
			p.MimeType == "application/pdf" &&
			p.Size == 1048576 &&
			p.S3Key == "uploads/abc123/lecture-notes.pdf"
	})).Return(files.File{
		ID:       fileID,
		UserID:   userID,
		Name:     "lecture-notes.pdf",
		Size:     1048576,
		MimeType: "application/pdf",
		Status:   "pending",
	}, nil)

	r := fileTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// New contract: 201 returns the bare FileResponse, no upload_url
	// wrapper. The Next.js server already has the s3_key it generated
	// (and presigned separately), so the API doesn't echo it.
	var resp api.FileResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "lecture-notes.pdf", resp.Name)
	assert.Equal(t, int64(1048576), resp.Size)
	assert.Equal(t, "application/pdf", resp.MimeType)
	assert.Equal(t, "pending", resp.Status)
}

func TestFileHandler_CreateFile_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		// Each case overrides exactly one field of the otherwise-
		// valid baseline so the error attribution is unambiguous.
		{"empty name", `{"name":"","mime_type":"application/pdf","size":100,"s3_key":"uploads/x/y.pdf"}`},
		{"whitespace-only name", `{"name":"   ","mime_type":"application/pdf","size":100,"s3_key":"uploads/x/y.pdf"}`},
		{"name exceeds 255 chars", `{"name":"` + strings.Repeat("a", 256) + `","mime_type":"application/pdf","size":100,"s3_key":"uploads/x/y.pdf"}`},
		{"invalid mime type", `{"name":"file.pdf","mime_type":"application/zip","size":100,"s3_key":"uploads/x/y.pdf"}`},
		{"zero size", `{"name":"file.pdf","mime_type":"application/pdf","size":0,"s3_key":"uploads/x/y.pdf"}`},
		{"negative size", `{"name":"file.pdf","mime_type":"application/pdf","size":-1,"s3_key":"uploads/x/y.pdf"}`},
		{"size exceeds max", `{"name":"file.pdf","mime_type":"application/pdf","size":104857601,"s3_key":"uploads/x/y.pdf"}`},
		{"missing s3_key", `{"name":"file.pdf","mime_type":"application/pdf","size":100}`},
		{"empty s3_key", `{"name":"file.pdf","mime_type":"application/pdf","size":100,"s3_key":""}`},
		{"s3_key exceeds 1024 chars", `{"name":"file.pdf","mime_type":"application/pdf","size":100,"s3_key":"` + strings.Repeat("a", 1025) + `"}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := mock_handlers.NewMockFileService(t)
			h := handlers.NewFileHandler(mockSvc, nil)

			userID := uuid.New()
			req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			ctx := authctx.WithUserID(req.Context(), userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			r := fileTestRouter(t, h)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestFileHandler_CreateFile_Unauthorized(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(validCreateFileBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := fileTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFileHandler_CreateFile_ServiceError(t *testing.T) {
	mockSvc := mock_handlers.NewMockFileService(t)
	h := handlers.NewFileHandler(mockSvc, nil)

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/files", strings.NewReader(validCreateFileBody))
	req.Header.Set("Content-Type", "application/json")
	ctx := authctx.WithUserID(req.Context(), userID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.EXPECT().
		CreateFile(mock.Anything, mock.Anything).
		Return(files.File{}, errors.New("internal error"))

	r := fileTestRouter(t, h)
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

	r := fileTestRouter(t, h)
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

	r := fileTestRouter(t, h)
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

	r := fileTestRouter(t, h)
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

	r := fileTestRouter(t, h)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp apperrors.AppError
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Details["name"], "/")
}

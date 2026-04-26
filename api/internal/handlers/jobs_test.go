package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	qstashclient "github.com/Ask-Atlas/AskAtlas/api/internal/qstash"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newExtractRequest(t *testing.T, msg qstashclient.ExtractFileMessage) *http.Request {
	t.Helper()
	body, err := json.Marshal(msg)
	require.NoError(t, err)
	return httptest.NewRequest(http.MethodPost, "/jobs/extract-file", bytes.NewReader(body))
}

func TestExtractFileJob_HappyPath(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor)

	fileID := uuid.New()
	extractor.EXPECT().Process(mock.Anything, fileID).Return(nil)

	req := newExtractRequest(t, qstashclient.ExtractFileMessage{
		FileID: fileID.String(), S3Key: "k", MimeType: "application/pdf",
	})
	w := httptest.NewRecorder()
	h.ExtractFileJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractFileJob_BadFileID(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor)

	req := newExtractRequest(t, qstashclient.ExtractFileMessage{FileID: "not-a-uuid"})
	w := httptest.NewRecorder()
	h.ExtractFileJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	extractor.AssertNotCalled(t, "Process", mock.Anything, mock.Anything)
}

func TestExtractFileJob_BadJSON(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor)

	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file",
		bytes.NewBufferString(`{"file_id`)) // truncated JSON
	w := httptest.NewRecorder()
	h.ExtractFileJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	extractor.AssertNotCalled(t, "Process", mock.Anything, mock.Anything)
}

func TestExtractFileJob_WorkerErrorYields500(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor)

	fileID := uuid.New()
	extractor.EXPECT().Process(mock.Anything, fileID).Return(errors.New("s3 down"))

	req := newExtractRequest(t, qstashclient.ExtractFileMessage{
		FileID: fileID.String(), S3Key: "k", MimeType: "text/plain",
	})
	w := httptest.NewRecorder()
	h.ExtractFileJob(w, req)

	// 5xx triggers QStash retry -- contract honored.
	assert.GreaterOrEqual(t, w.Code, 500)
}

func TestExtractFileFailedJob_MarksRowFailed(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor)

	fileID := uuid.New()
	extractor.EXPECT().MarkFailed(mock.Anything, fileID, mock.MatchedBy(func(reason string) bool {
		return reason != "" // contract: non-empty status_error
	})).Return(nil)

	body, _ := json.Marshal(qstashclient.ExtractFileMessage{
		FileID: fileID.String(), MimeType: "application/pdf",
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file-failed",
		bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.ExtractFileFailedJob(w, req)

	// Always 200 -- QStash has already given up; we don't want it
	// retrying the failure-callback itself.
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractFileFailedJob_BadJSONStill200(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor)

	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file-failed",
		bytes.NewBufferString(`{`))
	w := httptest.NewRecorder()
	h.ExtractFileFailedJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	extractor.AssertNotCalled(t, "MarkFailed", mock.Anything, mock.Anything, mock.Anything)
}

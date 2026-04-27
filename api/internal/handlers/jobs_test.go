package handlers_test

import (
	"bytes"
	"encoding/base64"
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

// wrapAsFailureCallback returns the QStash failure-callback envelope
// for msg. QStash POSTs this shape -- not the original message --
// to FailureCallback URLs:
//
//	{"sourceMessageId":"...", "sourceBody":"<base64>", "status":..., "retried":...}
//
// See https://upstash.com/docs/qstash/features/callbacks. The earlier
// version of these tests sent the raw message body, which mirrored the
// handler's incorrect assumption and masked the bug in production.
func wrapAsFailureCallback(t *testing.T, msg any, sourceMessageID string, status, retried int) []byte {
	t.Helper()
	inner, err := json.Marshal(msg)
	require.NoError(t, err)
	envelope, err := json.Marshal(map[string]any{
		"sourceMessageId": sourceMessageID,
		"sourceBody":      base64.StdEncoding.EncodeToString(inner),
		"status":          status,
		"retried":         retried,
	})
	require.NoError(t, err)
	return envelope
}

func newExtractRequest(t *testing.T, msg qstashclient.ExtractFileMessage) *http.Request {
	t.Helper()
	body, err := json.Marshal(msg)
	require.NoError(t, err)
	return httptest.NewRequest(http.MethodPost, "/jobs/extract-file", bytes.NewReader(body))
}

func TestExtractFileJob_HappyPath(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

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
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

	req := newExtractRequest(t, qstashclient.ExtractFileMessage{FileID: "not-a-uuid"})
	w := httptest.NewRecorder()
	h.ExtractFileJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	extractor.AssertNotCalled(t, "Process", mock.Anything, mock.Anything)
}

func TestExtractFileJob_BadJSON(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file",
		bytes.NewBufferString(`{"file_id`)) // truncated JSON
	w := httptest.NewRecorder()
	h.ExtractFileJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	extractor.AssertNotCalled(t, "Process", mock.Anything, mock.Anything)
}

func TestExtractFileJob_WorkerErrorYields500(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

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

func TestExtractFileFailedJob_UnwrapsEnvelopeAndMarksRowFailed(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

	fileID := uuid.New()
	// Reason should embed the source message id + final HTTP status
	// so ops triage can correlate to the QStash dashboard.
	extractor.EXPECT().MarkFailed(mock.Anything, fileID, mock.MatchedBy(func(reason string) bool {
		return reason != "" &&
			contains(reason, "source_msg=msg_xyz") &&
			contains(reason, "http_status=502") &&
			contains(reason, "mime=application/pdf")
	})).Return(nil)

	envelope := wrapAsFailureCallback(t, qstashclient.ExtractFileMessage{
		FileID: fileID.String(), MimeType: "application/pdf",
	}, "msg_xyz", 502, 3)

	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file-failed",
		bytes.NewReader(envelope))
	w := httptest.NewRecorder()
	h.ExtractFileFailedJob(w, req)

	// Always 200 -- QStash has already given up; we don't want it
	// retrying the failure-callback itself.
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestExtractFileFailedJob_BadEnvelopeJSONStill200(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file-failed",
		bytes.NewBufferString(`{`))
	w := httptest.NewRecorder()
	h.ExtractFileFailedJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	extractor.AssertNotCalled(t, "MarkFailed", mock.Anything, mock.Anything, mock.Anything)
}

// TestExtractFileFailedJob_BadInnerJSONStill200 covers the case where
// the envelope decodes fine but sourceBody is not valid JSON for our
// message type. Pre-fix the handler decoded the envelope itself as
// ExtractFileMessage (zero-value fields) and silently returned 200
// without ever calling MarkFailed -- the row was stuck in extracting
// forever. With the fix we reach the inner Unmarshal, which fails,
// and we still 200 (no retry on the callback) but explicitly do NOT
// invoke MarkFailed (which would mark with a zero-uuid file_id).
func TestExtractFileFailedJob_BadInnerJSONStill200(t *testing.T) {
	extractor := mock_handlers.NewMockFileExtractor(t)
	h := handlers.NewJobHandler(nil, nil, extractor, nil)

	// Envelope is valid JSON; sourceBody is base64 of garbage that
	// won't unmarshal into ExtractFileMessage. (Empty {} is technically
	// valid JSON for the message; use raw text instead.)
	envelope, _ := json.Marshal(map[string]any{
		"sourceMessageId": "msg_bad",
		"sourceBody":      base64.StdEncoding.EncodeToString([]byte("not-json")),
		"status":          500,
		"retried":         3,
	})
	req := httptest.NewRequest(http.MethodPost, "/jobs/extract-file-failed",
		bytes.NewReader(envelope))
	w := httptest.NewRecorder()
	h.ExtractFileFailedJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	extractor.AssertNotCalled(t, "MarkFailed", mock.Anything, mock.Anything, mock.Anything)
}

// contains is a tiny test helper to keep the matcher readable; we
// avoid pulling in strings.Contains directly in a mock.MatchedBy
// closure for clarity.
func contains(haystack, needle string) bool {
	return bytes.Contains([]byte(haystack), []byte(needle))
}

// --- ASK-221 ChunkEmbedFile* handler tests ---

func newChunkEmbedRequest(t *testing.T, msg qstashclient.ChunkEmbedFileMessage) *http.Request {
	t.Helper()
	body, err := json.Marshal(msg)
	require.NoError(t, err)
	return httptest.NewRequest(http.MethodPost, "/jobs/chunk-embed-file", bytes.NewReader(body))
}

func TestChunkEmbedFileJob_HappyPath(t *testing.T) {
	chunkEmbedder := mock_handlers.NewMockFileChunkEmbedder(t)
	h := handlers.NewJobHandler(nil, nil, nil, chunkEmbedder)

	fileID := uuid.New()
	chunkEmbedder.EXPECT().Process(mock.Anything, fileID).Return(nil)

	req := newChunkEmbedRequest(t, qstashclient.ChunkEmbedFileMessage{FileID: fileID.String()})
	w := httptest.NewRecorder()
	h.ChunkEmbedFileJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChunkEmbedFileJob_BadFileID(t *testing.T) {
	chunkEmbedder := mock_handlers.NewMockFileChunkEmbedder(t)
	h := handlers.NewJobHandler(nil, nil, nil, chunkEmbedder)

	req := newChunkEmbedRequest(t, qstashclient.ChunkEmbedFileMessage{FileID: "not-a-uuid"})
	w := httptest.NewRecorder()
	h.ChunkEmbedFileJob(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	chunkEmbedder.AssertNotCalled(t, "Process", mock.Anything, mock.Anything)
}

func TestChunkEmbedFileJob_WorkerErrorYields500(t *testing.T) {
	chunkEmbedder := mock_handlers.NewMockFileChunkEmbedder(t)
	h := handlers.NewJobHandler(nil, nil, nil, chunkEmbedder)

	fileID := uuid.New()
	chunkEmbedder.EXPECT().Process(mock.Anything, fileID).Return(errors.New("embed 503"))

	req := newChunkEmbedRequest(t, qstashclient.ChunkEmbedFileMessage{FileID: fileID.String()})
	w := httptest.NewRecorder()
	h.ChunkEmbedFileJob(w, req)

	assert.GreaterOrEqual(t, w.Code, 500)
}

func TestChunkEmbedFileFailedJob_UnwrapsEnvelopeAndMarksRowFailed(t *testing.T) {
	chunkEmbedder := mock_handlers.NewMockFileChunkEmbedder(t)
	h := handlers.NewJobHandler(nil, nil, nil, chunkEmbedder)

	fileID := uuid.New()
	chunkEmbedder.EXPECT().MarkFailed(mock.Anything, fileID, mock.MatchedBy(func(reason string) bool {
		return contains(reason, "source_msg=msg_ce") && contains(reason, "http_status=500")
	})).Return(nil)

	envelope := wrapAsFailureCallback(t, qstashclient.ChunkEmbedFileMessage{
		FileID: fileID.String(),
	}, "msg_ce", 500, 3)

	req := httptest.NewRequest(http.MethodPost, "/jobs/chunk-embed-file-failed",
		bytes.NewReader(envelope))
	w := httptest.NewRecorder()
	h.ChunkEmbedFileFailedJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestChunkEmbedFileFailedJob_BadEnvelopeStill200(t *testing.T) {
	chunkEmbedder := mock_handlers.NewMockFileChunkEmbedder(t)
	h := handlers.NewJobHandler(nil, nil, nil, chunkEmbedder)

	req := httptest.NewRequest(http.MethodPost, "/jobs/chunk-embed-file-failed",
		bytes.NewBufferString(`{`))
	w := httptest.NewRecorder()
	h.ChunkEmbedFileFailedJob(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	chunkEmbedder.AssertNotCalled(t, "MarkFailed", mock.Anything, mock.Anything, mock.Anything)
}

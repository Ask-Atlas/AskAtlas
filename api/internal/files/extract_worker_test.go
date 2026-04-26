package files_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeExtractRepo is a hand-rolled stub instead of a mockery-generated
// mock because the worker's narrow ExtractRepository interface is
// trivially small. The fake records every call so tests can assert
// state-transition ordering + idempotency without leaning on
// testify/mock's expectation DSL.
type fakeExtractRepo struct {
	row             db.GetFileForExtractionRow
	getErr          error
	statusErr       error
	upsertErr       error
	failErr         error
	transitions     []db.ProcessingStatus
	failedReasons   []string
	upsertedTexts   []string
	upsertedOffsets [][]int32
}

func (f *fakeExtractRepo) GetFileForExtraction(ctx context.Context, fileID uuid.UUID) (db.GetFileForExtractionRow, error) {
	if f.getErr != nil {
		return db.GetFileForExtractionRow{}, f.getErr
	}
	return f.row, nil
}

func (f *fakeExtractRepo) SetFileProcessingStatus(ctx context.Context, fileID uuid.UUID, status db.ProcessingStatus) error {
	if f.statusErr != nil {
		return f.statusErr
	}
	f.transitions = append(f.transitions, status)
	return nil
}

func (f *fakeExtractRepo) MarkFileProcessingFailed(ctx context.Context, fileID uuid.UUID, statusError string) error {
	if f.failErr != nil {
		return f.failErr
	}
	f.failedReasons = append(f.failedReasons, statusError)
	return nil
}

func (f *fakeExtractRepo) UpsertExtractedText(ctx context.Context, fileID uuid.UUID, text string, pageOffsets []int32) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.upsertedTexts = append(f.upsertedTexts, text)
	f.upsertedOffsets = append(f.upsertedOffsets, pageOffsets)
	return nil
}

type fakeDownloader struct {
	body []byte
	err  error
}

func (f *fakeDownloader) GetObject(ctx context.Context, key string) ([]byte, error) {
	return f.body, f.err
}

func newRow(s3Key, mime string, status db.ProcessingStatus) db.GetFileForExtractionRow {
	return db.GetFileForExtractionRow{
		ID:               utils.UUID(uuid.New()),
		UserID:           utils.UUID(uuid.New()),
		S3Key:            s3Key,
		MimeType:         mime,
		ProcessingStatus: status,
	}
}

func TestExtractWorker_Process_HappyPathPlainText(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "text/plain", db.ProcessingStatusUploaded),
	}
	dl := &fakeDownloader{body: []byte("Hello world.")}
	w := files.NewExtractWorker(repo, dl)

	require.NoError(t, w.Process(context.Background(), uuid.New()))

	assert.Equal(t,
		[]db.ProcessingStatus{db.ProcessingStatusExtracting, db.ProcessingStatusExtracted},
		repo.transitions,
		"worker should transition uploaded -> extracting -> extracted")
	require.Len(t, repo.upsertedTexts, 1)
	assert.Equal(t, "Hello world.", repo.upsertedTexts[0])
	assert.Empty(t, repo.failedReasons)
}

func TestExtractWorker_Process_HappyPathPDF(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "application/pdf", db.ProcessingStatusUploaded),
	}
	dl := &fakeDownloader{body: loadPDFFixture(t)}
	w := files.NewExtractWorker(repo, dl)

	require.NoError(t, w.Process(context.Background(), uuid.New()))

	require.Len(t, repo.upsertedTexts, 1)
	assert.NotEmpty(t, repo.upsertedTexts[0])
	require.Len(t, repo.upsertedOffsets, 1)
	assert.NotEmpty(t, repo.upsertedOffsets[0], "PDF persists per-page offsets")
}

func TestExtractWorker_Process_TerminalUnsupportedMime(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "application/zip", db.ProcessingStatusUploaded),
	}
	dl := &fakeDownloader{body: []byte("anything")}
	w := files.NewExtractWorker(repo, dl)

	// Returns nil so QStash does not retry.
	require.NoError(t, w.Process(context.Background(), uuid.New()))

	require.Len(t, repo.failedReasons, 1)
	assert.Contains(t, repo.failedReasons[0], "unsupported mime type")
	assert.Equal(t,
		[]db.ProcessingStatus{db.ProcessingStatusExtracting},
		repo.transitions,
		"failed path does NOT transition to extracted")
	assert.Empty(t, repo.upsertedTexts, "no text persisted on terminal failure")
}

func TestExtractWorker_Process_TerminalEmptyExtraction(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "text/plain", db.ProcessingStatusUploaded),
	}
	dl := &fakeDownloader{body: []byte("   \n  ")}
	w := files.NewExtractWorker(repo, dl)

	require.NoError(t, w.Process(context.Background(), uuid.New()))

	require.Len(t, repo.failedReasons, 1)
	assert.Contains(t, repo.failedReasons[0], "no extractable text")
}

func TestExtractWorker_Process_TransientS3Error(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "text/plain", db.ProcessingStatusUploaded),
	}
	dl := &fakeDownloader{err: errors.New("connection refused")}
	w := files.NewExtractWorker(repo, dl)

	err := w.Process(context.Background(), uuid.New())
	require.Error(t, err, "transient errors must propagate so QStash retries")
	assert.Contains(t, err.Error(), "download")
	assert.Empty(t, repo.failedReasons, "transient failure does NOT terminally mark failed")
	// extracting was set; no successful transition followed.
	assert.Equal(t,
		[]db.ProcessingStatus{db.ProcessingStatusExtracting},
		repo.transitions)
}

func TestExtractWorker_Process_IdempotentSkipsExtracted(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "text/plain", db.ProcessingStatusExtracted),
	}
	dl := &fakeDownloader{body: []byte("ignored")}
	w := files.NewExtractWorker(repo, dl)

	require.NoError(t, w.Process(context.Background(), uuid.New()))
	assert.Empty(t, repo.transitions, "no transitions when already extracted")
	assert.Empty(t, repo.upsertedTexts)
}

func TestExtractWorker_Process_IdempotentSkipsFailed(t *testing.T) {
	repo := &fakeExtractRepo{
		row: newRow("k", "application/zip", db.ProcessingStatusFailed),
	}
	w := files.NewExtractWorker(repo, &fakeDownloader{})

	require.NoError(t, w.Process(context.Background(), uuid.New()))
	assert.Empty(t, repo.transitions)
	assert.Empty(t, repo.failedReasons, "already-failed rows are not re-failed")
}

func TestExtractWorker_MarkFailed(t *testing.T) {
	repo := &fakeExtractRepo{}
	w := files.NewExtractWorker(repo, &fakeDownloader{})

	require.NoError(t, w.MarkFailed(context.Background(), uuid.New(), "qstash retries exhausted"))
	require.Len(t, repo.failedReasons, 1)
	assert.Equal(t, "qstash retries exhausted", repo.failedReasons[0])
}

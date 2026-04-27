package files_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeChunkEmbedRepo is a hand-rolled test double for the
// ChunkEmbedRepository interface. Tracks all calls so tests can
// assert state-transition order, persistence inputs, and idempotency.
type fakeChunkEmbedRepo struct {
	row             db.GetFileForExtractionRow
	rowErr          error
	extracted       db.GetExtractedTextRow
	extractedErr    error
	statusErr       error
	failErr         error
	persistErr      error
	deleteErr       error
	transitions     []db.ProcessingStatus
	failedReasons   []string
	persistedChunks [][]db.InsertStudyGuideFileChunkParams
	deleteCalls     int
}

func (f *fakeChunkEmbedRepo) GetFileForExtraction(ctx context.Context, fileID uuid.UUID) (db.GetFileForExtractionRow, error) {
	if f.rowErr != nil {
		return db.GetFileForExtractionRow{}, f.rowErr
	}
	return f.row, nil
}

func (f *fakeChunkEmbedRepo) GetExtractedText(ctx context.Context, fileID uuid.UUID) (db.GetExtractedTextRow, error) {
	if f.extractedErr != nil {
		return db.GetExtractedTextRow{}, f.extractedErr
	}
	return f.extracted, nil
}

func (f *fakeChunkEmbedRepo) SetFileProcessingStatus(ctx context.Context, fileID uuid.UUID, status db.ProcessingStatus) error {
	if f.statusErr != nil {
		return f.statusErr
	}
	f.transitions = append(f.transitions, status)
	return nil
}

func (f *fakeChunkEmbedRepo) MarkFileProcessingFailed(ctx context.Context, fileID uuid.UUID, statusError string) error {
	if f.failErr != nil {
		return f.failErr
	}
	f.failedReasons = append(f.failedReasons, statusError)
	return nil
}

func (f *fakeChunkEmbedRepo) PersistChunks(ctx context.Context, fileID uuid.UUID, params []db.InsertStudyGuideFileChunkParams) error {
	if f.persistErr != nil {
		return f.persistErr
	}
	f.persistedChunks = append(f.persistedChunks, params)
	return nil
}

func (f *fakeChunkEmbedRepo) DeleteExtractedText(ctx context.Context, fileID uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleteCalls++
	return nil
}

// fakeEmbedder returns one 1536-dim zero vector per input. The
// chunker's output count drives the expected vector count, so we
// dynamically size to match.
type fakeEmbedder struct {
	err    error
	tokens int64
}

func (f *fakeEmbedder) Embed(ctx context.Context, req ai.EmbedRequest) (ai.EmbedResponse, error) {
	if f.err != nil {
		return ai.EmbedResponse{}, f.err
	}
	out := make([][]float32, len(req.Inputs))
	for i := range out {
		out[i] = make([]float32, 1536)
	}
	return ai.EmbedResponse{
		Vectors: out,
		Usage:   ai.Usage{InputTokens: f.tokens},
	}, nil
}

func newCERow(status db.ProcessingStatus) db.GetFileForExtractionRow {
	return db.GetFileForExtractionRow{
		ID:               utils.UUID(uuid.New()),
		UserID:           utils.UUID(uuid.New()),
		S3Key:            "uploads/x.pdf",
		MimeType:         "application/pdf",
		ProcessingStatus: status,
	}
}

func TestChunkEmbedWorker_Process_HappyPath(t *testing.T) {
	repo := &fakeChunkEmbedRepo{
		row: newCERow(db.ProcessingStatusExtracted),
		extracted: db.GetExtractedTextRow{
			Text:        "First paragraph about recursion.\n\nSecond paragraph with more detail.",
			PageOffsets: []int32{0},
		},
	}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{tokens: 42})

	require.NoError(t, w.Process(context.Background(), uuid.New()))

	assert.Equal(t,
		[]db.ProcessingStatus{db.ProcessingStatusEmbedding, db.ProcessingStatusReady},
		repo.transitions,
		"happy path: extracted -> embedding -> ready")
	require.Len(t, repo.persistedChunks, 1, "exactly one persist call")
	assert.NotEmpty(t, repo.persistedChunks[0], "at least one chunk persisted")
	assert.Equal(t, 1, repo.deleteCalls, "extracted_text cleanup ran")
	assert.Empty(t, repo.failedReasons)
}

func TestChunkEmbedWorker_Process_VanishedFile(t *testing.T) {
	repo := &fakeChunkEmbedRepo{rowErr: apperrors.ErrNotFound}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{})

	require.NoError(t, w.Process(context.Background(), uuid.New()),
		"vanished file is terminal-success so QStash stops retrying")
	assert.Empty(t, repo.transitions)
	assert.Empty(t, repo.failedReasons)
}

func TestChunkEmbedWorker_Process_AlreadyReady(t *testing.T) {
	repo := &fakeChunkEmbedRepo{row: newCERow(db.ProcessingStatusReady)}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{})

	require.NoError(t, w.Process(context.Background(), uuid.New()))
	assert.Empty(t, repo.transitions)
	assert.Empty(t, repo.persistedChunks)
}

func TestChunkEmbedWorker_Process_AlreadyFailed(t *testing.T) {
	repo := &fakeChunkEmbedRepo{row: newCERow(db.ProcessingStatusFailed)}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{})

	require.NoError(t, w.Process(context.Background(), uuid.New()))
	assert.Empty(t, repo.transitions)
}

func TestChunkEmbedWorker_Process_MissingExtractedText_TerminalFail(t *testing.T) {
	repo := &fakeChunkEmbedRepo{
		row:          newCERow(db.ProcessingStatusExtracted),
		extractedErr: apperrors.ErrNotFound,
	}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{})

	require.NoError(t, w.Process(context.Background(), uuid.New()),
		"missing extracted text is terminal-failure, not transient")
	require.Len(t, repo.failedReasons, 1)
	assert.Contains(t, repo.failedReasons[0], "missing extracted text")
}

func TestChunkEmbedWorker_Process_EmbedTransientErrorPropagates(t *testing.T) {
	repo := &fakeChunkEmbedRepo{
		row:       newCERow(db.ProcessingStatusExtracted),
		extracted: db.GetExtractedTextRow{Text: "Hello world."},
	}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{err: errors.New("openai 503")})

	err := w.Process(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed")
	assert.Equal(t,
		[]db.ProcessingStatus{db.ProcessingStatusEmbedding},
		repo.transitions,
		"transient embed failure leaves row in 'embedding' for retry resume")
	assert.Empty(t, repo.persistedChunks)
}

func TestChunkEmbedWorker_Process_PersistErrorPropagates(t *testing.T) {
	repo := &fakeChunkEmbedRepo{
		row:        newCERow(db.ProcessingStatusExtracted),
		extracted:  db.GetExtractedTextRow{Text: "Some content here."},
		persistErr: errors.New("db down"),
	}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{})

	err := w.Process(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "persist")
}

func TestChunkEmbedWorker_MarkFailed(t *testing.T) {
	repo := &fakeChunkEmbedRepo{}
	w := files.NewChunkEmbedWorker(repo, &fakeEmbedder{})

	require.NoError(t, w.MarkFailed(context.Background(), uuid.New(), "qstash exhausted"))
	require.Len(t, repo.failedReasons, 1)
	assert.Equal(t, "qstash exhausted", repo.failedReasons[0])
}

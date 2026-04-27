package files_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadPDFFixture reads the small (~40KB) reference PDF stashed under
// testdata/. Sourced from ledongthuc/pdf's own examples; bundled here
// so the worker tests don't need a network or a Go module dance to
// run. The path resolves relative to this test file's directory --
// `go test` sets the package dir as CWD.
func loadPDFFixture(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "sample.pdf"))
	require.NoError(t, err, "load testdata/sample.pdf")
	return body
}

func TestExtractText_PlainText(t *testing.T) {
	doc, err := files.ExtractText([]byte("Hello, world!\n\nLine two."), "text/plain")
	require.NoError(t, err)
	assert.Equal(t, "Hello, world!\n\nLine two.", doc.Text)
	assert.Nil(t, doc.PageOffsets, "plain text has no page boundaries")
}

func TestExtractText_Markdown(t *testing.T) {
	doc, err := files.ExtractText([]byte("# Title\n\nBody."), "text/markdown")
	require.NoError(t, err)
	assert.Equal(t, "# Title\n\nBody.", doc.Text)
	assert.Nil(t, doc.PageOffsets)
}

func TestExtractText_PlainText_TrimsWhitespace(t *testing.T) {
	doc, err := files.ExtractText([]byte("   spaced  \n"), "text/plain")
	require.NoError(t, err)
	assert.Equal(t, "spaced", doc.Text)
}

func TestExtractText_PlainText_EmptyFails(t *testing.T) {
	_, err := files.ExtractText([]byte("   \n  "), "text/plain")
	require.Error(t, err)
	assert.True(t, errors.Is(err, files.ErrEmptyExtraction))
}

func TestExtractText_PlainText_InvalidUTF8(t *testing.T) {
	_, err := files.ExtractText([]byte{0xff, 0xfe, 0xfd}, "text/plain")
	require.Error(t, err)
	assert.False(t, errors.Is(err, files.ErrEmptyExtraction))
}

func TestExtractText_UnsupportedMime(t *testing.T) {
	_, err := files.ExtractText([]byte("anything"), "application/zip")
	require.Error(t, err)
	assert.True(t, errors.Is(err, files.ErrUnsupportedMimeType))
	assert.Contains(t, err.Error(), "application/zip", "error should name the rejected mime")
}

func TestExtractText_PDF(t *testing.T) {
	body := loadPDFFixture(t)
	doc, err := files.ExtractText(body, "application/pdf")
	require.NoError(t, err)
	assert.NotEmpty(t, doc.Text, "PDF should yield some text")
	require.NotNil(t, doc.PageOffsets, "PDF should have page boundaries")
	assert.GreaterOrEqual(t, len(doc.PageOffsets), 1)
	assert.Equal(t, int32(0), doc.PageOffsets[0], "first page starts at offset 0")
}

func TestExtractText_PDF_CorruptFails(t *testing.T) {
	// Bytes that don't start with "%PDF-" -- ledongthuc/pdf rejects.
	_, err := files.ExtractText([]byte("not a pdf"), "application/pdf")
	require.Error(t, err)
	// Whatever the underlying lib emits, we don't classify it as
	// terminal-by-mime; the worker's transient/terminal split is
	// based on the typed sentinels.
	assert.False(t, errors.Is(err, files.ErrUnsupportedMimeType))
}

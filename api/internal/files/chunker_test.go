package files_test

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunk_EmptyInput(t *testing.T) {
	assert.Nil(t, files.Chunk("", nil))
}

func TestChunk_ShortText_SingleChunk(t *testing.T) {
	chunks := files.Chunk("Hello world. Short doc.", nil)
	require.Len(t, chunks, 1)
	assert.Equal(t, int32(0), chunks[0].ChunkIdx)
	assert.Equal(t, "Hello world. Short doc.", chunks[0].Text)
	assert.Nil(t, chunks[0].Page, "no pageOffsets => no page tag")
	assert.Nil(t, chunks[0].Heading, "no markdown heading => nil")
}

func TestChunk_Markdown_HeadingAttached(t *testing.T) {
	text := "# Introduction\n\nFirst paragraph about recursion.\n\n## Background\n\nSecond paragraph with more detail."
	chunks := files.Chunk(text, nil)
	require.NotEmpty(t, chunks)
	require.NotNil(t, chunks[0].Heading, "first chunk under # Introduction should pick up the heading")
	assert.Equal(t, "Introduction", *chunks[0].Heading)
}

func TestChunk_LongText_SplitsAndOverlaps(t *testing.T) {
	// 50 paragraphs * 100 chars = ~5000 chars => well above target.
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		sb.WriteString(strings.Repeat("word ", 20))
		sb.WriteString("\n\n")
	}
	chunks := files.Chunk(sb.String(), nil)
	assert.Greater(t, len(chunks), 1, "long text should split into multiple chunks")
	for _, c := range chunks {
		runeLen := utf8.RuneCountInString(c.Text)
		// Allow some headroom over targetChars (overlap can push body beyond it).
		assert.LessOrEqual(t, runeLen, 1200, "chunk %d unexpectedly long: %d runes", c.ChunkIdx, runeLen)
	}
}

func TestChunk_PDF_PageTagged(t *testing.T) {
	text := "Page one content here.\n\nPage two starts later.\n\nPage three further on."
	// Synthetic page offsets — values chosen so each page has at
	// least one chunk.
	pageOffsets := []int32{0, 24, 50}
	chunks := files.Chunk(text, pageOffsets)
	require.NotEmpty(t, chunks)
	for _, c := range chunks {
		require.NotNil(t, c.Page, "chunk %d missing page tag", c.ChunkIdx)
		assert.GreaterOrEqual(t, *c.Page, int32(1))
		assert.LessOrEqual(t, *c.Page, int32(3))
	}
}

func TestChunk_OversizeParagraph_WordSplits(t *testing.T) {
	huge := strings.Repeat("Lorem ipsum dolor sit amet ", 200) // ~5400 chars, no \n\n
	chunks := files.Chunk(huge, nil)
	require.Greater(t, len(chunks), 1, "oversized single paragraph must split into multiple chunks")
}

func TestChunk_ChunkIdxIsSequential(t *testing.T) {
	huge := strings.Repeat("Lorem ipsum dolor sit amet ", 200)
	chunks := files.Chunk(huge, nil)
	for i, c := range chunks {
		assert.Equal(t, int32(i), c.ChunkIdx, "chunk[%d].ChunkIdx", i)
	}
}

func TestChunk_TokensApproxNonZero(t *testing.T) {
	chunks := files.Chunk("Hello world.", nil)
	require.Len(t, chunks, 1)
	// 12 chars / 4 ≈ 3 tokens.
	assert.GreaterOrEqual(t, chunks[0].Tokens, int32(2))
	assert.LessOrEqual(t, chunks[0].Tokens, int32(4))
}

package files_test

import (
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/files"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCursorEncoding(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	expected := files.Cursor{
		ID:        uuid.New(),
		UpdatedAt: &now,
	}

	encoded, err := files.EncodeCursor(expected)
	assert.NoError(t, err)

	decoded, err := files.DecodeCursor(encoded)
	assert.NoError(t, err)

	assert.Equal(t, expected.ID, decoded.ID)
	assert.True(t, expected.UpdatedAt.Equal(*decoded.UpdatedAt))
}

func TestDecodeCursorErrors(t *testing.T) {
	// Invalid base64
	_, err := files.DecodeCursor("invalid-base64---")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor encoding")

	// Invalid JSON payload
	b64 := "e2ludmFsaWQ6IGpzb259" // "{invalid: json}"
	_, err = files.DecodeCursor(b64)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cursor payload")
}

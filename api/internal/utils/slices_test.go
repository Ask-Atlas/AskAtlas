package utils_test

import (
	"encoding/json"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonNilStrings_NilInput_ReturnsNonNilEmpty(t *testing.T) {
	got := utils.NonNilStrings(nil)
	require.NotNil(t, got, "nil input must produce non-nil output (the whole point)")
	assert.Empty(t, got)
}

func TestNonNilStrings_EmptyInput_ReturnsNonNilEmpty(t *testing.T) {
	got := utils.NonNilStrings([]string{})
	require.NotNil(t, got)
	assert.Empty(t, got)
}

func TestNonNilStrings_PopulatedInput_CopiesElements(t *testing.T) {
	src := []string{"a", "b", "c"}
	got := utils.NonNilStrings(src)
	assert.Equal(t, src, got)
}

// The whole reason this helper exists: a nil string slice JSON-encodes
// as "null" rather than "[]". Pin that behavior so a future refactor
// doesn't accidentally regress and re-introduce the wire-shape bug
// that surfaced on PATCH /study-guides/{id} {"tags":[]} during PR
// #143's staging matrix.
func TestNonNilStrings_JSONEncodesAsEmptyArrayNotNull(t *testing.T) {
	type wire struct {
		Tags []string `json:"tags"`
	}

	bad, err := json.Marshal(wire{Tags: nil})
	require.NoError(t, err)
	assert.JSONEq(t, `{"tags":null}`, string(bad), "sanity: nil slice should encode as null")

	good, err := json.Marshal(wire{Tags: utils.NonNilStrings(nil)})
	require.NoError(t, err)
	assert.JSONEq(t, `{"tags":[]}`, string(good), "NonNilStrings result must encode as []")
}

// Defensive: callers should be safe to mutate the result without
// affecting the source. The make+copy implementation guarantees a
// fresh backing array.
func TestNonNilStrings_FreshBackingArray(t *testing.T) {
	src := []string{"a", "b"}
	got := utils.NonNilStrings(src)
	got[0] = "MUTATED"
	assert.Equal(t, "a", src[0], "source must not be aliased")
}

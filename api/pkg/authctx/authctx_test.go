package authctx_test

import (
	"context"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestWithUserID_RoundTrip(t *testing.T) {
	id := uuid.New()
	ctx := authctx.WithUserID(context.Background(), id)

	got, ok := authctx.UserIDFromContext(ctx)
	assert.True(t, ok, "expected ok to be true")
	assert.Equal(t, id, got)
}

func TestUserIDFromContext_MissingKey(t *testing.T) {
	got, ok := authctx.UserIDFromContext(context.Background())
	assert.False(t, ok, "expected ok to be false for empty context")
	assert.Equal(t, uuid.UUID{}, got)
}

func TestUserIDFromContext_WrongType(t *testing.T) {
	type key string
	ctx := context.WithValue(context.Background(), key("user_id"), "not-a-uuid")
	got, ok := authctx.UserIDFromContext(ctx)
	assert.False(t, ok, "expected ok to be false when key type differs")
	assert.Equal(t, uuid.UUID{}, got)
}

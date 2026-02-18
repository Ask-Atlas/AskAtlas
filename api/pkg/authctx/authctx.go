package authctx

import (
	"context"

	"github.com/google/uuid"
)

type UserIDResolver interface {
	GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error)
}

type contextKey struct{ name string }

var userIDKey = contextKey{name: "user_id"}

func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

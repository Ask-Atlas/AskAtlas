// Package authctx provides context extractors and utilities for handling user authentication states.
package authctx

import (
	"context"

	"github.com/google/uuid"
)

// UserIDResolver defines an interface for fetching internal user IDs mapped from external Clerk IDs.
type UserIDResolver interface {
	GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error)
}

type contextKey struct{ name string }

var userIDKey = contextKey{name: "user_id"}

// WithUserID injects the given user ID into the provided context.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserIDFromContext retrieves the user ID from the given context, if present.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(userIDKey).(uuid.UUID)
	return id, ok
}

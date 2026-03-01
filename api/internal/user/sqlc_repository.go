package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// sqlcRepository provides a Postgres implementation of the user Repository protocol.
type sqlcRepository struct {
	queries *db.Queries
}

// NewSQLCRepository creates a Postgres-backed Repository instance.
func NewSQLCRepository(queries *db.Queries) *sqlcRepository {
	return &sqlcRepository{queries: queries}
}

// UpsertClerkUser creates a new user or updates the fields of an existing user based on Clerk ID.
func (r *sqlcRepository) UpsertClerkUser(ctx context.Context, arg db.UpsertClerkUserParams) (db.User, error) {
	slog.Info("upserting clerk user", "clerk_id", arg.ClerkID, "email", arg.Email)
	user, err := r.queries.UpsertClerkUser(ctx, arg)
	if err != nil {
		return db.User{}, fmt.Errorf("failed to upsert clerk user: %w", err)
	}
	return user, nil
}

// SoftDeleteUserByClerkID marks the indicated user as deleted in the database.
func (r *sqlcRepository) SoftDeleteUserByClerkID(ctx context.Context, clerkID string) error {
	slog.Info("soft deleting user by clerk id", "clerk_id", clerkID)
	affectedRows, err := r.queries.SoftDeleteUserByClerkID(ctx, clerkID)
	if err != nil {
		return fmt.Errorf("failed to soft delete user by clerk id: %w", err)
	}

	if affectedRows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUserIDByClerkID fetches the internal UUID matching the provided Clerk external ID.
func (r *sqlcRepository) GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error) {
	pgID, err := r.queries.GetUserIDByClerkID(ctx, clerkID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, ErrUserNotFound
		}
		return uuid.UUID{}, fmt.Errorf("GetUserIDByClerkID: %w", err)
	}
	if !pgID.Valid {
		return uuid.UUID{}, fmt.Errorf("GetUserIDByClerkID: invalid/NULL UUID stored for user")
	}
	id, err := uuid.FromBytes(pgID.Bytes[:])
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("GetUserIDByClerkID: invalid UUID bytes: %w", err)
	}
	return id, nil
}

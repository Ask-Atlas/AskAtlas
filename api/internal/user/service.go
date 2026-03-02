package user

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
	"github.com/google/uuid"
)

// Repository defines the required data-access behaviors for managing users.
type Repository interface {
	UpsertClerkUser(ctx context.Context, arg db.UpsertClerkUserParams) (db.User, error)
	SoftDeleteUserByClerkID(ctx context.Context, clerkID string) error
	GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error)
}

// service implements the core user domain logic.
type service struct {
	repo Repository
}

// NewService creates a new configured User service.
func NewService(repo Repository) *service {
	return &service{repo: repo}
}

// UpsertClerkUser ensures the given user details are synchronized into the database.
func (s *service) UpsertClerkUser(ctx context.Context, payload UpsertUserPayload) (User, error) {
	arg, err := ToUpsertClerkUserParams(payload)
	if err != nil {
		return User{}, err
	}

	dbUser, err := s.repo.UpsertClerkUser(ctx, arg)
	if err != nil {
		return User{}, err
	}

	user, err := ToUser(dbUser)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// SoftDeleteUserByClerkID marks a user as deleted without entirely removing their relational data.
func (s *service) SoftDeleteUserByClerkID(ctx context.Context, clerkID string) error {
	return s.repo.SoftDeleteUserByClerkID(ctx, clerkID)
}

// GetUserIDByClerkID resolves the internal UUID mapping for a given external Clerk ID.
func (s *service) GetUserIDByClerkID(ctx context.Context, clerkID string) (uuid.UUID, error) {
	return s.repo.GetUserIDByClerkID(ctx, clerkID)
}

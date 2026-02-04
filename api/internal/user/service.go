package user

import (
	"context"

	"github.com/Ask-Atlas/AskAtlas/api/internal/db"
)

type Repository interface {
	UpsertClerkUser(ctx context.Context, arg db.UpsertClerkUserParams) (db.User, error)
	SoftDeleteUserByClerkID(ctx context.Context, clerkID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) *service {
	return &service{repo: repo}
}

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

func (s *service) SoftDeleteUserByClerkID(ctx context.Context, clerkID string) error {
	return s.repo.SoftDeleteUserByClerkID(ctx, clerkID)
}

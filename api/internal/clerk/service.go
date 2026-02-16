package clerk

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Ask-Atlas/AskAtlas/api/internal/user"
)

type UserService interface {
	UpsertClerkUser(ctx context.Context, arg user.UpsertUserPayload) (user.User, error)
	SoftDeleteUserByClerkID(ctx context.Context, clerkID string) error
}

type clerkService struct {
	userService UserService
}

func NewClerkService(userService UserService) *clerkService {
	return &clerkService{userService: userService}
}

func (cs *clerkService) HandleWebhookEvent(ctx context.Context, event Event) error {
	switch e := event.(type) {
	case UserCreatedEvent:
		return cs.handleUserCreated(ctx, e)
	case UserUpdateEvent:
		return cs.handleUserUpdated(ctx, e)
	case UserDeletedEvent:
		return cs.handleUserDeleted(ctx, e)
	default:
		slog.Warn("unknown event type", "type", e.GetType())
		return nil
	}
}

func (cs *clerkService) handleUserCreated(ctx context.Context, event UserCreatedEvent) error {
	return cs.handleUserUpdated(ctx, UserUpdateEvent(event))
}

func (cs *clerkService) handleUserUpdated(ctx context.Context, event UserUpdateEvent) error {
	payload, err := ToUpsertUserPayload(event.Data)
	if err != nil {
		return fmt.Errorf("failed to convert user to upsert payload: %w", err)
	}

	_, err = cs.userService.UpsertClerkUser(ctx, payload)
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	return nil
}

func (cs *clerkService) handleUserDeleted(ctx context.Context, event UserDeletedEvent) error {
	if err := cs.userService.SoftDeleteUserByClerkID(ctx, event.Data.ID); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return ErrUserNotFound
		}

		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

package handlers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/clerk"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

type ClerkService interface {
	HandleWebhookEvent(ctx context.Context, payload clerk.Event) error
}

type ClerkHandler struct {
	clerkService ClerkService
}

func NewClerkWebhookHandler(clerkService ClerkService) *ClerkHandler {
	return &ClerkHandler{clerkService: clerkService}
}

func (ch *ClerkHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read webhook event",
			"error", err,
		)
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Bad Request", nil))
		return
	}

	msgID := r.Header.Get("svix-id")
	event, err := clerk.ParseWebhookEvent(body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse webhook event",
			"error", err,
			"msgID", msgID,
		)
		apperrors.RespondWithError(w, &apperrors.AppError{
			Code:    http.StatusUnprocessableEntity,
			Status:  "Unprocessable Entity",
			Message: "Unprocessable Entity",
		})
		return
	}

	slog.InfoContext(ctx, "received webhook event",
		"msgID", msgID,
		"type", event.GetType(),
	)

	err = ch.clerkService.HandleWebhookEvent(ctx, event)
	if err != nil {
		slog.ErrorContext(ctx, "failed to handle webhook event",
			"error", err,
			"type", event.GetType(),
			"msgID", msgID,
		)

		if errors.Is(err, clerk.ErrUserNotFound) {
			apperrors.RespondWithError(w, apperrors.NewNotFound("Not Found"))
			return
		}

		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	w.WriteHeader(http.StatusOK)
}

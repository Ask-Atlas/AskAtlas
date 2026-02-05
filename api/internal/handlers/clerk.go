package handlers

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/Ask-Atlas/AskAtlas/api/internal/clerk"
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
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	msgID := r.Header.Get("svix-id")
	event, err := clerk.ParseWebhookEvent(body)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse webhook event",
			"error", err,
			"body", string(body),
			"msgID", msgID,
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/api"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

// defaultPingPrompt is what the smoke endpoint asks Claude when the
// caller doesn't override `prompt`. Kept short so a verification call
// burns ~10 input + ~3 output tokens.
const defaultPingPrompt = "Reply with the single word 'pong'."

// AIService is the slice of the ai package the handler depends on.
// Defined here (where it's used) per the codebase convention so the
// handler is unit-testable with a generated fake.
type AIService interface {
	Stream(ctx context.Context, req ai.StreamRequest) (<-chan ai.Event, error)
}

// AIHandler serves the /api/ai/* surface (ASK-213 ships /ai/ping;
// ASK-215+ add /ai/edit etc on top of the same handler struct).
type AIHandler struct {
	ai AIService
}

// NewAIHandler wires the handler over the given AI service.
func NewAIHandler(service AIService) *AIHandler {
	return &AIHandler{ai: service}
}

// AIPing handles POST /api/ai/ping (ASK-213). It builds a minimal
// Haiku request from the optional `prompt` body, then forwards the
// event channel to the SSE writer. Cancellation flows from the HTTP
// request context all the way to the Anthropic SDK.
//
// Errors that happen before SSE upgrade (auth, validation) are
// returned as JSON so the OpenAPI client + frontend understand them.
// Errors that happen mid-stream are emitted as `event: error` because
// the headers are already written.
func (h *AIHandler) AIPing(w http.ResponseWriter, r *http.Request) {
	viewerID, appErr := viewerIDFromContext(r)
	if appErr != nil {
		apperrors.RespondWithError(w, appErr)
		return
	}

	var body api.AIPingJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apperrors.RespondWithError(w, apperrors.NewBadRequest("Invalid request body", nil))
		return
	}

	prompt := defaultPingPrompt
	if body.Prompt != nil {
		trimmed := strings.TrimSpace(*body.Prompt)
		if trimmed != "" {
			prompt = trimmed
		}
	}

	events, err := h.ai.Stream(r.Context(), ai.StreamRequest{
		UserID:    viewerID,
		Feature:   ai.FeaturePing,
		Model:     ai.ModelCheap,
		MaxTokens: 64,
		Messages: []ai.Message{
			{
				Role:   ai.RoleUser,
				Blocks: []ai.Block{{Text: prompt}},
			},
		},
	})
	if err != nil {
		slog.Error("AIPing: ai.Stream failed", "error", err)
		apperrors.RespondWithError(w, apperrors.NewInternalError())
		return
	}

	if err := ai.WriteStream(w, r, events); err != nil {
		// Upgrade failure -- headers may not be written yet, but
		// WriteStream returns errors only for the upgrade step.
		// Log and bail; nothing recoverable to send.
		slog.Error("AIPing: sse write failed", "error", err)
	}
}

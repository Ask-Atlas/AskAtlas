package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/handlers"
	mock_handlers "github.com/Ask-Atlas/AskAtlas/api/internal/handlers/mocks"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func TestAIHandler_AIPing_Stream(t *testing.T) {
	t.Parallel()

	viewerID := uuid.New()
	mockSvc := mock_handlers.NewMockAIService(t)

	events := make(chan ai.Event, 4)
	events <- ai.Event{Kind: ai.EventDelta, Delta: "pong"}
	events <- ai.Event{Kind: ai.EventUsage, Usage: &ai.Usage{InputTokens: 10, OutputTokens: 1}}
	events <- ai.Event{Kind: ai.EventDone}
	close(events)

	mockSvc.EXPECT().
		Stream(mock.Anything, mock.MatchedBy(func(req ai.StreamRequest) bool {
			return req.UserID == viewerID && req.Feature == ai.FeaturePing && req.Model == ai.ModelCheap
		})).
		Return((<-chan ai.Event)(events), nil)

	h := handlers.NewAIHandler(mockSvc)

	body := strings.NewReader(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", body)
	req = req.WithContext(authctx.WithUserID(req.Context(), viewerID))

	rec := httptest.NewRecorder()
	h.AIPing(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200, body=%q", rec.Code, rec.Body.String())
	}
	got := rec.Body.String()
	for _, want := range []string{"event: delta", `"text":"pong"`, "event: usage", `"input_tokens":10`, "event: done"} {
		if !strings.Contains(got, want) {
			t.Errorf("response body missing %q\nbody:\n%s", want, got)
		}
	}
}

func TestAIHandler_AIPing_Unauthenticated(t *testing.T) {
	t.Parallel()

	mockSvc := mock_handlers.NewMockAIService(t)
	h := handlers.NewAIHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.AIPing(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401, body=%q", rec.Code, rec.Body.String())
	}
}

func TestAIHandler_AIPing_BadJSON(t *testing.T) {
	t.Parallel()

	viewerID := uuid.New()
	mockSvc := mock_handlers.NewMockAIService(t)
	h := handlers.NewAIHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", strings.NewReader(`{not json`))
	req = req.WithContext(authctx.WithUserID(req.Context(), viewerID))

	rec := httptest.NewRecorder()
	h.AIPing(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body=%q", rec.Code, rec.Body.String())
	}
}

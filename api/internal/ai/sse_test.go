package ai_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
)

func TestWriteStream_HappyPath(t *testing.T) {
	t.Parallel()

	events := make(chan ai.Event, 8)
	events <- ai.Event{Kind: ai.EventDelta, Delta: "Hel"}
	events <- ai.Event{Kind: ai.EventDelta, Delta: "lo"}
	events <- ai.Event{Kind: ai.EventUsage, Usage: &ai.Usage{InputTokens: 5, OutputTokens: 2}}
	events <- ai.Event{Kind: ai.EventDone}
	close(events)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", strings.NewReader(""))

	if err := ai.WriteStream(rec, req, events); err != nil {
		t.Fatalf("WriteStream returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: delta") {
		t.Errorf("missing delta event in body: %q", body)
	}
	if !strings.Contains(body, `"text":"Hel"`) {
		t.Errorf("missing first delta payload in body: %q", body)
	}
	if !strings.Contains(body, "event: usage") {
		t.Errorf("missing usage event in body: %q", body)
	}
	if !strings.Contains(body, `"input_tokens":5`) {
		t.Errorf("missing usage input_tokens in body: %q", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Errorf("missing done event in body: %q", body)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/event-stream") {
		t.Errorf("Content-Type = %q, want text/event-stream", got)
	}
}

func TestWriteStream_ErrorEvent(t *testing.T) {
	t.Parallel()

	events := make(chan ai.Event, 2)
	events <- ai.Event{Kind: ai.EventError, Err: errors.New("upstream model error")}
	close(events)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", strings.NewReader(""))

	if err := ai.WriteStream(rec, req, events); err != nil {
		t.Fatalf("WriteStream returned error: %v", err)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "event: error") {
		t.Errorf("missing error event in body: %q", body)
	}
	if !strings.Contains(body, `"message":"upstream model error"`) {
		t.Errorf("missing error message in body: %q", body)
	}
	if strings.Contains(body, "event: done") {
		t.Errorf("done event should not appear after error: %q", body)
	}
}

func TestWriteStream_ContextCancellation(t *testing.T) {
	t.Parallel()

	events := make(chan ai.Event) // unbuffered, never sent on
	ctx, cancel := context.WithCancel(context.Background())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", strings.NewReader("")).WithContext(ctx)

	done := make(chan struct{})
	go func() {
		_ = ai.WriteStream(rec, req, events)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("WriteStream did not return within 1s after ctx cancellation")
	}
	close(events)
}

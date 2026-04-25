package middleware_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Ask-Atlas/AskAtlas/api/internal/ai"
	"github.com/Ask-Atlas/AskAtlas/api/internal/middleware"
	"github.com/Ask-Atlas/AskAtlas/api/pkg/authctx"
	"github.com/google/uuid"
)

type fakeGate struct {
	err     error
	calls   int
	lastUID uuid.UUID
	lastFt  ai.Feature
}

func (f *fakeGate) CheckAndReserve(_ context.Context, userID uuid.UUID, feature ai.Feature) error {
	f.calls++
	f.lastUID = userID
	f.lastFt = feature
	return f.err
}

// nextHandler is a sentinel that flips a flag if invoked. Tests that
// expect short-circuit verify the flag stays false.
type nextHandler struct{ called bool }

func (h *nextHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.called = true
	w.WriteHeader(http.StatusOK)
}

func TestAIQuota_PassesThroughOutsideScope(t *testing.T) {
	t.Parallel()

	gate := &fakeGate{}
	next := &nextHandler{}
	mw := middleware.AIQuota(gate, "/api/ai/", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/files", nil)
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if !next.called {
		t.Fatal("next handler was not invoked for out-of-scope request")
	}
	if gate.calls != 0 {
		t.Errorf("gate called %d times for out-of-scope request, want 0", gate.calls)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAIQuota_NoAuthContext(t *testing.T) {
	t.Parallel()

	gate := &fakeGate{}
	next := &nextHandler{}
	mw := middleware.AIQuota(gate, "/api/ai/", nil)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", nil)
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if next.called {
		t.Error("next was invoked despite missing auth context")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	if gate.calls != 0 {
		t.Errorf("gate called %d times without auth, want 0", gate.calls)
	}
}

func TestAIQuota_UnderQuota(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	gate := &fakeGate{err: nil}
	next := &nextHandler{}
	mw := middleware.AIQuota(gate, "/api/ai/", nil)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", nil).
		WithContext(authctx.WithUserID(context.Background(), user))
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if !next.called {
		t.Error("next was not invoked for under-quota request")
	}
	if gate.calls != 1 {
		t.Errorf("gate called %d times, want 1", gate.calls)
	}
	if gate.lastUID != user {
		t.Errorf("gate user id = %v, want %v", gate.lastUID, user)
	}
	if gate.lastFt != ai.FeaturePing {
		t.Errorf("gate feature = %q, want %q", gate.lastFt, ai.FeaturePing)
	}
}

func TestAIQuota_OverQuota_429Envelope(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	resetAt := time.Now().Add(2 * time.Hour).UTC().Truncate(time.Second)
	gate := &fakeGate{err: &ai.QuotaExceededError{
		Feature: ai.FeatureEdit,
		Used:    50,
		Limit:   50,
		ResetAt: resetAt,
	}}
	next := &nextHandler{}
	mw := middleware.AIQuota(gate, "/api/ai/", func(string) ai.Feature { return ai.FeatureEdit })

	req := httptest.NewRequest(http.MethodPost, "/api/ai/edit", nil).
		WithContext(authctx.WithUserID(context.Background(), user))
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if next.called {
		t.Error("next was invoked despite quota exceeded")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}

	retryHeader := rec.Header().Get("Retry-After")
	retrySeconds, err := strconv.Atoi(retryHeader)
	if err != nil || retrySeconds < 1 {
		t.Errorf("Retry-After = %q, want positive integer", retryHeader)
	}

	var body struct {
		Code              int    `json:"code"`
		Status            string `json:"status"`
		Feature           string `json:"feature"`
		Used              int64  `json:"used"`
		Limit             int    `json:"limit"`
		ResetAt           string `json:"reset_at"`
		RetryAfterSeconds int    `json:"retry_after_seconds"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Code != 429 || body.Status != "QUOTA_EXCEEDED" {
		t.Errorf("envelope code/status = %d/%q, want 429/QUOTA_EXCEEDED", body.Code, body.Status)
	}
	if body.Feature != "edit" || body.Used != 50 || body.Limit != 50 {
		t.Errorf("envelope (feature=%q, used=%d, limit=%d), want (edit, 50, 50)", body.Feature, body.Used, body.Limit)
	}
	if body.RetryAfterSeconds != retrySeconds {
		t.Errorf("body retry_after_seconds=%d does not match header Retry-After=%d", body.RetryAfterSeconds, retrySeconds)
	}
}

func TestAIQuota_GenericError_500(t *testing.T) {
	t.Parallel()

	user := uuid.New()
	gate := &fakeGate{err: errors.New("postgres: connection refused")}
	next := &nextHandler{}
	mw := middleware.AIQuota(gate, "/api/ai/", nil)

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ping", nil).
		WithContext(authctx.WithUserID(context.Background(), user))
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)

	if next.called {
		t.Error("next was invoked despite gate error")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

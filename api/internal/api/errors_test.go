package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

func TestOAPIValidatorErrorHandler(t *testing.T) {
	handler := OAPIValidatorErrorHandler
	w := httptest.NewRecorder()

	message := `parameter "scope" in query has an error: value must be one of [owned]`
	handler(w, message, http.StatusBadRequest)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	var appErr apperrors.AppError
	if err := json.NewDecoder(res.Body).Decode(&appErr); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if msg, ok := appErr.Details["scope"]; !ok || msg != "value must be one of [owned]" {
		t.Errorf("expected details to contain scope: value must be one of [owned], got: %v", appErr.Details)
	}
}

func TestOAPIValidatorErrorHandler_Fallback(t *testing.T) {
	handler := OAPIValidatorErrorHandler
	w := httptest.NewRecorder()

	message := `some other error format`
	handler(w, message, http.StatusBadRequest)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}

	var appErr apperrors.AppError
	if err := json.NewDecoder(res.Body).Decode(&appErr); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if msg, ok := appErr.Details["validation"]; !ok || msg != message {
		t.Errorf("expected details to contain validation: %s, got: %v", message, appErr.Details)
	}
}

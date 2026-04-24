package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
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

	if appErr.Status != apperrors.StatusValidationError {
		t.Errorf("Status = %q, want %q", appErr.Status, apperrors.StatusValidationError)
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
	if appErr.Status != apperrors.StatusValidationError {
		t.Errorf("Status = %q, want %q", appErr.Status, apperrors.StatusValidationError)
	}
}

func TestBearerAuthFunc(t *testing.T) {
	bearerScheme := &openapi3.SecurityScheme{Type: "http", Scheme: "bearer"}
	apiKeyScheme := &openapi3.SecurityScheme{Type: "apiKey", In: "header", Name: "X-API-Key"}

	tests := []struct {
		name    string
		scheme  *openapi3.SecurityScheme
		header  string
		wantErr bool
	}{
		{name: "accepts bearer token", scheme: bearerScheme, header: "Bearer abc.def.ghi", wantErr: false},
		{name: "accepts lowercase bearer prefix", scheme: bearerScheme, header: "bearer abc.def.ghi", wantErr: false},
		{name: "rejects missing header", scheme: bearerScheme, header: "", wantErr: true},
		{name: "rejects non-bearer scheme prefix", scheme: bearerScheme, header: "Basic dXNlcjpwYXNz", wantErr: true},
		{name: "rejects prefix without token", scheme: bearerScheme, header: "Bearer ", wantErr: true},
		{name: "rejects unsupported security scheme", scheme: apiKeyScheme, header: "Bearer abc.def.ghi", wantErr: true},
		{name: "rejects nil security scheme", scheme: nil, header: "Bearer abc.def.ghi", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/me/files", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			input := &openapi3filter.AuthenticationInput{
				SecuritySchemeName:     "BearerAuth",
				SecurityScheme:         tc.scheme,
				RequestValidationInput: &openapi3filter.RequestValidationInput{Request: req},
			}
			err := BearerAuthFunc(context.Background(), input)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

package apperrors_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Ask-Atlas/AskAtlas/api/pkg/apperrors"
)

func TestToHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "ErrNotFound maps to 404",
			err:      apperrors.ErrNotFound,
			wantCode: http.StatusNotFound,
		},
		{
			name:     "ErrConflict maps to 409",
			err:      apperrors.ErrConflict,
			wantCode: http.StatusConflict,
		},
		{
			name:     "Wrapped sentinel error maps correctly",
			err:      fmt.Errorf("wrapped: %w", apperrors.ErrUnauthorized),
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "Unknown error maps to 500",
			err:      errors.New("some random db failure"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "Direct AppError is unwrapped",
			err: &apperrors.AppError{
				Code:    http.StatusTeapot,
				Message: "I am a teapot",
			},
			wantCode: http.StatusTeapot,
		},
		{
			name: "Wrapped AppError is unwrapped",
			err: fmt.Errorf("service failed: %w", &apperrors.AppError{
				Code:    http.StatusUnprocessableEntity,
				Message: "Custom validation failed",
			}),
			wantCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apperrors.ToHTTPError(tt.err)
			if got.Code != tt.wantCode {
				t.Errorf("ToHTTPError() code = %v, want %v", got.Code, tt.wantCode)
			}
		})
	}
}

// TestSymbolicStatusCodes pins the wire `status` string each
// constructor emits (ASK-110). This is a contract test -- the
// frontend reads the `status` field to drive error-boundary UI, so
// a regression to the old HTTP-phrase strings ("Bad Request",
// "Not Found", etc.) would break that surface silently. The
// explicit string literals here are the source of truth for the
// wire format; every new test + service callsite should match.
func TestSymbolicStatusCodes(t *testing.T) {
	t.Run("NewBadRequest emits VALIDATION_ERROR", func(t *testing.T) {
		if got := apperrors.NewBadRequest("x", nil).Status; got != "VALIDATION_ERROR" {
			t.Errorf("Status = %q, want VALIDATION_ERROR", got)
		}
	})
	t.Run("NewNotFound emits NOT_FOUND", func(t *testing.T) {
		if got := apperrors.NewNotFound("x").Status; got != "NOT_FOUND" {
			t.Errorf("Status = %q, want NOT_FOUND", got)
		}
	})
	t.Run("NewUnauthorized emits UNAUTHORIZED", func(t *testing.T) {
		if got := apperrors.NewUnauthorized().Status; got != "UNAUTHORIZED" {
			t.Errorf("Status = %q, want UNAUTHORIZED", got)
		}
	})
	t.Run("NewForbidden emits FORBIDDEN", func(t *testing.T) {
		if got := apperrors.NewForbidden().Status; got != "FORBIDDEN" {
			t.Errorf("Status = %q, want FORBIDDEN", got)
		}
	})
	t.Run("NewConflict emits CONFLICT", func(t *testing.T) {
		if got := apperrors.NewConflict("x").Status; got != "CONFLICT" {
			t.Errorf("Status = %q, want CONFLICT", got)
		}
	})
	t.Run("NewInternalError emits INTERNAL_ERROR", func(t *testing.T) {
		if got := apperrors.NewInternalError().Status; got != "INTERNAL_ERROR" {
			t.Errorf("Status = %q, want INTERNAL_ERROR", got)
		}
	})
	// Sentinel-path tests -- ToHTTPError on a wrapped sentinel must
	// produce the same symbolic status as the matching constructor.
	t.Run("ToHTTPError(ErrNotFound) -> NOT_FOUND", func(t *testing.T) {
		if got := apperrors.ToHTTPError(apperrors.ErrNotFound).Status; got != "NOT_FOUND" {
			t.Errorf("Status = %q, want NOT_FOUND", got)
		}
	})
	t.Run("ToHTTPError(ErrConflict) -> CONFLICT", func(t *testing.T) {
		if got := apperrors.ToHTTPError(apperrors.ErrConflict).Status; got != "CONFLICT" {
			t.Errorf("Status = %q, want CONFLICT", got)
		}
	})
	t.Run("ToHTTPError(unknown) -> INTERNAL_ERROR", func(t *testing.T) {
		if got := apperrors.ToHTTPError(errors.New("mystery")).Status; got != "INTERNAL_ERROR" {
			t.Errorf("Status = %q, want INTERNAL_ERROR", got)
		}
	})
}

// TestStatusForCodeFallback covers the fallback for AppErrors built
// without a Status value. A bare `&AppError{Code: 404}` must emit
// "NOT_FOUND" on the wire via the statusCodeMap lookup; an
// unrecognised code must fall through to "UNKNOWN_ERROR" so no
// wire consumer ever sees an empty Status.
func TestStatusForCodeFallback(t *testing.T) {
	cases := []struct {
		name string
		code int
		want string
	}{
		{"400 -> VALIDATION_ERROR", http.StatusBadRequest, "VALIDATION_ERROR"},
		{"401 -> UNAUTHORIZED", http.StatusUnauthorized, "UNAUTHORIZED"},
		{"403 -> FORBIDDEN", http.StatusForbidden, "FORBIDDEN"},
		{"404 -> NOT_FOUND", http.StatusNotFound, "NOT_FOUND"},
		{"409 -> CONFLICT", http.StatusConflict, "CONFLICT"},
		{"500 -> INTERNAL_ERROR", http.StatusInternalServerError, "INTERNAL_ERROR"},
		{"418 -> UNKNOWN_ERROR", http.StatusTeapot, "UNKNOWN_ERROR"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Bare AppError with no Status -- ToHTTPError's
			// `if appErr.Status == ""` branch fills it from
			// statusForCode.
			got := apperrors.ToHTTPError(&apperrors.AppError{Code: c.code})
			if got.Status != c.want {
				t.Errorf("Status = %q, want %q", got.Status, c.want)
			}
		})
	}
}

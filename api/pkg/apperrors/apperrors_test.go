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

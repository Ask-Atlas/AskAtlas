// Package apperrors defines standard application-level errors and HTTP mappings.
// It provides a unified way to represent, handle, and return errors across the API.
package apperrors

import (
	"cmp"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Standard predefined application errors.
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
)

// AppError represents an HTTP-friendly error containing a status code,
// a message, and optional validation details.
type AppError struct {
	Code    int               `json:"code"`
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Cause   error             `json:"-"`
}

// Error returns the underlying error message.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of the error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewBadRequest creates an AppError with an HTTP 400 Bad Request status.
func NewBadRequest(message string, details map[string]string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Status: "Bad Request", Message: message, Details: details}
}

// NewNotFound creates an AppError with an HTTP 404 Not Found status.
func NewNotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Status: "Not Found", Message: message}
}

// NewUnauthorized creates an AppError with an HTTP 401 Unauthorized status.
func NewUnauthorized() *AppError {
	return &AppError{Code: http.StatusUnauthorized, Status: "Unauthorized", Message: "Authentication required"}
}

// NewForbidden creates an AppError with an HTTP 403 Forbidden status.
func NewForbidden() *AppError {
	return &AppError{Code: http.StatusForbidden, Status: "Forbidden", Message: "You do not have permission"}
}

// NewConflict creates an AppError with an HTTP 409 Conflict status.
// First introduced for ASK-137 (SubmitAnswer on a completed
// session); reusable for any resource-state conflict (e.g. trying
// to vote twice, recommend the same guide twice).
func NewConflict(message string) *AppError {
	return &AppError{Code: http.StatusConflict, Status: "Conflict", Message: message}
}

// NewInternalError creates an AppError with an HTTP 500 Internal Server Error status.
func NewInternalError() *AppError {
	return &AppError{Code: http.StatusInternalServerError, Status: "Internal Server Error", Message: "Something went wrong"}
}

// ToHTTPError maps a sentinel error or existing AppError to an AppError
func ToHTTPError(err error) *AppError {
	if appErr := (*AppError)(nil); errors.As(err, &appErr) {
		if appErr == nil {
			return NewInternalError()
		}
		if appErr.Code < 100 || appErr.Code > 999 {
			appErr.Code = http.StatusInternalServerError
		}
		appErr.Status = cmp.Or(appErr.Status, http.StatusText(appErr.Code))
		return appErr
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return NewNotFound("Resource not found")
	case errors.Is(err, ErrConflict):
		return &AppError{Code: http.StatusConflict, Status: "Conflict", Message: "Resource already exists"}
	case errors.Is(err, ErrInvalidInput):
		return NewBadRequest(err.Error(), nil)
	case errors.Is(err, ErrUnauthorized):
		return NewUnauthorized()
	case errors.Is(err, ErrForbidden):
		return NewForbidden()
	default:
		return NewInternalError()
	}
}

// RespondWithError writes the given AppError as a JSON response to the client.
func RespondWithError(w http.ResponseWriter, appErr *AppError) {
	if appErr == nil {
		appErr = NewInternalError()
	}
	if appErr.Code < 100 || appErr.Code > 999 {
		appErr.Code = http.StatusInternalServerError
		appErr.Status = http.StatusText(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Code)
	if err := json.NewEncoder(w).Encode(appErr); err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

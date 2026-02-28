package apperrors

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("already exists")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
)

type AppError struct {
	Code    int               `json:"code"`
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Cause   error             `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewBadRequest(message string, details map[string]string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Status: "Bad Request", Message: message, Details: details}
}

func NewNotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Status: "Not Found", Message: message}
}

func NewUnauthorized() *AppError {
	return &AppError{Code: http.StatusUnauthorized, Status: "Unauthorized", Message: "Authentication required"}
}

func NewForbidden() *AppError {
	return &AppError{Code: http.StatusForbidden, Status: "Forbidden", Message: "You do not have permission"}
}

func NewInternalError() *AppError {
	return &AppError{Code: http.StatusInternalServerError, Status: "Internal Server Error", Message: "Something went wrong"}
}

// ToHTTPError maps a sentinel error or existing AppError to an AppError
func ToHTTPError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
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

func RespondWithError(w http.ResponseWriter, appErr *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Code)
	if err := json.NewEncoder(w).Encode(appErr); err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

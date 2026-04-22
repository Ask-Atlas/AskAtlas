// Package apperrors defines standard application-level errors and HTTP mappings.
// It provides a unified way to represent, handle, and return errors across the API.
//
// Status strings are SYMBOLIC (ASK-110): `VALIDATION_ERROR`,
// `NOT_FOUND`, etc. -- stable machine-readable identifiers rather
// than the HTTP phrase (`Bad Request`, `Not Found`). The integer
// `code` field is the canonical HTTP status code and is unchanged.
package apperrors

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Symbolic status strings shared across constructors + the
// RespondWithError / ToHTTPError fallback. Kept as named constants
// so a typo in a call site is a compile-time error rather than a
// silent wire-contract regression.
const (
	StatusValidationError = "VALIDATION_ERROR"
	StatusUnauthorized    = "UNAUTHORIZED"
	StatusForbidden       = "FORBIDDEN"
	StatusNotFound        = "NOT_FOUND"
	StatusConflict        = "CONFLICT"
	StatusInternalError   = "INTERNAL_ERROR"
	StatusUnknownError    = "UNKNOWN_ERROR"
)

// statusCodeMap resolves an HTTP code to the symbolic status string.
// Used as the fallback inside RespondWithError + ToHTTPError for
// AppErrors built without a Status value (bare literals, third-
// party code that surfaces only an integer code, etc.). An
// unrecognised code falls through to "UNKNOWN_ERROR" so a wire
// consumer never sees an empty string.
var statusCodeMap = map[int]string{
	http.StatusBadRequest:          StatusValidationError,
	http.StatusUnauthorized:        StatusUnauthorized,
	http.StatusForbidden:           StatusForbidden,
	http.StatusNotFound:            StatusNotFound,
	http.StatusConflict:            StatusConflict,
	http.StatusInternalServerError: StatusInternalError,
}

// statusForCode maps an HTTP status code onto its symbolic status
// string. Returns StatusUnknownError for codes not in the map so
// callers never surface an empty Status to the wire.
func statusForCode(code int) string {
	if s, ok := statusCodeMap[code]; ok {
		return s
	}
	return StatusUnknownError
}

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

// NewBadRequest creates an AppError with an HTTP 400 status. The
// symbolic Status is VALIDATION_ERROR -- every 400 the API emits
// is a request-validation failure (malformed body, bad query
// param, enum violation).
func NewBadRequest(message string, details map[string]string) *AppError {
	return &AppError{Code: http.StatusBadRequest, Status: StatusValidationError, Message: message, Details: details}
}

// NewNotFound creates an AppError with an HTTP 404 status and the
// symbolic Status NOT_FOUND.
func NewNotFound(message string) *AppError {
	return &AppError{Code: http.StatusNotFound, Status: StatusNotFound, Message: message}
}

// NewUnauthorized creates an AppError with an HTTP 401 status and
// the symbolic Status UNAUTHORIZED.
func NewUnauthorized() *AppError {
	return &AppError{Code: http.StatusUnauthorized, Status: StatusUnauthorized, Message: "Authentication required"}
}

// NewForbidden creates an AppError with an HTTP 403 status and the
// symbolic Status FORBIDDEN.
func NewForbidden() *AppError {
	return &AppError{Code: http.StatusForbidden, Status: StatusForbidden, Message: "You do not have permission"}
}

// NewConflict creates an AppError with an HTTP 409 status and the
// symbolic Status CONFLICT. First introduced for ASK-137
// (SubmitAnswer on a completed session); reusable for any
// resource-state conflict (e.g. trying to vote twice, recommend
// the same guide twice, duplicate file grant).
func NewConflict(message string) *AppError {
	return &AppError{Code: http.StatusConflict, Status: StatusConflict, Message: message}
}

// NewInternalError creates an AppError with an HTTP 500 status and
// the symbolic Status INTERNAL_ERROR.
func NewInternalError() *AppError {
	return &AppError{Code: http.StatusInternalServerError, Status: StatusInternalError, Message: "Something went wrong"}
}

// ToHTTPError maps a sentinel error or existing AppError to an AppError.
// AppErrors missing a Status value (legacy inline struct literals,
// third-party code) get the symbolic status looked up from their Code.
func ToHTTPError(err error) *AppError {
	if appErr := (*AppError)(nil); errors.As(err, &appErr) {
		if appErr == nil {
			return NewInternalError()
		}
		if appErr.Code < 100 || appErr.Code > 999 {
			appErr.Code = http.StatusInternalServerError
		}
		if appErr.Status == "" {
			appErr.Status = statusForCode(appErr.Code)
		}
		return appErr
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return NewNotFound("Resource not found")
	case errors.Is(err, ErrConflict):
		return NewConflict("Resource already exists")
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
// If Status is empty, fall back to the symbolic status for the Code so
// no wire response ever ships with a blank `status` field.
func RespondWithError(w http.ResponseWriter, appErr *AppError) {
	if appErr == nil {
		appErr = NewInternalError()
	}
	if appErr.Code < 100 || appErr.Code > 999 {
		appErr.Code = http.StatusInternalServerError
		appErr.Status = StatusInternalError
	}
	if appErr.Status == "" {
		appErr.Status = statusForCode(appErr.Code)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Code)
	if err := json.NewEncoder(w).Encode(appErr); err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

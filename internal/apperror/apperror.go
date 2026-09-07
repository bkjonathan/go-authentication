// Package apperror describes a failure in the terms the HTTP layer answers in,
// so a service can say what went wrong without importing gin.
package apperror

import (
	"errors"
	"net/http"
)

type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return e.Code + ": " + e.cause.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.cause }

// Wrap keeps the original error for the log without putting it in the response,
// which is how an internal failure stays diagnosable and still says nothing to
// the caller.
func (e *Error) Wrap(cause error) *Error {
	clone := *e
	clone.cause = cause
	return &clone
}

// WithFields attaches per-field validation messages.
func (e *Error) WithFields(fields map[string]string) *Error {
	clone := *e
	clone.Fields = fields
	return &clone
}

func New(status int, code, message string) *Error {
	return &Error{Status: status, Code: code, Message: message}
}

func BadRequest(code, message string) *Error {
	return New(http.StatusBadRequest, code, message)
}

func Unauthorized(code, message string) *Error {
	return New(http.StatusUnauthorized, code, message)
}

func Forbidden(code, message string) *Error {
	return New(http.StatusForbidden, code, message)
}

func NotFound(code, message string) *Error {
	return New(http.StatusNotFound, code, message)
}

func Conflict(code, message string) *Error {
	return New(http.StatusConflict, code, message)
}

func Locked(code, message string) *Error {
	return New(http.StatusLocked, code, message)
}

func Internal(cause error) *Error {
	return (&Error{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "Something went wrong.",
	}).Wrap(cause)
}

// From coerces any error into one the HTTP layer can render. Anything that is
// not already an *Error is a bug rather than a rule, so it becomes a 500 with
// its detail kept out of the response.
func From(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}

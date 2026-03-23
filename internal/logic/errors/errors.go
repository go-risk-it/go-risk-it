package domainerrors

import "fmt"

// ErrorCategory classifies domain errors for HTTP status mapping and metrics.
type ErrorCategory int

const (
	CategoryValidation   ErrorCategory = iota // 400
	CategoryUnauthorized                      // 401
	CategoryForbidden                         // 403
	CategoryNotFound                          // 404
	CategoryConflict                          // 409
)

func (c ErrorCategory) String() string {
	switch c {
	case CategoryValidation:
		return "VALIDATION_ERROR"
	case CategoryUnauthorized:
		return "UNAUTHORIZED"
	case CategoryForbidden:
		return "FORBIDDEN"
	case CategoryNotFound:
		return "NOT_FOUND"
	case CategoryConflict:
		return "CONFLICT"
	default:
		return "UNKNOWN"
	}
}

// Categorizable is implemented by all domain error types.
type Categorizable interface {
	error
	Category() ErrorCategory
}

// ValidationError indicates the client sent invalid input (HTTP 400).
type ValidationError struct {
	Msg   string
	Cause error
}

func (e *ValidationError) Error() string {
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}

	return e.Msg
}

func (e *ValidationError) Unwrap() error           { return e.Cause }
func (e *ValidationError) Category() ErrorCategory { return CategoryValidation }

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{Msg: msg}
}

func NewValidationErrorf(format string, args ...any) *ValidationError {
	return &ValidationError{Msg: fmt.Sprintf(format, args...)}
}

func WrapValidationError(cause error, msg string) *ValidationError {
	return &ValidationError{Msg: msg, Cause: cause}
}

func WrapValidationErrorf(cause error, format string, args ...any) *ValidationError {
	return &ValidationError{Msg: fmt.Sprintf(format, args...), Cause: cause}
}

// ConflictError indicates a state conflict, e.g. wrong phase or turn (HTTP 409).
type ConflictError struct {
	Msg   string
	Cause error
}

func (e *ConflictError) Error() string {
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}

	return e.Msg
}

func (e *ConflictError) Unwrap() error           { return e.Cause }
func (e *ConflictError) Category() ErrorCategory { return CategoryConflict }

func NewConflictError(msg string) *ConflictError {
	return &ConflictError{Msg: msg}
}

func NewConflictErrorf(format string, args ...any) *ConflictError {
	return &ConflictError{Msg: fmt.Sprintf(format, args...)}
}

func WrapConflictError(cause error, msg string) *ConflictError {
	return &ConflictError{Msg: msg, Cause: cause}
}

func WrapConflictErrorf(cause error, format string, args ...any) *ConflictError {
	return &ConflictError{Msg: fmt.Sprintf(format, args...), Cause: cause}
}

// ForbiddenError indicates the player is not authorized for this action (HTTP 403).
type ForbiddenError struct {
	Msg   string
	Cause error
}

func (e *ForbiddenError) Error() string {
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}

	return e.Msg
}

func (e *ForbiddenError) Unwrap() error           { return e.Cause }
func (e *ForbiddenError) Category() ErrorCategory { return CategoryForbidden }

func NewForbiddenError(msg string) *ForbiddenError {
	return &ForbiddenError{Msg: msg}
}

func WrapForbiddenError(cause error, msg string) *ForbiddenError {
	return &ForbiddenError{Msg: msg, Cause: cause}
}

func WrapForbiddenErrorf(cause error, format string, args ...any) *ForbiddenError {
	return &ForbiddenError{Msg: fmt.Sprintf(format, args...), Cause: cause}
}

// NotFoundError indicates the requested resource does not exist (HTTP 404).
type NotFoundError struct {
	Msg   string
	Cause error
}

func (e *NotFoundError) Error() string {
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}

	return e.Msg
}

func (e *NotFoundError) Unwrap() error           { return e.Cause }
func (e *NotFoundError) Category() ErrorCategory { return CategoryNotFound }

func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{Msg: msg}
}

func NewNotFoundErrorf(format string, args ...any) *NotFoundError {
	return &NotFoundError{Msg: fmt.Sprintf(format, args...)}
}

func WrapNotFoundError(cause error, msg string) *NotFoundError {
	return &NotFoundError{Msg: msg, Cause: cause}
}

// UnauthorizedError indicates missing or invalid credentials (HTTP 401).
type UnauthorizedError struct {
	Msg   string
	Cause error
}

func (e *UnauthorizedError) Error() string {
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}

	return e.Msg
}

func (e *UnauthorizedError) Unwrap() error           { return e.Cause }
func (e *UnauthorizedError) Category() ErrorCategory { return CategoryUnauthorized }

func NewUnauthorizedError(msg string) *UnauthorizedError {
	return &UnauthorizedError{Msg: msg}
}

func WrapUnauthorizedError(cause error, msg string) *UnauthorizedError {
	return &UnauthorizedError{Msg: msg, Cause: cause}
}

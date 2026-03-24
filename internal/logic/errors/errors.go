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

// DomainError is the single error type for all domain errors.
// The category field determines the HTTP status mapping.
type DomainError struct {
	Msg      string
	Cause    error
	category ErrorCategory
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}

	return e.Msg
}

func (e *DomainError) Unwrap() error           { return e.Cause }
func (e *DomainError) Category() ErrorCategory { return e.category }

// --- Validation (400) ---

func NewValidationError(msg string) *DomainError {
	return &DomainError{Msg: msg, category: CategoryValidation}
}

func NewValidationErrorf(format string, args ...any) *DomainError {
	return &DomainError{Msg: fmt.Sprintf(format, args...), category: CategoryValidation}
}

func WrapValidationError(cause error, msg string) *DomainError {
	return &DomainError{Msg: msg, Cause: cause, category: CategoryValidation}
}

func WrapValidationErrorf(cause error, format string, args ...any) *DomainError {
	return &DomainError{
		Msg:      fmt.Sprintf(format, args...),
		Cause:    cause,
		category: CategoryValidation,
	}
}

// --- Conflict (409) ---

func NewConflictError(msg string) *DomainError {
	return &DomainError{Msg: msg, category: CategoryConflict}
}

func NewConflictErrorf(format string, args ...any) *DomainError {
	return &DomainError{Msg: fmt.Sprintf(format, args...), category: CategoryConflict}
}

func WrapConflictError(cause error, msg string) *DomainError {
	return &DomainError{Msg: msg, Cause: cause, category: CategoryConflict}
}

func WrapConflictErrorf(cause error, format string, args ...any) *DomainError {
	return &DomainError{
		Msg:      fmt.Sprintf(format, args...),
		Cause:    cause,
		category: CategoryConflict,
	}
}

// --- Forbidden (403) ---

func NewForbiddenError(msg string) *DomainError {
	return &DomainError{Msg: msg, category: CategoryForbidden}
}

func WrapForbiddenError(cause error, msg string) *DomainError {
	return &DomainError{Msg: msg, Cause: cause, category: CategoryForbidden}
}

func WrapForbiddenErrorf(cause error, format string, args ...any) *DomainError {
	return &DomainError{
		Msg:      fmt.Sprintf(format, args...),
		Cause:    cause,
		category: CategoryForbidden,
	}
}

// --- NotFound (404) ---

func NewNotFoundError(msg string) *DomainError {
	return &DomainError{Msg: msg, category: CategoryNotFound}
}

func NewNotFoundErrorf(format string, args ...any) *DomainError {
	return &DomainError{Msg: fmt.Sprintf(format, args...), category: CategoryNotFound}
}

func WrapNotFoundError(cause error, msg string) *DomainError {
	return &DomainError{Msg: msg, Cause: cause, category: CategoryNotFound}
}

// --- Unauthorized (401) ---

func NewUnauthorizedError(msg string) *DomainError {
	return &DomainError{Msg: msg, category: CategoryUnauthorized}
}

func WrapUnauthorizedError(cause error, msg string) *DomainError {
	return &DomainError{Msg: msg, Cause: cause, category: CategoryUnauthorized}
}

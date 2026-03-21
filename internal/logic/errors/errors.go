package domainerrors

import "fmt"

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

func (e *ValidationError) Unwrap() error { return e.Cause }

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

func (e *ConflictError) Unwrap() error { return e.Cause }

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

func (e *ForbiddenError) Unwrap() error { return e.Cause }

func NewForbiddenError(msg string) *ForbiddenError {
	return &ForbiddenError{Msg: msg}
}

func WrapForbiddenError(cause error, msg string) *ForbiddenError {
	return &ForbiddenError{Msg: msg, Cause: cause}
}

func WrapForbiddenErrorf(cause error, format string, args ...any) *ForbiddenError {
	return &ForbiddenError{Msg: fmt.Sprintf(format, args...), Cause: cause}
}

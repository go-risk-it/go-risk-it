package domainerrors

import "fmt"

// ValidationError indicates the client sent invalid input (HTTP 400).
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{Msg: msg}
}

func NewValidationErrorf(format string, args ...any) *ValidationError {
	return &ValidationError{Msg: fmt.Sprintf(format, args...)}
}

// ConflictError indicates a state conflict, e.g. wrong phase or turn (HTTP 409).
type ConflictError struct {
	Msg string
}

func (e *ConflictError) Error() string {
	return e.Msg
}

func NewConflictError(msg string) *ConflictError {
	return &ConflictError{Msg: msg}
}

func NewConflictErrorf(format string, args ...any) *ConflictError {
	return &ConflictError{Msg: fmt.Sprintf(format, args...)}
}

// ForbiddenError indicates the player is not authorized for this action (HTTP 403).
type ForbiddenError struct {
	Msg string
}

func (e *ForbiddenError) Error() string {
	return e.Msg
}

func NewForbiddenError(msg string) *ForbiddenError {
	return &ForbiddenError{Msg: msg}
}

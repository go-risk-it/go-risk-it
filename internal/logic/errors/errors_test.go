package domainerrors_test

import (
	"errors"
	"testing"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSentinel = errors.New("underlying cause")

func TestValidationError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewValidationError("bad input")
	assert.Equal(t, "bad input", err.Error())
}

func TestValidationError_Error_WithCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapValidationError(errSentinel, "bad input")
	assert.Equal(t, "bad input: underlying cause", err.Error())
}

func TestValidationError_Unwrap_NilCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewValidationError("bad input")
	assert.NoError(t, err.Unwrap())
}

func TestValidationError_Unwrap_PreservesCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapValidationError(errSentinel, "bad input")
	assert.ErrorIs(t, err, errSentinel)
}

func TestValidationError_ErrorsAs(t *testing.T) {
	t.Parallel()

	wrapped := domainerrors.WrapValidationError(errSentinel, "bad input")
	outerErr := errors.Join(errors.New("context"), wrapped)

	var validationErr *domainerrors.ValidationError

	require.ErrorAs(t, outerErr, &validationErr)
	assert.Equal(t, "bad input: underlying cause", validationErr.Error())
	assert.ErrorIs(t, validationErr, errSentinel)
}

func TestValidationErrorf_Wrap(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapValidationErrorf(errSentinel, "field %s invalid", "name")
	assert.Equal(t, "field name invalid: underlying cause", err.Error())
	assert.ErrorIs(t, err, errSentinel)
}

func TestConflictError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewConflictError("wrong phase")
	assert.Equal(t, "wrong phase", err.Error())
}

func TestConflictError_Error_WithCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapConflictError(errSentinel, "wrong phase")
	assert.Equal(t, "wrong phase: underlying cause", err.Error())
}

func TestConflictError_Unwrap_PreservesCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapConflictError(errSentinel, "wrong phase")
	assert.ErrorIs(t, err, errSentinel)
}

func TestConflictError_ErrorsAs(t *testing.T) {
	t.Parallel()

	wrapped := domainerrors.WrapConflictErrorf(errSentinel, "phase %s", "deploy")
	outerErr := errors.Join(errors.New("context"), wrapped)

	var conflictErr *domainerrors.ConflictError

	require.ErrorAs(t, outerErr, &conflictErr)
	assert.Equal(t, "phase deploy: underlying cause", conflictErr.Error())
}

func TestForbiddenError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewForbiddenError("not your turn")
	assert.Equal(t, "not your turn", err.Error())
}

func TestForbiddenError_Error_WithCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapForbiddenError(errSentinel, "not your turn")
	assert.Equal(t, "not your turn: underlying cause", err.Error())
}

func TestForbiddenError_Unwrap_PreservesCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapForbiddenError(errSentinel, "not your turn")
	assert.ErrorIs(t, err, errSentinel)
}

func TestForbiddenError_ErrorsAs(t *testing.T) {
	t.Parallel()

	wrapped := domainerrors.WrapForbiddenErrorf(errSentinel, "player %s", "alice")
	outerErr := errors.Join(errors.New("context"), wrapped)

	var forbiddenErr *domainerrors.ForbiddenError

	require.ErrorAs(t, outerErr, &forbiddenErr)
	assert.Equal(t, "player alice: underlying cause", forbiddenErr.Error())
}

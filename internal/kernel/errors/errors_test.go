package domainerrors_test

import (
	"errors"
	"fmt"
	"testing"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
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

	var domainErr *domainerrors.DomainError

	require.ErrorAs(t, outerErr, &domainErr)
	assert.Equal(t, "bad input: underlying cause", domainErr.Error())
	require.ErrorIs(t, domainErr, errSentinel)
	assert.Equal(t, domainerrors.CategoryValidation, domainErr.Category())
}

func TestValidationErrorf(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewValidationErrorf("field %s invalid", "name")
	assert.Equal(t, "field name invalid", err.Error())
	assert.Equal(t, domainerrors.CategoryValidation, err.Category())
}

func TestConflictError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewConflictError("wrong phase")
	assert.Equal(t, "wrong phase", err.Error())
}

func TestConflictError_Errorf(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewConflictErrorf("phase %s", "deploy")
	assert.Equal(t, "phase deploy", err.Error())
	assert.Equal(t, domainerrors.CategoryConflict, err.Category())
}

func TestForbiddenError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewForbiddenError("not your turn")
	assert.Equal(t, "not your turn", err.Error())
}

// --- Category tests ---

func TestValidationError_Category(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewValidationError("bad input")
	assert.Equal(t, domainerrors.CategoryValidation, err.Category())
	assert.Equal(t, "VALIDATION_ERROR", err.Category().String())
}

func TestConflictError_Category(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewConflictError("wrong phase")
	assert.Equal(t, domainerrors.CategoryConflict, err.Category())
	assert.Equal(t, "CONFLICT", err.Category().String())
}

func TestForbiddenError_Category(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewForbiddenError("not allowed")
	assert.Equal(t, domainerrors.CategoryForbidden, err.Category())
	assert.Equal(t, "FORBIDDEN", err.Category().String())
}

func TestNotFoundError_Category(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewNotFoundError("game not found")
	assert.Equal(t, domainerrors.CategoryNotFound, err.Category())
	assert.Equal(t, "NOT_FOUND", err.Category().String())
}

func TestNotFoundError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewNotFoundError("game not found")
	assert.Equal(t, "game not found", err.Error())
}

// --- UnauthorizedError tests ---

func TestUnauthorizedError_Error_WithoutCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewUnauthorizedError("invalid token")
	assert.Equal(t, "invalid token", err.Error())
}

func TestUnauthorizedError_Error_WithCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapUnauthorizedError(errSentinel, "authentication failed")
	assert.Equal(t, "authentication failed: underlying cause", err.Error())
}

func TestUnauthorizedError_Unwrap_PreservesCause(t *testing.T) {
	t.Parallel()

	err := domainerrors.WrapUnauthorizedError(errSentinel, "authentication failed")
	assert.ErrorIs(t, err, errSentinel)
}

func TestUnauthorizedError_Category(t *testing.T) {
	t.Parallel()

	err := domainerrors.NewUnauthorizedError("invalid token")
	assert.Equal(t, domainerrors.CategoryUnauthorized, err.Category())
	assert.Equal(t, "UNAUTHORIZED", err.Category().String())
}

// --- Categorizable interface tests ---

func TestCategorizable_AllTypesImplement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      domainerrors.Categorizable
		category domainerrors.ErrorCategory
	}{
		{"ValidationError", domainerrors.NewValidationError("x"), domainerrors.CategoryValidation},
		{
			"UnauthorizedError",
			domainerrors.NewUnauthorizedError("x"),
			domainerrors.CategoryUnauthorized,
		},
		{"ConflictError", domainerrors.NewConflictError("x"), domainerrors.CategoryConflict},
		{"ForbiddenError", domainerrors.NewForbiddenError("x"), domainerrors.CategoryForbidden},
		{"NotFoundError", domainerrors.NewNotFoundError("x"), domainerrors.CategoryNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.category, tt.err.Category())
		})
	}
}

// --- Category accessible through fmt.Errorf wrapping ---

func TestCategory_AccessibleAfterFmtErrorfWrapping(t *testing.T) {
	t.Parallel()

	original := domainerrors.NewValidationError("bad input")
	wrapped := fmt.Errorf("handler failed: %w", original)

	var categorizable domainerrors.Categorizable
	require.ErrorAs(t, wrapped, &categorizable)
	assert.Equal(t, domainerrors.CategoryValidation, categorizable.Category())
}

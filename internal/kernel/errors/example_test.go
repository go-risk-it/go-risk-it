package domainerrors_test

import (
	"errors"
	"fmt"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

func ExampleNewValidationError() {
	err := domainerrors.NewValidationError("region not owned by player")
	fmt.Println(err)
	fmt.Println(err.Category())
	// Output:
	// region not owned by player
	// VALIDATION_ERROR
}

func ExampleWrapValidationError() {
	cause := errors.New("connection refused")
	err := domainerrors.WrapValidationError(cause, "lookup failed")

	fmt.Println(err)
	fmt.Println(errors.Is(err, cause))
	// Output:
	// lookup failed: connection refused
	// true
}

func ExampleNewConflictError() {
	err := domainerrors.NewConflictError("game is already over")

	var categorizable domainerrors.Categorizable
	if errors.As(err, &categorizable) {
		fmt.Println(categorizable.Category())
	}

	wrapped := fmt.Errorf("handler: %w", err)

	var extracted domainerrors.Categorizable
	if errors.As(wrapped, &extracted) {
		fmt.Println(extracted.Category())
	}
	// Output:
	// CONFLICT
	// CONFLICT
}

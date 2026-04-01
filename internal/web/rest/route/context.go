package route

import (
	"fmt"
	"net/http"
	"strconv"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

// ExtractID parses the {id} path parameter from the request as an int64.
func ExtractID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, fmt.Errorf("invalid id: %w", err)
	}

	return int64(id), nil
}

// ExtractUserContext asserts the request context implements [kernelctx.UserContext].
// Returns an UnauthorizedError when the assertion fails.
func ExtractUserContext(r *http.Request) (kernelctx.UserContext, error) {
	uc, ok := r.Context().(kernelctx.UserContext)
	if !ok {
		return nil, domainerrors.NewUnauthorizedError("user context not found")
	}

	return uc, nil
}

// BuildDomainContext composes [ExtractUserContext] and [ExtractID], then passes
// both to withID to produce a domain-specific context C. This eliminates the
// duplicated user-context + ID extraction pattern in game and lobby routes.
func BuildDomainContext[C any](
	request *http.Request,
	withID func(kernelctx.UserContext, int64) C,
) (C, error) {
	var zero C

	userCtx, err := ExtractUserContext(request)
	if err != nil {
		return zero, err
	}

	id, err := ExtractID(request)
	if err != nil {
		return zero, domainerrors.WrapValidationError(err, "invalid path parameter")
	}

	return withID(userCtx, id), nil
}

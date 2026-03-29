package routes

import (
	"errors"
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

// BuildGameContext extracts UserContext and {id} from the request to build a GameContext.
func BuildGameContext(request *http.Request) (ctx.GameContext, error) {
	userContext, ok := request.Context().(kernelctx.UserContext)
	if !ok {
		return nil, errors.New("user context not found")
	}

	id, err := route.ExtractID(request)
	if err != nil {
		return nil, domainerrors.WrapValidationError(err, "invalid path parameter")
	}

	return ctx.WithGameID(userContext, id), nil
}

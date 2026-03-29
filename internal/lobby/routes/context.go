package routes

import (
	"errors"
	"net/http"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

// BuildLobbyContext extracts UserContext and {id} from the request to build a LobbyContext.
func BuildLobbyContext(request *http.Request) (ctx.LobbyContext, error) {
	userContext, ok := request.Context().(kernelctx.UserContext)
	if !ok {
		return nil, errors.New("user context not found")
	}

	id, err := route.ExtractID(request)
	if err != nil {
		return nil, domainerrors.WrapValidationError(err, "invalid path parameter")
	}

	return ctx.WithLobbyID(userContext, id), nil
}

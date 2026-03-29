package route

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

func extractID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, fmt.Errorf("invalid id: %w", err)
	}

	return int64(id), nil
}

// BuildGameContext extracts UserContext and {id} from the request to build a GameContext.
func BuildGameContext(request *http.Request) (ctx.GameContext, error) {
	userContext, ok := request.Context().(ctx.UserContext)
	if !ok {
		return nil, errors.New("user context not found")
	}

	id, err := extractID(request)
	if err != nil {
		return nil, domainerrors.WrapValidationError(err, "invalid path parameter")
	}

	return ctx.WithGameID(userContext, id), nil
}

// BuildLobbyContext extracts UserContext and {id} from the request to build a LobbyContext.
func BuildLobbyContext(request *http.Request) (ctx.LobbyContext, error) {
	userContext, ok := request.Context().(ctx.UserContext)
	if !ok {
		return nil, errors.New("user context not found")
	}

	id, err := extractID(request)
	if err != nil {
		return nil, domainerrors.WrapValidationError(err, "invalid path parameter")
	}

	return ctx.WithLobbyID(userContext, id), nil
}

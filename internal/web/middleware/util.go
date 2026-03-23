package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"go.uber.org/zap"
)

func buildDomainContext[T ctx.UserContext](
	log *zap.SugaredLogger,
	routeToWrap route.Route,
	domain string,
	contextFunc func(ctx ctx.UserContext, ID int64) T,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Debugf("applying %s middleware", domain)

		result, err := buildContext[T](request, contextFunc)
		if err != nil {
			_ = restutils.WriteError(writer, err)

			return
		}

		routeToWrap.ServeHTTP(
			writer,
			request.WithContext(result),
		)
	}
}

func buildContext[T ctx.UserContext](
	request *http.Request,
	contextFunc func(ctx ctx.UserContext, ID int64) T,
) (T, error) {
	var result T

	extractedID, err := extractID(request)
	if err != nil {
		return result, domainerrors.WrapValidationError(err, "invalid path parameter")
	}

	userContext, ok := request.Context().(ctx.UserContext)
	if !ok {
		return result, errors.New("user context not found")
	}

	result = contextFunc(userContext, extractedID)

	return result, nil
}

func extractID(req *http.Request) (int64, error) {
	IDStr := req.PathValue("id")

	ID, err := strconv.Atoi(IDStr)
	if err != nil {
		return -1, fmt.Errorf("invalid id: %w", err)
	}

	return int64(ID), nil
}

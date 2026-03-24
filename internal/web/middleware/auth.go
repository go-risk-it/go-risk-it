package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware interface {
	Middleware
}

type AuthMiddlewareImpl struct {
	jwtConfig config.JwtConfig
}

var _ AuthMiddleware = (*AuthMiddlewareImpl)(nil)

func NewAuthMiddleware(jwtConfig config.JwtConfig) *AuthMiddlewareImpl {
	return &AuthMiddlewareImpl{jwtConfig: jwtConfig}
}

func (m *AuthMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	if !routeToWrap.RequiresAuth() {
		return routeToWrap
	}

	return route.NewRoute(
		routeToWrap.Pattern(),
		routeToWrap.RequiresAuth(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			traceContext, ok := request.Context().(ctx.TraceContext)
			if !ok {
				_ = restutils.WriteError(writer, errors.New("invalid trace context"))

				return
			}

			traceContext.Log().Debug("applying auth middleware")

			subject, err := m.verifyJWT(request)
			if err != nil {
				_ = restutils.WriteError(
					writer,
					domainerrors.WrapUnauthorizedError(err, "authentication failed"),
				)

				return
			}

			traceContext.Log().Debugw("Auth token is valid")

			userContext := ctx.WithUserID(traceContext, subject)

			routeToWrap.ServeHTTP(
				writer,
				request.WithContext(userContext),
			)
		}))
}

func (m *AuthMiddlewareImpl) verifyJWT(request *http.Request) (string, error) {
	authHeader := request.Header.Get("Authorization") // Bearer <token>
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return m.jwtConfig.Secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return "", errors.New("failed to parse claims")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return "", errors.New("failed to extract UserID")
	}

	return subject, nil
}

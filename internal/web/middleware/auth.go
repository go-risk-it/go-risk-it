package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthMiddleware interface {
	Middleware
}

type AuthMiddlewareImpl struct {
	log       *zap.SugaredLogger
	jwtConfig config.JwtConfig
}

var _ AuthMiddleware = (*AuthMiddlewareImpl)(nil)

func NewAuthMiddleware(log *zap.SugaredLogger, jwtConfig config.JwtConfig) AuthMiddleware {
	return &AuthMiddlewareImpl{
		log:       log,
		jwtConfig: jwtConfig,
	}
}

func (m *AuthMiddlewareImpl) Wrap(routeToWrap route.Route) route.Route {
	return route.NewRoute(
		routeToWrap.Pattern(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			m.log.Debug("Applying auth middleware")

			subject, err := m.verifyJWT(request)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusUnauthorized)

				return
			}

			m.log.Debugw("Auth token is valid")

			logContext, ok := request.Context().(ctx.LogContext)
			if !ok {
				http.Error(writer, "invalid log context", http.StatusInternalServerError)

				return
			}

			userContext := ctx.WithUserID(
				ctx.WithLog(
					request.Context(),
					logContext.Log().With("userID", subject),
				),
				subject,
			)

			routeToWrap.ServeHTTP(
				writer,
				request.WithContext(userContext),
			)
		}))
}

func (m *AuthMiddlewareImpl) verifyJWT(request *http.Request) (string, error) {
	authHeader := request.Header.Get("Authorization") // Bearer <token>
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return m.jwtConfig.Secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok {
		return "", fmt.Errorf("failed to parse claims")
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("failed to extract UserID")
	}

	return subject, nil
}

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthMiddleware interface {
	Wrap(route rest.Route) rest.Route
}

type AuthMiddlewareImpl struct {
	log       *zap.SugaredLogger
	jwtConfig config.JwtConfig
}

func NewAuthMiddleware(log *zap.SugaredLogger, jwtConfig config.JwtConfig) AuthMiddleware {
	return &AuthMiddlewareImpl{
		log:       log,
		jwtConfig: jwtConfig,
	}
}

func (m *AuthMiddlewareImpl) Wrap(route rest.Route) rest.Route {
	return rest.NewRoute(
		route.Pattern(),
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			authHeader := request.Header.Get("Authorization") // Bearer <token>
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return m.jwtConfig.Secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil {
				m.log.Errorw("failed to parse token", "token", tokenString, "err", err)
				rest.WriteResponse(writer, marshalError(err), http.StatusUnauthorized)

				return
			}

			if !token.Valid {
				m.log.Errorw("invalid token", "token", tokenString)
				rest.WriteResponse(
					writer,
					marshalError(fmt.Errorf("invalid token")),
					http.StatusUnauthorized,
				)

				return
			}

			m.log.Debugw("Auth token is valid")

			if _, ok := token.Claims.(jwt.MapClaims); !ok {
				m.log.Error("failed to parse claims")
				rest.WriteResponse(
					writer,
					marshalError(fmt.Errorf("failed to parse claims")),
					http.StatusUnauthorized,
				)

				return
			}

			subject, err := token.Claims.GetSubject()
			if err != nil {
				m.log.Errorw("failed to get subject", "err", err)
				rest.WriteResponse(
					writer,
					marshalError(fmt.Errorf("failed to extract UserID")),
					http.StatusUnauthorized,
				)
			}

			route.ServeHTTP(
				writer,
				request.WithContext(
					context.WithValue(request.Context(), riskcontext.UserIDKey, subject),
				),
			)
		}))
}

func marshalError(err error) []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
}

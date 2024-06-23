package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type AuthMiddleware interface {
	Wrap(handler http.Handler) http.Handler
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

func (m *AuthMiddlewareImpl) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization") // Bearer <token>
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return m.jwtConfig.Secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			m.log.Errorw("failed to parse token", "token", tokenString, "err", err)
			rest.WriteResponse(writer, m.log, marshalError(err), http.StatusUnauthorized)

			return
		}

		if !token.Valid {
			m.log.Errorw("invalid token", "token", tokenString)
			rest.WriteResponse(
				writer,
				m.log,
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
				m.log,
				marshalError(fmt.Errorf("failed to parse claims")),
				http.StatusUnauthorized,
			)

			return
		}

		handler.ServeHTTP(writer, request)
	})
}

func marshalError(err error) []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error()))
}

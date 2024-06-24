package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/riskcontext"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
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
			authHeader := request.Header.Get("Authorization") // Bearer <token>
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return m.jwtConfig.Secret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
			if err != nil {
				m.log.Errorw("failed to parse token", "token", tokenString, "err", err)
				restutils.WriteResponse(writer, marshalError(err), http.StatusUnauthorized)

				return
			}

			if !token.Valid {
				m.log.Errorw("invalid token", "token", tokenString)
				restutils.WriteResponse(
					writer,
					marshalError(fmt.Errorf("invalid token")),
					http.StatusUnauthorized,
				)

				return
			}

			m.log.Debugw("Auth token is valid")

			if _, ok := token.Claims.(jwt.MapClaims); !ok {
				m.log.Error("failed to parse claims")
				restutils.WriteResponse(
					writer,
					marshalError(fmt.Errorf("failed to parse claims")),
					http.StatusUnauthorized,
				)

				return
			}

			subject, err := token.Claims.GetSubject()
			if err != nil {
				m.log.Errorw("failed to get subject", "err", err)
				restutils.WriteResponse(
					writer,
					marshalError(fmt.Errorf("failed to extract UserID")),
					http.StatusUnauthorized,
				)
			}

			routeToWrap.ServeHTTP(
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

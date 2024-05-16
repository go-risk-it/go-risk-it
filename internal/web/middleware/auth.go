package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/config"
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
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return m.jwtConfig.Secret, nil
		})
		if err != nil {
			m.log.Errorw("Failed to parse token : ", "token", tokenString, "err", err)

			return
		}

		m.log.Debugf("Successfully validated token: %+v", token)

		// if claims, ok := token.Claims.(jwt.MapClaims); ok {
		//	fmt.Println(claims["foo"], claims["nbf"])
		// } else {
		//	fmt.Println(err)
		//}
		// ctx := context.WithValue(request.Context(), session.CurrentUserKey, user)
		// handler.ServeHTTP(writer, request.WithContext(ctx))
		handler.ServeHTTP(writer, request)
	})
}

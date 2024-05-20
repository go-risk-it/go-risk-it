package middleware

import (
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

		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return m.jwtConfig.Secret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			m.log.Errorw("Failed to parse token : ", "token", tokenString, "err", err)

			return
		}

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

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setup(t *testing.T) (middleware.AuthMiddleware, *httptest.ResponseRecorder) {
	t.Helper()

	jwtConfig := config.JwtConfig{
		Secret: []byte("secret"),
	}
	middleware := middleware.NewAuthMiddleware(jwtConfig)

	responseWriter := httptest.NewRecorder()

	return middleware, responseWriter
}

func TestAuthMiddleware_Wrap(t *testing.T) {
	t.Parallel()

	type inputType struct {
		name          string
		token         string
		expectedError string
		expectedCode  int
	}

	tests := []inputType{
		{
			"Should fail when token can't be parsed",
			"asd",
			"failed to parse token: token is malformed: token contains an invalid number of segments\n",
			http.StatusUnauthorized,
		},
		{
			"Should fail when token is invalid",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.AEqWPXS_88UL5a0bTWDj9OZdd83fZV03xsNMUdPZeg8",
			"failed to parse token: token signature is invalid: signature is invalid\n",
			http.StatusUnauthorized,
		},
		{
			"Should fail when token is expired",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjk2MTc2MTE3M30.qg2HxtFJf72fWP12IGVsUsbwNLaOSI9Kr3Ws-cjrlPo",
			"failed to parse token: token has invalid claims: token is expired\n",
			http.StatusUnauthorized,
		},
		{
			"Should succeed when token is valid",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.XbPfbIHMI6arZ3Y922BhjWgQzWXcXNrz0ogtVhfEd2o",
			"",
			http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			middleware, responseWriter := setup(t)

			wrappedHandler := middleware.Wrap(
				route.NewRoute(
					"/",
					true,
					http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
						writer.WriteHeader(http.StatusOK)
					})))

			request, _ := http.NewRequestWithContext(
				ctx.WithLog(context.Background(), zap.NewNop().Sugar()),
				http.MethodGet,
				"/",
				nil,
			)

			request.Header.Set("Authorization", "Bearer "+test.token)

			wrappedHandler.ServeHTTP(responseWriter, request)

			require.Equal(t, test.expectedError, responseWriter.Body.String())
			require.Equal(t, test.expectedCode, responseWriter.Code)
		})
	}
}

package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

func setup(t *testing.T) (*middleware.AuthMiddleware, *httptest.ResponseRecorder) {
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
		name            string
		token           string
		expectedCode    int
		expectedErrCode string
		errorContains   string
	}

	tests := []inputType{
		{
			"Should fail when token can't be parsed",
			"asd",
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"token is malformed",
		},
		{
			"Should fail when token is invalid",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.AEqWPXS_88UL5a0bTWDj9OZdd83fZV03xsNMUdPZeg8",
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"signature is invalid",
		},
		{
			"Should fail when token is expired",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjk2MTc2MTE3M30.qg2HxtFJf72fWP12IGVsUsbwNLaOSI9Kr3Ws-cjrlPo",
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"token is expired",
		},
		{
			"Should fail when token has no expiration",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.XbPfbIHMI6arZ3Y922BhjWgQzWXcXNrz0ogtVhfEd2o",
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"exp claim is required",
		},
		{
			"Should succeed when token is valid",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQxMDI0NDQ4MDAsImlhdCI6MTUxNjIzOTAyMiwibmFtZSI6IkpvaG4gRG9lIiwic3ViIjoiMTIzNDU2Nzg5MCJ9.AEzCmT-_46lhDrK0X-eUkUO8SDuxBvVcoR8STh9NvaE",
			http.StatusOK,
			"",
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			authMiddleware, responseWriter := setup(t)

			wrappedHandler := authMiddleware.Wrap(
				route.New(
					"/",
					true,
					http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
						writer.WriteHeader(http.StatusOK)
					})))

			request, _ := http.NewRequestWithContext(
				ctx.WithSpan(ctx.WithLog(t.Context(), zap.NewNop().Sugar()), noop.Span{}),
				http.MethodGet,
				"/",
				nil,
			)

			request.Header.Set("Authorization", "Bearer "+test.token)

			wrappedHandler.ServeHTTP(responseWriter, request)

			require.Equal(t, test.expectedCode, responseWriter.Code)

			if test.expectedErrCode != "" {
				var resp restutils.ErrorResponse
				require.NoError(t, json.Unmarshal(responseWriter.Body.Bytes(), &resp))
				assert.Equal(t, test.expectedErrCode, resp.Code)
				assert.Contains(t, resp.Error, test.errorContains)
			}
		})
	}
}

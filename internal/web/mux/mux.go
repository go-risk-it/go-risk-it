package mux

import (
	"net/http"

	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

func NewServeMux(
	routes []*route.Route,
	authMiddleware *middleware.AuthMiddleware,
	corsMiddleware *middleware.CorsMiddleware,
	logMiddleware *middleware.LogMiddleware,
	otelMiddleware *middleware.OTelMiddleware,
) http.Handler {
	mux := http.NewServeMux()

	for _, route := range routes {
		mux.Handle(
			route.Pattern(),
			logMiddleware.Wrap(
				otelMiddleware.Wrap(
					authMiddleware.Wrap(
						route,
					),
				),
			),
		)
	}

	// CORS wraps the entire mux so OPTIONS preflight requests are handled
	// before Go's ServeMux does method matching (which would return 405).
	//
	// otelhttp.NewHandler was intentionally removed: OTelMiddleware already
	// creates per-route spans with accurate http.route attributes. The
	// mux-level handler added a redundant root span with route="/", which
	// polluted spanmetrics with an uninformative aggregation.
	return corsMiddleware.WrapHandler(mux)
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewServeMux,
			fx.ParamTags(`group:"routes"`),
		),
	),
)

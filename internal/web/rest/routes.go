package rest

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
)

func ProvideRoutes(
	healthRoute *route.Route,
) []*route.Route {
	return []*route.Route{
		healthRoute,
	}
}

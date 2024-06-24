package game

import (
	"github.com/go-risk-it/go-risk-it/internal/web/rest/game/move"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"go.uber.org/fx"
)

var Module = fx.Options(
	move.Module,
	fx.Provide(
		route.AsRoute(NewGameHandler),
	),
)

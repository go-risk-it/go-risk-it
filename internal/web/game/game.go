package game

import (
	"github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/game/publisher"
	"github.com/go-risk-it/go-risk-it/internal/web/game/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/game/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	publisher.Module,
	controller.Module,
	rest.Module,
	ws.Module,
)

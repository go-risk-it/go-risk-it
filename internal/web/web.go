package web

import (
	"github.com/go-risk-it/go-risk-it/internal/web/game"
	gamecontroller "github.com/go-risk-it/go-risk-it/internal/web/game/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby"
	lobbycontroller "github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/mux"
	"github.com/go-risk-it/go-risk-it/internal/web/nbio"
	"github.com/go-risk-it/go-risk-it/internal/web/otel"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	game.Module,
	lobby.Module,
	middleware.Module,
	mux.Module,
	nbio.Module,
	otel.Module,
	rest.Module,
	ws.Module,
	fx.Provide(
		func(gc *gamecontroller.GameController) lobbycontroller.GameCreator { return gc },
	),
)

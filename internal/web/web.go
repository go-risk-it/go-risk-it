package web

import (
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/fetchers"
	"github.com/go-risk-it/go-risk-it/internal/web/handlers"
	"github.com/go-risk-it/go-risk-it/internal/web/nbio"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"github.com/go-risk-it/go-risk-it/internal/web/ws/connection"
	"go.uber.org/fx"
)

var Module = fx.Options(
	nbio.Module,
	controller.Module,
	fetchers.Module,
	ws.Module,
	rest.Module,
	fx.Provide(
		NewServeMux,
		connection.NewWebSocketHandler,
	),
	fx.Invoke(
		handlers.HandleBoardStateChanged,
		handlers.HandleGameStateChanged,
		handlers.HandlePlayerStateChanged,
		handlers.HandlePlayerConnected,
	),
)

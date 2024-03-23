package web

import (
	"github.com/tomfran/go-risk-it/internal/web/controller"
	"github.com/tomfran/go-risk-it/internal/web/fetchers"
	"github.com/tomfran/go-risk-it/internal/web/handlers"
	"github.com/tomfran/go-risk-it/internal/web/nbio"
	"github.com/tomfran/go-risk-it/internal/web/ws"
	"github.com/tomfran/go-risk-it/internal/web/ws/connection"
	"go.uber.org/fx"
)

var Module = fx.Options(
	nbio.Module,
	controller.Module,
	fetchers.Module,
	ws.Module,
	fx.Provide(
		NewServeMux,
		connection.NewWebSocketHandler,
	),
	fx.Invoke(
		handlers.HandleBoardStateChanged,
		handlers.HandlePlayerConnected,
	),
)

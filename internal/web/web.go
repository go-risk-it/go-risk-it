package web

import (
	"github.com/go-risk-it/go-risk-it/internal/web/game"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/mux"
	"github.com/go-risk-it/go-risk-it/internal/web/nbio"
	"github.com/go-risk-it/go-risk-it/internal/web/otel"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
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
)

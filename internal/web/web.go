package web

import (
	"github.com/go-risk-it/go-risk-it/internal/kernel/otelsetup"
	"github.com/go-risk-it/go-risk-it/internal/kernel/router"
	"github.com/go-risk-it/go-risk-it/internal/web/middleware"
	"github.com/go-risk-it/go-risk-it/internal/web/mux"
	"github.com/go-risk-it/go-risk-it/internal/web/nbio"
	"github.com/go-risk-it/go-risk-it/internal/web/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	middleware.Module,
	mux.Module,
	nbio.Module,
	otelsetup.Module,
	rest.Module,
	router.Module,
	ws.Module,
)

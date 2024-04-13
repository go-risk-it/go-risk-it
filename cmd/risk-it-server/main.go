package main

import (
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data"
	"github.com/go-risk-it/go-risk-it/internal/loggerfx"
	"github.com/go-risk-it/go-risk-it/internal/logic"
	"github.com/go-risk-it/go-risk-it/internal/web"
	"github.com/lesismal/nbio/nbhttp"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		loggerfx.Module,
		logic.Module,
		data.Module,
		web.Module,
		fx.Invoke(func(engine *nbhttp.Engine) {}),
	).Run()
}

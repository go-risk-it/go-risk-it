package main

import (
	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/db"
	"github.com/tomfran/go-risk-it/internal/loggerfx"
	"github.com/tomfran/go-risk-it/internal/logic"
	"github.com/tomfran/go-risk-it/internal/web"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		loggerfx.Module,
		logic.Module,
		db.Module,
		web.Module,
		fx.Invoke(func(engine *nbhttp.Engine) {}),
	).Run()
}

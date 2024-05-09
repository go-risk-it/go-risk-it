package main

import (
	"github.com/go-risk-it/go-risk-it/cmd/risk-it-server/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(app.Module).Run()
}

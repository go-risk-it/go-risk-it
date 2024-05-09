package main

import (
	"github.com/go-risk-it/go-risk-it/cmd/risk-it-server/app"
	"github.com/go-risk-it/go-risk-it/internal/testonly"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		app.Module,
		testonly.Module,
	).Run()
}

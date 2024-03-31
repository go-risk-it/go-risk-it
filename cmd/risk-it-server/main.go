package main

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game/rest/request"
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/data"
	"github.com/go-risk-it/go-risk-it/internal/loggerfx"
	"github.com/go-risk-it/go-risk-it/internal/logic"
	"github.com/go-risk-it/go-risk-it/internal/web"
	"github.com/go-risk-it/go-risk-it/internal/web/controller"
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
		fx.Invoke(func(gameController controller.GameController) error {
			ctx := context.TODO()

			_, err := gameController.CreateGame(ctx, request.CreateGame{
				Players: []string{"gabriele", "giovanni", "francesco", "vasilii"},
			})
			if err != nil {
				return fmt.Errorf("failed to create game: %w", err)
			}

			return nil
		}),
	).Run()
}

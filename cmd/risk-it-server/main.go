package main

import (
	"context"
	"fmt"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/tomfran/go-risk-it/internal/data"
	"github.com/tomfran/go-risk-it/internal/loggerfx"
	"github.com/tomfran/go-risk-it/internal/logic"
	"github.com/tomfran/go-risk-it/internal/logic/board"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/web"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		loggerfx.Module,
		logic.Module,
		data.Module,
		web.Module,
		fx.Invoke(func(engine *nbhttp.Engine) {}),
		fx.Invoke(func(boardService board.Service, gameService game.Service) error {
			ctx := context.TODO()

			gameBoard, err := boardService.FetchFromFile()
			if err != nil {
				return fmt.Errorf("failed to fetch board from file: %w", err)
			}

			err = gameService.CreateGameWithTx(
				ctx,
				gameBoard,
				[]string{"gabriele", "giovanni", "francesco", "vasilii"},
			)
			if err != nil {
				return fmt.Errorf("failed to create game: %w", err)
			}

			return nil
		}),
	).Run()
}

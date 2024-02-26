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

var gameBoard = &board.Board{
	Regions: []board.Region{
		{
			ExternalReference: "alaska",
			Name:              "Alaska",
			ContinentID:       1,
		},
		{
			ExternalReference: "northwest_territory",
			Name:              "Northwest Territory",
			ContinentID:       1,
		},
	},
	Continents: []board.Continent{
		{
			ExternalReference: "north_america",
			Name:              "North America",
			BonusTroops:       5,
		},
	},
	Borders: nil,
}

func main() {
	fx.New(
		loggerfx.Module,
		logic.Module,
		data.Module,
		web.Module,
		fx.Invoke(func(engine *nbhttp.Engine) {}),
		fx.Invoke(func(gameService game.Service) error {
			ctx := context.TODO()
			err := gameService.CreateGameWithTx(ctx, gameBoard, []string{"test"})
			if err != nil {
				return fmt.Errorf("failed to create game: %w", err)
			}

			return nil
		}),
	).Run()
}

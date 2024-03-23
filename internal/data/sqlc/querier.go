// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"context"
)

type Querier interface {
	DecreaseDeployableTroops(ctx context.Context, arg DecreaseDeployableTroopsParams) error
	GetGame(ctx context.Context, id int64) (Game, error)
	GetPlayerByUserId(ctx context.Context, userID string) (Player, error)
	GetPlayersByGame(ctx context.Context, gameID int64) ([]Player, error)
	GetRegionsByGame(ctx context.Context, id int64) ([]GetRegionsByGameRow, error)
	IncreaseRegionTroops(ctx context.Context, arg IncreaseRegionTroopsParams) error
	InsertGame(ctx context.Context) (int64, error)
	InsertPlayers(ctx context.Context, arg []InsertPlayersParams) (int64, error)
	InsertRegions(ctx context.Context, arg []InsertRegionsParams) (int64, error)
	SetGamePhase(ctx context.Context, arg SetGamePhaseParams) error
}

var _ Querier = (*Queries)(nil)

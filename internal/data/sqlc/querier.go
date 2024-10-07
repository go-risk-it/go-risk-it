// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package sqlc

import (
	"context"
)

type Querier interface {
	DecreaseDeployableTroops(ctx context.Context, arg DecreaseDeployableTroopsParams) error
	DrawCard(ctx context.Context, arg DrawCardParams) error
	GetAvailableCards(ctx context.Context, id int64) ([]Card, error)
	GetConquerPhaseState(ctx context.Context, id int64) (GetConquerPhaseStateRow, error)
	GetDeployableTroops(ctx context.Context, id int64) (int64, error)
	GetGame(ctx context.Context, id int64) (GetGameRow, error)
	GetPlayerByUserId(ctx context.Context, userID string) (Player, error)
	GetPlayersByGame(ctx context.Context, gameID int64) ([]Player, error)
	GetPlayersState(ctx context.Context, gameID int64) ([]GetPlayersStateRow, error)
	GetRegionsByGame(ctx context.Context, id int64) ([]GetRegionsByGameRow, error)
	HasConqueredInTurn(ctx context.Context, arg HasConqueredInTurnParams) (bool, error)
	IncreaseRegionTroops(ctx context.Context, arg IncreaseRegionTroopsParams) error
	InsertCards(ctx context.Context, arg []InsertCardsParams) (int64, error)
	InsertConquerPhase(ctx context.Context, arg InsertConquerPhaseParams) (ConquerPhase, error)
	InsertDeployPhase(ctx context.Context, arg InsertDeployPhaseParams) (DeployPhase, error)
	InsertGame(ctx context.Context) (Game, error)
	InsertPhase(ctx context.Context, arg InsertPhaseParams) (Phase, error)
	InsertPlayers(ctx context.Context, arg []InsertPlayersParams) (int64, error)
	InsertRegions(ctx context.Context, arg []InsertRegionsParams) (int64, error)
	SetGamePhase(ctx context.Context, arg SetGamePhaseParams) error
	UpdateRegionOwner(ctx context.Context, arg UpdateRegionOwnerParams) error
}

var _ Querier = (*Queries)(nil)

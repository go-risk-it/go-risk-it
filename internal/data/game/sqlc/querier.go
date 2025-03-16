// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package sqlc

import (
	"context"
)

type Querier interface {
	AssignGameWinner(ctx context.Context, arg AssignGameWinnerParams) error
	CreateMoveLog(ctx context.Context, arg CreateMoveLogParams) (GameMoveLog, error)
	DecreaseDeployableTroops(ctx context.Context, arg DecreaseDeployableTroopsParams) error
	DeleteSpuriousEliminatePlayerMissions(ctx context.Context, gameID int64) error
	DrawCard(ctx context.Context, arg DrawCardParams) error
	GetAvailableCards(ctx context.Context, id int64) ([]GameCard, error)
	GetCardsForPlayer(ctx context.Context, arg GetCardsForPlayerParams) ([]GetCardsForPlayerRow, error)
	GetConquerPhaseState(ctx context.Context, id int64) (GetConquerPhaseStateRow, error)
	GetCurrentPhase(ctx context.Context, id int64) (GamePhaseType, error)
	GetCurrentPlayer(ctx context.Context, gameID int64) (GamePlayer, error)
	GetDeployableTroops(ctx context.Context, id int64) (int64, error)
	GetEliminatePlayerMission(ctx context.Context, missionID int64) (GameEliminatePlayerMission, error)
	GetGame(ctx context.Context, id int64) (GetGameRow, error)
	GetMission(ctx context.Context, arg GetMissionParams) (GameMission, error)
	GetMoveLogs(ctx context.Context, arg GetMoveLogsParams) ([]GetMoveLogsRow, error)
	GetNextPlayer(ctx context.Context, gameID int64) (GamePlayer, error)
	GetPlayerByUserId(ctx context.Context, userID string) (GamePlayer, error)
	GetPlayerToEliminate(ctx context.Context, missionID int64) (string, error)
	GetPlayersByGame(ctx context.Context, gameID int64) ([]GamePlayer, error)
	GetPlayersState(ctx context.Context, gameID int64) ([]GetPlayersStateRow, error)
	GetRegionsByGame(ctx context.Context, id int64) ([]GetRegionsByGameRow, error)
	GetRegionsByPlayer(ctx context.Context, id int64) ([]GameRegion, error)
	GetTwoContinentsMission(ctx context.Context, missionID int64) (GameTwoContinentsMission, error)
	GetTwoContinentsPlusOneMission(ctx context.Context, missionID int64) (GameTwoContinentsPlusOneMission, error)
	GetUserGames(ctx context.Context, userID string) ([]int64, error)
	GrantRegionTroops(ctx context.Context, arg GrantRegionTroopsParams) error
	HasConqueredInTurn(ctx context.Context, arg HasConqueredInTurnParams) (bool, error)
	IncreaseRegionTroops(ctx context.Context, arg IncreaseRegionTroopsParams) error
	InsertCards(ctx context.Context, arg []InsertCardsParams) (int64, error)
	InsertConquerPhase(ctx context.Context, arg InsertConquerPhaseParams) (GameConquerPhase, error)
	InsertDeployPhase(ctx context.Context, arg InsertDeployPhaseParams) (GameDeployPhase, error)
	InsertEliminatePlayerMission(ctx context.Context, arg InsertEliminatePlayerMissionParams) error
	InsertGame(ctx context.Context) (GameGame, error)
	InsertMission(ctx context.Context, arg InsertMissionParams) (int64, error)
	InsertPhase(ctx context.Context, arg InsertPhaseParams) (GamePhase, error)
	InsertPlayers(ctx context.Context, arg []InsertPlayersParams) (int64, error)
	InsertRegions(ctx context.Context, arg []InsertRegionsParams) (int64, error)
	InsertTwoContinentsMission(ctx context.Context, arg InsertTwoContinentsMissionParams) error
	InsertTwoContinentsPlusOneMission(ctx context.Context, arg InsertTwoContinentsPlusOneMissionParams) error
	ReassignMissions(ctx context.Context, arg ReassignMissionsParams) error
	SetGamePhase(ctx context.Context, arg SetGamePhaseParams) error
	TransferCardsOwnership(ctx context.Context, arg TransferCardsOwnershipParams) error
	UnlinkCardsFromOwner(ctx context.Context, cards []int64) error
	UpdateRegionOwner(ctx context.Context, arg UpdateRegionOwnerParams) (int64, error)
}

var _ Querier = (*Queries)(nil)

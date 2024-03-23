package move

import (
	"context"
	"fmt"

	"github.com/tomfran/go-risk-it/internal/data/db"
	"github.com/tomfran/go-risk-it/internal/data/sqlc"
	"github.com/tomfran/go-risk-it/internal/logic/game"
	"github.com/tomfran/go-risk-it/internal/logic/player"
	"github.com/tomfran/go-risk-it/internal/logic/region"
	"github.com/tomfran/go-risk-it/internal/signals"
	"go.uber.org/zap"
)

type Service interface {
	PerformDeployMoveWithTx(
		ctx context.Context,
		gameID int64,
		player string,
		region string,
		troops int,
	) error
	PerformDeployMoveQ(
		ctx context.Context,
		querier db.Querier,
		gameID int64,
		player string,
		region string,
		troops int,
	) error
}

type ServiceImpl struct {
	log                     *zap.SugaredLogger
	querier                 db.Querier
	gameService             game.Service
	playerService           player.Service
	regionService           region.Service
	boardStateChangedSignal signals.BoardStateChangedSignal
}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService game.Service,
	playerService player.Service,
	regionService region.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                 que,
		log:                     log,
		gameService:             gameService,
		playerService:           playerService,
		regionService:           regionService,
		boardStateChangedSignal: boardStateChangedSignal,
	}
}

func (s *ServiceImpl) PerformDeployMoveWithTx(
	ctx context.Context,
	gameID int64,
	userID string,
	region string,
	troops int,
) error {
	if err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) error {
		return s.PerformDeployMoveQ(ctx, qtx, gameID, userID, region, troops)
	}); err != nil {
		return fmt.Errorf("failed to perform deploy move: %w", err)
	}

	return nil
}

func (s *ServiceImpl) PerformDeployMoveQ(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	userID string,
	region string,
	troops int,
) error {
	players, err := s.playerService.GetPlayersQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	playerState := getPlayerFrom(userID, players)
	if playerState == nil {
		return fmt.Errorf("player is not in game")
	}

	if err2 := s.checkTurn(ctx, querier, gameID, players, playerState.TurnIndex); err2 != nil {
		return fmt.Errorf("turn check failed: %w", err2)
	}

	regionState, err := s.getRegion(ctx, querier, gameID, region, playerState)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	err = s.playerService.DecreaseDeployableTroopsQ(ctx, querier, playerState, int64(troops))
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	err = s.regionService.IncreaseTroopsInRegion(ctx, querier, regionState.ID, int64(troops))
	if err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	if playerState.DeployableTroops == int64(troops) {
		err = s.gameService.SetGamePhaseQ(ctx, querier, gameID, sqlc.PhaseATTACK)
		if err != nil {
			return fmt.Errorf("failed to set game phase: %w", err)
		}
	}

	s.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: gameID,
	})

	return nil
}

func (s *ServiceImpl) getRegion(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	region string,
	playerState *sqlc.Player,
) (*sqlc.GetRegionsByGameRow, error) {
	regions, err := s.regionService.GetRegionsQ(ctx, querier, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get regions: %w", err)
	}

	regionState := getRegionFrom(region, regions)
	if regionState == nil {
		return nil, fmt.Errorf("region is not in game")
	}

	if regionState.PlayerName != playerState.UserID {
		return nil, fmt.Errorf("region is not owned by player")
	}

	return regionState, nil
}

func (s *ServiceImpl) checkTurn(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	players []sqlc.Player,
	playerTurn int64,
) error {
	gameState, err := s.gameService.GetGameStateQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	if gameState.Phase != sqlc.PhaseDEPLOY {
		return fmt.Errorf("game is not in deploy phase")
	}

	if gameState.Turn%int64(len(players)) != playerTurn {
		return fmt.Errorf("it is not the player's turn")
	}

	return nil
}

func getRegionFrom(region string, regions []sqlc.GetRegionsByGameRow) *sqlc.GetRegionsByGameRow {
	for _, r := range regions {
		if r.ExternalReference == region {
			return &r
		}
	}

	return nil
}

func getPlayerFrom(player string, players []sqlc.Player) *sqlc.Player {
	for _, p := range players {
		if p.UserID == player {
			return &p
		}
	}

	return nil
}

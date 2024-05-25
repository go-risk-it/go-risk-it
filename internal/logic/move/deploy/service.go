package deploy

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/signals"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type Service interface {
	PerformDeployMoveWithTx(
		ctx context.Context,
		gameID int64,
		userID string,
		region string,
		currentTroops int64,
		desiredTroops int64,
	) error
	PerformDeployMoveQ(
		ctx context.Context,
		querier db.Querier,
		gameID int64,
		userID string,
		region string,
		currentTroops int64,
		desiredTroops int64,
	) error
}

type ServiceImpl struct {
	log                      *zap.SugaredLogger
	querier                  db.Querier
	gameService              game.Service
	playerService            player.Service
	regionService            region.Service
	boardStateChangedSignal  signals.BoardStateChangedSignal
	playerStateChangedSignal signals.PlayerStateChangedSignal
	gameStateChangedSignal   signals.GameStateChangedSignal
}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService game.Service,
	playerService player.Service,
	regionService region.Service,
	boardStateChangedSignal signals.BoardStateChangedSignal,
	playerStateChangedSignal signals.PlayerStateChangedSignal,
	gameStateChangedSignal signals.GameStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                  que,
		log:                      log,
		gameService:              gameService,
		playerService:            playerService,
		regionService:            regionService,
		boardStateChangedSignal:  boardStateChangedSignal,
		playerStateChangedSignal: playerStateChangedSignal,
		gameStateChangedSignal:   gameStateChangedSignal,
	}
}

func (s *ServiceImpl) PerformDeployMoveWithTx(
	ctx context.Context,
	gameID int64,
	userID string,
	region string,
	currentTroops int64,
	desiredTroops int64,
) error {
	_, err := s.querier.ExecuteInTransactionWithIsolation(
		ctx,
		pgx.RepeatableRead,
		func(qtx db.Querier) (interface{}, error) {
			return nil, s.PerformDeployMoveQ(
				ctx,
				qtx,
				gameID,
				userID,
				region,
				currentTroops,
				desiredTroops,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to perform deploy move: %w", err)
	}

	go s.boardStateChangedSignal.Emit(ctx, signals.BoardStateChangedData{
		GameID: gameID,
	})
	go s.playerStateChangedSignal.Emit(ctx, signals.PlayerStateChangedData{
		GameID: gameID,
	})
	go s.gameStateChangedSignal.Emit(ctx, signals.GameStateChangedData{
		GameID: gameID,
	})

	return nil
}

func (s *ServiceImpl) PerformDeployMoveQ(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	userID string,
	region string,
	currentTroops int64,
	desiredTroops int64,
) error {
	s.log.Infow(
		"performing deploy move",
		"gameID",
		gameID,
		"userID",
		userID,
		"region",
		region,
		"currentTroops",
		currentTroops,
		"desiredTroops",
		desiredTroops,
	)

	players, err := s.playerService.GetPlayersQ(ctx, querier, gameID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	playerState := extractPlayerFrom(userID, players)
	if playerState == nil {
		return fmt.Errorf("player is not in game")
	}

	if err := s.checkTurn(ctx, querier, gameID, players, playerState.TurnIndex); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	troops := desiredTroops - currentTroops
	if playerState.DeployableTroops < troops {
		return fmt.Errorf("not enough deployable troops")
	}

	regionState, err := s.getRegion(ctx, querier, gameID, region, userID)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	if regionState.Troops != currentTroops {
		return fmt.Errorf("region has different number of troops than declared")
	}

	err = s.executeDeploy(ctx, querier, gameID, playerState, regionState, troops)
	if err != nil {
		return fmt.Errorf("failed to execute deploy: %w", err)
	}

	return nil
}

func (s *ServiceImpl) executeDeploy(ctx context.Context,
	querier db.Querier,
	gameID int64,
	player *sqlc.Player,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	s.log.Infow(
		"deploying",
		"player",
		player.UserID,
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)
	// player has 5 deployable troops
	// req1 -> decrease deployable troops by 3 (5 - 3 = 2)
	// req2 -> decrease deployable troops by 4 (5 - 4 = 1)
	// req 1 commit
	// req 2 commit
	// inconsistent state

	err := s.playerService.DecreaseDeployableTroopsQ(ctx, querier, player, troops)
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	err = s.regionService.IncreaseTroopsInRegion(ctx, querier, region.ID, troops)
	if err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	if player.DeployableTroops == troops {
		s.log.Infow("all deployable troops were deployed, advancing game phase", "gameID", gameID)

		err = s.gameService.SetGamePhaseQ(ctx, querier, gameID, sqlc.PhaseATTACK)
		if err != nil {
			return fmt.Errorf("failed to set game phase: %w", err)
		}
	}

	return nil
}

func (s *ServiceImpl) getRegion(
	ctx context.Context,
	querier db.Querier,
	gameID int64,
	region string,
	userID string,
) (*sqlc.GetRegionsByGameRow, error) {
	result, err := s.regionService.GetRegionQ(ctx, querier, gameID, region)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	if result.UserID != userID {
		return nil, fmt.Errorf("region is not owned by player")
	}

	return result, nil
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

func extractPlayerFrom(userID string, players []sqlc.Player) *sqlc.Player {
	for _, p := range players {
		if p.UserID == userID {
			return &p
		}
	}

	return nil
}

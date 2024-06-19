package deploy

import (
	"context"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/region"
	"go.uber.org/zap"
)

type MoveData struct {
	RegionID      string
	CurrentTroops int64
	DesiredTroops int64
}

type Service interface {
	move.Service[MoveData]
	PerformDeployMoveQ(
		ctx context.Context,
		querier db.Querier,
		move move.Move[MoveData],
	) error
}

type ServiceImpl struct {
	log           *zap.SugaredLogger
	querier       db.Querier
	gameService   game.Service
	playerService player.Service
	regionService region.Service
}

func NewService(
	que db.Querier,
	log *zap.SugaredLogger,
	gameService game.Service,
	playerService player.Service,
	regionService region.Service,
) *ServiceImpl {
	return &ServiceImpl{
		querier:       que,
		log:           log,
		gameService:   gameService,
		playerService: playerService,
		regionService: regionService,
	}
}

func (s *ServiceImpl) MustAdvanceQ(
	ctx context.Context,
	querier db.Querier,
	game *sqlc.Game,
) bool {
	return game.DeployableTroops == 0
}

func (s *ServiceImpl) PerformQ(
	ctx context.Context,
	querier db.Querier,
	move move.Move[MoveData],
) error {
	err := s.PerformDeployMoveQ(ctx, querier, move)
	if err != nil {
		return fmt.Errorf("failed to perform deploy move: %w", err)
	}

	return nil
}

func (s *ServiceImpl) PerformDeployMoveQ(
	ctx context.Context,
	querier db.Querier,
	move move.Move[MoveData],
) error {
	s.log.Infow(
		"performing deploy move",
		"gameID",
		move.GameID,
		"userID",
		move.UserID,
		"move",
		move,
	)

	game, err := s.gameService.GetGameStateQ(ctx, querier, move.GameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	players, err := s.playerService.GetPlayersQ(ctx, querier, move.GameID)
	if err != nil {
		return fmt.Errorf("failed to get players: %w", err)
	}

	thisPlayer := extractPlayerFrom(players, move.UserID)
	if thisPlayer == nil {
		return fmt.Errorf("player is not in game")
	}

	if err := checkTurn(
		game,
		int64(len(players)),
		thisPlayer.TurnIndex,
		sqlc.PhaseDEPLOY); err != nil {
		return fmt.Errorf("turn check failed: %w", err)
	}

	troops := move.Payload.DesiredTroops - move.Payload.CurrentTroops
	if game.DeployableTroops < troops {
		return fmt.Errorf("not enough deployable troops")
	}

	regionState, err := s.getRegion(ctx, querier, move.GameID, move.Payload.RegionID, move.UserID)
	if err != nil {
		return fmt.Errorf("failed to get region: %w", err)
	}

	if regionState.Troops != move.Payload.CurrentTroops {
		return fmt.Errorf("region has different number of troops than declared")
	}

	err = s.executeDeploy(ctx, querier, game, regionState, troops)
	if err != nil {
		return fmt.Errorf("failed to execute deploy: %w", err)
	}

	return nil
}

func extractPlayerFrom(players []sqlc.Player, userID string) *sqlc.Player {
	for _, p := range players {
		if p.UserID == userID {
			return &p
		}
	}

	return nil
}

func (s *ServiceImpl) executeDeploy(
	ctx context.Context,
	querier db.Querier,
	game *sqlc.Game,
	region *sqlc.GetRegionsByGameRow,
	troops int64,
) error {
	s.log.Infow(
		"deploying",
		"gameID",
		game.ID,
		"region",
		region.ExternalReference,
		"troops",
		troops,
	)

	err := s.gameService.DecreaseDeployableTroopsQ(ctx, querier, game, troops)
	if err != nil {
		return fmt.Errorf("failed to decrease deployable troops: %w", err)
	}

	err = s.regionService.IncreaseTroopsInRegion(ctx, querier, region.ID, troops)
	if err != nil {
		return fmt.Errorf("failed to increase region troops: %w", err)
	}

	return nil
}

func checkTurn(
	game *sqlc.Game,
	playersInGame int64,
	playerTurn int64,
	phase sqlc.Phase,
) error {
	// move to specific services
	if game.Phase != phase {
		return fmt.Errorf("game is not in %s phase", phase)
	}

	if game.Turn%playersInGame != playerTurn {
		return fmt.Errorf("it is not the player's turn")
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

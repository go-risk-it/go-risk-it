package snapshot

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
)

// PublicSnapshot contains all publicly visible game state aggregated from
// multiple sqlc queries into a single read model.
type PublicSnapshot struct {
	Game    sqlc.GetGameRow
	Phase   PhaseState
	Board   []sqlc.GetRegionsByGameRow
	Players []sqlc.GetPlayersStateRow
}

// PhaseState holds phase-specific data, discriminated by Type.
// Only the pointer field matching Type is non-nil.
type PhaseState struct {
	Type         sqlc.GamePhaseType
	DeployState  *DeployState
	ConquerState *sqlc.GetConquerPhaseStateRow
}

// DeployState holds deploy phase data.
// This is a snapshot-local type because the sqlc query GetDeployableTroops
// returns a bare int64 rather than a row struct.
type DeployState struct {
	DeployableTroops int64
}

// PrivateSnapshot contains per-player private state.
type PrivateSnapshot struct {
	Cards       []sqlc.GetCardsForPlayerRow
	MissionType sqlc.GameMissionType
	MissionID   int64
}

// Service provides aggregated game state snapshots for efficient read operations.
//
// Contract: All methods execute sequential queries on a single db.Querier connection.
// No goroutines are spawned within any method. This invariant is the architectural
// foundation that fixes DB pool starvation in signal handlers.
//
// Consistency caveat: Sequential queries may read across state boundaries if a
// concurrent move commits between queries. This is pre-existing behavior (same as
// the current scattered fetcher approach) and will be addressed by read-only
// transactions in future work.
type Service interface {
	// GetPublicSnapshot returns the full publicly visible game state, including
	// the game metadata, phase-specific state, board regions, and player summaries.
	GetPublicSnapshot(ctx ctx.GameContext) (*PublicSnapshot, error)

	// GetPrivateSnapshotsByUser returns private state for every player in the game,
	// keyed by user ID (string). Each snapshot contains the player's
	// card hand and base mission information. Internally maps player_id → user_id
	// via GetPlayersByGame, so callers need not perform the mapping themselves.
	GetPrivateSnapshotsByUser(ctx ctx.GameContext) (map[string]*PrivateSnapshot, error)
}

type service struct {
	querier db.Querier
}

var _ Service = (*service)(nil)

// NewService creates a new snapshot Service backed by the given querier.
func NewService(querier db.Querier) Service {
	return &service{
		querier: querier,
	}
}

func (s *service) GetPublicSnapshot(ctx ctx.GameContext) (*PublicSnapshot, error) {
	game, err := s.querier.GetGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("getting game: %w", err)
	}

	phaseState, err := s.getPhaseState(ctx, game.CurrentPhase)
	if err != nil {
		return nil, err
	}

	regions, err := s.querier.GetRegionsByGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("getting regions: %w", err)
	}

	players, err := s.querier.GetPlayersState(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("getting players state: %w", err)
	}

	return &PublicSnapshot{
		Game:    game,
		Phase:   phaseState,
		Board:   regions,
		Players: players,
	}, nil
}

func (s *service) getPhaseState(
	ctx ctx.GameContext,
	phaseType sqlc.GamePhaseType,
) (PhaseState, error) {
	switch phaseType {
	case sqlc.GamePhaseTypeDEPLOY:
		troops, err := s.querier.GetDeployableTroops(ctx, ctx.GameID())
		if err != nil {
			return PhaseState{}, fmt.Errorf("getting deploy phase state: %w", err)
		}

		return PhaseState{
			Type:        phaseType,
			DeployState: &DeployState{DeployableTroops: troops},
		}, nil
	case sqlc.GamePhaseTypeCONQUER:
		conquerState, err := s.querier.GetConquerPhaseState(ctx, ctx.GameID())
		if err != nil {
			return PhaseState{}, fmt.Errorf("getting conquer phase state: %w", err)
		}

		return PhaseState{
			Type:         phaseType,
			ConquerState: &conquerState,
		}, nil
	default:
		return PhaseState{Type: phaseType}, nil
	}
}

func (s *service) GetPrivateSnapshotsByUser(
	ctx ctx.GameContext,
) (map[string]*PrivateSnapshot, error) {
	players, err := s.querier.GetPlayersByGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("getting players by game: %w", err)
	}

	cards, err := s.querier.GetAllCardsForGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("getting all cards for game: %w", err)
	}

	missions, err := s.querier.GetAllMissionsForGame(ctx, ctx.GameID())
	if err != nil {
		return nil, fmt.Errorf("getting all missions for game: %w", err)
	}

	playerToUser := make(map[int64]string, len(players))
	for _, player := range players {
		playerToUser[player.ID] = player.UserID
	}

	snapshots := make(map[string]*PrivateSnapshot, len(missions))

	for _, mission := range missions {
		userID, ok := playerToUser[mission.PlayerID]
		if !ok {
			continue
		}

		snapshots[userID] = &PrivateSnapshot{
			MissionType: mission.Type,
			MissionID:   mission.ID,
		}
	}

	for _, card := range cards {
		playerID := card.PlayerID.Int64

		userID, found := playerToUser[playerID]
		if !found {
			continue
		}

		snapshot, ok := snapshots[userID]
		if !ok {
			continue
		}

		snapshot.Cards = append(snapshot.Cards, sqlc.GetCardsForPlayerRow{
			ID:       card.ID,
			CardType: card.CardType,
			Region:   card.Region,
		})
	}

	return snapshots, nil
}

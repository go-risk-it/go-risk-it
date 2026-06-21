package phase

import (
	"fmt"

	apisnapshot "github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/jackc/pgx/v5/pgtype"
)

// PhaseInsertParams carries all pre-computed inputs for InsertPhase so it
// performs zero database reads. Turn must reflect the current turn; getNextTurn
// computes the actual next turn (including dead-player skip) from these values.
type PhaseInsertParams struct {
	PhaseType    sqlc.GamePhaseType
	CurrentPhase sqlc.GamePhaseType
	Turn         int64
	Players      []apisnapshot.PlayerState
}

type Service interface {
	InsertPhase(
		ctx ctx.GameContext,
		querier db.Querier,
		params PhaseInsertParams,
	) (*sqlc.GamePhase, error)
}

type service struct{}

var _ Service = (*service)(nil)

func NewService() Service {
	return &service{}
}

func (s *service) InsertPhase(
	ctx ctx.GameContext,
	querier db.Querier,
	params PhaseInsertParams,
) (*sqlc.GamePhase, error) {
	if params.PhaseType == params.CurrentPhase {
		return nil, fmt.Errorf("game already in desired phase: %v", params.PhaseType)
	}

	turn := getNextTurn(params)

	phase, err := insertPhase(ctx, querier, ctx.GameID(), params.PhaseType, turn)
	if err != nil {
		return nil, fmt.Errorf("failed to create new phase: %w", err)
	}

	err = setGamePhase(ctx, querier, ctx.GameID(), phase.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to set game phase: %w", err)
	}

	return phase, nil
}

// getNextTurn computes the next turn from pre-computed state. When the current
// phase is REINFORCE, the turn advances and skips dead players.
func getNextTurn(params PhaseInsertParams) int64 {
	turn := params.Turn

	if params.CurrentPhase == sqlc.GamePhaseTypeREINFORCE {
		turn++

		players := int64(len(params.Players))
		for params.Players[turn%players].Status == apisnapshot.PlayerDead {
			turn++
		}
	}

	return turn
}

func insertPhase(
	ctx kernelctx.UserContext,
	querier db.Querier,
	gameID int64,
	phaseType sqlc.GamePhaseType,
	turn int64,
) (*sqlc.GamePhase, error) {
	phase, err := querier.InsertPhase(ctx, sqlc.InsertPhaseParams{
		GameID: gameID,
		Type:   phaseType,
		Turn:   turn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create phase: %w", err)
	}

	return &phase, nil
}

func setGamePhase(
	ctx kernelctx.UserContext,
	querier db.Querier,
	gameID int64,
	phaseID int64,
) error {
	err := querier.SetGamePhase(ctx, sqlc.SetGamePhaseParams{
		ID:             gameID,
		CurrentPhaseID: pgtype.Int8{Int64: phaseID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to set phase: %w", err)
	}

	return nil
}

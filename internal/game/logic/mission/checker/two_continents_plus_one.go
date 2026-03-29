package checker

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
)

type TwoContinentsPlusOneChecker struct {
	boardService board.Service
}

var _ MissionChecker = (*TwoContinentsPlusOneChecker)(nil)

func NewTwoContinentsPlusOneChecker(boardService board.Service) *TwoContinentsPlusOneChecker {
	return &TwoContinentsPlusOneChecker{boardService: boardService}
}

func (c *TwoContinentsPlusOneChecker) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeTWOCONTINENTSPLUSONE
}

func (c *TwoContinentsPlusOneChecker) Check(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	mission, err := querier.GetTwoContinentsPlusOneMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get two continents plus one mission: %w", err)
	}

	continents, err := c.boardService.GetContinentsControlledByPlayer(
		ctx,
		querier,
		baseMission.PlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get continents controlled by player: %w", err)
	}

	playerControlsTwoMandatoryContinents := slices.ContainsFunc(
		continents,
		continentEquals(mission.Continent1),
	) &&
		slices.ContainsFunc(continents, continentEquals(mission.Continent2))

	return playerControlsTwoMandatoryContinents && len(continents) > 2, nil
}

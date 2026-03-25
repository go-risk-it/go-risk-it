package checker

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/db"
	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
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

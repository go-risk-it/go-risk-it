package checker

import (
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
)

type TwoContinentsChecker struct {
	boardService board.Service
}

var _ MissionChecker = (*TwoContinentsChecker)(nil)

func NewTwoContinentsChecker(boardService board.Service) *TwoContinentsChecker {
	return &TwoContinentsChecker{boardService: boardService}
}

func (c *TwoContinentsChecker) Type() sqlc.GameMissionType {
	return sqlc.GameMissionTypeTWOCONTINENTS
}

func (c *TwoContinentsChecker) Check(
	ctx ctx.GameContext,
	querier db.Querier,
	baseMission sqlc.GameMission,
) (bool, error) {
	mission, err := querier.GetTwoContinentsMission(ctx, baseMission.ID)
	if err != nil {
		return false, fmt.Errorf("failed to get two continents mission: %w", err)
	}

	continents, err := c.boardService.GetContinentsControlledByPlayer(
		ctx,
		querier,
		baseMission.PlayerID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to get continents controlled by player: %w", err)
	}

	return slices.ContainsFunc(continents, continentEquals(mission.Continent1)) &&
		slices.ContainsFunc(continents, continentEquals(mission.Continent2)), nil
}

func continentEquals(cont string) func(continent *board.Continent) bool {
	return func(continent *board.Continent) bool {
		return continent.ExternalReference == cont
	}
}

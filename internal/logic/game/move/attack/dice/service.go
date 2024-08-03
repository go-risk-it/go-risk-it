package dice

import (
	"math/rand"

	"github.com/go-risk-it/go-risk-it/internal/config"
	roller2 "github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice/roller"
)

type Service interface {
	Roll(n int) []int
}

type ServiceImpl struct {
	roller roller2.Roller
}

var _ Service = (*ServiceImpl)(nil)

func (s *ServiceImpl) Roll(dices int) []int {
	result := make([]int, 0, dices)

	for i := 0; i < dices; i++ {
		result = append(result, s.roller.Roll())
	}

	return result
}

func New(diceConfig config.DiceConfig) *ServiceImpl {
	return &ServiceImpl{roller: getDiceRoller(diceConfig)}
}

func getDiceRoller(diceConfig config.DiceConfig) roller2.Roller {
	switch diceConfig.RollStrategy {
	case "fixed":
		return roller2.WithSequence([]int{1})
	case "random":
		return roller2.WithRandomSource(rand.NewSource(0))
	default:
		panic("unknown roll strategy: " + diceConfig.RollStrategy)
	}
}

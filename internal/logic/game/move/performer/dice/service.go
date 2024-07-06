package dice

import (
	"math/rand"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/performer/dice/roller"
)

type Service interface {
	Roll(n int) []int
}

type ServiceImpl struct {
	roller roller.Roller
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

func getDiceRoller(diceConfig config.DiceConfig) roller.Roller {
	switch diceConfig.RollStrategy {
	case "fixed":
		return roller.WithSequence([]int{1})
	case "random":
		return roller.WithRandomSource(rand.NewSource(0))
	default:
		panic("unknown roll strategy: " + diceConfig.RollStrategy)
	}
}

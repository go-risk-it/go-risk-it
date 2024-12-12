package dice

import (
	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice/roller"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type Service interface {
	RollAttackingDices(n int) []int
	RollDefendingDices(n int) []int
}

type ServiceImpl struct {
	attackingRoller roller.Roller
	defendingRoller roller.Roller
}

var _ Service = (*ServiceImpl)(nil)

func (s *ServiceImpl) RollAttackingDices(dices int) []int {
	return roll(dices, s.attackingRoller)
}

func (s *ServiceImpl) RollDefendingDices(n int) []int {
	return roll(n, s.defendingRoller)
}

func roll(dices int, roller roller.Roller) []int {
	result := make([]int, 0, dices)

	for range dices {
		result = append(result, roller.Roll())
	}

	return result
}

func New(diceConfig config.DiceConfig, rng rand.RNG) *ServiceImpl {
	attackingRoller, defendingRoller := getDiceRollers(diceConfig, rng)

	return &ServiceImpl{
		attackingRoller: attackingRoller,
		defendingRoller: defendingRoller,
	}
}

func getDiceRollers(diceConfig config.DiceConfig, rng rand.RNG) (roller.Roller, roller.Roller) {
	switch diceConfig.RollStrategy {
	case "attacker_always_wins":
		return roller.WithSequence([]int{6}), roller.WithSequence([]int{1})
	case "attacker_always_loses":
		return roller.WithSequence([]int{1}), roller.WithSequence([]int{6})
	case "random":
		return roller.WithRandomSource(rng), roller.WithRandomSource(rng)
	default:
		panic("unknown roll strategy: " + diceConfig.RollStrategy)
	}
}

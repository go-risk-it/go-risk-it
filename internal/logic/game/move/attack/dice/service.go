package dice

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice/roller"
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type Service interface {
	RollAttackingDices(n int) []int
	RollDefendingDices(n int) []int
}

type service struct {
	attackingRoller roller.Roller
	defendingRoller roller.Roller
}

var _ Service = (*service)(nil)

func (s *service) RollAttackingDices(dices int) []int {
	return roll(dices, s.attackingRoller)
}

func (s *service) RollDefendingDices(n int) []int {
	return roll(n, s.defendingRoller)
}

func roll(dices int, roller roller.Roller) []int {
	result := make([]int, 0, dices)

	for range dices {
		result = append(result, roller.Roll())
	}

	return result
}

func New(diceConfig config.DiceConfig, rng rand.RNG) (Service, error) {
	attackingRoller, defendingRoller, err := getDiceRollers(diceConfig, rng)
	if err != nil {
		return nil, err
	}

	return &service{
		attackingRoller: attackingRoller,
		defendingRoller: defendingRoller,
	}, nil
}

func getDiceRollers(
	diceConfig config.DiceConfig,
	rng rand.RNG,
) (roller.Roller, roller.Roller, error) {
	switch diceConfig.RollStrategy {
	case "attacker_always_wins":
		return roller.WithSequence([]int{6}), roller.WithSequence([]int{1}), nil
	case "attacker_always_loses":
		return roller.WithSequence([]int{1}), roller.WithSequence([]int{6}), nil
	case "random":
		return roller.WithRandomSource(rng), roller.WithRandomSource(rng), nil
	default:
		return nil, nil, fmt.Errorf("unknown roll strategy: %s", diceConfig.RollStrategy)
	}
}

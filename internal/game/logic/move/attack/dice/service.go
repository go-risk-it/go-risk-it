package dice

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/config"
	"github.com/go-risk-it/go-risk-it/internal/game/rand"
)

type Service interface {
	RollAttackingDices(n int) []int
	RollDefendingDices(n int) []int
}

type service struct {
	attackingRoller Roller
	defendingRoller Roller
}

var _ Service = (*service)(nil)

func (s *service) RollAttackingDices(dices int) []int {
	return roll(dices, s.attackingRoller)
}

func (s *service) RollDefendingDices(n int) []int {
	return roll(n, s.defendingRoller)
}

func roll(dices int, r Roller) []int {
	result := make([]int, 0, dices)

	for range dices {
		result = append(result, r.Roll())
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
) (Roller, Roller, error) {
	switch diceConfig.RollStrategy {
	case "attacker_always_wins":
		return WithSequence([]int{6}), WithSequence([]int{1}), nil
	case "attacker_always_loses":
		return WithSequence([]int{1}), WithSequence([]int{6}), nil
	case "random":
		return WithRandomSource(rng), WithRandomSource(rng), nil
	default:
		return nil, nil, fmt.Errorf("unknown roll strategy: %s", diceConfig.RollStrategy)
	}
}

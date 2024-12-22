package rand

import (
	"math/rand/v2"

	"go.uber.org/fx"
)

type RNG interface {
	Shuffle(n int, swap func(i, j int))
	IntN(n int) int
}

func NewRNG() RNG {
	return rand.New(rand.NewPCG(420, 69))
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			NewRNG,
			fx.As(new(RNG)),
		),
	),
)

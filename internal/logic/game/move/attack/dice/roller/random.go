package roller

import (
	"github.com/go-risk-it/go-risk-it/internal/rand"
)

type Random struct {
	rng rand.RNG
}

var _ Roller = (*Random)(nil)

func WithRandomSource(src rand.RNG) Roller {
	return &Random{rng: src}
}

func (r *Random) Roll() int {
	return r.rng.IntN(6) + 1
}

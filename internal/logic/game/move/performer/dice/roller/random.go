package roller

import (
	"math/rand"
)

type Random struct {
	r *rand.Rand
}

var _ Roller = (*Random)(nil)

func WithRandomSource(src rand.Source) Roller {
	return &Random{rand.New(src)}
}

func (r *Random) Roll() int {
	return r.r.Intn(6) + 1
}

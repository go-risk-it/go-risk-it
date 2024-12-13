package roller_test

import (
	"math/rand/v2"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice/roller"
)

func Test_Random_Roll_ReturnsFromSource(t *testing.T) {
	t.Parallel()

	source := rand.New(rand.NewPCG(69, 420))
	roller := roller.WithRandomSource(source)

	testSource := rand.New(rand.NewPCG(69, 420))

	testRand := rand.New(testSource)

	for range 100 {
		expected := testRand.IntN(6) + 1

		actual := roller.Roll()
		if expected != actual {
			t.Errorf("Random.Roll expected %v, got %v", expected, actual)
		}
	}
}

package roller_test

import (
	"math/rand"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/dice/roller"
)

func Test_Random_Roll_ReturnsFromSource(t *testing.T) {
	t.Parallel()

	source := rand.NewSource(42)
	roller := roller.WithRandomSource(source)

	testSource := rand.NewSource(42)
	testRand := rand.New(testSource)

	for i := 0; i < 100; i++ {
		expected := testRand.Intn(6) + 1

		actual := roller.Roll()
		if expected != actual {
			t.Errorf("Random.Roll expected %v, got %v", expected, actual)
		}
	}
}

package roller_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack/dice/roller"
)

func Test_Sequence_Roll_ReturnsSequence(t *testing.T) {
	t.Parallel()

	seq := []int{1, 2, 3, 4, 5}

	roller := roller.WithSequence(seq)

	for _, expected := range seq {
		actual := roller.Roll()
		if expected != actual {
			t.Errorf("Sequence.Roll expected %v, got %v", expected, actual)
		}
	}
}

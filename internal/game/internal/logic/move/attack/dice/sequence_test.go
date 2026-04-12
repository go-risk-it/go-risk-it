package dice_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/attack/dice"
)

func Test_Sequence_Roll_ReturnsSequence(t *testing.T) {
	t.Parallel()

	seq := []int{1, 2, 3, 4, 5}

	r := dice.WithSequence(seq)

	for _, expected := range seq {
		actual := r.Roll()
		if expected != actual {
			t.Errorf("Sequence.Roll expected %v, got %v", expected, actual)
		}
	}
}

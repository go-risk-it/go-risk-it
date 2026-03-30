package dice_test

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack/dice"
)

func ExampleWithSequence() {
	r := dice.WithSequence([]int{6, 3, 1})

	fmt.Println(r.Roll())
	fmt.Println(r.Roll())
	fmt.Println(r.Roll())
	fmt.Println(r.Roll()) // wraps around to start
	// Output:
	// 6
	// 3
	// 1
	// 6
}

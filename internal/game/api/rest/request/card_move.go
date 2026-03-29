package request

import (
	"errors"
	"fmt"
)

type CardCombination struct {
	CardIDs []int64 `json:"cardIds"`
}

type CardsMove struct {
	Combinations []CardCombination `json:"combinations"`
}

func (r CardsMove) Validate() error {
	if len(r.Combinations) == 0 {
		return errors.New("combinations must not be empty")
	}

	for idx, combo := range r.Combinations {
		if len(combo.CardIDs) == 0 {
			return fmt.Errorf("combinations[%d].cardIds must not be empty", idx)
		}
	}

	return nil
}

package request

import "errors"

type ConquerMove struct {
	Troops int64 `json:"troops"`
}

func (r ConquerMove) Validate() error {
	if r.Troops <= 0 {
		return errors.New("troops must be positive")
	}

	if r.Troops > maxTroops {
		return errors.New("troops exceeds maximum")
	}

	return nil
}

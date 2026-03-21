package request

import "errors"

type ConquerMove struct {
	Troops int64 `json:"troops"`
}

func (r ConquerMove) Validate() error {
	if r.Troops <= 0 {
		return errors.New("troops must be positive")
	}

	return nil
}

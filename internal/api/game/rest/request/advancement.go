package request

import (
	"errors"

	"github.com/go-risk-it/go-risk-it/internal/api/game"
)

type Advancement struct {
	CurrentPhase game.PhaseType `json:"currentPhase"`
}

func (r Advancement) Validate() error {
	if r.CurrentPhase == "" {
		return errors.New("currentPhase must not be empty")
	}

	return nil
}

package request

import (
	"github.com/go-risk-it/go-risk-it/internal/api/game"
)

type Advancement struct {
	Phase game.PhaseType `json:"type"`
}

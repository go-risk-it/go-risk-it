package signals

import (
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/maniartech/signals"
)

type MovePerformedData struct {
	MoveLog sqlc.MoveLog
}

type MovePerformedSignal interface {
	signals.Signal[MovePerformedData]
}

func NewMovePerformedSignal() MovePerformedSignal {
	return signals.New[MovePerformedData]()
}

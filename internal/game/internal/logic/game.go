package game

import (
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/card"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/creation"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/internal/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"go.uber.org/fx"
)

var Module = fx.Options(
	board.Module,
	card.Module,
	creation.Module,
	bus.Module,
	gamemetrics.Module,
	mission.Module,
	move.Module,
	phase.Module,
	player.Module,
	region.Module,
	state.Module,
)

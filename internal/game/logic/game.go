package game

import (
	"github.com/go-risk-it/go-risk-it/internal/game/logic/advancement"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/card"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/creation"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"go.uber.org/fx"
)

var Module = fx.Options(
	advancement.Module,
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

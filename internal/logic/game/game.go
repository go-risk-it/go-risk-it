package game

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/advancement"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/card"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/creation"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/mission"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/phase"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/state"
	"go.uber.org/fx"
)

var Module = fx.Options(
	advancement.Module,
	board.Module,
	card.Module,
	creation.Module,
	mission.Module,
	move.Module,
	phase.Module,
	player.Module,
	signals.Module,
	region.Module,
	state.Module,
)

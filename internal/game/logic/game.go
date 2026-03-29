package game

import (
	"github.com/go-risk-it/go-risk-it/internal/game/logic/advancement"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/card"
	gameconfig "github.com/go-risk-it/go-risk-it/internal/game/logic/config"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/headlines"
	gamemetrics "github.com/go-risk-it/go-risk-it/internal/game/logic/metrics"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/mission"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/phase"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/region"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/timing"
	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/logger"
	"go.uber.org/fx"
)

var Module = fx.Options(
	advancement.Module,
	board.Module,
	card.Module,
	gameconfig.Module,
	creation.Module,
	bus.Module,
	gamemetrics.Module,
	headlines.Module,
	logger.Module,
	mission.Module,
	move.Module,
	phase.Module,
	player.Module,
	snapshot.Module,
	region.Module,
	state.Module,
	timing.Module,
)

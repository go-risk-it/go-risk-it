package game

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/game/board"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/player"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/region"
	"go.uber.org/fx"
)

var Module = fx.Options(
	board.Module,
	gamestate.Module,
	move.Module,
	region.Module,
	player.Module,
)

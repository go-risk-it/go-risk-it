package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/creation"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/management"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/start"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/state"
	"go.uber.org/fx"
)

var Module = fx.Options(
	creation.Module,
	management.Module,
	signals.Module,
	state.Module,
	start.Module,
)

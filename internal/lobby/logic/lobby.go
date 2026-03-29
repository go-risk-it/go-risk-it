package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/management"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/start"
	"github.com/go-risk-it/go-risk-it/internal/lobby/logic/state"
	"go.uber.org/fx"
)

var Module = fx.Options(
	creation.Module,
	management.Module,
	state.Module,
	start.Module,
)

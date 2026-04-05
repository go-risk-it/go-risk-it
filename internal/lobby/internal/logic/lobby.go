package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/management"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/start"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/state"
	"go.uber.org/fx"
)

var Module = fx.Options(
	creation.Module,
	management.Module,
	state.Module,
	start.Module,
)

package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/creation"
	"go.uber.org/fx"
)

var Module = fx.Options(
	creation.Module,
)

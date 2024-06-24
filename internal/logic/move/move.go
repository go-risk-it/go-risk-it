package move

import (
	"github.com/go-risk-it/go-risk-it/internal/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/logic/move/performer"
	"go.uber.org/fx"
)

var Module = fx.Options(
	performer.Module,
	orchestration.Module,
)

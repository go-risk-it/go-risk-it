package lobby

import (
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/controller"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/rest"
	"github.com/go-risk-it/go-risk-it/internal/web/lobby/ws"
	"go.uber.org/fx"
)

var Module = fx.Options(
	controller.Module,
	rest.Module,
	ws.Module,
)

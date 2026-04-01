package runner

import (
	"github.com/go-risk-it/go-risk-it/perf-test/internal/chaos"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/client"
)

// ChaosHandler injects fault scenarios after successful moves.
type ChaosHandler struct {
	injector *chaos.Injector
	gameCtx  *GameSession
	// For testing: overrides the default handle logic.
	maybeDisconnectFn func(players []*PlayerInfo, activeIdx int)
}

// NewChaosHandler creates a ChaosHandler. If injector is nil, the handler is a no-op.
func NewChaosHandler(injector *chaos.Injector, gameCtx *GameSession) *ChaosHandler {
	return &ChaosHandler{injector: injector, gameCtx: gameCtx}
}

// Register subscribes to EventMoveSucceeded if an injector is configured.
func (h *ChaosHandler) Register(bus *Bus) {
	if h.injector == nil && h.maybeDisconnectFn == nil {
		return
	}

	bus.On(EventMoveSucceeded, h.handle)
}

func (h *ChaosHandler) handle(_ *Bus, _ Event) {
	if h.maybeDisconnectFn != nil {
		h.maybeDisconnectFn(h.gameCtx.Players, 0)

		return
	}

	handles := make([]chaos.PlayerHandle, len(h.gameCtx.Players))

	for i, p := range h.gameCtx.Players {
		ws, ok := p.WS.(*client.WS)
		if !ok {
			return
		}

		handles[i] = chaos.PlayerHandle{Name: p.Name, WS: ws}
	}

	h.injector.MaybeDisconnect(handles, 0)
}

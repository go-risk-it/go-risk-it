package runner

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"go.opentelemetry.io/otel/attribute"
)

// ProtocolHandler handles GameStarted: signs up players, creates a game,
// connects WebSockets, and waits for initial state.
type ProtocolHandler struct {
	baseURL  string
	wsURL    string
	anonKey  string
	timeouts Timeouts
	gameCtx  *GameSession

	// Factories for testability.
	newAuth func(baseURL, anonKey string) AuthClient
	newREST func(baseURL, token string, collector *metrics.StepAccumulator) RESTClient
	//nolint:lll // long string literal
	newWS func(wsURL string, gameID int64, token string, collector *metrics.StepAccumulator) (WSClient, error)
}

// Register subscribes to EventGameStarted.
func (h *ProtocolHandler) Register(bus *Bus) {
	bus.On(EventGameStarted, h.handle)
}

//nolint:funlen // sequential CLI/report logic
func (h *ProtocolHandler) handle(bus *Bus, e Event) {
	evt := e.(GameStartedEvent) //nolint:forcetypeassert // event bus guarantees type
	gameCtx := h.gameCtx
	gameCtx.GameIndex = evt.GameIndex

	emitFatal := func(err error) {
		bus.Emit(GameCompleteEvent{Result: GameResult{
			GameIndex:  gameCtx.GameIndex,
			FatalError: err,
		}})
	}

	// Shared transport for connection pooling across all players.
	transport := &http.Transport{
		MaxIdleConnsPerHost: evt.NumPlayers,
		IdleConnTimeout:     90 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
	_ = transport // used by default newREST factory

	// 1. Sign up players.
	auth := h.newAuth(h.baseURL, h.anonKey)
	players := make([]*PlayerInfo, evt.NumPlayers)

	for i := range evt.NumPlayers {
		email := fmt.Sprintf(
			"perf-g%dp%d-%d@test.local",
			evt.GameIndex, i, time.Now().UnixNano(),
		)
		password := "perftest123"

		authResult, err := auth.Signup(email, password)
		if err != nil {
			emitFatal(fmt.Errorf("signup player %d: %w", i, err))

			return
		}

		players[i] = &PlayerInfo{
			UserID: authResult.UserID,
			Name:   fmt.Sprintf("bot-%d-%d", evt.GameIndex, i),
			Auth:   authResult,
			REST:   h.newREST(h.baseURL, authResult.AccessToken, gameCtx.Accumulator),
		}
	}

	observe.Info(context.Background(), "players authenticated",
		attribute.Int("gameIndex", evt.GameIndex),
		attribute.Int("numPlayers", evt.NumPlayers),
	)

	// 2. Create game via first player.
	gamePlayers := make([]client.CreateGamePlayer, evt.NumPlayers)
	for i, p := range players {
		gamePlayers[i] = client.CreateGamePlayer{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	gameID, err := players[0].REST.CreateGame(
		context.Background(),
		client.CreateGameRequest{Players: gamePlayers},
	)
	if err != nil {
		emitFatal(fmt.Errorf("create game: %w", err))

		return
	}

	observe.Info(context.Background(), "game created",
		attribute.Int("gameIndex", evt.GameIndex),
		attribute.Int64("game_id", gameID),
	)

	// 3. All players connect WebSocket.
	for i, p := range players {
		ws, err := h.newWS(h.wsURL, gameID, p.Auth.AccessToken, gameCtx.Accumulator)
		if err != nil {
			emitFatal(fmt.Errorf("ws connect player %d: %w", i, err))

			return
		}

		p.WS = ws
	}

	observe.Info(context.Background(), "all players connected via WebSocket",
		attribute.Int("gameIndex", evt.GameIndex),
	)

	// 4. Wait for initial state.
	time.Sleep(h.timeouts.InitialStateWait)

	// 5. Populate game context.
	userIndex := make(map[string]int)
	for i, p := range players {
		userIndex[p.UserID] = i
	}

	gameCtx.GameID = gameID
	gameCtx.Players = players
	gameCtx.UserIndex = userIndex

	// 6. Emit initial state.
	snap := players[0].WS.View().Snapshot()
	bus.Emit(StateReceivedEvent{
		Snapshot:  snap,
		Timestamp: time.Now(),
	})
}

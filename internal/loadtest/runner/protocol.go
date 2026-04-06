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
	"github.com/go-risk-it/go-risk-it/internal/loadtest/userpool"
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
	userPool *userpool.Pool // nil = legacy per-game signup

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

	players, err := h.signupPlayers(evt)
	if err != nil {
		emitFatal(err)

		return
	}

	gameID, err := h.createGame(evt, players)
	if err != nil {
		emitFatal(err)

		return
	}

	if err := h.connectWebSockets(evt, players, gameID); err != nil {
		emitFatal(err)

		return
	}

	h.populateSession(gameID, players)
	// No initial state emission — the barrier-driven game loop handles the
	// first move via the initial playerView WS notification from game creation.
}

// signupPlayers authenticates all players and returns their info.
// When a UserPool is configured, users are acquired from the pool instead of
// creating fresh Supabase accounts, eliminating signup churn at high concurrency.
func (h *ProtocolHandler) signupPlayers(
	evt GameStartedEvent,
) ([]*PlayerInfo, error) {
	if h.userPool != nil {
		return h.acquireFromPool(evt)
	}

	return h.signupFresh(evt)
}

func (h *ProtocolHandler) acquireFromPool(evt GameStartedEvent) ([]*PlayerInfo, error) {
	entries, err := h.userPool.Acquire(h.gameCtx.Ctx, evt.NumPlayers)
	if err != nil {
		return nil, fmt.Errorf("acquire users from pool: %w", err)
	}

	h.gameCtx.AcquiredUsers = entries
	players := make([]*PlayerInfo, evt.NumPlayers)

	for i, entry := range entries {
		players[i] = &PlayerInfo{
			UserID: entry.Auth.UserID,
			Name:   fmt.Sprintf("bot-%d-%d", evt.GameIndex, i),
			Auth:   entry.Auth,
			REST:   h.newREST(h.baseURL, entry.Auth.AccessToken, h.gameCtx.Accumulator),
		}
	}

	observe.Info(context.Background(), "players acquired from pool",
		attribute.Int("gameIndex", evt.GameIndex),
		attribute.Int("numPlayers", evt.NumPlayers),
	)

	return players, nil
}

func (h *ProtocolHandler) signupFresh(evt GameStartedEvent) ([]*PlayerInfo, error) {
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
			return nil, fmt.Errorf("signup player %d: %w", i, err)
		}

		players[i] = &PlayerInfo{
			UserID: authResult.UserID,
			Name:   fmt.Sprintf("bot-%d-%d", evt.GameIndex, i),
			Auth:   authResult,
			REST:   h.newREST(h.baseURL, authResult.AccessToken, h.gameCtx.Accumulator),
		}
	}

	observe.Info(context.Background(), "players authenticated",
		attribute.Int("gameIndex", evt.GameIndex),
		attribute.Int("numPlayers", evt.NumPlayers),
	)

	return players, nil
}

// createGame creates a game via the first player and returns its ID.
func (h *ProtocolHandler) createGame(
	evt GameStartedEvent,
	players []*PlayerInfo,
) (int64, error) {
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
		return 0, fmt.Errorf("create game: %w", err)
	}

	observe.Info(context.Background(), "game created",
		attribute.Int("gameIndex", evt.GameIndex),
		attribute.Int64("game_id", gameID),
	)

	return gameID, nil
}

// connectWebSockets establishes WS connections for all players.
func (h *ProtocolHandler) connectWebSockets(
	evt GameStartedEvent,
	players []*PlayerInfo,
	gameID int64,
) error {
	for i, p := range players {
		ws, err := h.newWS(h.wsURL, gameID, p.Auth.AccessToken, h.gameCtx.Accumulator)
		if err != nil {
			return fmt.Errorf("ws connect player %d: %w", i, err)
		}

		p.WS = ws
	}

	observe.Info(context.Background(), "all players connected via WebSocket",
		attribute.Int("gameIndex", evt.GameIndex),
	)

	return nil
}

// populateSession fills the game context with session state.
func (h *ProtocolHandler) populateSession(gameID int64, players []*PlayerInfo) {
	time.Sleep(h.timeouts.InitialStateWait)

	userIndex := make(map[string]int)
	for i, p := range players {
		userIndex[p.UserID] = i
	}

	h.gameCtx.GameID = gameID
	h.gameCtx.Players = players
	h.gameCtx.UserIndex = userIndex
}

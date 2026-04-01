package runner

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
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
	newREST func(baseURL, token string, collector *metrics.Collector) RESTClient
	newWS   func(wsURL string, gameID int64, token string, collector *metrics.Collector) (WSClient, error)
}

// Register subscribes to EventGameStarted.
func (h *ProtocolHandler) Register(bus *Bus) {
	bus.On(EventGameStarted, h.handle)
}

func (h *ProtocolHandler) handle(bus *Bus, e Event) {
	evt := e.(GameStartedEvent)
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
			REST:   h.newREST(h.baseURL, authResult.AccessToken, gameCtx.Collector),
		}
	}

	log.Printf("[game %d] %d players authenticated", evt.GameIndex, evt.NumPlayers)

	// 2. Create game via first player.
	gamePlayers := make([]client.CreateGamePlayer, evt.NumPlayers)
	for i, p := range players {
		gamePlayers[i] = client.CreateGamePlayer{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	gameID, err := players[0].REST.CreateGame(client.CreateGameRequest{Players: gamePlayers})
	if err != nil {
		emitFatal(fmt.Errorf("create game: %w", err))

		return
	}

	log.Printf("[game %d] created game %d", evt.GameIndex, gameID)

	// 3. All players connect WebSocket.
	for i, p := range players {
		ws, err := h.newWS(h.wsURL, gameID, p.Auth.AccessToken, gameCtx.Collector)
		if err != nil {
			emitFatal(fmt.Errorf("ws connect player %d: %w", i, err))

			return
		}

		p.WS = ws
	}

	log.Printf("[game %d] all players connected via WebSocket", evt.GameIndex)

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

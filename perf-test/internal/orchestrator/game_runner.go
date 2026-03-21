package orchestrator

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/client"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/gamestate"
	"github.com/go-risk-it/go-risk-it/perf-test/internal/player"
)

// PlayerInfo holds all state for a single player in a game.
type PlayerInfo struct {
	UserID string
	Name   string
	Auth   *client.AuthResult
	REST   *client.REST
	WS     *client.WS
}

// GameRunner runs a single game to completion.
type GameRunner struct {
	baseURL  string
	wsURL    string
	anonKey  string
	strategy player.Strategy
	timeout  time.Duration
}

func NewGameRunner(
	baseURL, wsURL, anonKey string,
	strategy player.Strategy,
	timeout time.Duration,
) *GameRunner {
	return &GameRunner{
		baseURL:  baseURL,
		wsURL:    wsURL,
		anonKey:  anonKey,
		strategy: strategy,
		timeout:  timeout,
	}
}

// GameResult holds stats from a completed game.
type GameResult struct {
	Duration   time.Duration
	Moves      int
	Errors     int
	Winner     string
	TimedOut   bool
	FatalError error
}

// Run executes a single game with the given number of players.
func (gr *GameRunner) Run(gameIndex, numPlayers int) GameResult {
	start := time.Now()
	result := GameResult{}

	// 1. Create and authenticate players.
	auth := client.NewAuth(gr.baseURL, gr.anonKey)
	players := make([]*PlayerInfo, numPlayers)

	for i := 0; i < numPlayers; i++ {
		email := fmt.Sprintf("perf-g%dp%d-%d@test.local", gameIndex, i, time.Now().UnixNano())
		password := "perftest123"

		authResult, err := auth.Signup(email, password)
		if err != nil {
			result.FatalError = fmt.Errorf("signup player %d: %w", i, err)
			result.Duration = time.Since(start)

			return result
		}

		players[i] = &PlayerInfo{
			UserID: authResult.UserID,
			Name:   fmt.Sprintf("bot-%d-%d", gameIndex, i),
			Auth:   authResult,
			REST:   client.NewREST(gr.baseURL, authResult.AccessToken),
		}
	}

	log.Printf("[game %d] %d players authenticated", gameIndex, numPlayers)

	// 2. Create game via first player.
	gamePlayers := make([]client.CreateGamePlayer, numPlayers)
	for i, p := range players {
		gamePlayers[i] = client.CreateGamePlayer{
			UserID: p.UserID,
			Name:   p.Name,
		}
	}

	gameID, err := players[0].REST.CreateGame(client.CreateGameRequest{Players: gamePlayers})
	if err != nil {
		result.FatalError = fmt.Errorf("create game: %w", err)
		result.Duration = time.Since(start)

		return result
	}

	log.Printf("[game %d] created game %d", gameIndex, gameID)

	// 3. All players connect WebSocket.
	for i, p := range players {
		ws, err := client.ConnectWS(gr.wsURL, gameID, p.Auth.AccessToken)
		if err != nil {
			result.FatalError = fmt.Errorf("ws connect player %d: %w", i, err)
			result.Duration = time.Since(start)

			return result
		}

		p.WS = ws
		defer ws.Close()
	}

	log.Printf("[game %d] all players connected via WebSocket", gameIndex)

	// 4. Wait for initial state to arrive on all connections.
	time.Sleep(1 * time.Second)

	// Build a userID→player index for turn lookup.
	userIndex := make(map[string]int)
	for i, p := range players {
		userIndex[p.UserID] = i
	}

	// 5. Event-driven game loop.
	// Use player 0's view as the canonical state source. After each move,
	// wait for player 0's view to update before deciding the next action.
	deadline := time.After(gr.timeout)
	consecutiveErrors := 0

	for {
		select {
		case <-deadline:
			result.TimedOut = true
			result.Duration = time.Since(start)

			log.Printf(
				"[game %d] timed out after %v (%d moves, %d errors)",
				gameIndex,
				result.Duration,
				result.Moves,
				result.Errors,
			)

			return result
		default:
		}

		// Use the active player's own view — they have the most accurate card state.
		// First, figure out whose turn it is from player 0's view.
		refSnap := players[0].WS.View().Snapshot()

		if refSnap.IsGameOver() {
			result.Winner = refSnap.GameState.WinnerUserID
			result.Duration = time.Since(start)

			log.Printf("[game %d] finished in %v (%d moves, %d errors, winner: %s)",
				gameIndex, result.Duration, result.Moves, result.Errors, abbreviate(result.Winner))

			return result
		}

		if refSnap.GameState == nil || refSnap.PlayersState == nil {
			// State not yet received — wait.
			waitForAnyUpdate(players, 2*time.Second)

			continue
		}

		// Find the current player by turn index.
		activePlayer := findActivePlayer(refSnap, players, userIndex)
		if activePlayer == nil {
			// Turn doesn't map to any player (shouldn't happen) — wait.
			log.Printf("[game %d] no active player for turn %d", gameIndex, refSnap.GameState.Turn)
			waitForAnyUpdate(players, 2*time.Second)

			continue
		}

		// Use the active player's own snapshot (has their card state).
		snap := activePlayer.WS.View().Snapshot()

		// Decide and execute.
		action, err := gr.strategy.DecideMove(snap, activePlayer.UserID)
		if err != nil {
			log.Printf("[game %d] strategy error for %s (phase=%s): %v",
				gameIndex, activePlayer.Name, snap.CurrentPhase(), err)
			result.Errors++
			consecutiveErrors++

			if consecutiveErrors > 20 {
				result.FatalError = fmt.Errorf("too many consecutive errors")
				result.Duration = time.Since(start)

				return result
			}

			waitForAnyUpdate(players, 1*time.Second)

			continue
		}

		logAction(gameIndex, activePlayer.Name, action)

		if err := executeAction(activePlayer.REST, gameID, action); err != nil {
			log.Printf("[game %d] execute error for %s: %v", gameIndex, activePlayer.Name, err)
			result.Errors++
			consecutiveErrors++

			if consecutiveErrors > 20 {
				result.FatalError = fmt.Errorf("too many consecutive errors")
				result.Duration = time.Since(start)

				return result
			}

			// If card play failed, advance past cards phase instead of retrying.
			if action.Type == player.ActionPlayCards {
				log.Printf("[game %d] card play failed, advancing past cards phase", gameIndex)

				if advErr := activePlayer.REST.Advance(gameID, string(gamestate.Cards)); advErr != nil {
					log.Printf("[game %d] advance past cards also failed: %v", gameIndex, advErr)
				} else {
					result.Moves++
				}

				waitForAnyUpdate(players, 3*time.Second)
				time.Sleep(50 * time.Millisecond)

				continue
			}

			// On 409 (stale state), wait for a WS update and retry.
			waitForAnyUpdate(players, 2*time.Second)

			continue
		}

		result.Moves++
		consecutiveErrors = 0

		// Wait for state to propagate. After a move, multiple WS messages arrive
		// (boardState, gameState, playerState, etc.). Wait for a burst to settle.
		waitForAnyUpdate(players, 3*time.Second)
		// Small extra delay to let remaining burst messages arrive.
		time.Sleep(50 * time.Millisecond)
	}
}

// findActivePlayer finds the player whose turn it is.
// Turn is a monotonically increasing counter; current player = Turn % numPlayers.
func findActivePlayer(
	snap gamestate.ViewSnapshot,
	players []*PlayerInfo,
	userIndex map[string]int,
) *PlayerInfo {
	if snap.GameState == nil || snap.PlayersState == nil {
		return nil
	}

	numPlayers := int64(len(snap.PlayersState.Players))
	if numPlayers == 0 {
		return nil
	}

	currentIndex := snap.GameState.Turn % numPlayers
	for _, p := range snap.PlayersState.Players {
		if p.Index == currentIndex {
			if idx, ok := userIndex[p.UserID]; ok {
				return players[idx]
			}
		}
	}

	return nil
}

// waitForAnyUpdate waits for a state update from any player's WS connection,
// or returns if all connections are closed.
func waitForAnyUpdate(players []*PlayerInfo, timeout time.Duration) {
	timer := time.After(timeout)

	signal := make(chan struct{}, 1)
	for _, p := range players {
		go func(updated <-chan struct{}, wsDone <-chan struct{}) {
			select {
			case <-updated:
				select {
				case signal <- struct{}{}:
				default:
				}
			case <-wsDone:
				select {
				case signal <- struct{}{}:
				default:
				}
			case <-timer:
			}
		}(p.WS.View().Updated(), p.WS.Done())
	}

	select {
	case <-signal:
	case <-timer:
	}
}

func logAction(gameIndex int, playerName string, action *player.Action) {
	switch action.Type {
	case player.ActionDeploy:
		log.Printf(
			"[game %d] %s: deploy %d→%d on %s",
			gameIndex,
			playerName,
			action.Deploy.CurrentTroops,
			action.Deploy.DesiredTroops,
			action.Deploy.RegionID,
		)
	case player.ActionAttack:
		log.Printf(
			"[game %d] %s: attack %s→%s (%d troops)",
			gameIndex,
			playerName,
			action.Attack.SourceRegionID,
			action.Attack.TargetRegionID,
			action.Attack.AttackingTroops,
		)
	case player.ActionConquer:
		log.Printf("[game %d] %s: conquer (move %d troops)",
			gameIndex, playerName, action.Conquer.Troops)
	case player.ActionReinforce:
		log.Printf(
			"[game %d] %s: reinforce %s→%s (%d troops)",
			gameIndex,
			playerName,
			action.Reinforce.SourceRegionID,
			action.Reinforce.TargetRegionID,
			action.Reinforce.MovingTroops,
		)
	case player.ActionPlayCards:
		log.Printf("[game %d] %s: play cards", gameIndex, playerName)
	case player.ActionAdvance:
		log.Printf(
			"[game %d] %s: advance from %s",
			gameIndex,
			playerName,
			action.Advance.CurrentPhase,
		)
	}
}

func executeAction(rest *client.REST, gameID int64, action *player.Action) error {
	switch action.Type {
	case player.ActionDeploy:
		a := action.Deploy

		return rest.Deploy(gameID, client.DeployMove{
			RegionID:      a.RegionID,
			CurrentTroops: a.CurrentTroops,
			DesiredTroops: a.DesiredTroops,
		})
	case player.ActionAttack:
		a := action.Attack

		return rest.Attack(gameID, client.AttackMove{
			SourceRegionID:  a.SourceRegionID,
			TargetRegionID:  a.TargetRegionID,
			TroopsInSource:  a.TroopsInSource,
			TroopsInTarget:  a.TroopsInTarget,
			AttackingTroops: a.AttackingTroops,
		})
	case player.ActionConquer:
		return rest.Conquer(gameID, client.ConquerMove{
			Troops: action.Conquer.Troops,
		})
	case player.ActionReinforce:
		a := action.Reinforce

		return rest.Reinforce(gameID, client.ReinforceMove{
			SourceRegionID: a.SourceRegionID,
			TargetRegionID: a.TargetRegionID,
			TroopsInSource: a.TroopsInSource,
			TroopsInTarget: a.TroopsInTarget,
			MovingTroops:   a.MovingTroops,
		})
	case player.ActionPlayCards:
		combos := make([]client.CardCombination, len(action.Cards.Combinations))
		for i, ids := range action.Cards.Combinations {
			combos[i] = client.CardCombination{CardIDs: ids}
		}

		return rest.PlayCards(gameID, client.CardsMove{Combinations: combos})
	case player.ActionAdvance:
		return rest.Advance(gameID, action.Advance.CurrentPhase)
	default:
		return fmt.Errorf("unknown action type: %d", action.Type)
	}
}

func abbreviate(s string) string {
	if len(s) > 8 {
		return s[:8]
	}

	return strings.TrimSpace(s)
}

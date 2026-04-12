package runner

import (
	"context"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/userpool"
)

// GameResult holds stats from a completed game.
type GameResult struct {
	GameIndex  int
	Duration   time.Duration
	Moves      int
	Errors     int
	Winner     string
	TimedOut   bool
	Cancelled  bool
	FatalError error
}

// Timeouts holds all configurable timing parameters for the game loop.
type Timeouts struct {
	InitialStateWait  time.Duration
	UpdateWait        time.Duration
	PhaseChangeWait   time.Duration
	MaxConsecutiveErr int
}

// DefaultTimeouts returns sensible defaults for all timing parameters.
func DefaultTimeouts() Timeouts {
	return Timeouts{
		InitialStateWait:  1 * time.Second,
		UpdateWait:        3 * time.Second,
		PhaseChangeWait:   3 * time.Second,
		MaxConsecutiveErr: 20,
	}
}

// PlayerInfo holds all state for a single player in a game.
type PlayerInfo struct {
	UserID string
	Name   string
	Auth   *client.AuthResult
	REST   RESTClient
	WS     WSClient
}

// GameSession holds shared mutable state for a single game.
type GameSession struct {
	Ctx           context.Context //nolint:containedctx // game session carries context by design
	GameIndex     int
	GameID        int64
	Players       []*PlayerInfo
	UserIndex     map[string]int
	StartTime     time.Time
	Accumulator   *metrics.StepAccumulator
	AcquiredUsers []*userpool.Entry // set by protocol handler, released by runner
}

// AuthClient abstracts client.Auth for testability.
type AuthClient interface {
	Signup(email, password string) (*client.AuthResult, error)
}

// RESTClient abstracts client.REST for testability.
type RESTClient interface {
	CreateGame(ctx context.Context, req client.CreateGameRequest) (int64, error)
	Deploy(ctx context.Context, gameID int64, move client.DeployMove) error
	Attack(ctx context.Context, gameID int64, move client.AttackMove) error
	Conquer(ctx context.Context, gameID int64, move client.ConquerMove) error
	Reinforce(ctx context.Context, gameID int64, move client.ReinforceMove) error
	PlayCards(ctx context.Context, gameID int64, move client.CardsMove) error
	Advance(ctx context.Context, gameID int64, currentPhase string) error
}

// WSClient abstracts client.WS for testability.
type WSClient interface {
	View() *gamestate.View
	Done() <-chan struct{}
	Close() error
	Disrupt()
}

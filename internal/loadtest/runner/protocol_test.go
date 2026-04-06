package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAuth implements AuthClient for testing.
type fakeAuth struct {
	results []*client.AuthResult
	err     error
	failAt  int // -1 = never fail
	calls   []string
}

func newFakeAuth(n int) *fakeAuth { //nolint:unparam // interface conformance / future use
	results := make([]*client.AuthResult, n)
	for i := range n {
		results[i] = &client.AuthResult{
			UserID:      fmt.Sprintf("user-%d", i),
			AccessToken: fmt.Sprintf("token-%d", i),
		}
	}

	return &fakeAuth{results: results, failAt: -1}
}

func (f *fakeAuth) Signup(email, password string) (*client.AuthResult, error) {
	f.calls = append(f.calls, email)

	idx := len(f.calls) - 1
	if f.failAt >= 0 && idx >= f.failAt {
		return nil, f.err
	}

	if idx < len(f.results) {
		return f.results[idx], nil
	}

	return nil, fmt.Errorf("no result for index %d", idx)
}

// fakeRESTForProtocol implements RESTClient for protocol tests.
type fakeRESTForProtocol struct {
	gameID    int64
	createErr error
}

func (f *fakeRESTForProtocol) CreateGame(
	_ context.Context,
	_ client.CreateGameRequest,
) (int64, error) {
	if f.createErr != nil {
		return 0, f.createErr
	}

	return f.gameID, nil
}

func (f *fakeRESTForProtocol) Deploy(
	context.Context,
	int64,
	client.DeployMove,
) error {
	return nil
}

func (f *fakeRESTForProtocol) Attack(
	context.Context,
	int64,
	client.AttackMove,
) error {
	return nil
}

func (f *fakeRESTForProtocol) Conquer(
	context.Context,
	int64,
	client.ConquerMove,
) error {
	return nil
}

func (f *fakeRESTForProtocol) Reinforce(context.Context, int64, client.ReinforceMove) error {
	return nil
}

func (f *fakeRESTForProtocol) PlayCards(
	context.Context,
	int64,
	client.CardsMove,
) error {
	return nil
}

func (f *fakeRESTForProtocol) Advance(
	context.Context,
	int64,
	string,
) error {
	return nil
}

// fakeWSForProtocol implements WSClient for protocol tests.
type fakeWSForProtocol struct {
	view   *gamestate.View
	done   chan struct{}
	closed bool
}

func newFakeWS() *fakeWSForProtocol {
	v := gamestate.NewView()

	return &fakeWSForProtocol{view: v, done: make(chan struct{})}
}

// newFakeWSWithState creates a fakeWS pre-populated with game state.
func newFakeWSWithState(snap gamestate.ViewSnapshot) *fakeWSForProtocol {
	v := gamestate.NewView()

	// Apply state by marshaling the PlayerView and applying as a playerView message.
	if snap.PlayerView != nil {
		data, _ := json.Marshal(snap.PlayerView) //nolint:errchkjson // known safe type
		_ = v.Apply(gamestate.WSMessage{Type: "playerView", Payload: data})
	}

	return &fakeWSForProtocol{view: v, done: make(chan struct{})}
}

func (f *fakeWSForProtocol) View() *gamestate.View { return f.view }
func (f *fakeWSForProtocol) Done() <-chan struct{} { return f.done }
func (f *fakeWSForProtocol) Close() error {
	f.closed = true

	return nil
}
func (f *fakeWSForProtocol) Disrupt() {}

func makeProtocolHandler(
	auth AuthClient,
	rest *fakeRESTForProtocol,
	wsErr error,
) (*ProtocolHandler, *GameSession) {
	gameCtx := &GameSession{
		GameIndex:   1,
		StartTime:   time.Now(),
		Accumulator: metrics.NewStepAccumulator(0),
	}

	h := &ProtocolHandler{
		baseURL:  "http://localhost",
		wsURL:    "ws://localhost",
		anonKey:  "test-key",
		timeouts: DefaultTimeouts(),
		gameCtx:  gameCtx,
		newAuth: func(_, _ string) AuthClient {
			return auth
		},
		newREST: func(_ string, token string, _ *metrics.StepAccumulator) RESTClient {
			return rest
		},
		newWS: func(_ string, _ int64, _ string, _ *metrics.StepAccumulator) (WSClient, error) {
			if wsErr != nil {
				return nil, wsErr
			}

			return newFakeWS(), nil
		},
	}

	return h, gameCtx
}

func TestProtocol_HappyPath_PopulatesSession(t *testing.T) {
	t.Parallel()

	auth := newFakeAuth(4)
	rest := &fakeRESTForProtocol{gameID: 42}

	h, _ := makeProtocolHandler(auth, rest, nil)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 1, NumPlayers: 4})

	// Protocol handler no longer emits StateReceived — the barrier loop does.
	stateEvents := bus.EmittedOfType(EventStateReceived)
	assert.Empty(t, stateEvents)

	// Session should be populated.
	assert.Equal(t, int64(42), h.gameCtx.GameID)
	assert.Len(t, h.gameCtx.Players, 4)

	completeEvents := bus.EmittedOfType(EventGameComplete)
	assert.Empty(t, completeEvents)
}

func TestProtocol_SignupFails_EmitsGameComplete(t *testing.T) {
	t.Parallel()

	auth := newFakeAuth(4)
	auth.failAt = 1
	auth.err = errors.New("signup failed")

	rest := &fakeRESTForProtocol{gameID: 42}
	h, _ := makeProtocolHandler(auth, rest, nil)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 1, NumPlayers: 4})

	completeEvents := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completeEvents, 1)

	//nolint:forcetypeassert // test assertion
	result := completeEvents[0].(GameCompleteEvent).Result
	require.Error(t, result.FatalError)
	assert.Contains(t, result.FatalError.Error(), "signup")

	stateEvents := bus.EmittedOfType(EventStateReceived)
	assert.Empty(t, stateEvents)
}

func TestProtocol_CreateGameFails_EmitsGameComplete(t *testing.T) {
	t.Parallel()

	auth := newFakeAuth(4)
	rest := &fakeRESTForProtocol{createErr: errors.New("create failed")}

	h, _ := makeProtocolHandler(auth, rest, nil)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 1, NumPlayers: 4})

	completeEvents := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completeEvents, 1)
	//nolint:forcetypeassert // test assertion
	require.Error(t, completeEvents[0].(GameCompleteEvent).Result.FatalError)
}

func TestProtocol_WSConnectFails_EmitsGameComplete(t *testing.T) {
	t.Parallel()

	auth := newFakeAuth(4)
	rest := &fakeRESTForProtocol{gameID: 42}

	h, _ := makeProtocolHandler(auth, rest, errors.New("ws dial failed"))
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 1, NumPlayers: 4})

	completeEvents := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completeEvents, 1)
	//nolint:forcetypeassert // test assertion
	require.Error(t, completeEvents[0].(GameCompleteEvent).Result.FatalError)
}

func TestProtocol_PopulatesGameSession(t *testing.T) {
	t.Parallel()

	auth := newFakeAuth(4)
	rest := &fakeRESTForProtocol{gameID: 42}

	h, gameCtx := makeProtocolHandler(auth, rest, nil)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 1, NumPlayers: 4})

	assert.Equal(t, int64(42), gameCtx.GameID)
	assert.Len(t, gameCtx.Players, 4)
	assert.Len(t, gameCtx.UserIndex, 4)

	for i, p := range gameCtx.Players {
		assert.Equal(t, fmt.Sprintf("user-%d", i), p.UserID)
		assert.NotNil(t, p.REST)
		assert.NotNil(t, p.WS)
	}
}

func TestProtocol_PlayerNaming(t *testing.T) {
	t.Parallel()

	auth := newFakeAuth(4)
	rest := &fakeRESTForProtocol{gameID: 42}

	h, _ := makeProtocolHandler(auth, rest, nil)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(GameStartedEvent{GameIndex: 7, NumPlayers: 4})

	// Verify signup emails follow the naming pattern.
	require.Len(t, auth.calls, 4)

	for i, email := range auth.calls {
		assert.Contains(t, email, fmt.Sprintf("perf-g7p%d-", i))
	}
}

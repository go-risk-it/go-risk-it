package runner

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/gamestate"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRESTForError tracks Advance calls.
type fakeRESTForError struct {
	advanceCalls int
	advanceErr   error
}

func (f *fakeRESTForError) CreateGame(context.Context, client.CreateGameRequest) (int64, error) {
	return 0, nil
}
func (f *fakeRESTForError) Deploy(context.Context, int64, client.DeployMove) error   { return nil }
func (f *fakeRESTForError) Attack(context.Context, int64, client.AttackMove) error   { return nil }
func (f *fakeRESTForError) Conquer(context.Context, int64, client.ConquerMove) error { return nil }
func (f *fakeRESTForError) Reinforce(context.Context, int64, client.ReinforceMove) error {
	return nil
}
func (f *fakeRESTForError) PlayCards(context.Context, int64, client.CardsMove) error { return nil }

func (f *fakeRESTForError) Advance(_ context.Context, _ int64, _ string) error {
	f.advanceCalls++

	return f.advanceErr
}

func makeErrorHandler(
	ctx context.Context,
	rest *fakeRESTForError,
) (*ErrorHandler, *GameResult) {
	snap := mkSnap(0, gamestate.Deploy, "")
	ws := newFakeWSWithState(snap)

	gameCtx := &GameSession{
		Ctx:       ctx,
		GameIndex: 1,
		GameID:    42,
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", REST: rest, WS: ws},
		},
		UserIndex: map[string]int{"u0": 0},
	}

	result := &GameResult{GameIndex: 1}

	h := &ErrorHandler{
		gameCtx: gameCtx,
		timeouts: Timeouts{
			UpdateWait:        10 * time.Millisecond,
			PostMoveSettle:    1 * time.Millisecond,
			MaxConsecutiveErr: 20,
		},
		result:          result,
		maxStaleRetries: 5,
		maxAdvanceFails: 3,
	}

	return h, result
}

func signalWSUpdate(p *PlayerInfo) {
	data, _ := json.Marshal(
		&gamestate.GameState{Turn: 1, Phase: gamestate.Phase{Type: gamestate.Deploy}},
	)
	_ = p.WS.View().Apply(gamestate.WSMessage{Type: "gameState", Payload: data})
}

func TestError_StaleState_RetriesUnderThreshold(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(5 * time.Millisecond)
		signalWSUpdate(h.gameCtx.Players[0])
	}()

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     &client.StaleStateError{Message: "stale"},
		ErrType: "stale_state",
	})

	assert.Equal(t, 1, h.consecutiveStaleErrors)

	stateEvents := bus.EmittedOfType(EventStateReceived)
	require.Len(t, stateEvents, 1)

	completeEvents := bus.EmittedOfType(EventGameComplete)
	assert.Empty(t, completeEvents)
}

func TestError_StaleState_ExhaustedRetries_Advances(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	h.consecutiveStaleErrors = 4 // Next will be 5th (>= maxStaleRetries)
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(5 * time.Millisecond)
		signalWSUpdate(h.gameCtx.Players[0])
	}()

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     &client.StaleStateError{Message: "stale"},
		ErrType: "stale_state",
	})

	assert.Equal(t, 1, rest.advanceCalls)
}

func TestError_StaleState_AdvanceSucceeds_ResetsCounters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, result := makeErrorHandler(ctx, rest)
	h.consecutiveStaleErrors = 4
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(5 * time.Millisecond)
		signalWSUpdate(h.gameCtx.Players[0])
	}()

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     &client.StaleStateError{Message: "stale"},
		ErrType: "stale_state",
	})

	assert.Equal(t, 0, h.consecutiveStaleErrors)
	assert.Equal(t, 0, h.consecutiveAdvanceFails)
	assert.Equal(t, 1, result.Moves)
}

func TestError_StaleState_AdvanceFails3x_Fatal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{advanceErr: errors.New("advance failed")}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	// Each iteration: staleErrors=4, emit stale → triggers advance → fails.
	for i := range 3 {
		h.consecutiveStaleErrors = 4

		go func() {
			time.Sleep(5 * time.Millisecond)
			signalWSUpdate(h.gameCtx.Players[0])
		}()

		bus.Emit(MoveFailedEvent{
			Action:  &player.Action{Type: player.ActionDeploy},
			Err:     &client.StaleStateError{Message: "stale"},
			ErrType: "stale_state",
		})

		if i < 2 {
			// Not yet fatal, should have emitted StateReceived.
			assert.Equal(t, i+1, h.consecutiveAdvanceFails)
		}
	}

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	assert.Error(t, completes[0].(GameCompleteEvent).Result.FatalError)
}

func TestError_Transient_CountsAndRetries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(5 * time.Millisecond)
		signalWSUpdate(h.gameCtx.Players[0])
	}()

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("503"),
		ErrType: "transient",
	})

	assert.Equal(t, 1, h.consecutiveErrors)
	assert.Len(t, bus.EmittedOfType(EventStateReceived), 1)
}

func TestError_Execution_PlayCards_TriesAdvance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(5 * time.Millisecond)
		signalWSUpdate(h.gameCtx.Players[0])
	}()

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionPlayCards, Cards: &player.CardsAction{}},
		Err:     errors.New("card play failed"),
		ErrType: "execution",
	})

	assert.Equal(t, 1, rest.advanceCalls)
}

func TestError_ConsecutiveExceedMax_Fatal(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	h.consecutiveErrors = 20
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("error"),
		ErrType: "execution",
	})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	assert.Error(t, completes[0].(GameCompleteEvent).Result.FatalError)
}

func TestError_MoveSucceeded_ResetsCounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	h.consecutiveErrors = 5
	h.consecutiveStaleErrors = 3
	h.consecutiveAdvanceFails = 2
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTLatency: time.Millisecond,
		RESTEndTime: time.Now(),
	})

	assert.Equal(t, 0, h.consecutiveErrors)
	assert.Equal(t, 0, h.consecutiveStaleErrors)
	assert.Equal(t, 0, h.consecutiveAdvanceFails)
}

func TestError_Strategy_CountsAndRetries(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	go func() {
		time.Sleep(5 * time.Millisecond)
		signalWSUpdate(h.gameCtx.Players[0])
	}()

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("strategy error"),
		ErrType: "strategy",
	})

	assert.Equal(t, 1, h.consecutiveErrors)
	assert.Len(t, bus.EmittedOfType(EventStateReceived), 1)
}

func TestError_FatalFlag_SkipsRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("fatal error"),
		ErrType: "execution",
		Fatal:   true,
	})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	assert.Error(t, completes[0].(GameCompleteEvent).Result.FatalError)
}

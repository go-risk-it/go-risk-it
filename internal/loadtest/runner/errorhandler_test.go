package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/client"
	"github.com/go-risk-it/go-risk-it/internal/loadtest/metrics"
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
	result := &GameResult{}
	gameCtx := &GameSession{
		Ctx:         ctx,
		GameID:      42,
		Accumulator: metrics.NewStepAccumulator(0),
		Players: []*PlayerInfo{
			{UserID: "u0", Name: "p0", REST: rest},
		},
		UserIndex: map[string]int{"u0": 0},
	}

	return &ErrorHandler{
		gameCtx:           gameCtx,
		result:            result,
		maxConsecutiveErr: 20,
	}, result
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
		Fatal:   true,
		ErrType: "execution",
	})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	//nolint:forcetypeassert // test assertion
	assert.Error(t, completes[0].(GameCompleteEvent).Result.FatalError)
}

func TestError_MoveSucceeded_ResetsCounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	h.consecutiveErrors = 5
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTEndTime: time.Now(),
	})

	assert.Equal(t, 0, h.consecutiveErrors)
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
		Err:     errors.New("some error"),
		ErrType: "stale_state",
	})

	completes := bus.EmittedOfType(EventGameComplete)
	require.Len(t, completes, 1)
	//nolint:forcetypeassert // test assertion
	assert.Error(t, completes[0].(GameCompleteEvent).Result.FatalError)
}

func TestError_NonFatal_CountsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, result := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("some error"),
		ErrType: "stale_state",
	})

	assert.Equal(t, 1, h.consecutiveErrors)
	assert.Equal(t, 1, result.Errors)

	// No GameComplete emitted — just returns for next WS update.
	assert.Empty(t, bus.EmittedOfType(EventGameComplete))
}

func TestError_Transient_CountsAndReturns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionDeploy},
		Err:     errors.New("503"),
		ErrType: "transient",
	})

	assert.Equal(t, 1, h.consecutiveErrors)

	// Non-blocking: no StateReceived emitted, no GameComplete.
	assert.Empty(t, bus.EmittedOfType(EventStateReceived))
	assert.Empty(t, bus.EmittedOfType(EventGameComplete))
}

func TestError_CardPlayFailed_TriesAdvance(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	// The view snapshot has nil PlayerView (no phase set), so currentPhase
	// won't match "cards" — advance should NOT be called.
	bus.Emit(MoveFailedEvent{
		Action:  &player.Action{Type: player.ActionPlayCards, Cards: &player.CardsAction{}},
		Err:     errors.New("card play failed"),
		ErrType: "execution",
	})

	assert.Equal(t, 0, rest.advanceCalls, "should not advance when not in CARDS phase")
}

func TestError_SucceededResetsAfterErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rest := &fakeRESTForError{}
	h, _ := makeErrorHandler(ctx, rest)
	bus := NewTestBus()
	h.Register(bus)

	// Accumulate errors.
	for range 5 {
		bus.Emit(MoveFailedEvent{
			Action:  &player.Action{Type: player.ActionDeploy},
			Err:     errors.New("error"),
			ErrType: "stale_state",
		})
	}

	assert.Equal(t, 5, h.consecutiveErrors)

	// Success resets.
	bus.Emit(MoveSucceededEvent{
		Action:      &player.Action{Type: player.ActionDeploy},
		RESTEndTime: time.Now(),
	})

	assert.Equal(t, 0, h.consecutiveErrors)
}

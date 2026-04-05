package game //nolint:testpackage // Tests unexported fx adapters

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	gamectx "github.com/go-risk-it/go-risk-it/internal/game/ctx"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	wsmock "github.com/go-risk-it/go-risk-it/mocks/internal_/game/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestStatePublisherAdapter_PublishState(t *testing.T) {
	t.Parallel()

	const (
		gameID     = int64(42)
		playerUser = "player-1@test.com"
	)

	view := &snapshot.PlayerView{
		Game: snapshot.GameMeta{Turn: 3},
		Regions: []snapshot.RegionState{
			{ID: "alaska", Troops: 5, OwnerID: playerUser},
		},
		Players: []snapshot.PlayerState{
			{UserID: playerUser, Name: "Alice", Index: 0},
		},
	}

	writer := wsmock.NewWriter(t)

	adapter := &statePublisherAdapter{writer: writer}

	// Build the expected envelope for comparison.
	wantRaw, err := messaging.BuildMessage(messaging.PlayerViewType, view)
	require.NoError(t, err)

	writer.EXPECT().
		WriteMessage(mock.MatchedBy(func(ctx gamectx.GameContext) bool {
			return ctx.GameID() == gameID && ctx.UserID() == playerUser
		}), wantRaw).
		Return()

	ctx := gamectx.WithGameID(
		kernelctx.WithUserID(
			kernelctx.WithSpan(context.Background(), noop.Span{}),
			"original-user",
		),
		gameID,
	)

	err = adapter.PublishState(ctx, playerUser, view)
	require.NoError(t, err)
}

func TestStatePublisherAdapter_PublishState_EnvelopeShape(t *testing.T) {
	t.Parallel()

	const (
		gameID     = int64(7)
		playerUser = "bob@test.com"
	)

	view := &snapshot.PlayerView{
		Game: snapshot.GameMeta{Turn: 1},
		Phase: snapshot.Phase{
			Type:  snapshot.PhaseAttack,
			State: snapshot.EmptyPhaseState{},
		},
	}

	writer := wsmock.NewWriter(t)

	adapter := &statePublisherAdapter{writer: writer}

	var captured json.RawMessage

	writer.EXPECT().
		WriteMessage(mock.Anything, mock.Anything).
		Run(func(ctx gamectx.GameContext, message json.RawMessage) {
			captured = message
		}).
		Return()

	ctx := gamectx.WithGameID(
		kernelctx.WithUserID(
			kernelctx.WithSpan(context.Background(), noop.Span{}),
			"",
		),
		gameID,
	)

	err := adapter.PublishState(ctx, playerUser, view)
	require.NoError(t, err)

	// Verify envelope has the correct type without full deserialization of
	// nested discriminated types (Phase, Mission).
	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(captured, &envelope))
	assert.Equal(t, "playerView", envelope.Type)
	assert.NotEmpty(t, envelope.Data)
}

func TestStatePublisherAdapter_PublishState_NonGameContext(t *testing.T) {
	t.Parallel()

	writer := wsmock.NewWriter(t)
	adapter := &statePublisherAdapter{writer: writer}

	err := adapter.PublishState(context.Background(), "user", &snapshot.PlayerView{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected GameContext")
}

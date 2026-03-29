package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/events/logger"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/headlines"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/cards"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	riskslog "github.com/go-risk-it/go-risk-it/internal/kernel/slog"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/fx"
)

// nopLifecycle satisfies fx.Lifecycle by discarding hooks. Used to construct a
// real Bus without a full fx.App — the test manages shutdown explicitly via Close.
type nopLifecycle struct{}

func (nopLifecycle) Append(fx.Hook) {}

// fixedTime provides a deterministic timestamp for all test events.
var fixedTime = time.Date(2026, 3, 27, 14, 0, 0, 0, time.UTC)

// parsedLog holds the unmarshaled JSON output from slog.
type parsedLog struct {
	Level          string         `json:"level"`
	Msg            string         `json:"msg"`
	GameID         json.Number    `json:"gameId"`
	EventType      string         `json:"eventType"`
	EventTimestamp string         `json:"eventTimestamp"`
	Payload        map[string]any `json:"payload"`
}

// emitAndParse registers the logger handler, emits the event through a TestBus,
// and returns the parsed JSON log line.
func emitAndParse(t *testing.T, event eventbus.Event) parsedLog {
	t.Helper()

	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	bus := eventbus.NewTestBus()
	logger.Register(logger.Params{
		Bus:    bus,
		Logger: slog.New(handler),
	})

	bus.Emit(context.Background(), event)

	var parsed parsedLog
	err := json.Unmarshal(buf.Bytes(), &parsed)
	require.NoError(t, err, "failed to unmarshal log output: %s", buf.String())

	return parsed
}

func TestRegister_LogsAllEventTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		event             eventbus.Event
		expectedGameID    int64
		expectedEventType string
		expectedPayload   map[string]any
	}{
		{
			name: "MoveExecuted/deploy_no_result",
			event: gameevt.NewMoveExecuted(
				42, "player1", fixedTime,
				sqlc.GamePhaseTypeDEPLOY,
				sqlc.GameMoveLog{ID: 101},
				sqlc.GamePhaseTypeDEPLOY,
				false, 2,
				nil, nil,
			),
			expectedGameID:    42,
			expectedEventType: gameevt.TypeMoveExecuted,
			expectedPayload: map[string]any{
				"event_type":   gameevt.TypeMoveExecuted,
				"game_id":      float64(42),
				"user_id":      "player1",
				"action_type":  string(sqlc.GamePhaseTypeDEPLOY),
				"target_phase": string(sqlc.GamePhaseTypeDEPLOY),
				"game_over":    false,
				"turn":         float64(2),
				"move_log_id":  float64(101),
			},
		},
		{
			name: "MoveExecuted/attack_result",
			event: gameevt.NewMoveExecuted(
				42, "attacker", fixedTime,
				sqlc.GamePhaseTypeATTACK,
				sqlc.GameMoveLog{ID: 99},
				sqlc.GamePhaseTypeATTACK,
				false, 5,
				&attack.MoveResult{
					AttackingRegionID: "brazil",
					DefendingRegionID: "argentina",
					ConqueringTroops:  3,
				},
				nil,
			),
			expectedGameID:    42,
			expectedEventType: gameevt.TypeMoveExecuted,
			expectedPayload: map[string]any{
				"attacking_region_id": "brazil",
				"defending_region_id": "argentina",
				"conquering_troops":   float64(3),
			},
		},
		{
			name: "MoveExecuted/cards_result",
			event: gameevt.NewMoveExecuted(
				42, "player1", fixedTime,
				sqlc.GamePhaseTypeCARDS,
				sqlc.GameMoveLog{ID: 100},
				sqlc.GamePhaseTypeDEPLOY,
				false, 3,
				nil,
				&cards.MoveResult{
					ExtraDeployableTroops: 6,
					RegionTroopGrants: []cards.RegionTroopGrant{
						{RegionID: 1, RegionExternalReference: "brazil"},
						{RegionID: 2, RegionExternalReference: "argentina"},
					},
				},
			),
			expectedGameID:    42,
			expectedEventType: gameevt.TypeMoveExecuted,
			expectedPayload: map[string]any{
				"extra_deployable_troops": float64(6),
				"region_troop_grants":     float64(2),
			},
		},
		{
			name: "PhaseTransitioned",
			event: gameevt.NewPhaseTransitioned(
				42, "player1", fixedTime,
				sqlc.GamePhaseTypeDEPLOY,
				sqlc.GamePhaseTypeATTACK,
				5,
			),
			expectedGameID:    42,
			expectedEventType: gameevt.TypePhaseTransitioned,
			expectedPayload: map[string]any{
				"event_type": gameevt.TypePhaseTransitioned,
				"game_id":    float64(42),
				"user_id":    "player1",
				"from_phase": string(sqlc.GamePhaseTypeDEPLOY),
				"to_phase":   string(sqlc.GamePhaseTypeATTACK),
				"turn":       float64(5),
			},
		},
		{
			name:              "GameCompleted",
			event:             gameevt.NewGameCompleted(42, "winner", fixedTime, 10),
			expectedGameID:    42,
			expectedEventType: gameevt.TypeGameCompleted,
			expectedPayload: map[string]any{
				"event_type":     gameevt.TypeGameCompleted,
				"game_id":        float64(42),
				"winner_user_id": "winner",
				"turn":           float64(10),
			},
		},
		{
			name:              "GameCreated",
			event:             gameevt.NewGameCreated(42, fixedTime, 4),
			expectedGameID:    42,
			expectedEventType: gameevt.TypeGameCreated,
			expectedPayload: map[string]any{
				"event_type":  gameevt.TypeGameCreated,
				"game_id":     float64(42),
				"num_players": float64(4),
			},
		},
		{
			name:              "PlayerConnected",
			event:             gameevt.NewPlayerConnected(42, "player1", fixedTime),
			expectedGameID:    42,
			expectedEventType: gameevt.TypePlayerConnected,
			expectedPayload: map[string]any{
				"event_type": gameevt.TypePlayerConnected,
				"game_id":    float64(42),
				"user_id":    "player1",
			},
		},
		{
			name: "PlayerEliminated",
			event: headlines.NewPlayerEliminated(
				42,
				"victim",
				"attacker",
				fixedTime,
				7,
			),
			expectedGameID:    42,
			expectedEventType: headlines.TypePlayerEliminated,
			expectedPayload: map[string]any{
				"event_type":         headlines.TypePlayerEliminated,
				"game_id":            float64(42),
				"eliminated_user_id": "victim",
				"eliminator_user_id": "attacker",
				"turn":               float64(7),
			},
		},
		{
			name: "ContinentCaptured",
			event: headlines.NewContinentCaptured(
				42,
				"player1",
				fixedTime,
				"europe",
				3,
			),
			expectedGameID:    42,
			expectedEventType: headlines.TypeContinentCaptured,
			expectedPayload: map[string]any{
				"event_type":   headlines.TypeContinentCaptured,
				"game_id":      float64(42),
				"user_id":      "player1",
				"continent_id": "europe",
				"turn":         float64(3),
			},
		},
		{
			name:              "ContinentLost",
			event:             headlines.NewContinentLost(42, "player1", fixedTime, "asia", 5),
			expectedGameID:    42,
			expectedEventType: headlines.TypeContinentLost,
			expectedPayload: map[string]any{
				"event_type":   headlines.TypeContinentLost,
				"game_id":      float64(42),
				"user_id":      "player1",
				"continent_id": "asia",
				"turn":         float64(5),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parsed := emitAndParse(t, testCase.event)

			// Verify log level is INFO
			require.Equal(t, "INFO", parsed.Level)

			// Verify consistent message
			require.Equal(t, "game_event", parsed.Msg)

			// Verify top-level attributes
			gameID, err := parsed.GameID.Int64()
			require.NoError(t, err)
			require.Equal(t, testCase.expectedGameID, gameID)

			require.Equal(t, testCase.expectedEventType, parsed.EventType)
			require.Equal(t, fixedTime.Format(time.RFC3339), parsed.EventTimestamp)

			// Verify payload is a nested object (not nil)
			require.NotNil(t, parsed.Payload, "payload group must be present")

			// Verify expected payload keys are present with correct values
			for key, expectedVal := range testCase.expectedPayload {
				actualVal, ok := parsed.Payload[key]
				require.True(t, ok, "payload missing key: %s", key)
				require.Equal(t, expectedVal, actualVal, "payload key %s mismatch", key)
			}
		})
	}
}

//nolint:paralleltest // swaps global TracerProvider
func TestRegister_LogsTraceIDFromLinkedSpan(
	t *testing.T,
) {
	// Setup: real TracerProvider + InMemoryExporter so spans are recorded.
	exporter := tracetest.NewInMemoryExporter()
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() { _ = tracerProvider.Shutdown(context.Background()) })

	// Swap global TracerProvider — the bus's startLinkedSpan uses otel.GetTracerProvider().
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tracerProvider)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	// Create a parent span simulating the HTTP handler trace (Trace 1).
	tracer := tracerProvider.Tracer("test")
	parentCtx, parentSpan := tracer.Start(context.Background(), "http-handler")
	parentTraceID := parentSpan.SpanContext().TraceID()
	parentSpan.End()

	// Build a GameContext carrying the parent span so DetachOnto copies domain metadata.
	traceCtx := ctx.WithSpan(parentCtx, parentSpan)
	userCtx := ctx.WithUserID(traceCtx, "player1")
	gameCtx := ctx.WithGameID(userCtx, 42)

	// Create a ContextHandler-wrapped JSONHandler writing to a buffer.
	// ContextHandler extracts domain fields (userID, gameID); trace context
	// (traceID/spanID) is handled by the otelslog bridge in production.
	var buf bytes.Buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	contextHandler := riskslog.NewContextHandler(jsonHandler, slog.LevelInfo)

	// Use the real async bus which runs detachContext + startLinkedSpan.
	bus := eventbus.NewBus(nopLifecycle{}, nil)

	logger.Register(logger.Params{
		Bus:    bus,
		Logger: slog.New(contextHandler),
	})

	// Register a second OnAll handler to signal when handlers complete.
	// The signal fires after the logger handler (OnAll dispatch is sequential).
	done := make(chan struct{}, 1)
	bus.OnAll(func(_ context.Context, _ eventbus.Event) {
		done <- struct{}{}
	})

	// Emit a game event with the GameContext as parent.
	evt := gameevt.NewGameCreated(42, fixedTime, 4)
	bus.Emit(gameCtx, evt)

	// Wait for handlers to complete.
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not complete within timeout")
	}

	// Close the bus to ensure the dispatch goroutine finishes (deferred cancel ends
	// the linked span, syncing it to the in-memory exporter via WithSyncer).
	require.NoError(t, bus.Close(context.Background()))

	// Parse the JSON log line.
	var result map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result),
		"failed to unmarshal log output: %s", buf.String())

	// Assert: domain fields are present in log output (extracted by ContextHandler).
	require.Equal(t, "player1", result["userID"], "userID must be present from detached context")
	require.InDelta(
		t,
		float64(42),
		result["gameID"],
		0,
		"gameID must be present from detached context",
	)

	// Note: traceID/spanID are NOT in log output because ContextHandler no longer
	// extracts them — the otelslog bridge handles that in production. This test
	// uses a plain JSONHandler as inner, so trace fields are absent.
	require.NotContains(t, result, "traceID")
	require.NotContains(t, result, "spanID")

	// Assert: the linked span was created and links back to the parent trace.
	stubs := exporter.GetSpans()

	var linkedStub *tracetest.SpanStub
	for i := range stubs {
		if stubs[i].Name == "bus:game_created" {
			linkedStub = &stubs[i]

			break
		}
	}

	require.NotNil(t, linkedStub, "bus:game_created span must be in recorded spans")

	// Verify the linked span has a link back to the parent trace.
	require.Len(t, linkedStub.Links, 1, "linked span must have exactly 1 link")
	require.Equal(t, parentTraceID, linkedStub.Links[0].SpanContext.TraceID(),
		"linked span's link must reference the parent HTTP trace")
}

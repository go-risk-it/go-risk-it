package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/data/game/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/events"
	gameevt "github.com/go-risk-it/go-risk-it/internal/events/game"
	"github.com/go-risk-it/go-risk-it/internal/events/logger"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/headlines"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/attack"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/cards"
	"github.com/stretchr/testify/require"
)

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
func emitAndParse(t *testing.T, event events.Event) parsedLog {
	t.Helper()

	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	bus := events.NewTestBus()
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
		event             events.Event
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

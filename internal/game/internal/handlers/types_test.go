package handlers_test

import (
	"testing"
	"time"

	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance guards.
var (
	_ gameevt.GameEvent = (*gameevt.PlayerEliminated)(nil)
	_ gameevt.GameEvent = (*gameevt.ContinentCaptured)(nil)
	_ gameevt.GameEvent = (*gameevt.ContinentLost)(nil)
)

func TestEventTypes_NilPointerSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		callType func() string
	}{
		{
			name:     "PlayerEliminated nil pointer",
			callType: (*gameevt.PlayerEliminated)(nil).EventType,
		},
		{
			name:     "ContinentCaptured nil pointer",
			callType: (*gameevt.ContinentCaptured)(nil).EventType,
		},
		{
			name:     "ContinentLost nil pointer",
			callType: (*gameevt.ContinentLost)(nil).EventType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				_ = test.callType()
			})
		})
	}
}

func TestEventTypes_EventType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    gameevt.GameEvent
		expected string
	}{
		{
			name:     "PlayerEliminated",
			event:    &gameevt.PlayerEliminated{},
			expected: gameevt.TypePlayerEliminated,
		},
		{
			name:     "ContinentCaptured",
			event:    &gameevt.ContinentCaptured{},
			expected: gameevt.TypeContinentCaptured,
		},
		{
			name:     "ContinentLost",
			event:    &gameevt.ContinentLost{},
			expected: gameevt.TypeContinentLost,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.expected, test.event.EventType())
		})
	}
}

func TestEventTypes_GameIDAndTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		event     gameevt.GameEvent
		gameID    int64
		timestamp time.Time
	}{
		{
			name:      "PlayerEliminated",
			event:     gameevt.NewPlayerEliminated(42, "eliminated", "eliminator", now, 7),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "ContinentCaptured",
			event:     gameevt.NewContinentCaptured(42, "user1", now, "europe", 3),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "ContinentLost",
			event:     gameevt.NewContinentLost(42, "user1", now, "asia", 5),
			gameID:    42,
			timestamp: now,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.gameID, test.event.GameID())
			require.Equal(t, test.timestamp, test.event.EventTimestamp())
		})
	}
}

func TestPlayerEliminated_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewPlayerEliminated(42, "victim", "attacker", now, 7)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypePlayerEliminated, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "victim", record["eliminated_user_id"])
	require.Equal(t, "attacker", record["eliminator_user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, int64(7), record["turn"])
}

func TestContinentCaptured_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewContinentCaptured(42, "player1", now, "europe", 3)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeContinentCaptured, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, "europe", record["continent_id"])
	require.Equal(t, int64(3), record["turn"])
}

func TestContinentLost_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := gameevt.NewContinentLost(42, "player1", now, "asia", 5)

	record := event.ToRecord()

	require.Equal(t, gameevt.TypeContinentLost, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, "asia", record["continent_id"])
	require.Equal(t, int64(5), record["turn"])
}

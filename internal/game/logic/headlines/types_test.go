package headlines_test

import (
	"testing"
	"time"

	gameevt "github.com/go-risk-it/go-risk-it/internal/game/events"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/headlines"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance guards.
var (
	_ gameevt.GameEvent = (*headlines.PlayerEliminated)(nil)
	_ gameevt.GameEvent = (*headlines.ContinentCaptured)(nil)
	_ gameevt.GameEvent = (*headlines.ContinentLost)(nil)
)

func TestEventTypes_NilPointerSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		callType func() string
	}{
		{
			name:     "PlayerEliminated nil pointer",
			callType: (*headlines.PlayerEliminated)(nil).EventType,
		},
		{
			name:     "ContinentCaptured nil pointer",
			callType: (*headlines.ContinentCaptured)(nil).EventType,
		},
		{
			name:     "ContinentLost nil pointer",
			callType: (*headlines.ContinentLost)(nil).EventType,
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
			event:    &headlines.PlayerEliminated{},
			expected: headlines.TypePlayerEliminated,
		},
		{
			name:     "ContinentCaptured",
			event:    &headlines.ContinentCaptured{},
			expected: headlines.TypeContinentCaptured,
		},
		{
			name:     "ContinentLost",
			event:    &headlines.ContinentLost{},
			expected: headlines.TypeContinentLost,
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
			event:     headlines.NewPlayerEliminated(42, "eliminated", "eliminator", now, 7),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "ContinentCaptured",
			event:     headlines.NewContinentCaptured(42, "user1", now, "europe", 3),
			gameID:    42,
			timestamp: now,
		},
		{
			name:      "ContinentLost",
			event:     headlines.NewContinentLost(42, "user1", now, "asia", 5),
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
	event := headlines.NewPlayerEliminated(42, "victim", "attacker", now, 7)

	record := event.ToRecord()

	require.Equal(t, headlines.TypePlayerEliminated, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "victim", record["eliminated_user_id"])
	require.Equal(t, "attacker", record["eliminator_user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, int64(7), record["turn"])
}

func TestContinentCaptured_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := headlines.NewContinentCaptured(42, "player1", now, "europe", 3)

	record := event.ToRecord()

	require.Equal(t, headlines.TypeContinentCaptured, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, "europe", record["continent_id"])
	require.Equal(t, int64(3), record["turn"])
}

func TestContinentLost_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC)
	event := headlines.NewContinentLost(42, "player1", now, "asia", 5)

	record := event.ToRecord()

	require.Equal(t, headlines.TypeContinentLost, record["event_type"])
	require.Equal(t, int64(42), record["game_id"])
	require.Equal(t, "player1", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, "asia", record["continent_id"])
	require.Equal(t, int64(5), record["turn"])
}

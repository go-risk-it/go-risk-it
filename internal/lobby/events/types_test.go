package lobby_test

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance guards.
var (
	_ lobbyevt.LobbyEvent = (*lobbyevt.LobbyStateChanged)(nil)
	_ lobbyevt.LobbyEvent = (*lobbyevt.LobbyPlayerConnected)(nil)
	_ lobbyevt.LobbyEvent = (*lobbyevt.CreateGameRequested)(nil)

	// LobbyEvent embeds bus.Event, so both must also satisfy Event.
	_ bus.Event = (*lobbyevt.LobbyStateChanged)(nil)
	_ bus.Event = (*lobbyevt.LobbyPlayerConnected)(nil)
	_ bus.Event = (*lobbyevt.CreateGameRequested)(nil)
)

func TestEventTypes_NilPointerSafety(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		callType func() string
	}{
		{
			name:     "LobbyStateChanged nil pointer",
			callType: (*lobbyevt.LobbyStateChanged)(nil).EventType,
		},
		{
			name:     "LobbyPlayerConnected nil pointer",
			callType: (*lobbyevt.LobbyPlayerConnected)(nil).EventType,
		},
		{
			name:     "CreateGameRequested nil pointer",
			callType: (*lobbyevt.CreateGameRequested)(nil).EventType,
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
		event    lobbyevt.LobbyEvent
		expected string
	}{
		{
			name:     "LobbyStateChanged",
			event:    &lobbyevt.LobbyStateChanged{},
			expected: lobbyevt.TypeLobbyStateChanged,
		},
		{
			name:     "LobbyPlayerConnected",
			event:    &lobbyevt.LobbyPlayerConnected{},
			expected: lobbyevt.TypeLobbyPlayerConnected,
		},
		{
			name:     "CreateGameRequested",
			event:    &lobbyevt.CreateGameRequested{},
			expected: lobbyevt.TypeCreateGameRequested,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.expected, test.event.EventType())
		})
	}
}

func TestEventTypes_LobbyIDAndTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		event   lobbyevt.LobbyEvent
		lobbyID int64
	}{
		{
			name:    "LobbyStateChanged",
			event:   lobbyevt.NewLobbyStateChanged(42, "user-7", nil),
			lobbyID: 42,
		},
		{
			name:    "LobbyPlayerConnected",
			event:   lobbyevt.NewLobbyPlayerConnected(42, "user-7"),
			lobbyID: 42,
		},
		{
			name: "CreateGameRequested",
			event: lobbyevt.NewCreateGameRequested(
				42,
				"user-7",
				time.Now(),
				[]lobbyevt.LobbyPlayer{
					{UserID: "user-1", Name: "Alice"},
				},
			),
			lobbyID: 42,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.lobbyID, test.event.LobbyID())

			// Timestamp should be recent (within last second, set by constructor).
			require.WithinDuration(t, time.Now(), test.event.EventTimestamp(), 1*time.Second)
		})
	}
}

func TestLobbyStateChanged_UserID(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewLobbyStateChanged(10, "user-99", nil)

	require.Equal(t, "user-99", event.UserID())
}

func TestLobbyPlayerConnected_UserID(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewLobbyPlayerConnected(10, "user-99")

	require.Equal(t, "user-99", event.UserID())
}

func TestLobbyStateChanged_ToRecord(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewLobbyStateChanged(42, "user-7", nil)

	record := event.ToRecord()

	require.Equal(t, lobbyevt.TypeLobbyStateChanged, record["event_type"])
	require.Equal(t, int64(42), record["lobby_id"])
	require.Equal(t, "user-7", record["user_id"])
	require.Contains(t, record, "timestamp")

	// Verify timestamp is valid RFC3339.
	ts, ok := record["timestamp"].(string)
	require.True(t, ok, "timestamp should be a string")

	_, err := time.Parse(time.RFC3339, ts)
	require.NoError(t, err, "timestamp should be valid RFC3339")
}

func TestLobbyPlayerConnected_ToRecord(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewLobbyPlayerConnected(42, "user-7")

	record := event.ToRecord()

	require.Equal(t, lobbyevt.TypeLobbyPlayerConnected, record["event_type"])
	require.Equal(t, int64(42), record["lobby_id"])
	require.Equal(t, "user-7", record["user_id"])
	require.Contains(t, record, "timestamp")

	// Verify timestamp is valid RFC3339.
	ts, ok := record["timestamp"].(string)
	require.True(t, ok, "timestamp should be a string")

	_, err := time.Parse(time.RFC3339, ts)
	require.NoError(t, err, "timestamp should be valid RFC3339")
}

func TestCreateGameRequested_UserID(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewCreateGameRequested(10, "user-99", time.Now(), nil)

	require.Equal(t, "user-99", event.UserID())
}

func TestCreateGameRequested_Players(t *testing.T) {
	t.Parallel()

	players := []lobbyevt.LobbyPlayer{
		{UserID: "user-1", Name: "Alice"},
		{UserID: "user-2", Name: "Bob"},
		{UserID: "user-3", Name: "Charlie"},
	}

	event := lobbyevt.NewCreateGameRequested(42, "user-1", time.Now(), players)

	require.Len(t, event.Players, 3)
	require.Equal(t, "user-1", event.Players[0].UserID)
	require.Equal(t, "Alice", event.Players[0].Name)
	require.Equal(t, "user-3", event.Players[2].UserID)
}

func TestCreateGameRequested_ToRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	players := []lobbyevt.LobbyPlayer{
		{UserID: "user-1", Name: "Alice"},
		{UserID: "user-2", Name: "Bob"},
	}

	event := lobbyevt.NewCreateGameRequested(42, "user-7", now, players)

	record := event.ToRecord()

	require.Equal(t, lobbyevt.TypeCreateGameRequested, record["event_type"])
	require.Equal(t, int64(42), record["lobby_id"])
	require.Equal(t, "user-7", record["user_id"])
	require.Equal(t, now.Format(time.RFC3339), record["timestamp"])
	require.Equal(t, 2, record["num_players"])
}

func TestCreateGameRequested_Timestamp(t *testing.T) {
	t.Parallel()

	// CreateGameRequested uses caller-provided timestamp (not time.Now())
	now := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)
	event := lobbyevt.NewCreateGameRequested(42, "user-1", now, nil)

	require.Equal(t, now, event.EventTimestamp())
}

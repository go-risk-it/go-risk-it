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

	// LobbyEvent embeds bus.Event, so both must also satisfy Event.
	_ bus.Event = (*lobbyevt.LobbyStateChanged)(nil)
	_ bus.Event = (*lobbyevt.LobbyPlayerConnected)(nil)
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
			event:   lobbyevt.NewLobbyStateChanged(42, "user-7"),
			lobbyID: 42,
		},
		{
			name:    "LobbyPlayerConnected",
			event:   lobbyevt.NewLobbyPlayerConnected(42, "user-7"),
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

	event := lobbyevt.NewLobbyStateChanged(10, "user-99")

	require.Equal(t, "user-99", event.UserID())
}

func TestLobbyPlayerConnected_UserID(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewLobbyPlayerConnected(10, "user-99")

	require.Equal(t, "user-99", event.UserID())
}

func TestLobbyStateChanged_ToRecord(t *testing.T) {
	t.Parallel()

	event := lobbyevt.NewLobbyStateChanged(42, "user-7")

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

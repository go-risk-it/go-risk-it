package runner //nolint:testpackage // whitebox tests access unexported helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Compile-time interface checks.
var (
	_ Event = GameStartedEvent{}
	_ Event = StateReceivedEvent{}
	_ Event = MoveDecidedEvent{}
	_ Event = TurnSkippedEvent{}
	_ Event = MoveSucceededEvent{}
	_ Event = MoveConflictEvent{}
	_ Event = MoveFailedEvent{}
	_ Event = GameCompleteEvent{}
)

func TestEventTypes_UniqueConstants(t *testing.T) {
	t.Parallel()

	all := []EventType{
		EventGameStarted,
		EventStateReceived,
		EventMoveDecided,
		EventTurnSkipped,
		EventMoveSucceeded,
		EventMoveConflict,
		EventMoveFailed,
		EventGameComplete,
	}

	seen := make(map[EventType]bool)
	for _, et := range all {
		assert.False(t, seen[et], "duplicate event type: %s", et)
		seen[et] = true
	}

	assert.Len(t, seen, 8)
}

func TestEventTypes_CorrectType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		event    Event
		expected EventType
	}{
		{GameStartedEvent{}, EventGameStarted},
		{StateReceivedEvent{}, EventStateReceived},
		{MoveDecidedEvent{}, EventMoveDecided},
		{TurnSkippedEvent{}, EventTurnSkipped},
		{MoveSucceededEvent{}, EventMoveSucceeded},
		{MoveConflictEvent{}, EventMoveConflict},
		{MoveFailedEvent{}, EventMoveFailed},
		{GameCompleteEvent{}, EventGameComplete},
	}

	for _, tc := range tests {
		t.Run(string(tc.expected), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.event.Type())
		})
	}
}

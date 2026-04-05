package snapshot_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/lobby/api/snapshot"
	"github.com/stretchr/testify/require"
)

func TestLobbySnapshot_JSON(t *testing.T) {
	t.Parallel()

	original := snapshot.LobbySnapshot{
		ID: 42,
		Participants: []snapshot.Participant{
			{UserID: "alice"},
			{UserID: "bob"},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	require.JSONEq(t, `{
		"id": 42,
		"participants": [
			{"userId": "alice"},
			{"userId": "bob"}
		]
	}`, string(data))

	var decoded snapshot.LobbySnapshot
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, original, decoded)
}

func TestLobbySnapshot_JSON_EmptyParticipants(t *testing.T) {
	t.Parallel()

	original := snapshot.LobbySnapshot{
		ID:           1,
		Participants: []snapshot.Participant{},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	require.JSONEq(t, `{"id": 1, "participants": []}`, string(data))

	var decoded snapshot.LobbySnapshot
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Equal(t, original, decoded)
}

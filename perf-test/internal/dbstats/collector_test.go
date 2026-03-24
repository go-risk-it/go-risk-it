package dbstats_test

import (
	"encoding/json"
	"testing"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/dbstats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStepDBStats_JSONRoundtrip(t *testing.T) {
	t.Parallel()

	original := dbstats.StepDBStats{
		TopQueries: []dbstats.QueryFingerprint{
			{
				Query:       "SELECT * FROM games WHERE id = $1",
				Calls:       42,
				TotalTimeMs: 150.5,
				MeanTimeMs:  3.58,
				MaxTimeMs:   25.0,
			},
			{
				Query:       "INSERT INTO moves (game_id, player_id) VALUES ($1, $2)",
				Calls:       100,
				TotalTimeMs: 80.2,
				MeanTimeMs:  0.802,
				MaxTimeMs:   12.5,
			},
		},
		TotalQueryTimeMs:  230.7,
		ActiveConnections: 5,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded dbstats.StepDBStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original, decoded)
}

func TestStepDBStats_EmptyQueries(t *testing.T) {
	t.Parallel()

	stats := dbstats.StepDBStats{
		TopQueries:        nil,
		TotalQueryTimeMs:  0,
		ActiveConnections: 0,
	}

	data, err := json.Marshal(stats)
	require.NoError(t, err)

	var decoded dbstats.StepDBStats
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Nil(t, decoded.TopQueries)
	assert.Equal(t, float64(0), decoded.TotalQueryTimeMs)
}

func TestQueryFingerprint_JSONFields(t *testing.T) {
	t.Parallel()

	qf := dbstats.QueryFingerprint{
		Query:       "SELECT 1",
		Calls:       10,
		TotalTimeMs: 5.5,
		MeanTimeMs:  0.55,
		MaxTimeMs:   2.1,
	}

	data, err := json.Marshal(qf)
	require.NoError(t, err)

	// Verify JSON field names.
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)

	assert.Contains(t, m, "query")
	assert.Contains(t, m, "calls")
	assert.Contains(t, m, "total_time_ms")
	assert.Contains(t, m, "mean_time_ms")
	assert.Contains(t, m, "max_time_ms")
}

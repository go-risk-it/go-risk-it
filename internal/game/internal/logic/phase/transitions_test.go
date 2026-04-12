package phase_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/phase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTransition_AllowedTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from sqlc.GamePhaseType
		to   sqlc.GamePhaseType
	}{
		{sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeDEPLOY},
		{sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeATTACK},
		{sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeCONQUER},
		{sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeREINFORCE},
		{sqlc.GamePhaseTypeCONQUER, sqlc.GamePhaseTypeATTACK},
		{sqlc.GamePhaseTypeCONQUER, sqlc.GamePhaseTypeREINFORCE},
		{sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeCARDS},
		{sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeDEPLOY},
	}

	for _, testCase := range tests {
		t.Run(string(testCase.from)+"->"+string(testCase.to), func(t *testing.T) {
			t.Parallel()

			err := phase.ValidateTransition(testCase.from, testCase.to)
			require.NoError(t, err)
		})
	}
}

func TestValidateTransition_DisallowedTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from sqlc.GamePhaseType
		to   sqlc.GamePhaseType
	}{
		{sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeATTACK},
		{sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeCONQUER},
		{sqlc.GamePhaseTypeCARDS, sqlc.GamePhaseTypeREINFORCE},
		{sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeCARDS},
		{sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeCONQUER},
		{sqlc.GamePhaseTypeDEPLOY, sqlc.GamePhaseTypeREINFORCE},
		{sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeCARDS},
		{sqlc.GamePhaseTypeATTACK, sqlc.GamePhaseTypeDEPLOY},
		{sqlc.GamePhaseTypeCONQUER, sqlc.GamePhaseTypeCARDS},
		{sqlc.GamePhaseTypeCONQUER, sqlc.GamePhaseTypeDEPLOY},
		{sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeATTACK},
		{sqlc.GamePhaseTypeREINFORCE, sqlc.GamePhaseTypeCONQUER},
	}

	for _, testCase := range tests {
		t.Run(string(testCase.from)+"->"+string(testCase.to), func(t *testing.T) {
			t.Parallel()

			err := phase.ValidateTransition(testCase.from, testCase.to)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot advance from")
		})
	}
}

func TestValidateTransition_SelfTransitionsDisallowed(t *testing.T) {
	t.Parallel()

	phases := []sqlc.GamePhaseType{
		sqlc.GamePhaseTypeCARDS,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GamePhaseTypeATTACK,
		sqlc.GamePhaseTypeCONQUER,
		sqlc.GamePhaseTypeREINFORCE,
	}

	for _, phaseType := range phases {
		t.Run(string(phaseType)+"->self", func(t *testing.T) {
			t.Parallel()

			err := phase.ValidateTransition(phaseType, phaseType)
			require.Error(t, err)
		})
	}
}

func TestValidateTransition_UnknownPhase(t *testing.T) {
	t.Parallel()

	err := phase.ValidateTransition("NONEXISTENT", sqlc.GamePhaseTypeDEPLOY)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown phase")
}

func TestValidTransitions_AllPhasesHaveEntries(t *testing.T) {
	t.Parallel()

	phases := []sqlc.GamePhaseType{
		sqlc.GamePhaseTypeCARDS,
		sqlc.GamePhaseTypeDEPLOY,
		sqlc.GamePhaseTypeATTACK,
		sqlc.GamePhaseTypeCONQUER,
		sqlc.GamePhaseTypeREINFORCE,
	}

	for _, phaseType := range phases {
		t.Run(string(phaseType), func(t *testing.T) {
			t.Parallel()

			targets, ok := phase.ValidTransitions[phaseType]
			assert.True(t, ok, "phase %s missing from ValidTransitions", phaseType)
			assert.NotEmpty(t, targets, "phase %s has no valid transitions", phaseType)
		})
	}
}

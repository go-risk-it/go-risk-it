package validation_test

import (
	"context"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/validation"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func gameContext(userID string) ctx.GameContext {
	userContext := kernelctx.WithUserID(
		kernelctx.WithSpan(context.Background(), noop.Span{}),
		userID,
	)

	return ctx.WithGameID(userContext, 1)
}

func TestCheckSourceOwnedByPlayer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		userID        string
		regionOwner   string
		regionLabel   string
		expectErr     bool
		expectedError string
	}{
		{
			name:        "succeeds when source region is owned by player",
			userID:      "giovanni",
			regionOwner: "giovanni",
			regionLabel: "source",
			expectErr:   false,
		},
		{
			name:          "fails when attacking region is not owned by player (attack label)",
			userID:        "giovanni",
			regionOwner:   "gabriele",
			regionLabel:   "attacking",
			expectErr:     true,
			expectedError: "attacking region is not owned by player",
		},
		{
			name:          "fails when source region is not owned by player (reinforce label)",
			userID:        "giovanni",
			regionOwner:   "gabriele",
			regionLabel:   "source",
			expectErr:     true,
			expectedError: "source region is not owned by player",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gameCtx := gameContext(test.userID)
			region := &sqlc.GetRegionsByGameRow{
				UserID: test.regionOwner,
			}

			err := validation.CheckSourceOwnedByPlayer(gameCtx, region, test.regionLabel)

			if !test.expectErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)

			var domainErr *domainerrors.DomainError
			require.ErrorAs(t, err, &domainErr)
			require.Equal(t, domainerrors.CategoryValidation, domainErr.Category())
		})
	}
}

func TestCheckTargetNotOwnedByPlayer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		userID        string
		regionOwner   string
		expectErr     bool
		expectedError string
	}{
		{
			name:        "succeeds when target region is not owned by player",
			userID:      "giovanni",
			regionOwner: "gabriele",
			expectErr:   false,
		},
		{
			name:          "fails when target region is owned by player",
			userID:        "giovanni",
			regionOwner:   "giovanni",
			expectErr:     true,
			expectedError: "cannot attack your own region",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gameCtx := gameContext(test.userID)
			region := &sqlc.GetRegionsByGameRow{
				UserID: test.regionOwner,
			}

			err := validation.CheckTargetNotOwnedByPlayer(gameCtx, region)

			if !test.expectErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)

			var domainErr *domainerrors.DomainError
			require.ErrorAs(t, err, &domainErr)
			require.Equal(t, domainerrors.CategoryValidation, domainErr.Category())
		})
	}
}

func TestCheckTargetOwnedByPlayer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		userID        string
		regionOwner   string
		expectErr     bool
		expectedError string
	}{
		{
			name:        "succeeds when target region is owned by player",
			userID:      "giovanni",
			regionOwner: "giovanni",
			expectErr:   false,
		},
		{
			name:          "fails when target region is not owned by player",
			userID:        "giovanni",
			regionOwner:   "gabriele",
			expectErr:     true,
			expectedError: "target region is not owned by player",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gameCtx := gameContext(test.userID)
			region := &sqlc.GetRegionsByGameRow{
				UserID: test.regionOwner,
			}

			err := validation.CheckTargetOwnedByPlayer(gameCtx, region)

			if !test.expectErr {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			require.EqualError(t, err, test.expectedError)

			var domainErr *domainerrors.DomainError
			require.ErrorAs(t, err, &domainErr)
			require.Equal(t, domainerrors.CategoryValidation, domainErr.Category())
		})
	}
}

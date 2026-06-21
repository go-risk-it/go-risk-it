package service_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/logic/move/service"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/stretchr/testify/require"
)

func TestFindRegion_Found(t *testing.T) {
	t.Parallel()

	regions := []snapshot.RegionState{
		{InternalID: 1, ID: "alaska", OwnerID: "user-1", Troops: 3},
		{InternalID: 2, ID: "brazil", OwnerID: "user-2", Troops: 5},
		{InternalID: 3, ID: "congo", OwnerID: "user-1", Troops: 1},
	}

	result, err := service.FindRegion(regions, "brazil")

	require.NoError(t, err)
	require.Equal(t, snapshot.RegionState{
		InternalID: 2,
		ID:         "brazil",
		OwnerID:    "user-2",
		Troops:     5,
	}, result)
}

func TestFindRegion_NotFound(t *testing.T) {
	t.Parallel()

	regions := []snapshot.RegionState{
		{InternalID: 1, ID: "alaska", OwnerID: "user-1", Troops: 3},
		{InternalID: 2, ID: "brazil", OwnerID: "user-2", Troops: 5},
	}

	result, err := service.FindRegion(regions, "congo")

	require.Error(t, err)

	var domErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domErr)
	require.Equal(t, domainerrors.CategoryNotFound, domErr.Category())
	require.Contains(t, err.Error(), "congo")
	require.Equal(t, snapshot.RegionState{}, result)
}

func TestFindRegion_EmptySlice(t *testing.T) {
	t.Parallel()

	result, err := service.FindRegion(nil, "alaska")

	require.Error(t, err)

	var domErr *domainerrors.DomainError
	require.ErrorAs(t, err, &domErr)
	require.Equal(t, domainerrors.CategoryNotFound, domErr.Category())
	require.Contains(t, err.Error(), "alaska")
	require.Equal(t, snapshot.RegionState{}, result)
}

func TestToDBRegion_MapsCorrectly(t *testing.T) {
	t.Parallel()

	region := snapshot.RegionState{
		InternalID: 42,
		ID:         "brazil",
		OwnerID:    "user-7",
		Troops:     5,
	}

	result := service.ToDBRegion(region)

	require.NotNil(t, result)
	require.Equal(t, int64(42), result.ID)
	require.Equal(t, "brazil", result.ExternalReference)
	require.Equal(t, "user-7", result.UserID)
	require.Equal(t, int64(5), result.Troops)
}

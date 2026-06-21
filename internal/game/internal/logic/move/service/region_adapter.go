package service

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/game/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/game/internal/data/sqlc"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

// FindRegion returns the RegionState matching externalRef, or a NotFoundError
// if no region in the slice has that ID.
func FindRegion(regions []snapshot.RegionState, externalRef string) (snapshot.RegionState, error) {
	for _, r := range regions {
		if r.ID == externalRef {
			return r, nil
		}
	}

	return snapshot.RegionState{}, domainerrors.NewNotFoundError(
		fmt.Sprintf("region %s not found in cached state", externalRef),
	)
}

// ToDBRegion maps a snapshot RegionState back to the sqlc row type expected by
// write methods. The mapping is purely structural — no validation is performed.
func ToDBRegion(r snapshot.RegionState) *sqlc.GetRegionsByGameRow {
	return &sqlc.GetRegionsByGameRow{
		ID:                r.InternalID,
		ExternalReference: r.ID,
		UserID:            r.OwnerID,
		Troops:            r.Troops,
	}
}

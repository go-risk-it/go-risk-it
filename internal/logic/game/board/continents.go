package board

import (
	"fmt"
	"slices"

	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
)

type Continent struct {
	ExternalReference string
	BonusTroops       int
	regions           []string
}

// Regions returns a copy of the continent's region identifiers.
func (c *Continent) Regions() []string {
	result := make([]string, len(c.regions))
	copy(result, c.regions)

	return result
}

type Continents interface {
	GetContinentsControlledBy(regions []string) []*Continent
	All() []*Continent
}

type continentsImpl struct {
	continents []*Continent
}

var _ Continents = (*continentsImpl)(nil)

func (c *continentsImpl) All() []*Continent {
	result := make([]*Continent, len(c.continents))
	copy(result, c.continents)

	return result
}

func (c *continentsImpl) GetContinentsControlledBy(regions []string) []*Continent {
	result := make([]*Continent, 0)

	for _, continent := range c.continents {
		if allRegionsContained(continent, regions) {
			result = append(result, continent)
		}
	}

	return result
}

func allRegionsContained(continent *Continent, regions []string) bool {
	for _, region := range continent.regions {
		if !slices.Contains(regions, region) {
			return false
		}
	}

	return true
}

func validateContinents(board *BoardDto) error {
	if len(board.Regions) == 0 {
		return domainerrors.NewValidationError("no regions")
	}

	if len(board.Continents) == 0 {
		return domainerrors.NewValidationError("no continents")
	}

	continentNames := make(map[string]struct{})
	for _, continent := range board.Continents {
		if _, ok := continentNames[continent.ExternalReference]; ok {
			return domainerrors.NewValidationError(
				"duplicate continent id: " + continent.ExternalReference,
			)
		}

		continentNames[continent.ExternalReference] = struct{}{}
	}

	return nil
}

func NewContinents(board *BoardDto) (Continents, error) {
	if err := validateContinents(board); err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	continents := make([]*Continent, len(board.Continents))

	for i, continent := range board.Continents {
		continents[i] = &Continent{
			ExternalReference: continent.ExternalReference,
			BonusTroops:       continent.BonusTroops,
			regions:           make([]string, 0),
		}
	}

	for _, region := range board.Regions {
		for _, continent := range continents {
			if region.Continent == continent.ExternalReference {
				continent.regions = append(continent.regions, region.ExternalReference)
			}
		}
	}

	return &continentsImpl{continents: continents}, nil
}

package continents

import (
	"errors"
	"fmt"
	"slices"

	"github.com/go-risk-it/go-risk-it/internal/logic/game/board/dto"
)

type Continent struct {
	ExternalReference string
	BonusTroops       int
	regions           []string
}

type Continents interface {
	GetContinentsControlledBy(regions []string) []*Continent
}

type ContinentsImpl struct {
	continents []*Continent
}

var _ Continents = (*ContinentsImpl)(nil)

func (c *ContinentsImpl) GetContinentsControlledBy(regions []string) []*Continent {
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

var _ Continents = (*ContinentsImpl)(nil)

func validate(board *dto.Board) error {
	if len(board.Regions) == 0 {
		return errors.New("no regions")
	}

	if len(board.Continents) == 0 {
		return errors.New("no continents")
	}

	continentNames := make(map[string]struct{})
	for _, continent := range board.Continents {
		if _, ok := continentNames[continent.ExternalReference]; ok {
			return fmt.Errorf("duplicate continent id: %s", continent.ExternalReference)
		}

		continentNames[continent.ExternalReference] = struct{}{}
	}

	return nil
}

func New(board *dto.Board) (*ContinentsImpl, error) {
	if err := validate(board); err != nil {
		return nil, fmt.Errorf("invalid board: %w", err)
	}

	continents := make([]*Continent, len(board.Continents))

	for i, continent := range board.Continents {
		continents[i] = &Continent{
			ExternalReference: continent.Name,
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

	return &ContinentsImpl{continents: continents}, nil
}

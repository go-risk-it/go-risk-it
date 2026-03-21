package request

import "errors"

type ReinforceMove struct {
	SourceRegionID string `json:"sourceRegionId"`
	TargetRegionID string `json:"targetRegionId"`
	TroopsInSource int64  `json:"troopsInSource"`
	TroopsInTarget int64  `json:"troopsInTarget"`
	MovingTroops   int64  `json:"movingTroops"`
}

func (r ReinforceMove) Validate() error {
	if r.SourceRegionID == "" {
		return errors.New("sourceRegionId must not be empty")
	}

	if r.TargetRegionID == "" {
		return errors.New("targetRegionId must not be empty")
	}

	if r.SourceRegionID == r.TargetRegionID {
		return errors.New("sourceRegionId and targetRegionId must be different")
	}

	if r.MovingTroops <= 0 {
		return errors.New("movingTroops must be positive")
	}

	return nil
}

package request

import "errors"

type AttackMove struct {
	SourceRegionID  string `json:"sourceRegionId"`
	TargetRegionID  string `json:"targetRegionId"`
	TroopsInSource  int64  `json:"troopsInSource"`
	TroopsInTarget  int64  `json:"troopsInTarget"`
	AttackingTroops int64  `json:"attackingTroops"`
}

func (r AttackMove) Validate() error {
	if r.SourceRegionID == "" {
		return errors.New("sourceRegionId must not be empty")
	}

	if r.TargetRegionID == "" {
		return errors.New("targetRegionId must not be empty")
	}

	if r.SourceRegionID == r.TargetRegionID {
		return errors.New("sourceRegionId and targetRegionId must be different")
	}

	if r.AttackingTroops <= 0 {
		return errors.New("attackingTroops must be positive")
	}

	if r.AttackingTroops > maxTroops {
		return errors.New("attackingTroops exceeds maximum")
	}

	return nil
}

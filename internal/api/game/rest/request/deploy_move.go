package request

import "errors"

type DeployMove struct {
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}

func (r DeployMove) Validate() error {
	if r.RegionID == "" {
		return errors.New("regionId must not be empty")
	}

	if r.DesiredTroops <= 0 {
		return errors.New("desiredTroops must be positive")
	}

	if r.DesiredTroops > maxTroops {
		return errors.New("desiredTroops exceeds maximum")
	}

	return nil
}

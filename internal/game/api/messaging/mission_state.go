package messaging

type MissionType string

const (
	TwoContinents                MissionType = "twoContinents"
	TwoContinentsPlusOne         MissionType = "twoContinentsPlusOne"
	EighteenTerritoriesTwoTroops MissionType = "eighteenTerritoriesTwoTroops"
	TwentyFourTerritories        MissionType = "twentyFourTerritories"
	EliminatePlayer              MissionType = "eliminatePlayer"
)

type EighteenTerritoriesTwoTroopsMission struct{}

type TwentyFourTerritoriesMission struct{}

type TwoContinentsMission struct {
	Continent1 string `json:"continent1"`
	Continent2 string `json:"continent2"`
}

type TwoContinentsPlusOneMission struct {
	Continent1 string `json:"continent1"`
	Continent2 string `json:"continent2"`
}

type EliminatePlayerMission struct {
	TargetUserID string `json:"targetUserId"`
}

type MissionDetails interface {
	TwoContinentsMission |
		TwoContinentsPlusOneMission |
		EighteenTerritoriesTwoTroopsMission |
		TwentyFourTerritoriesMission |
		EliminatePlayerMission
}

type MissionState[T MissionDetails] struct {
	Type    MissionType `json:"type"`
	Details T           `json:"details"`
}

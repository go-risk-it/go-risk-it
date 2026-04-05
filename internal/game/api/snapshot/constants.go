package snapshot

// PhaseType identifies the current game phase.
type PhaseType string

const (
	PhaseCards     PhaseType = "cards"
	PhaseDeploy    PhaseType = "deploy"
	PhaseAttack    PhaseType = "attack"
	PhaseConquer   PhaseType = "conquer"
	PhaseReinforce PhaseType = "reinforce"
)

// CardType identifies the type of a game card.
type CardType string

const (
	CardInfantry  CardType = "infantry"
	CardCavalry   CardType = "cavalry"
	CardArtillery CardType = "artillery"
	CardJolly     CardType = "jolly"
)

// PlayerStatus indicates whether a player is alive or eliminated.
type PlayerStatus string

const (
	PlayerAlive PlayerStatus = "alive"
	PlayerDead  PlayerStatus = "dead"
)

// MissionType identifies the kind of mission assigned to a player.
type MissionType string

const (
	MissionTwoContinents                MissionType = "twoContinents"
	MissionTwoContinentsPlusOne         MissionType = "twoContinentsPlusOne"
	MissionEighteenTerritoriesTwoTroops MissionType = "eighteenTerritoriesTwoTroops"
	MissionTwentyFourTerritories        MissionType = "twentyFourTerritories"
	MissionEliminatePlayer              MissionType = "eliminatePlayer"
)

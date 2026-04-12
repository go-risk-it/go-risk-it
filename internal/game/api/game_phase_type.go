package game

// GamePhaseType represents a game phase as stored in the database.
// These constants use UPPERCASE values matching the DB enum.
// For API/JSON representation, see PhaseType (lowercase).
type GamePhaseType string

const (
	GamePhaseTypeCARDS     GamePhaseType = "CARDS"
	GamePhaseTypeDEPLOY    GamePhaseType = "DEPLOY"
	GamePhaseTypeATTACK    GamePhaseType = "ATTACK"
	GamePhaseTypeCONQUER   GamePhaseType = "CONQUER"
	GamePhaseTypeREINFORCE GamePhaseType = "REINFORCE"
)

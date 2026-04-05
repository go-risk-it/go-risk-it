package config

// DiceConfig controls the dice rolling strategy for attack resolution.
type DiceConfig struct {
	RollStrategy string `koanf:"roll_strategy"`
}

// RegionassignmentConfig controls the region assignment strategy for game
// creation.
type RegionassignmentConfig struct {
	AssignmentStrategy string `koanf:"assignment_strategy"`
}

// HistoryConfig controls how many recent move log entries are sent to
// connecting players.
type HistoryConfig struct {
	Size int64 `koanf:"size"`
}

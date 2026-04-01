package health

import "time"

// Status classifies a game's health state.
type Status string

const (
	StatusHealthy Status = "healthy"
	StatusSlow    Status = "slow"
	StatusStalled Status = "stalled"
	StatusZombie  Status = "zombie"
)

// Thresholds configures health classification boundaries.
type Thresholds struct {
	// SlowMultiplier: time since last move > SlowMultiplier * mean interval → slow.
	SlowMultiplier float64 // default 1.5
	// StalledMultiplier: time since last move > StalledMultiplier * mean interval → stalled.
	StalledMultiplier float64 // default 3.0
	// ZombieAge: game older than this → zombie. If zero, uses 3x mean completed game duration.
	ZombieAge time.Duration
}

// DefaultThresholds returns sensible defaults.
func DefaultThresholds() Thresholds {
	return Thresholds{
		SlowMultiplier:    1.5,
		StalledMultiplier: 3.0,
	}
}

// Distribution summarizes game health across a pool.
type Distribution struct {
	Healthy int `json:"healthy"`
	Slow    int `json:"slow"`
	Stalled int `json:"stalled"`
	Zombie  int `json:"zombie"`
	Total   int `json:"total"`
}

// EffectiveConcurrency returns games that are actually making progress.
func (d Distribution) EffectiveConcurrency() int {
	return d.Healthy + d.Slow
}

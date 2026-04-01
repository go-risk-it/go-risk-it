package chaos

import (
	"errors"
	"time"
)

// Config controls fault injection behavior during load tests.
type Config struct {
	// Strategy chaos: probability [0.0, 1.0] per move.
	SlowMoveRate  float64       // chance of adding artificial delay before a move
	SlowMoveDelay time.Duration // how long to delay (default 2s)
	ErrorMoveRate float64       // chance of returning a strategy error

	// Player chaos: probability per game loop iteration.
	DisconnectRate float64       // chance of forcibly closing a player's WS
	ReconnectDelay time.Duration // how long to wait before reconnecting (0 = no reconnect)
}

// Validate clamps rates to [0,1] and returns an error for nonsensical combos.
func (c *Config) Validate() error {
	c.SlowMoveRate = clamp(c.SlowMoveRate)
	c.ErrorMoveRate = clamp(c.ErrorMoveRate)
	c.DisconnectRate = clamp(c.DisconnectRate)

	if c.SlowMoveRate > 0 && c.SlowMoveDelay <= 0 {
		c.SlowMoveDelay = 2 * time.Second
	}

	if c.DisconnectRate > 0 && c.ReconnectDelay < 0 {
		return errors.New("chaos: ReconnectDelay cannot be negative when DisconnectRate > 0")
	}

	if c.ErrorMoveRate+c.SlowMoveRate > 1.0 {
		return errors.New("chaos: ErrorMoveRate + SlowMoveRate exceeds 1.0")
	}

	return nil
}

// Enabled returns true if any chaos injection is configured.
func (c *Config) Enabled() bool {
	return c.SlowMoveRate > 0 || c.ErrorMoveRate > 0 || c.DisconnectRate > 0
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return v
}

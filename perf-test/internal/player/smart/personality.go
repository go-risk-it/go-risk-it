package smart

import "time"

// Personality configures bot behavior: aggression, greed, risk tolerance, and noise.
// Phase 1 uses only ThinkTime and Name. Scoring weights are used in Phase 2.
type Personality struct {
	Name          string
	Aggression    float64       // 0-1: attack frequency & risk appetite
	Greed         float64       // 0-1: territory expansion vs conservation
	RiskTolerance float64       // 0-1: min advantage ratio for attacks
	Noise         float64       // 0-1: random suboptimality (higher = weaker play)
	ThinkTimeMin  time.Duration // simulated human delay range
	ThinkTimeMax  time.Duration
}

// Beginner returns a cautious, noisy personality.
func Beginner() Personality {
	return Personality{
		Name:          "beginner",
		Aggression:    0.3,
		Greed:         0.4,
		RiskTolerance: 0.3,
		Noise:         0.6,
		ThinkTimeMin:  500 * time.Millisecond,
		ThinkTimeMax:  2 * time.Second,
	}
}

// Normal returns a balanced personality.
func Normal() Personality {
	return Personality{
		Name:          "normal",
		Aggression:    0.5,
		Greed:         0.5,
		RiskTolerance: 0.5,
		Noise:         0.2,
		ThinkTimeMin:  200 * time.Millisecond,
		ThinkTimeMax:  800 * time.Millisecond,
	}
}

// Expert returns an aggressive, precise personality.
func Expert() Personality {
	return Personality{
		Name:          "expert",
		Aggression:    0.7,
		Greed:         0.6,
		RiskTolerance: 0.8,
		Noise:         0.0,
		ThinkTimeMin:  50 * time.Millisecond,
		ThinkTimeMax:  200 * time.Millisecond,
	}
}

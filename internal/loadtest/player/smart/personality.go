package smart

import "time"

// Personality configures bot behavior: name, simulated think time, and combat strategy.
type Personality struct {
	Name         string
	ThinkTimeMin time.Duration // simulated human delay range
	ThinkTimeMax time.Duration

	Aggression        float64 // 0.0 (cautious) to 1.0 (aggressive): controls conquer troop ratio
	BaseAttackRatio   float64 // minimum troop ratio to attack; lower = more aggressive
	ContinentWeight   float64 // multiplier for continent completion bonus in attack scoring
	LargeArmyDiscount float64 // discount on attack threshold at high troop counts
	ContinueAfterCard float64 // 0.0=stop after card, 1.0=never stop
}

// Beginner returns a cautious personality with long think times.
func Beginner() Personality {
	return Personality{
		Name:              "beginner",
		ThinkTimeMin:      500 * time.Millisecond,
		ThinkTimeMax:      2 * time.Second,
		Aggression:        0.3,
		BaseAttackRatio:   1.5,
		ContinentWeight:   1.0,
		LargeArmyDiscount: 1.0,
		ContinueAfterCard: 0.0,
	}
}

// Normal returns a balanced personality with moderate think times.
func Normal() Personality {
	return Personality{
		Name:              "normal",
		ThinkTimeMin:      200 * time.Millisecond,
		ThinkTimeMax:      800 * time.Millisecond,
		Aggression:        0.7,
		BaseAttackRatio:   1.3,
		ContinentWeight:   1.5,
		LargeArmyDiscount: 0.9,
		ContinueAfterCard: 0.5,
	}
}

// Expert returns a fast personality with short think times.
func Expert() Personality {
	return Personality{
		Name:              "expert",
		ThinkTimeMin:      50 * time.Millisecond,
		ThinkTimeMax:      200 * time.Millisecond,
		Aggression:        0.9,
		BaseAttackRatio:   1.0,
		ContinentWeight:   2.0,
		LargeArmyDiscount: 0.8,
		ContinueAfterCard: 1.0,
	}
}

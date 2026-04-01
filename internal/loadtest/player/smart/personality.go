package smart

import "time"

// Personality configures bot behavior: name and simulated think time.
type Personality struct {
	Name         string
	ThinkTimeMin time.Duration // simulated human delay range
	ThinkTimeMax time.Duration
}

// Beginner returns a cautious personality with long think times.
func Beginner() Personality {
	return Personality{
		Name:         "beginner",
		ThinkTimeMin: 500 * time.Millisecond,
		ThinkTimeMax: 2 * time.Second,
	}
}

// Normal returns a balanced personality with moderate think times.
func Normal() Personality {
	return Personality{
		Name:         "normal",
		ThinkTimeMin: 200 * time.Millisecond,
		ThinkTimeMax: 800 * time.Millisecond,
	}
}

// Expert returns a fast personality with short think times.
func Expert() Personality {
	return Personality{
		Name:         "expert",
		ThinkTimeMin: 50 * time.Millisecond,
		ThinkTimeMax: 200 * time.Millisecond,
	}
}

package session

import "time"

// Session tracks an optimization workflow on a branch.
type Session struct {
	ID            string    `json:"id"`
	Branch        string    `json:"branch"`
	BaselineEntry string    `json:"baselineEntry"`
	StartedAt     time.Time `json:"startedAt"`
	Runs          []RunRef  `json:"runs"`
	Status        string    `json:"status"` // "active", "merged", "abandoned"
}

// RunRef records a single staircase run within a session.
type RunRef struct {
	EntryPath    string    `json:"entryPath"`
	CommitSHA    string    `json:"commitSha"`
	CeilingGames int       `json:"ceilingGames"`
	CeilingDelta int       `json:"ceilingDelta"`
	Hypothesis   string    `json:"hypothesis,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// LastCeiling returns the ceiling from the most recent run, or 0 if no runs.
func (s *Session) LastCeiling() int {
	if len(s.Runs) == 0 {
		return 0
	}

	return s.Runs[len(s.Runs)-1].CeilingGames
}

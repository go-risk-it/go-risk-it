package session

import "time"

// Session tracks an optimization workflow on a branch.
type Session struct {
	ID            string    `json:"id"`
	Branch        string    `json:"branch"`
	BaselineEntry string    `json:"baseline_entry"`
	StartedAt     time.Time `json:"started_at"`
	Runs          []RunRef  `json:"runs"`
	Status        string    `json:"status"` // "active", "merged", "abandoned"
}

// RunRef records a single staircase run within a session.
type RunRef struct {
	EntryPath    string    `json:"entry_path"`
	CommitSHA    string    `json:"commit_sha"`
	CeilingGames int       `json:"ceiling_games"`
	CeilingDelta int       `json:"ceiling_delta"`
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

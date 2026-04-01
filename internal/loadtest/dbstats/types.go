package dbstats

// QueryFingerprint captures aggregate stats for a single query pattern.
type QueryFingerprint struct {
	Query       string  `json:"query"`
	Calls       int64   `json:"calls"`
	TotalTimeMs float64 `json:"totalTimeMs"`
	MeanTimeMs  float64 `json:"meanTimeMs"`
	MaxTimeMs   float64 `json:"maxTimeMs"`
}

// StepDBStats holds database statistics for a single staircase step.
type StepDBStats struct {
	TopQueries        []QueryFingerprint `json:"topQueries"`
	TotalQueryTimeMs  float64            `json:"totalQueryTimeMs"`
	ActiveConnections int                `json:"activeConnections"`
}

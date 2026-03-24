package dbstats

// QueryFingerprint captures aggregate stats for a single query pattern.
type QueryFingerprint struct {
	Query       string  `json:"query"`
	Calls       int64   `json:"calls"`
	TotalTimeMs float64 `json:"total_time_ms"`
	MeanTimeMs  float64 `json:"mean_time_ms"`
	MaxTimeMs   float64 `json:"max_time_ms"`
}

// StepDBStats holds database statistics for a single staircase step.
type StepDBStats struct {
	TopQueries        []QueryFingerprint `json:"top_queries"`
	TotalQueryTimeMs  float64            `json:"total_query_time_ms"`
	ActiveConnections int                `json:"active_connections"`
}

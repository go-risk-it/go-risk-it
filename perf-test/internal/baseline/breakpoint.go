package baseline

// LevelResult holds the metrics captured at a particular concurrency level.
type LevelResult struct {
	Games   int
	Metrics MetricsSnapshot
}

// FindBreakingPoints analyzes escalating-load results to find where each SLO
// first fails. Only the first breach per SLO is reported.
func FindBreakingPoints(runs []LevelResult, slos SLOSet) []BreakingPoint {
	allSLOs := append(slos.UserExperience, slos.BoundaryHealth...)

	// Track last good value per SLO.
	type sloTracker struct {
		lastGoodValue float64
		broken        bool
	}

	trackers := make(map[string]*sloTracker, len(allSLOs))
	for _, slo := range allSLOs {
		trackers[slo.Name] = &sloTracker{}
	}

	var breakingPoints []BreakingPoint

	values := func(snap MetricsSnapshot) map[string]float64 {
		return metricValues(snap)
	}

	for _, run := range runs {
		runValues := values(run.Metrics)

		for _, slo := range allSLOs {
			tracker := trackers[slo.Name]
			if tracker.broken {
				continue
			}

			actual, exists := runValues[slo.Metric]
			if !exists {
				continue
			}

			if actual > slo.Threshold {
				tracker.broken = true
				breakingPoints = append(breakingPoints, BreakingPoint{
					SLOName:       slo.Name,
					BreaksAtGames: run.Games,
					LastGoodValue: tracker.lastGoodValue,
					BreakValue:    actual,
				})
			} else {
				tracker.lastGoodValue = actual
			}
		}
	}

	return breakingPoints
}

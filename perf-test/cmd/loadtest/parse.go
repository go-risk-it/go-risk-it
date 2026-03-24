package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/orchestrator"
)

// parseSteps parses a comma-separated list of step counts.
func parseSteps(s string) ([]int, error) {
	if s == "" {
		return nil, nil
	}

	parts := strings.Split(s, ",")
	steps := make([]int, 0, len(parts))

	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid step count %q: %w", p, err)
		}

		steps = append(steps, n)
	}

	return steps, nil
}

// estimateRampDuration estimates the total ramp duration for throughput bucket sizing.
func estimateRampDuration(rampCfg *orchestrator.RampConfig) time.Duration {
	if rampCfg.Multiplier <= 0 {
		return time.Duration(
			rampCfg.MaxGames/max(rampCfg.GamesPerMinute, 1),
		) * time.Minute
	}

	rate := float64(rampCfg.GamesPerMinute)
	total := 0
	steps := 0

	step := rampCfg.StepInterval
	if step <= 0 {
		step = time.Minute
	}

	for total < rampCfg.MaxGames {
		total += int(rate)
		steps++
		rate *= rampCfg.Multiplier
	}

	return time.Duration(steps) * step
}

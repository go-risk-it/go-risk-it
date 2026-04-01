package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// ContainerStats holds resource usage for a single Docker container.
type ContainerStats struct {
	CPUPercent  float64 `json:"cpuPercent"`
	MemoryMB    float64 `json:"memoryMb"`
	MemoryLimit float64 `json:"memoryLimitMb"`
}

// ServerResources holds resource snapshots for all monitored containers.
type ServerResources struct {
	RiskIt ContainerStats `json:"riskIt"`
	DB     ContainerStats `json:"db"`
}

// StatsFunc abstracts the Docker stats command for testing.
type StatsFunc func(containerName string) ([]byte, error)

// dockerStatsJSON matches the JSON output of docker stats --format '{{json .}}'.
type dockerStatsJSON struct {
	CPUPerc  string `json:"cpuPerc"`
	MemUsage string `json:"memUsage"`
}

// DefaultStatsFunc shells out to: docker stats <name> --no-stream --format '{{json .}}'.
func DefaultStatsFunc(containerName string) ([]byte, error) {
	out, err := exec.CommandContext(
		context.Background(),
		"docker", "stats", containerName,
		"--no-stream",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("docker stats: %w", err)
	}

	return out, nil
}

// ParseDockerStats parses raw docker stats JSON into ContainerStats.
// Docker stats returns CPUPerc as "2.50%" and MemUsage as "150MiB / 512MiB".
func ParseDockerStats(raw []byte) (ContainerStats, error) {
	var d dockerStatsJSON
	if err := json.Unmarshal(raw, &d); err != nil {
		return ContainerStats{}, fmt.Errorf("unmarshal docker stats: %w", err)
	}

	var stats ContainerStats

	// Parse CPU percentage: "2.50%" → 2.50.
	if cpu := strings.TrimSuffix(d.CPUPerc, "%"); cpu != "" {
		if v, err := strconv.ParseFloat(cpu, 64); err == nil {
			stats.CPUPercent = v
		}
	}

	// Parse memory usage: "150MiB / 512MiB" → 150.0, 512.0.
	if d.MemUsage != "" {
		parts := strings.Split(d.MemUsage, "/")
		if len(parts) == 2 {
			stats.MemoryMB = parseMemoryValue(strings.TrimSpace(parts[0]))
			stats.MemoryLimit = parseMemoryValue(strings.TrimSpace(parts[1]))
		}
	}

	return stats, nil
}

// parseMemoryValue extracts a numeric MB value from strings like "150MiB", "1.5GiB", "512B".
func parseMemoryValue(s string) float64 {
	s = strings.TrimSpace(s)

	multiplier := 1.0

	switch {
	case strings.HasSuffix(s, "GiB"):
		s = strings.TrimSuffix(s, "GiB")
		multiplier = 1024.0
	case strings.HasSuffix(s, "MiB"):
		s = strings.TrimSuffix(s, "MiB")
	case strings.HasSuffix(s, "KiB"):
		s = strings.TrimSuffix(s, "KiB")
		multiplier = 1.0 / 1024.0
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
		multiplier = 1.0 / (1024.0 * 1024.0)
	}

	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}

	return v * multiplier
}

// CollectServerResources collects stats for risk-it and db containers.
// Best-effort: returns zero stats for containers that fail, never errors.
func CollectServerResources(statsFn StatsFunc) ServerResources {
	var res ServerResources

	ctx := context.Background()

	if raw, err := statsFn("go-risk-it-risk-it-1"); err != nil {
		observe.Warn(ctx, "docker stats failed",
			attribute.String("container", "go-risk-it-risk-it-1"),
			attribute.String("error", err.Error()),
		)
	} else if stats, err := ParseDockerStats(raw); err != nil {
		observe.Warn(ctx, "docker stats parse failed",
			attribute.String("container", "go-risk-it-risk-it-1"),
			attribute.String("error", err.Error()),
		)
	} else {
		res.RiskIt = stats
	}

	if raw, err := statsFn("supabase-db"); err != nil {
		observe.Warn(ctx, "docker stats failed",
			attribute.String("container", "supabase-db"),
			attribute.String("error", err.Error()),
		)
	} else if stats, err := ParseDockerStats(raw); err != nil {
		observe.Warn(ctx, "docker stats parse failed",
			attribute.String("container", "supabase-db"),
			attribute.String("error", err.Error()),
		)
	} else {
		res.DB = stats
	}

	return res
}

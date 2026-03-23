package resources

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// ContainerStats holds resource usage for a single Docker container.
type ContainerStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryMB    float64 `json:"memory_mb"`
	MemoryLimit float64 `json:"memory_limit_mb"`
}

// ServerResources holds resource snapshots for all monitored containers.
type ServerResources struct {
	RiskIt ContainerStats `json:"risk_it"`
	DB     ContainerStats `json:"db"`
}

// StatsFunc abstracts the Docker stats command for testing.
type StatsFunc func(containerName string) ([]byte, error)

// dockerStatsJSON matches the JSON output of docker stats --format '{{json .}}'.
type dockerStatsJSON struct {
	CPUPerc  string `json:"CPUPerc"`
	MemUsage string `json:"MemUsage"`
}

// DefaultStatsFunc shells out to: docker stats <name> --no-stream --format '{{json .}}'.
func DefaultStatsFunc(containerName string) ([]byte, error) {
	return exec.Command(
		"docker", "stats", containerName,
		"--no-stream",
		"--format", "{{json .}}",
	).Output()
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

	if raw, err := statsFn("risk-it"); err != nil {
		log.Printf("docker stats risk-it: %v", err)
	} else if stats, err := ParseDockerStats(raw); err != nil {
		log.Printf("parse docker stats risk-it: %v", err)
	} else {
		res.RiskIt = stats
	}

	if raw, err := statsFn("db"); err != nil {
		log.Printf("docker stats db: %v", err)
	} else if stats, err := ParseDockerStats(raw); err != nil {
		log.Printf("parse docker stats db: %v", err)
	} else {
		res.DB = stats
	}

	return res
}

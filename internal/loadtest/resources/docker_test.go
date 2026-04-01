package resources_test

import (
	"errors"
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDockerStats_ValidJSON(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"CPUPerc":"2.50%","MemUsage":"150MiB / 512MiB"}`)
	stats, err := resources.ParseDockerStats(raw)
	require.NoError(t, err)
	assert.InDelta(t, 2.50, stats.CPUPercent, 0.01)
	assert.InDelta(t, 150.0, stats.MemoryMB, 0.1)
	assert.InDelta(t, 512.0, stats.MemoryLimit, 0.1)
}

func TestParseDockerStats_ZeroCPU(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"CPUPerc":"0.00%","MemUsage":"64MiB / 256MiB"}`)
	stats, err := resources.ParseDockerStats(raw)
	require.NoError(t, err)
	assert.InDelta(t, 0.0, stats.CPUPercent, 0.01)
	assert.InDelta(t, 64.0, stats.MemoryMB, 0.1)
}

func TestParseDockerStats_GiBMemory(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"CPUPerc":"10.00%","MemUsage":"1.5GiB / 4GiB"}`)
	stats, err := resources.ParseDockerStats(raw)
	require.NoError(t, err)
	assert.InDelta(t, 10.0, stats.CPUPercent, 0.01)
	assert.InDelta(t, 1536.0, stats.MemoryMB, 0.1)    // 1.5 * 1024
	assert.InDelta(t, 4096.0, stats.MemoryLimit, 0.1) // 4 * 1024
}

func TestParseDockerStats_InvalidJSON(t *testing.T) {
	t.Parallel()

	raw := []byte(`not json`)
	_, err := resources.ParseDockerStats(raw)
	require.Error(t, err)
}

func TestParseDockerStats_MissingFields(t *testing.T) {
	t.Parallel()

	raw := []byte(`{}`)
	stats, err := resources.ParseDockerStats(raw)
	require.NoError(t, err)
	assert.Equal(t, 0.0, stats.CPUPercent)
	assert.Equal(t, 0.0, stats.MemoryMB)
	assert.Equal(t, 0.0, stats.MemoryLimit)
}

func TestCollectServerResources_BothContainers(t *testing.T) {
	t.Parallel()

	fake := func(name string) ([]byte, error) {
		return []byte(`{"CPUPerc":"5.00%","MemUsage":"200MiB / 1024MiB"}`), nil
	}
	res := resources.CollectServerResources(fake)
	assert.InDelta(t, 5.0, res.RiskIt.CPUPercent, 0.01)
	assert.InDelta(t, 200.0, res.RiskIt.MemoryMB, 0.1)
	assert.InDelta(t, 5.0, res.DB.CPUPercent, 0.01)
	assert.InDelta(t, 200.0, res.DB.MemoryMB, 0.1)
}

func TestCollectServerResources_ContainerDown(t *testing.T) {
	t.Parallel()

	fake := func(name string) ([]byte, error) {
		if name == "go-risk-it-risk-it-1" {
			return nil, errors.New("container not running")
		}

		return []byte(`{"CPUPerc":"3.00%","MemUsage":"100MiB / 512MiB"}`), nil
	}
	res := resources.CollectServerResources(fake)
	assert.Equal(t, 0.0, res.RiskIt.CPUPercent)
	assert.Equal(t, 0.0, res.RiskIt.MemoryMB)
	assert.InDelta(t, 3.0, res.DB.CPUPercent, 0.01)
	assert.InDelta(t, 100.0, res.DB.MemoryMB, 0.1)
}

package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// maxLatencyMs is the upper bound for histogram recording (30 seconds in ms).
const maxLatencyMs = 30_000

// Collector aggregates latency histograms and counters across concurrent games.
// All methods are safe for concurrent use.
type Collector struct {
	mu sync.Mutex

	// Per-operation REST latency histograms (keyed by action type name).
	restLatency map[string]*hdrhistogram.Histogram

	// WS delivery latency: time from REST response to WS state update.
	wsDelivery *hdrhistogram.Histogram

	// End-to-end move latency: time from before executeAction to after WS update.
	e2eMove *hdrhistogram.Histogram

	// Atomic counters.
	totalMoves     atomic.Int64
	totalErrors    atomic.Int64
	gamesCompleted atomic.Int64
	gamesTimedOut  atomic.Int64
}

// NewCollector creates a new metrics collector with initialized histograms.
func NewCollector() *Collector {
	return &Collector{
		restLatency: make(map[string]*hdrhistogram.Histogram),
		wsDelivery:  hdrhistogram.New(1, maxLatencyMs, 3),
		e2eMove:     hdrhistogram.New(1, maxLatencyMs, 3),
	}
}

// getOrCreateHist returns the histogram for the given action type, creating it if needed.
func (c *Collector) getOrCreateHist(actionType string) *hdrhistogram.Histogram {
	h, ok := c.restLatency[actionType]
	if !ok {
		h = hdrhistogram.New(1, maxLatencyMs, 3)
		c.restLatency[actionType] = h
	}

	return h
}

// RecordREST records a REST API call latency for the given action type.
func (c *Collector) RecordREST(actionType string, d time.Duration) {
	ms := d.Milliseconds()
	if ms < 1 {
		ms = 1
	}

	if ms > maxLatencyMs {
		ms = maxLatencyMs
	}

	c.mu.Lock()
	_ = c.getOrCreateHist(actionType).RecordValue(ms)
	c.mu.Unlock()
}

// RecordWSDelivery records the latency from REST response to WS state update arriving.
func (c *Collector) RecordWSDelivery(d time.Duration) {
	ms := d.Milliseconds()
	if ms < 1 {
		ms = 1
	}

	if ms > maxLatencyMs {
		ms = maxLatencyMs
	}

	c.mu.Lock()
	_ = c.wsDelivery.RecordValue(ms)
	c.mu.Unlock()
}

// RecordE2E records end-to-end move latency (from before action to after WS update).
func (c *Collector) RecordE2E(d time.Duration) {
	ms := d.Milliseconds()
	if ms < 1 {
		ms = 1
	}

	if ms > maxLatencyMs {
		ms = maxLatencyMs
	}

	c.mu.Lock()
	_ = c.e2eMove.RecordValue(ms)
	c.mu.Unlock()
}

// RecordMove increments the total moves counter.
func (c *Collector) RecordMove() {
	c.totalMoves.Add(1)
}

// RecordError increments the total errors counter.
func (c *Collector) RecordError() {
	c.totalErrors.Add(1)
}

// RecordGameComplete increments the games completed counter.
func (c *Collector) RecordGameComplete() {
	c.gamesCompleted.Add(1)
}

// RecordGameTimedOut increments the games timed out counter.
func (c *Collector) RecordGameTimedOut() {
	c.gamesTimedOut.Add(1)
}

// Snapshot returns a point-in-time copy of all metrics for reporting.
func (c *Collector) Snapshot() *Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	rest := make(map[string]HistogramSnapshot, len(c.restLatency))
	for name, h := range c.restLatency {
		rest[name] = snapshotHist(h)
	}

	return &Snapshot{
		RESTLatency:    rest,
		WSDelivery:     snapshotHist(c.wsDelivery),
		E2EMove:        snapshotHist(c.e2eMove),
		TotalMoves:     c.totalMoves.Load(),
		TotalErrors:    c.totalErrors.Load(),
		GamesCompleted: c.gamesCompleted.Load(),
		GamesTimedOut:  c.gamesTimedOut.Load(),
	}
}

// Snapshot is a point-in-time copy of collector metrics.
type Snapshot struct {
	RESTLatency    map[string]HistogramSnapshot
	WSDelivery     HistogramSnapshot
	E2EMove        HistogramSnapshot
	TotalMoves     int64
	TotalErrors    int64
	GamesCompleted int64
	GamesTimedOut  int64
}

// HistogramSnapshot holds percentile values from an HDR histogram.
type HistogramSnapshot struct {
	P50   int64
	P95   int64
	P99   int64
	Max   int64
	Count int64
}

func snapshotHist(h *hdrhistogram.Histogram) HistogramSnapshot {
	return HistogramSnapshot{
		P50:   h.ValueAtQuantile(50),
		P95:   h.ValueAtQuantile(95),
		P99:   h.ValueAtQuantile(99),
		Max:   h.Max(),
		Count: h.TotalCount(),
	}
}

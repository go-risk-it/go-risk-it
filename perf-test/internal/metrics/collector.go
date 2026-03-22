package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

// maxLatencyMs is the upper bound for histogram recording (30 seconds in ms).
const maxLatencyMs = 30_000

// throughputBucketSize is the fixed time window for throughput measurement.
const throughputBucketSize = 5 * time.Second

// knownPhases are the 5 game phases, pre-initialized to avoid map contention.
var knownPhases = []string{"cards", "deploy", "attack", "conquer", "reinforce"}

// Collector aggregates latency histograms and counters across concurrent games.
// All methods are safe for concurrent use.
type Collector struct {
	mu sync.Mutex

	// Per-operation REST latency histograms (keyed by action type name).
	restLatency map[string]*hdrhistogram.Histogram

	// Per-phase E2E latency histograms (mutex-protected, same as restLatency).
	phaseLatency map[string]*hdrhistogram.Histogram

	// WS delivery latency: time from REST response to WS state update.
	wsDelivery *hdrhistogram.Histogram

	// End-to-end move latency: time from before executeAction to after WS update.
	e2eMove *hdrhistogram.Histogram

	// HTTP status code distribution (mutex-protected, lazily created).
	httpStatusCounts map[int]*atomic.Int64

	// Atomic counters.
	totalMoves     atomic.Int64
	totalErrors    atomic.Int64
	gamesCompleted atomic.Int64
	gamesTimedOut  atomic.Int64
	gamesFatal     atomic.Int64

	// Resilience counters.
	totalRetries           atomic.Int64
	totalConflicts         atomic.Int64
	totalReconnects        atomic.Int64
	totalReconnectFailures atomic.Int64

	// Phase transition counters (pre-initialized for all 5 phases).
	phaseEntries map[string]*atomic.Int64
	phaseMoves   map[string]*atomic.Int64

	// Error breakdown by category (pre-initialized).
	errorCounts map[string]*atomic.Int64

	// Chaos event counters (pre-initialized).
	chaosEvents map[string]*atomic.Int64

	// Throughput time-series buckets.
	startTime   time.Time
	moveBuckets []atomic.Int64
}

// NewCollector creates a new metrics collector with initialized histograms.
// maxDuration sizes the throughput bucket slice.
func NewCollector(maxDuration time.Duration) *Collector {
	numBuckets := int(maxDuration/throughputBucketSize) + 1

	phaseEntries := make(map[string]*atomic.Int64, len(knownPhases))
	phaseMoves := make(map[string]*atomic.Int64, len(knownPhases))

	for _, p := range knownPhases {
		phaseEntries[p] = &atomic.Int64{}
		phaseMoves[p] = &atomic.Int64{}
	}

	errorCounts := map[string]*atomic.Int64{
		"strategy":  {},
		"execution": {},
		"transient": {},
		"timeout":   {},
	}

	chaosEvents := map[string]*atomic.Int64{
		"disconnect": {},
		"reconnect":  {},
		"slow_move":  {},
		"error_move": {},
	}

	return &Collector{
		restLatency:      make(map[string]*hdrhistogram.Histogram),
		phaseLatency:     make(map[string]*hdrhistogram.Histogram),
		wsDelivery:       hdrhistogram.New(1, maxLatencyMs, 3),
		e2eMove:          hdrhistogram.New(1, maxLatencyMs, 3),
		httpStatusCounts: make(map[int]*atomic.Int64),
		phaseEntries:     phaseEntries,
		phaseMoves:       phaseMoves,
		errorCounts:      errorCounts,
		chaosEvents:      chaosEvents,
		startTime:        time.Now(),
		moveBuckets:      make([]atomic.Int64, numBuckets),
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

// getOrCreatePhaseHist returns the phase histogram, creating it if needed.
func (c *Collector) getOrCreatePhaseHist(phase string) *hdrhistogram.Histogram {
	h, ok := c.phaseLatency[phase]
	if !ok {
		h = hdrhistogram.New(1, maxLatencyMs, 3)
		c.phaseLatency[phase] = h
	}

	return h
}

// clampMs clamps a duration to a valid millisecond range for histogram recording.
func clampMs(d time.Duration) int64 {
	ms := d.Milliseconds()
	if ms < 1 {
		ms = 1
	}

	if ms > maxLatencyMs {
		ms = maxLatencyMs
	}

	return ms
}

// RecordREST records a REST API call latency for the given action type.
func (c *Collector) RecordREST(actionType string, d time.Duration) {
	c.mu.Lock()
	_ = c.getOrCreateHist(actionType).RecordValue(clampMs(d))
	c.mu.Unlock()
}

// RecordWSDelivery records the latency from REST response to WS state update arriving.
func (c *Collector) RecordWSDelivery(d time.Duration) {
	c.mu.Lock()
	_ = c.wsDelivery.RecordValue(clampMs(d))
	c.mu.Unlock()
}

// RecordE2E records end-to-end move latency (from before action to after WS update).
func (c *Collector) RecordE2E(d time.Duration) {
	c.mu.Lock()
	_ = c.e2eMove.RecordValue(clampMs(d))
	c.mu.Unlock()
}

// RecordPhaseLatency records E2E latency tagged by game phase.
func (c *Collector) RecordPhaseLatency(phase string, d time.Duration) {
	c.mu.Lock()
	_ = c.getOrCreatePhaseHist(phase).RecordValue(clampMs(d))
	c.mu.Unlock()
}

// RecordPhaseEntry increments the phase entry count.
func (c *Collector) RecordPhaseEntry(phase string) {
	if counter, ok := c.phaseEntries[phase]; ok {
		counter.Add(1)
	}
}

// RecordPhaseMove increments the phase move count.
func (c *Collector) RecordPhaseMove(phase string) {
	if counter, ok := c.phaseMoves[phase]; ok {
		counter.Add(1)
	}
}

// RecordHTTPStatus increments the counter for the given HTTP status code.
func (c *Collector) RecordHTTPStatus(statusCode int) {
	c.mu.Lock()
	counter, ok := c.httpStatusCounts[statusCode]
	if !ok {
		counter = &atomic.Int64{}
		c.httpStatusCounts[statusCode] = counter
	}
	c.mu.Unlock()

	counter.Add(1)
}

// RecordErrorType increments the specific error category counter.
func (c *Collector) RecordErrorType(errorType string) {
	if counter, ok := c.errorCounts[errorType]; ok {
		counter.Add(1)
	}
}

// RecordTimedMove records a move in the throughput time-series bucket.
func (c *Collector) RecordTimedMove() {
	elapsed := time.Since(c.startTime)
	idx := int(elapsed / throughputBucketSize)

	if idx >= 0 && idx < len(c.moveBuckets) {
		c.moveBuckets[idx].Add(1)
	}
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

// RecordGameFatal increments the games fatal error counter.
func (c *Collector) RecordGameFatal() {
	c.gamesFatal.Add(1)
}

// RecordRetry increments the REST retry counter.
func (c *Collector) RecordRetry() {
	c.totalRetries.Add(1)
}

// RecordConflict increments the 409 conflict counter.
func (c *Collector) RecordConflict() {
	c.totalConflicts.Add(1)
}

// RecordReconnect increments the WS reconnection attempt counter.
func (c *Collector) RecordReconnect() {
	c.totalReconnects.Add(1)
}

// RecordReconnectFailure increments the WS reconnection failure counter.
func (c *Collector) RecordReconnectFailure() {
	c.totalReconnectFailures.Add(1)
}

// RecordChaosEvent increments the counter for the given chaos event type.
func (c *Collector) RecordChaosEvent(eventType string) {
	if counter, ok := c.chaosEvents[eventType]; ok {
		counter.Add(1)
	}
}

// ThroughputBucket represents moves in a fixed time window.
type ThroughputBucket struct {
	OffsetSec float64
	Moves     int64
}

// Snapshot returns a point-in-time copy of all metrics for reporting.
func (c *Collector) Snapshot() *Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	rest := make(map[string]HistogramSnapshot, len(c.restLatency))
	for name, h := range c.restLatency {
		rest[name] = snapshotHist(h)
	}

	phaseLatency := make(map[string]HistogramSnapshot, len(c.phaseLatency))
	for name, h := range c.phaseLatency {
		phaseLatency[name] = snapshotHist(h)
	}

	phaseEntries := make(map[string]int64, len(c.phaseEntries))
	for name, counter := range c.phaseEntries {
		phaseEntries[name] = counter.Load()
	}

	phaseMoves := make(map[string]int64, len(c.phaseMoves))
	for name, counter := range c.phaseMoves {
		phaseMoves[name] = counter.Load()
	}

	httpStatusCounts := make(map[int]int64, len(c.httpStatusCounts))
	for code, counter := range c.httpStatusCounts {
		httpStatusCounts[code] = counter.Load()
	}

	errorBreakdown := make(map[string]int64, len(c.errorCounts))
	for name, counter := range c.errorCounts {
		errorBreakdown[name] = counter.Load()
	}

	chaosEvents := make(map[string]int64, len(c.chaosEvents))
	for name, counter := range c.chaosEvents {
		v := counter.Load()
		if v > 0 {
			chaosEvents[name] = v
		}
	}

	// Only include non-zero throughput buckets.
	var throughputBuckets []ThroughputBucket
	for i := range c.moveBuckets {
		moves := c.moveBuckets[i].Load()
		if moves > 0 {
			throughputBuckets = append(throughputBuckets, ThroughputBucket{
				OffsetSec: float64(i) * throughputBucketSize.Seconds(),
				Moves:     moves,
			})
		}
	}

	return &Snapshot{
		RESTLatency:            rest,
		WSDelivery:             snapshotHist(c.wsDelivery),
		E2EMove:                snapshotHist(c.e2eMove),
		TotalMoves:             c.totalMoves.Load(),
		TotalErrors:            c.totalErrors.Load(),
		GamesCompleted:         c.gamesCompleted.Load(),
		GamesTimedOut:          c.gamesTimedOut.Load(),
		GamesFatal:             c.gamesFatal.Load(),
		TotalRetries:           c.totalRetries.Load(),
		TotalConflicts:         c.totalConflicts.Load(),
		TotalReconnects:        c.totalReconnects.Load(),
		TotalReconnectFailures: c.totalReconnectFailures.Load(),
		PhaseLatency:           phaseLatency,
		PhaseEntries:           phaseEntries,
		PhaseMoves:             phaseMoves,
		HTTPStatusCounts:       httpStatusCounts,
		ErrorBreakdown:         errorBreakdown,
		ChaosEvents:            chaosEvents,
		ThroughputBuckets:      throughputBuckets,
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
	GamesFatal     int64

	// Resilience counters.
	TotalRetries           int64
	TotalConflicts         int64
	TotalReconnects        int64
	TotalReconnectFailures int64

	// Phase metrics.
	PhaseLatency map[string]HistogramSnapshot
	PhaseEntries map[string]int64
	PhaseMoves   map[string]int64

	// HTTP status distribution.
	HTTPStatusCounts map[int]int64

	// Error breakdown by category.
	ErrorBreakdown map[string]int64

	// Chaos event counters (only non-zero events).
	ChaosEvents map[string]int64

	// Throughput time-series.
	ThroughputBuckets []ThroughputBucket
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

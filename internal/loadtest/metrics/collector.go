package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
)

// maxLatencyMs is the upper bound for histogram recording (30 seconds in ms).
const maxLatencyMs = 30_000

// throughputBucketSize is the fixed time window for throughput measurement.
const throughputBucketSize = 5 * time.Second

// Phase represents a game phase for metrics recording.
type Phase string

const (
	PhaseCards     Phase = "cards"
	PhaseDeploy    Phase = "deploy"
	PhaseAttack    Phase = "attack"
	PhaseConquer   Phase = "conquer"
	PhaseReinforce Phase = "reinforce"
)

// knownPhases are the 5 game phases, pre-initialized to avoid map contention.
var knownPhases = []Phase{PhaseCards, PhaseDeploy, PhaseAttack, PhaseConquer, PhaseReinforce}

// ErrorType represents a categorized error for metrics recording.
type ErrorType string

const (
	ErrorTypeStrategy   ErrorType = "strategy"
	ErrorTypeExecution  ErrorType = "execution"
	ErrorTypeTransient  ErrorType = "transient"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeStaleState ErrorType = "stale_state"
	ErrorTypeConflict   ErrorType = "conflict"
)

// ChaosEvent represents a chaos engineering event type.
type ChaosEvent string

const (
	ChaosEventDisconnect ChaosEvent = "disconnect"
	ChaosEventReconnect  ChaosEvent = "reconnect"
	ChaosEventSlowMove   ChaosEvent = "slow_move"
	ChaosEventErrorMove  ChaosEvent = "error_move"
)

// StepAccumulator aggregates latency histograms and counters for a single
// staircase step. One instance is created per step for clean per-step
// percentiles. All methods are safe for concurrent use.
type StepAccumulator struct {
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
	gamesCancelled atomic.Int64

	// Resilience counters.
	totalRetries           atomic.Int64
	totalConflicts         atomic.Int64
	totalReconnects        atomic.Int64
	totalReconnectFailures atomic.Int64

	// Phase transition counters (pre-initialized for all 5 phases).
	phaseEntries map[Phase]*atomic.Int64
	phaseMoves   map[Phase]*atomic.Int64

	// Error breakdown by category (pre-initialized).
	errorCounts map[ErrorType]*atomic.Int64

	// Chaos event counters (pre-initialized).
	chaosEvents map[ChaosEvent]*atomic.Int64

	// Throughput time-series buckets.
	startTime   time.Time
	moveBuckets []atomic.Int64

	// Warm-up filtering: gate histogram recording until MarkWarmUpDone is called.
	warmUpConfigured atomic.Bool
	warmUpDone       atomic.Bool
}

// NewStepAccumulator creates a new per-step metrics accumulator with initialized
// histograms. maxDuration sizes the throughput bucket slice.
func NewStepAccumulator(maxDuration time.Duration) *StepAccumulator {
	numBuckets := int(maxDuration/throughputBucketSize) + 1

	phaseEntries := make(map[Phase]*atomic.Int64, len(knownPhases))
	phaseMoves := make(map[Phase]*atomic.Int64, len(knownPhases))

	for _, p := range knownPhases {
		phaseEntries[p] = &atomic.Int64{}
		phaseMoves[p] = &atomic.Int64{}
	}

	errorCounts := map[ErrorType]*atomic.Int64{
		ErrorTypeStrategy:   {},
		ErrorTypeExecution:  {},
		ErrorTypeTransient:  {},
		ErrorTypeTimeout:    {},
		ErrorTypeStaleState: {},
		ErrorTypeConflict:   {},
	}

	chaosEvents := map[ChaosEvent]*atomic.Int64{
		ChaosEventDisconnect: {},
		ChaosEventReconnect:  {},
		ChaosEventSlowMove:   {},
		ChaosEventErrorMove:  {},
	}

	return &StepAccumulator{
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
func (a *StepAccumulator) getOrCreateHist(actionType string) *hdrhistogram.Histogram {
	h, ok := a.restLatency[actionType]
	if !ok {
		h = hdrhistogram.New(1, maxLatencyMs, 3)
		a.restLatency[actionType] = h
	}

	return h
}

// getOrCreatePhaseHist returns the phase histogram, creating it if needed.
func (a *StepAccumulator) getOrCreatePhaseHist(phase string) *hdrhistogram.Histogram {
	h, ok := a.phaseLatency[phase]
	if !ok {
		h = hdrhistogram.New(1, maxLatencyMs, 3)
		a.phaseLatency[phase] = h
	}

	return h
}

// clampMs clamps a duration to a valid millisecond range for histogram recording.
func clampMs(d time.Duration) int64 {
	ms := min(max(d.Milliseconds(), 1), maxLatencyMs)

	return ms
}

// recordToHist records d to h, gated by warm-up status.
// Must NOT be called while a.mu is held (acquires a.mu internally).
func (a *StepAccumulator) recordToHist(h *hdrhistogram.Histogram, d time.Duration) {
	if !a.isWarmUpDone() {
		return
	}

	a.mu.Lock()
	_ = h.RecordValue(clampMs(d))
	a.mu.Unlock()
}

// RecordREST records a REST API call latency for the given action type.
func (a *StepAccumulator) RecordREST(actionType string, d time.Duration) {
	if a.isWarmUpDone() {
		a.mu.Lock()
		_ = a.getOrCreateHist(actionType).RecordValue(clampMs(d))
		a.mu.Unlock()
	}
}

// RecordWSDelivery records the latency from REST response to WS state update arriving.
func (a *StepAccumulator) RecordWSDelivery(d time.Duration) {
	a.recordToHist(a.wsDelivery, d)
}

// RecordE2E records end-to-end move latency (from before action to after WS update).
func (a *StepAccumulator) RecordE2E(d time.Duration) {
	a.recordToHist(a.e2eMove, d)
}

// RecordPhaseLatency records E2E latency tagged by game phase.
func (a *StepAccumulator) RecordPhaseLatency(phase string, d time.Duration) {
	if a.isWarmUpDone() {
		a.mu.Lock()
		_ = a.getOrCreatePhaseHist(phase).RecordValue(clampMs(d))
		a.mu.Unlock()
	}
}

// RecordPhaseEntry increments the phase entry count.
func (a *StepAccumulator) RecordPhaseEntry(phase Phase) {
	if counter, ok := a.phaseEntries[phase]; ok {
		counter.Add(1)
	}
}

// RecordPhaseMove increments the phase move count.
func (a *StepAccumulator) RecordPhaseMove(phase Phase) {
	if counter, ok := a.phaseMoves[phase]; ok {
		counter.Add(1)
	}
}

// RecordHTTPStatus increments the counter for the given HTTP status code.
func (a *StepAccumulator) RecordHTTPStatus(statusCode int) {
	a.mu.Lock()
	counter, ok := a.httpStatusCounts[statusCode]
	if !ok {
		counter = &atomic.Int64{}
		a.httpStatusCounts[statusCode] = counter
	}
	a.mu.Unlock()

	counter.Add(1)
}

// RecordErrorType increments the specific error category counter.
func (a *StepAccumulator) RecordErrorType(errorType ErrorType) {
	if counter, ok := a.errorCounts[errorType]; ok {
		counter.Add(1)
	}
}

// RecordTimedMove records a move in the throughput time-series bucket.
func (a *StepAccumulator) RecordTimedMove() {
	elapsed := time.Since(a.startTime)
	idx := int(elapsed / throughputBucketSize)

	if idx >= 0 && idx < len(a.moveBuckets) {
		a.moveBuckets[idx].Add(1)
	} else {
		observe.Warn(context.Background(), "move exceeds max duration, not counted in throughput",
			attribute.String("elapsed", elapsed.String()),
		)
	}
}

// RecordMove increments the total moves counter.
func (a *StepAccumulator) RecordMove() {
	a.totalMoves.Add(1)
}

// RecordError increments the total errors counter.
func (a *StepAccumulator) RecordError() {
	a.totalErrors.Add(1)
}

// RecordGameComplete increments the games completed counter.
//
//nolint:unparam // interface conformance / future use
func (a *StepAccumulator) RecordGameComplete(duration time.Duration, moves int) {
	a.gamesCompleted.Add(1)
}

// RecordGameTimedOut increments the games timed out counter.
//
//nolint:unparam // interface conformance / future use
func (a *StepAccumulator) RecordGameTimedOut(duration time.Duration, moves int) {
	a.gamesTimedOut.Add(1)
}

// RecordGameFatal increments the games fatal error counter.
func (a *StepAccumulator) RecordGameFatal() {
	a.gamesFatal.Add(1)
}

// RecordGameCancelled increments the games cancelled counter (step transition).
func (a *StepAccumulator) RecordGameCancelled() {
	a.gamesCancelled.Add(1)
}

// RecordRetry increments the REST retry counter.
func (a *StepAccumulator) RecordRetry() {
	a.totalRetries.Add(1)
}

// RecordConflict increments the 409 conflict counter.
func (a *StepAccumulator) RecordConflict() {
	a.totalConflicts.Add(1)
	a.RecordErrorType(ErrorTypeConflict)
}

// RecordReconnect increments the WS reconnection attempt counter.
func (a *StepAccumulator) RecordReconnect() {
	a.totalReconnects.Add(1)
}

// RecordReconnectFailure increments the WS reconnection failure counter.
func (a *StepAccumulator) RecordReconnectFailure() {
	a.totalReconnectFailures.Add(1)
}

// RecordChaosEvent increments the counter for the given chaos event type.
func (a *StepAccumulator) RecordChaosEvent(eventType ChaosEvent) {
	if counter, ok := a.chaosEvents[eventType]; ok {
		counter.Add(1)
	}
}

// ThroughputBucket represents moves in a fixed time window.
type ThroughputBucket struct {
	OffsetSec float64
	Moves     int64
}

// Snapshot returns a point-in-time copy of all metrics for reporting.
//
//nolint:cyclop,funlen // sequential map iteration snapshot
func (a *StepAccumulator) Snapshot() *Snapshot {
	a.mu.Lock()
	defer a.mu.Unlock()

	rest := make(map[string]HistogramSnapshot, len(a.restLatency))
	for name, h := range a.restLatency {
		rest[name] = snapshotHist(h)
	}

	phaseLatency := make(map[string]HistogramSnapshot, len(a.phaseLatency))
	for name, h := range a.phaseLatency {
		phaseLatency[name] = snapshotHist(h)
	}

	phaseEntries := make(map[string]int64, len(a.phaseEntries))
	for name, counter := range a.phaseEntries {
		phaseEntries[string(name)] = counter.Load()
	}

	phaseMoves := make(map[string]int64, len(a.phaseMoves))
	for name, counter := range a.phaseMoves {
		phaseMoves[string(name)] = counter.Load()
	}

	httpStatusCounts := make(map[int]int64, len(a.httpStatusCounts))
	for code, counter := range a.httpStatusCounts {
		httpStatusCounts[code] = counter.Load()
	}

	errorBreakdown := make(map[string]int64, len(a.errorCounts))
	for name, counter := range a.errorCounts {
		errorBreakdown[string(name)] = counter.Load()
	}

	chaosEvents := make(map[string]int64, len(a.chaosEvents))
	for name, counter := range a.chaosEvents {
		v := counter.Load()
		if v > 0 {
			chaosEvents[string(name)] = v
		}
	}

	// Only include non-zero throughput buckets.
	var throughputBuckets []ThroughputBucket
	for i := range a.moveBuckets {
		moves := a.moveBuckets[i].Load()
		if moves > 0 {
			throughputBuckets = append(throughputBuckets, ThroughputBucket{
				OffsetSec: float64(i) * throughputBucketSize.Seconds(),
				Moves:     moves,
			})
		}
	}

	return &Snapshot{
		RESTLatency:            rest,
		WSDelivery:             snapshotHist(a.wsDelivery),
		E2EMove:                snapshotHist(a.e2eMove),
		TotalMoves:             a.totalMoves.Load(),
		TotalErrors:            a.totalErrors.Load(),
		GamesCompleted:         a.gamesCompleted.Load(),
		GamesTimedOut:          a.gamesTimedOut.Load(),
		GamesFatal:             a.gamesFatal.Load(),
		GamesCancelled:         a.gamesCancelled.Load(),
		TotalRetries:           a.totalRetries.Load(),
		TotalConflicts:         a.totalConflicts.Load(),
		TotalReconnects:        a.totalReconnects.Load(),
		TotalReconnectFailures: a.totalReconnectFailures.Load(),
		PhaseLatency:           phaseLatency,
		PhaseEntries:           phaseEntries,
		PhaseMoves:             phaseMoves,
		HTTPStatusCounts:       httpStatusCounts,
		ErrorBreakdown:         errorBreakdown,
		ChaosEvents:            chaosEvents,
		ThroughputBuckets:      throughputBuckets,
		WarmUpComplete:         a.isWarmUpDone(),
	}
}

// Snapshot is a point-in-time copy of accumulator metrics.
type Snapshot struct {
	RESTLatency    map[string]HistogramSnapshot
	WSDelivery     HistogramSnapshot
	E2EMove        HistogramSnapshot
	TotalMoves     int64
	TotalErrors    int64
	GamesCompleted int64
	GamesTimedOut  int64
	GamesFatal     int64
	GamesCancelled int64

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

	// Warm-up status.
	WarmUpComplete bool
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

package master

import (
	"fmt"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// runSummary holds aggregated statistics for one "run" (a distinct combination
// of cache algorithm + node selector + workload).
// ─────────────────────────────────────────────────────────────────────────────

type runSummary struct {
	RunID        string             `json:"runId"`
	CacheAlgo    string             `json:"cacheAlgo"`
	SelectorAlgo string             `json:"selectorAlgo"`
	WorkloadID   string             `json:"workloadId"`
	Concurrency  int                `json:"concurrency"`
	StartedAt    time.Time          `json:"startedAt"`

	// Request counts & latency
	TotalRequests  int64   `json:"totalRequests"`
	SuccessCount   int64   `json:"successCount"`
	FailureCount   int64   `json:"failureCount"`
	TotalLatencyMs float64 `json:"totalLatencyMs"`
	AvgLatencyMs   float64 `json:"avgLatencyMs"`

	// Per-operation breakdown
	OpCounts   map[string]int64   `json:"opCounts"`
	OpLatency  map[string]float64 `json:"opLatency"`

	// Cache
	CacheHits   int64   `json:"cacheHits"`
	CacheMisses int64   `json:"cacheMisses"`
	CacheHitPct float64 `json:"cacheHitPct"`

	// Throughput
	TotalBytesTransferred int64   `json:"totalBytesTransferred"`
	ThroughputBps         float64 `json:"throughputBps"`

	// Node load
	NodeRequestCounts map[string]int64 `json:"nodeRequestCounts"`
	NodeBytesServed   map[string]int64 `json:"nodeBytesServed"`
}

// ─────────────────────────────────────────────────────────────────────────────
// performanceTracker is the concrete metrics implementation.
// ─────────────────────────────────────────────────────────────────────────────

type performanceTracker struct {
	mu      sync.Mutex
	current runSummary
	history []runSummary
}

// Metrics is the package-level singleton used everywhere in the master package.
var Metrics = &performanceTracker{}

// ── Init / Reset ──────────────────────────────────────────────────────────────

// Init initialises the first run. Called once at startup.
func (p *performanceTracker) Init(cacheAlgo, selectorAlgo string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = newRun(cacheAlgo, selectorAlgo, "default", 1)
}

// StartNewRun archives the current run and begins a fresh one.
func (p *performanceTracker) StartNewRun(cacheAlgo, selectorAlgo, workloadID string, concurrency int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.finalise(&p.current)
	p.history = append(p.history, p.current)
	p.current = newRun(cacheAlgo, selectorAlgo, workloadID, concurrency)
}

// Reset archives the current run and starts fresh (used by /metrics/reset).
func (p *performanceTracker) Reset(cacheAlgo, selectorAlgo, workloadID string, concurrency int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.finalise(&p.current)
	p.history = append(p.history, p.current)
	p.current = newRun(cacheAlgo, selectorAlgo, workloadID, concurrency)
}

// ── Recording ─────────────────────────────────────────────────────────────────

// RecordRequest records one completed request.
func (p *performanceTracker) RecordRequest(op string, success bool, elapsed time.Duration, bytes int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ms := float64(elapsed.Milliseconds())
	p.current.TotalRequests++
	p.current.TotalLatencyMs += ms
	p.current.TotalBytesTransferred += int64(bytes)

	if success {
		p.current.SuccessCount++
	} else {
		p.current.FailureCount++
	}

	p.current.OpCounts[op]++
	p.current.OpLatency[op] += ms
}

// RecordCacheResult records one cache lookup outcome.
func (p *performanceTracker) RecordCacheResult(hit bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if hit {
		p.current.CacheHits++
	} else {
		p.current.CacheMisses++
	}
}

// RecordNodeRequest records bytes served by a specific node port.
func (p *performanceTracker) RecordNodeRequest(port string, bytes int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current.NodeRequestCounts[port]++
	p.current.NodeBytesServed[port] += int64(bytes)
}

// ── Snapshot ─────────────────────────────────────────────────────────────────

// Snapshot returns a copy of the current run (with derived stats computed)
// and a copy of the history slice.
func (p *performanceTracker) Snapshot() (runSummary, []runSummary) {
	p.mu.Lock()
	defer p.mu.Unlock()

	snap := p.current
	p.finalise(&snap)

	hist := make([]runSummary, len(p.history))
	copy(hist, p.history)
	return snap, hist
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func newRun(cacheAlgo, selectorAlgo, workloadID string, concurrency int) runSummary {
	return runSummary{
		RunID:             fmt.Sprintf("%s-%s-%d", cacheAlgo, selectorAlgo, time.Now().UnixMilli()),
		CacheAlgo:         cacheAlgo,
		SelectorAlgo:      selectorAlgo,
		WorkloadID:        workloadID,
		Concurrency:       concurrency,
		StartedAt:         time.Now(),
		OpCounts:          make(map[string]int64),
		OpLatency:         make(map[string]float64),
		NodeRequestCounts: make(map[string]int64),
		NodeBytesServed:   make(map[string]int64),
	}
}

// finalise recomputes derived averages / percentages in-place.
func (p *performanceTracker) finalise(r *runSummary) {
	if r.TotalRequests > 0 {
		r.AvgLatencyMs = r.TotalLatencyMs / float64(r.TotalRequests)
	}

	total := r.CacheHits + r.CacheMisses
	if total > 0 {
		r.CacheHitPct = float64(r.CacheHits) / float64(total) * 100
	}

	elapsed := time.Since(r.StartedAt).Seconds()
	if elapsed > 0 {
		r.ThroughputBps = float64(r.TotalBytesTransferred) / elapsed
	}
}

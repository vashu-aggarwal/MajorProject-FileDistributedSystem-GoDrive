package master

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type nodeMetric struct {
	NodeID      string `json:"nodeId"`
	Requests    int    `json:"requests"`
	BytesServed int64  `json:"bytesServed"`
}

type runSummary struct {
	RunID            string         `json:"runId"`
	CacheAlgorithm   string         `json:"cacheAlgorithm"`
	NodeSelectorAlgo string         `json:"nodeSelectorAlgo"`
	WorkloadID       string         `json:"workloadId"`
	Concurrency      int            `json:"concurrency"`
	StartedAt        string         `json:"startedAt"`
	EndedAt          string         `json:"endedAt,omitempty"`
	DurationSeconds  float64        `json:"durationSeconds"`
	TotalRequests    int            `json:"totalRequests"`
	Successful       int            `json:"successful"`
	Failed           int            `json:"failed"`
	CacheHits        int            `json:"cacheHits"`
	CacheMisses      int            `json:"cacheMisses"`
	CacheHitRatio    float64        `json:"cacheHitRatio"`
	AvgLatencyMs     float64        `json:"avgLatencyMs"`
	P95LatencyMs     float64        `json:"p95LatencyMs"`
	ThroughputReqSec float64        `json:"throughputReqSec"`
	ThroughputMBSec  float64        `json:"throughputMBSec"`
	OperationCount   map[string]int `json:"operationCount"`
	NodeMetrics      []nodeMetric   `json:"nodeMetrics"`
}

type runMetrics struct {
	runID            string
	cacheAlgorithm   string
	nodeSelectorAlgo string
	workloadID       string
	concurrency      int
	startedAt        time.Time
	endedAt          time.Time

	totalRequests int
	successful    int
	failed        int
	cacheHits     int
	cacheMisses   int
	totalBytes    int64
	latencyMs     []float64

	operationCount map[string]int
	nodeRequests   map[string]int
	nodeBytes      map[string]int64
}

type MetricsManager struct {
	mu      sync.Mutex
	seq     uint64
	current *runMetrics
	history []runSummary
}

var Metrics = &MetricsManager{}

func (m *MetricsManager) startRun(cacheAlgorithm string, nodeSelectorAlgo string, workloadID string, concurrency int) {
	runNumber := atomic.AddUint64(&m.seq, 1)
	m.current = &runMetrics{
		runID:            fmt.Sprintf("run-%04d", runNumber),
		cacheAlgorithm:   cacheAlgorithm,
		nodeSelectorAlgo: nodeSelectorAlgo,
		workloadID:       workloadID,
		concurrency:      concurrency,
		startedAt:        time.Now(),
		operationCount: map[string]int{
			"upload":   0,
			"download": 0,
			"update":   0,
			"delete":   0,
		},
		nodeRequests: make(map[string]int),
		nodeBytes:    make(map[string]int64),
	}
}

func (m *MetricsManager) Init(cacheAlgorithm string, nodeSelectorAlgo string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = nil
	m.startRun(cacheAlgorithm, nodeSelectorAlgo, "default", 1)
}

func (m *MetricsManager) StartNewRun(cacheAlgorithm string, nodeSelectorAlgo string, workloadID string, concurrency int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current != nil {
		m.current.endedAt = time.Now()
		m.history = append(m.history, toSummary(m.current, false))
	}
	m.startRun(cacheAlgorithm, nodeSelectorAlgo, workloadID, concurrency)
}

func (m *MetricsManager) Reset(cacheAlgorithm string, nodeSelectorAlgo string, workloadID string, concurrency int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = nil
	m.startRun(cacheAlgorithm, nodeSelectorAlgo, workloadID, concurrency)
}

func (m *MetricsManager) RecordRequest(op string, success bool, latency time.Duration, bytesTransferred int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil {
		return
	}
	m.current.totalRequests++
	if success {
		m.current.successful++
	} else {
		m.current.failed++
	}
	if _, ok := m.current.operationCount[op]; !ok {
		m.current.operationCount[op] = 0
	}
	m.current.operationCount[op]++
	m.current.totalBytes += int64(bytesTransferred)
	m.current.latencyMs = append(m.current.latencyMs, float64(latency.Milliseconds()))
}

func (m *MetricsManager) RecordCacheResult(hit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil {
		return
	}
	if hit {
		m.current.cacheHits++
		return
	}
	m.current.cacheMisses++
}

func (m *MetricsManager) RecordNodeRequest(nodeID string, bytes int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil || nodeID == "" {
		return
	}
	m.current.nodeRequests[nodeID]++
	m.current.nodeBytes[nodeID] += int64(bytes)
}

func (m *MetricsManager) Snapshot() (runSummary, []runSummary) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentSummary := runSummary{}
	if m.current != nil {
		currentSummary = toSummary(m.current, true)
	}

	historyCopy := make([]runSummary, len(m.history))
	copy(historyCopy, m.history)
	return currentSummary, historyCopy
}

func toSummary(run *runMetrics, isCurrent bool) runSummary {
	endedAt := run.endedAt
	if isCurrent {
		endedAt = time.Now()
	}
	duration := endedAt.Sub(run.startedAt).Seconds()
	if duration <= 0 {
		duration = 1
	}

	latencies := make([]float64, len(run.latencyMs))
	copy(latencies, run.latencyMs)
	sort.Float64s(latencies)

	avgLatency := 0.0
	if len(latencies) > 0 {
		sum := 0.0
		for _, l := range latencies {
			sum += l
		}
		avgLatency = sum / float64(len(latencies))
	}

	p95 := 0.0
	if len(latencies) > 0 {
		p95Index := int(0.95 * float64(len(latencies)-1))
		p95 = latencies[p95Index]
	}

	hitRatio := 0.0
	totalCacheAccess := run.cacheHits + run.cacheMisses
	if totalCacheAccess > 0 {
		hitRatio = float64(run.cacheHits) / float64(totalCacheAccess)
	}

	nodes := make([]nodeMetric, 0, len(run.nodeRequests))
	for nodeID, requests := range run.nodeRequests {
		nodes = append(nodes, nodeMetric{
			NodeID:      nodeID,
			Requests:    requests,
			BytesServed: run.nodeBytes[nodeID],
		})
	}
	sort.Slice(nodes, func(i int, j int) bool {
		return nodes[i].NodeID < nodes[j].NodeID
	})

	opCounts := make(map[string]int, len(run.operationCount))
	for k, v := range run.operationCount {
		opCounts[k] = v
	}

	result := runSummary{
		RunID:            run.runID,
		CacheAlgorithm:   run.cacheAlgorithm,
		NodeSelectorAlgo: run.nodeSelectorAlgo,
		WorkloadID:       run.workloadID,
		Concurrency:      run.concurrency,
		StartedAt:        run.startedAt.Format(time.RFC3339),
		DurationSeconds:  duration,
		TotalRequests:    run.totalRequests,
		Successful:       run.successful,
		Failed:           run.failed,
		CacheHits:        run.cacheHits,
		CacheMisses:      run.cacheMisses,
		CacheHitRatio:    hitRatio,
		AvgLatencyMs:     avgLatency,
		P95LatencyMs:     p95,
		ThroughputReqSec: float64(run.totalRequests) / duration,
		ThroughputMBSec:  (float64(run.totalBytes) / 1024.0 / 1024.0) / duration,
		OperationCount:   opCounts,
		NodeMetrics:      nodes,
	}

	if !run.endedAt.IsZero() {
		result.EndedAt = run.endedAt.Format(time.RFC3339)
	}

	return result
}

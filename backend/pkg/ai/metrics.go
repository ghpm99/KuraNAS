package ai

import (
	"sync/atomic"
	"time"
)

// usageCounters accumulates process-wide AI request metrics. It is intentionally
// in-memory (reset on restart) and lock-free: the dashboard wants a cheap live
// pulse of AI activity, not durable accounting. Durable per-call detail lives in
// the forensic file log.
type usageCounters struct {
	total          atomic.Uint64
	success        atomic.Uint64
	failure        atomic.Uint64
	totalLatencyMs atomic.Uint64
	totalTokens    atomic.Uint64
}

var usage usageCounters

// UsageSnapshot is an immutable view of the AI usage counters at one instant.
type UsageSnapshot struct {
	Total        uint64 `json:"total"`
	Success      uint64 `json:"success"`
	Failure      uint64 `json:"failure"`
	TotalTokens  uint64 `json:"total_tokens"`
	AvgLatencyMs int64  `json:"avg_latency_ms"`
}

// recordUsage tallies one logical AI request (one Execute call), regardless of
// how many provider attempts it took. latency is the wall-clock time of the
// whole request; tokens is best-effort (0 when the provider does not report it).
func recordUsage(success bool, latency time.Duration, tokens int) {
	usage.total.Add(1)
	if success {
		usage.success.Add(1)
	} else {
		usage.failure.Add(1)
	}
	if latency > 0 {
		usage.totalLatencyMs.Add(uint64(latency.Milliseconds()))
	}
	if tokens > 0 {
		usage.totalTokens.Add(uint64(tokens))
	}
}

// UsageMetrics returns a snapshot of the AI usage counters, with average latency
// derived over all recorded requests.
func UsageMetrics() UsageSnapshot {
	total := usage.total.Load()
	var avg int64
	if total > 0 {
		avg = int64(usage.totalLatencyMs.Load() / total)
	}
	return UsageSnapshot{
		Total:        total,
		Success:      usage.success.Load(),
		Failure:      usage.failure.Load(),
		TotalTokens:  usage.totalTokens.Load(),
		AvgLatencyMs: avg,
	}
}

// ResetUsageMetrics zeroes the counters. Intended for tests.
func ResetUsageMetrics() {
	usage.total.Store(0)
	usage.success.Store(0)
	usage.failure.Store(0)
	usage.totalLatencyMs.Store(0)
	usage.totalTokens.Store(0)
}

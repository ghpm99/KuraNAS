package ai

import (
	"testing"
	"time"
)

func TestUsageMetricsAccumulates(t *testing.T) {
	ResetUsageMetrics()
	t.Cleanup(ResetUsageMetrics)

	recordUsage(true, 100*time.Millisecond, 30)
	recordUsage(true, 300*time.Millisecond, 10)
	recordUsage(false, 200*time.Millisecond, 0)

	got := UsageMetrics()
	if got.Total != 3 {
		t.Fatalf("expected total 3, got %d", got.Total)
	}
	if got.Success != 2 || got.Failure != 1 {
		t.Fatalf("expected 2 success / 1 failure, got %d/%d", got.Success, got.Failure)
	}
	if got.TotalTokens != 40 {
		t.Fatalf("expected 40 tokens, got %d", got.TotalTokens)
	}
	// (100+300+200)/3 = 200
	if got.AvgLatencyMs != 200 {
		t.Fatalf("expected avg latency 200ms, got %d", got.AvgLatencyMs)
	}
}

func TestUsageMetricsEmptyHasZeroAverage(t *testing.T) {
	ResetUsageMetrics()
	t.Cleanup(ResetUsageMetrics)

	got := UsageMetrics()
	if got.Total != 0 || got.AvgLatencyMs != 0 {
		t.Fatalf("expected zeroed metrics, got %+v", got)
	}
}

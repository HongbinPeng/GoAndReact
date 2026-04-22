package main

import (
	"testing"
	"time"
)

func TestBuildSummary(t *testing.T) {
	results := []ProbeResult{
		{Name: "A", Success: true, Latency: 50 * time.Millisecond},
		{Name: "B", Success: false, Latency: 300 * time.Millisecond},
		{Name: "C", Success: true, Latency: 700 * time.Millisecond},
		{Name: "D", Success: false, Latency: 2 * time.Second},
	}

	summary := buildSummary(results)

	if summary.Total != 4 {
		t.Fatalf("Total = %d，期望 4", summary.Total)
	}
	if summary.SuccessCount != 2 {
		t.Fatalf("SuccessCount = %d，期望 2", summary.SuccessCount)
	}
	if summary.FailureCount != 2 {
		t.Fatalf("FailureCount = %d，期望 2", summary.FailureCount)
	}
	if summary.LatencyBuckets["<100ms"] != 1 {
		t.Fatalf("<100ms 桶数量 = %d，期望 1", summary.LatencyBuckets["<100ms"])
	}
	if summary.LatencyBuckets["100ms~500ms"] != 1 {
		t.Fatalf("100ms~500ms 桶数量 = %d，期望 1", summary.LatencyBuckets["100ms~500ms"])
	}
	if summary.LatencyBuckets["500ms~1s"] != 1 {
		t.Fatalf("500ms~1s 桶数量 = %d，期望 1", summary.LatencyBuckets["500ms~1s"])
	}
	if summary.LatencyBuckets[">1s"] != 1 {
		t.Fatalf(">1s 桶数量 = %d，期望 1", summary.LatencyBuckets[">1s"])
	}
	if len(summary.Slowest) == 0 || summary.Slowest[0].Name != "D" {
		t.Fatalf("最慢服务计算错误，实际结果：%v", summary.Slowest)
	}
}

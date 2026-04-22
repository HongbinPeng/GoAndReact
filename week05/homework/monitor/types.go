package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Config 表示整个配置文件。
type Config struct {
	Targets []Target `json:"targets"`
}

// Expectation 表示某个探测目标的预期结果。
type Expectation struct {
	StatusCode *int   `json:"status_code,omitempty"`
	Contains   string `json:"contains,omitempty"`
	Connected  *bool  `json:"connected,omitempty"`
}

// Target 表示一个探测目标。
type Target struct {
	Name       string      `json:"name"`
	Protocol   string      `json:"protocol"`
	Address    string      `json:"address"`
	Expect     Expectation `json:"expect"`
	RetryCount int         `json:"retry_count,omitempty"`
	Index      int         `json:"-"`
}

// Options 表示命令行参数解析后的结果。
type Options struct {
	ConfigPath string
	Timeout    time.Duration
	Verbose    bool
}

// ProbeResult 表示单个目标的最终探测结果。
type ProbeResult struct {
	Index    int
	Name     string
	Protocol string
	Address  string

	Success  bool
	Expected string
	Observed string
	Error    string

	Latency  time.Duration
	Attempts int
}

// Summary 表示最终汇总统计。
type Summary struct {
	Total          int
	SuccessCount   int
	FailureCount   int
	SuccessRate    float64
	FailureRate    float64
	AverageLatency time.Duration
	LatencyBuckets map[string]int
	Slowest        []ProbeResult
}

func (t Target) expectedDescription() string {
	switch t.Protocol {
	case "tcp":
		return "Connected"
	case "http":
		parts := make([]string, 0, 2)
		if t.Expect.StatusCode != nil {
			parts = append(parts, fmt.Sprintf("%d %s", *t.Expect.StatusCode, http.StatusText(*t.Expect.StatusCode)))
		} else {
			parts = append(parts, "2xx status")
		}
		if t.Expect.Contains != "" {
			parts = append(parts, fmt.Sprintf("Contains %q", t.Expect.Contains))
		}
		return strings.Join(parts, " + ")
	default:
		return "N/A"
	}
}

func (r ProbeResult) statusLabel() string {
	if r.Success {
		return "OK"
	}
	return "FAIL"
}

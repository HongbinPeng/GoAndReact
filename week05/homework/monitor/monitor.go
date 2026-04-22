package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

const maxHTTPBodyBytes int64 = 1 << 20

var artificialTaskDelay = 1 * time.Second

type Monitor struct {
	timeout    time.Duration
	verbose    bool
	logger     *log.Logger
	taskDelay  time.Duration
	httpClient *httpClient
}

type httpClient struct {
	do func(*httpRequest) (*httpResponse, error)
}

type httpRequest struct {
	ctx     context.Context
	method  string
	url     string
	headers map[string]string
}

type httpResponse struct {
	status     string
	statusCode int
	body       []byte
	close      func() error
}

func NewMonitor(timeout time.Duration, verbose bool) *Monitor {
	return &Monitor{
		timeout:    timeout,
		verbose:    verbose,
		logger:     log.New(os.Stdout, "[verbose] ", log.LstdFlags),
		taskDelay:  artificialTaskDelay,
		httpClient: newHTTPClient(timeout),
	}
}

func (m *Monitor) Run(targets []Target) []ProbeResult {
	resultsCh := make(chan ProbeResult, len(targets))
	var wg sync.WaitGroup

	for _, target := range targets {
		target := target
		wg.Add(1)

		go func() {
			defer wg.Done()
			resultsCh <- m.runTarget(target)
		}()
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	results := make([]ProbeResult, 0, len(targets))
	for result := range resultsCh {
		results = append(results, result)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	return results
}

func (m *Monitor) runTarget(target Target) ProbeResult {
	m.logf("开始探测 [%s] 协议=%s 地址=%s", target.Name, target.Protocol, target.Address)

	// 按作业要求，为每个探测任务人为加入 1 秒延迟。
	// 由于所有任务是并发启动的，这个延迟不会让总耗时线性叠加。
	if m.taskDelay > 0 {
		time.Sleep(m.taskDelay)
	}

	result := m.executeWithRetry(target, m.probeOnce)

	if result.Success {
		m.logf("探测完成 [%s] 成功 耗时=%v 尝试次数=%d 实际结果=%s", result.Name, result.Latency, result.Attempts, result.Observed)
	} else {
		m.logf("探测完成 [%s] 失败 耗时=%v 尝试次数=%d 错误=%s", result.Name, result.Latency, result.Attempts, result.Error)
	}

	return result
}

func (m *Monitor) executeWithRetry(target Target, probe func(Target) ProbeResult) ProbeResult {
	totalAttempts := target.RetryCount + 1
	var lastResult ProbeResult
	var totalLatency time.Duration

	for attempt := 1; attempt <= totalAttempts; attempt++ {
		if attempt > 1 {
			m.logf("目标 [%s] 开始第 %d 次重试", target.Name, attempt-1)
		}

		result := probe(target)
		result.Attempts = attempt
		totalLatency += result.Latency
		lastResult = result

		if result.Success {
			result.Latency = totalLatency
			return result
		}

		if attempt < totalAttempts {
			m.logf("目标 [%s] 第 %d 次探测失败，准备重试，原因：%s", target.Name, attempt, result.Error)
		}
	}

	lastResult.Latency = totalLatency
	return lastResult
}

func (m *Monitor) logf(format string, args ...any) {
	if m.verbose {
		m.logger.Printf(format, args...)
	}
}

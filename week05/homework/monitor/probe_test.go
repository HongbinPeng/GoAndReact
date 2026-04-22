package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeHTTPSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "Go monitor is healthy")
	}))
	defer server.Close()

	monitor := NewMonitor(2*time.Second, false)
	monitor.taskDelay = 0

	expectCode := http.StatusOK
	target := Target{
		Name:     "测试 HTTP 成功",
		Protocol: "http",
		Address:  server.URL,
		Expect: Expectation{
			StatusCode: &expectCode,
			Contains:   "Go",
		},
	}

	result := monitor.probeHTTP(target)
	if !result.Success {
		t.Fatalf("probeHTTP 应该成功，实际失败：%s", result.Error)
	}
}

func TestProbeHTTPContainsFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "hello world")
	}))
	defer server.Close()

	monitor := NewMonitor(2*time.Second, false)
	monitor.taskDelay = 0

	target := Target{
		Name:     "测试 HTTP 内容失败",
		Protocol: "http",
		Address:  server.URL,
		Expect: Expectation{
			Contains: "Go",
		},
	}

	result := monitor.probeHTTP(target)
	if result.Success {
		t.Fatalf("probeHTTP 应该失败，但实际成功")
	}
}

func TestProbeTCPSuccess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("启动 TCP 测试服务失败：%v", err)
	}
	defer listener.Close()

	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
	}()

	monitor := NewMonitor(2*time.Second, false)
	monitor.taskDelay = 0

	target := Target{
		Name:     "测试 TCP 成功",
		Protocol: "tcp",
		Address:  listener.Addr().String(),
	}

	result := monitor.probeTCP(target)
	if !result.Success {
		t.Fatalf("probeTCP 应该成功，实际失败：%s", result.Error)
	}
}

func TestExecuteWithRetrySucceedsOnSecondAttempt(t *testing.T) {
	monitor := NewMonitor(2*time.Second, false)
	monitor.taskDelay = 0

	attempts := 0
	target := Target{
		Name:       "测试重试",
		Protocol:   "http",
		Address:    "https://example.com",
		RetryCount: 2,
	}

	result := monitor.executeWithRetry(target, func(target Target) ProbeResult {
		attempts++
		if attempts < 2 {
			return ProbeResult{
				Name:     target.Name,
				Protocol: target.Protocol,
				Address:  target.Address,
				Expected: target.expectedDescription(),
				Observed: "500 Internal Server Error",
				Error:    "第一次模拟失败",
				Latency:  100 * time.Millisecond,
			}
		}
		return ProbeResult{
			Name:     target.Name,
			Protocol: target.Protocol,
			Address:  target.Address,
			Expected: target.expectedDescription(),
			Observed: "200 OK",
			Success:  true,
			Latency:  50 * time.Millisecond,
		}
	})

	if !result.Success {
		t.Fatalf("重试后应该成功，实际失败：%s", result.Error)
	}
	if attempts != 2 {
		t.Fatalf("实际尝试次数 = %d，期望 2", attempts)
	}
	if result.Attempts != 2 {
		t.Fatalf("结果中的 Attempts = %d，期望 2", result.Attempts)
	}
	if result.Latency != 150*time.Millisecond {
		t.Fatalf("累计耗时 = %v，期望 %v", result.Latency, 150*time.Millisecond)
	}
}

func TestCustomHTTPClientTimeoutFailure(t *testing.T) {
	monitor := NewMonitor(100*time.Millisecond, false)
	monitor.taskDelay = 0
	monitor.httpClient = &httpClient{
		do: func(req *httpRequest) (*httpResponse, error) {
			select {
			case <-time.After(300 * time.Millisecond):
				return &httpResponse{
					status:     "200 OK",
					statusCode: 200,
					body:       []byte("ok"),
					close:      func() error { return nil },
				}, nil
			case <-req.ctx.Done():
				return nil, req.ctx.Err()
			}
		},
	}

	target := Target{
		Name:     "超时 HTTP",
		Protocol: "http",
		Address:  "https://slow.example.com",
	}

	result := monitor.probeHTTP(target)
	if result.Success {
		t.Fatalf("超时请求不应该成功")
	}
}

func TestHTTPClientRequestCarriesContext(t *testing.T) {
	monitor := NewMonitor(2*time.Second, false)
	monitor.taskDelay = 0

	monitor.httpClient = &httpClient{
		do: func(req *httpRequest) (*httpResponse, error) {
			if req.ctx == nil {
				t.Fatalf("httpRequest 缺少 context")
			}
			if req.method != http.MethodGet {
				t.Fatalf("HTTP 方法 = %s，期望 GET", req.method)
			}
			select {
			case <-req.ctx.Done():
				return nil, req.ctx.Err()
			default:
			}
			return &httpResponse{
				status:     "200 OK",
				statusCode: 200,
				body:       []byte("Go"),
				close:      func() error { return nil },
			}, nil
		},
	}

	expectCode := 200
	target := Target{
		Name:     "上下文注入测试",
		Protocol: "http",
		Address:  "https://example.com",
		Expect: Expectation{
			StatusCode: &expectCode,
			Contains:   "Go",
		},
	}

	result := monitor.probeHTTP(target)
	if !result.Success {
		t.Fatalf("probeHTTP 应该成功，实际失败：%s", result.Error)
	}
}

func TestProbeTCPFailure(t *testing.T) {
	monitor := NewMonitor(50*time.Millisecond, false)
	monitor.taskDelay = 0

	target := Target{
		Name:     "不可连接 TCP",
		Protocol: "tcp",
		Address:  "127.0.0.1:1",
	}

	result := monitor.probeTCP(target)
	if result.Success {
		t.Fatalf("不可连接 TCP 不应该成功")
	}
}

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func newHTTPClient(timeout time.Duration) *httpClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = dialer.DialContext
	transport.ResponseHeaderTimeout = timeout
	transport.TLSHandshakeTimeout = timeout

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &httpClient{
		do: func(req *httpRequest) (*httpResponse, error) {
			httpReq, err := http.NewRequestWithContext(req.ctx, req.method, req.url, nil)
			if err != nil {
				return nil, err
			}
			for key, value := range req.headers {
				httpReq.Header.Set(key, value)
			}

			resp, err := client.Do(httpReq)
			if err != nil {
				return nil, err
			}

			body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBodyBytes))
			if readErr != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("读取 HTTP 响应体失败: %w", readErr)
			}

			return &httpResponse{
				status:     resp.Status,
				statusCode: resp.StatusCode,
				body:       body,
				close:      resp.Body.Close,
			}, nil
		},
	}
}

func (m *Monitor) probeOnce(target Target) ProbeResult {
	switch target.Protocol {
	case "http":
		return m.probeHTTP(target)
	case "tcp":
		return m.probeTCP(target)
	default:
		return ProbeResult{
			Index:    target.Index,
			Name:     target.Name,
			Protocol: target.Protocol,
			Address:  target.Address,
			Expected: target.expectedDescription(),
			Observed: "Unsupported protocol",
			Error:    fmt.Sprintf("不支持的 protocol: %s", target.Protocol),
		}
	}
}

func (m *Monitor) probeHTTP(target Target) ProbeResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	response, err := m.httpClient.do(&httpRequest{
		ctx:    ctx,
		method: http.MethodGet,
		url:    target.Address,
		headers: map[string]string{
			"User-Agent": "monitor/1.0",
			"Accept":     "*/*",
		},
	})
	latency := time.Since(start)

	result := ProbeResult{
		Index:    target.Index,
		Name:     target.Name,
		Protocol: target.Protocol,
		Address:  target.Address,
		Expected: target.expectedDescription(),
		Latency:  latency,
	}

	if err != nil {
		result.Observed = "Request failed"
		result.Error = fmt.Sprintf("HTTP 请求失败: %v", err)
		return result
	}
	defer response.close()

	checkErrors := make([]string, 0, 2)

	if target.Expect.StatusCode != nil {
		if response.statusCode != *target.Expect.StatusCode {
			checkErrors = append(checkErrors, fmt.Sprintf("期望状态码 %d，实际为 %d", *target.Expect.StatusCode, response.statusCode))
		}
	} else if response.statusCode < http.StatusOK || response.statusCode >= http.StatusMultipleChoices {
		checkErrors = append(checkErrors, fmt.Sprintf("HTTP 状态码不是 2xx，实际为 %s", response.status))
	}

	containsMatched := true
	if target.Expect.Contains != "" {
		containsMatched = bytes.Contains(response.body, []byte(target.Expect.Contains))
		if !containsMatched {
			checkErrors = append(checkErrors, fmt.Sprintf("响应体不包含 %q", target.Expect.Contains))
		}
	}

	observedParts := []string{response.status}
	if target.Expect.Contains != "" {
		observedParts = append(observedParts, fmt.Sprintf("Contains %q=%t", target.Expect.Contains, containsMatched))
	}
	result.Observed = strings.Join(observedParts, ", ")

	if len(checkErrors) == 0 {
		result.Success = true
		return result
	}

	result.Error = strings.Join(checkErrors, "; ")
	return result
}

func (m *Monitor) probeTCP(target Target) ProbeResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	dialer := &net.Dialer{Timeout: m.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", target.Address)
	latency := time.Since(start)

	result := ProbeResult{
		Index:    target.Index,
		Name:     target.Name,
		Protocol: target.Protocol,
		Address:  target.Address,
		Expected: target.expectedDescription(),
		Latency:  latency,
	}

	if err != nil {
		result.Observed = "Disconnected"
		result.Error = fmt.Sprintf("TCP 连接失败: %v", err)
		return result
	}
	defer conn.Close()

	result.Success = true
	result.Observed = "Connected"
	return result
}

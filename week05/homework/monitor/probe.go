package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const containsScanChunkSize = 32 * 1024 //设置扫描块大小，32KB，用于读取响应体并检查是否包含指定字符串

// 构建一个带超时，可选代理，可复用连接配置的HTTP客户端
func newHTTPClient(timeout time.Duration, proxy string) *httpClient {
	transport := http.DefaultTransport.(*http.Transport).Clone() //克隆默认传输以避免修改全局状态
	/*
		这里之所以要使用Clone，是因为http.DefaultTransport是一个全局共享的变量，
		如果直接修改它的字段，可能会影响到程序中其他使用http.DefaultTransport的部分，
		导致不可预期的行为。通过Clone方法，我们创建了一个新的Transport实例，
		这样我们就可以安全地修改它的字段，而不会影响到其他地方使用http.DefaultTransport的代码。
	*/
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	} //这里是底层的TCP连接器，用于创建与目标服务器的TCP连接。设置TCP连接的超时时间，保持连接的时间30秒。
	transport.DialContext = dialer.DialContext  //这里设置TCP连接的超时和保持连接的时间
	transport.ResponseHeaderTimeout = timeout   //设置等待服务器响应头的超时时间，不包括TCP连接时间，TLS握手时间
	transport.TLSHandshakeTimeout = timeout     //设置TLS握手超时时间，用于HTTPS连接
	transport.Proxy = http.ProxyFromEnvironment //这里从环境变量中读取代理设置

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL) //生成一个固定的代理函数，始终返回相同的代理URL
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout, //设置请求的总超时，包括TCP连接时间、TLS握手时间、发送请求时间和等待响应时间，读取响应体的时间不包括在内
	} //构造一个HTTP客户端，设置好了超时时间，可选代理

	return &httpClient{
		do: func(req *httpRequest) (*httpResponse, error) {
			httpReq, err := http.NewRequestWithContext(req.ctx, req.method, req.url, nil) //创建真正的请求对象
			if err != nil {
				return nil, err
			}
			for key, value := range req.headers {
				httpReq.Header.Set(key, value)
			}
			resp, err := client.Do(httpReq)
			/*
				这里发送HTTP请求，这里经历：
				代理选择
				TCP 建连
				TLS 握手
				发请求
				收响应头
				如果报错，这里常见原因：
					超时
					DNS 解析失败
					代理不可用
					TLS 握手失败
					连接被拒绝
			*/
			if err != nil {
				return nil, err
			}
			return &httpResponse{
				status:     resp.Status,
				statusCode: resp.StatusCode,
				body:       resp.Body,
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
	if response.body != nil {
		defer response.body.Close()
	}

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
		containsMatched, err = responseBodyContains(response.body, target.Expect.Contains, maxHTTPBodyBytes)
		if err != nil {
			result.Observed = "Request failed"
			result.Error = fmt.Sprintf("读取 HTTP 响应体失败: %v", err)
			return result
		}
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

func responseBodyContains(reader io.Reader, needle string, maxBytes int64) (bool, error) {
	if needle == "" {
		return true, nil
	}

	limitedReader := io.LimitReader(reader, maxBytes)
	bufferedReader := bufio.NewReaderSize(limitedReader, containsScanChunkSize)
	target := []byte(needle)

	chunk := make([]byte, containsScanChunkSize)
	overlapSize := len(target) - 1
	if overlapSize < 0 {
		overlapSize = 0
	}
	tail := make([]byte, 0, overlapSize)
	for {
		n, err := bufferedReader.Read(chunk)
		if n > 0 {
			window := make([]byte, 0, len(tail)+n)
			window = append(window, tail...)
			window = append(window, chunk[:n]...)

			if bytes.Contains(window, target) {
				return true, nil
			}

			if overlapSize > 0 {
				if len(window) > overlapSize {
					tail = append(tail[:0], window[len(window)-overlapSize:]...)
				} else {
					tail = append(tail[:0], window...)
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
	}
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

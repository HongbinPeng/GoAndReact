package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// 运行方式：
// go test -v
//
// 这份示例用于学习 httptest。
// 它的核心价值是：不依赖外网，也能稳定地测试 HTTP 逻辑。

func TestHealthEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Go monitor is healthy"))
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("请求测试服务器失败：%v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("读取响应体失败：%v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d，期望 %d", resp.StatusCode, http.StatusOK)
	}

	if !strings.Contains(string(body), "Go") {
		t.Fatalf("响应体中没有包含 Go，实际内容：%q", string(body))
	}
}

func TestSlowEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(120 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("slow response"))
	}))
	defer server.Close()

	start := time.Now()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("请求慢接口失败：%v", err)
	}
	defer resp.Body.Close()

	if time.Since(start) < 100*time.Millisecond {
		t.Fatalf("慢响应测试没有达到预期的延时效果")
	}
}

package main

// ============================================================
// 第 11 课：httptest — 写 HTTP 客户端的单元测试
// ============================================================
//
// 【为什么用 httptest？】
//   1. 不依赖外网 — 离线也能跑测试
//   2. 完全可控 — 可以模拟各种异常情况
//   3. 速度快 — 本地回环比访问外网快得多
//   4. 可重复 — 每次测试结果一致
//
// 【httptest 的核心】
//   httptest.NewServer(handler) — 启动一个本地临时 HTTP 服务器
//   server.URL                  — 服务器地址，如 "http://127.0.0.1:12345"
//   server.Close()              — 测试结束后关闭服务器
//   httptest.NewRecorder()      — 记录 HTTP 响应，用于测试服务端代码
//
// 【运行方式】
//   go run 11_httptest/main.go
//   或者：go test ./11_httptest/  （作为测试运行）

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

// 定义我们要测试的函数：模拟监控客户端的探测逻辑

// ProbeResult 探测结果
type ProbeResult struct {
	OK       bool   // 是否成功
	StatusCode int  // HTTP 状态码
	Latency  string // 耗时
	Body     string // 响应体内容
	ErrorMsg string // 错误信息
}

// ProbeHTTP 发送 HTTP 请求并返回探测结果
// 这就是你作业中需要实现的函数的简化版
func ProbeHTTP(targetURL string, timeout time.Duration) ProbeResult {
	result := ProbeResult{}
	start := time.Now()

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 创建客户端
	client := &http.Client{Timeout: timeout}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("创建请求失败: %v", err)
		return result
	}

	// 发送请求
	resp, err := client.Do(req)
	result.Latency = time.Since(start).String()

	if err != nil {
		result.ErrorMsg = fmt.Sprintf("请求失败: %v", err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// 检查状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.ErrorMsg = fmt.Sprintf("状态码异常: %d", resp.StatusCode)
		return result
	}

	// 读取响应体（限制最多 1KB）
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		result.ErrorMsg = fmt.Sprintf("读取响应体失败: %v", err)
		return result
	}

	result.Body = string(body)
	result.OK = true
	return result
}

func main() {
	// ================================================================
	// 测试用例 1：正常 200 响应
	// ================================================================

	fmt.Println("===== 测试用例 1：正常 200 响应 =====")

	// 创建一个模拟服务器，返回 200 OK
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 可以验证请求的参数是否正确
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("system is healthy"))
	}))
	defer server1.Close()

	fmt.Println("模拟服务器:", server1.URL)

	result1 := ProbeHTTP(server1.URL, 5*time.Second)
	fmt.Printf("OK: %v, 状态码: %d, 耗时: %s, 响应体: %s, 错误: %s\n",
		result1.OK, result1.StatusCode, result1.Latency, result1.Body, result1.ErrorMsg)

	// 断言：应该成功
	if !result1.OK {
		fmt.Println("❌ 测试失败：期望成功，实际失败")
	} else {
		fmt.Println("✅ 测试通过")
	}

	// ================================================================
	// 测试用例 2：404 Not Found
	// ================================================================

	fmt.Println("\n===== 测试用例 2：404 Not Found =====")

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server2.Close()

	result2 := ProbeHTTP(server2.URL+"/not-found", 5*time.Second)
	fmt.Printf("OK: %v, 状态码: %d, 错误: %s\n",
		result2.OK, result2.StatusCode, result2.ErrorMsg)

	if result2.StatusCode == 404 {
		fmt.Println("✅ 测试通过：正确识别 404")
	}

	// ================================================================
	// 测试用例 3：500 Internal Server Error
	// ================================================================

	fmt.Println("\n===== 测试用例 3：500 Internal Server Error =====")

	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "database connection failed", http.StatusInternalServerError)
	}))
	defer server3.Close()

	result3 := ProbeHTTP(server3.URL, 5*time.Second)
	fmt.Printf("OK: %v, 状态码: %d, 错误: %s\n",
		result3.OK, result3.StatusCode, result3.ErrorMsg)

	if result3.StatusCode == 500 {
		fmt.Println("✅ 测试通过：正确识别 500")
	}

	// ================================================================
	// 测试用例 4：超时
	// ================================================================

	fmt.Println("\n===== 测试用例 4：超时 =====")

	// 模拟一个很慢的服务器（3 秒后才响应）
	server4 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.Write([]byte("slow response"))
	}))
	defer server4.Close()

	// 但客户端只等 100ms
	result4 := ProbeHTTP(server4.URL, 100*time.Millisecond)
	fmt.Printf("OK: %v, 错误: %s, 耗时: %s\n",
		result4.OK, result4.ErrorMsg, result4.Latency)

	if result4.ErrorMsg != "" && result4.Latency != "" {
		fmt.Println("✅ 测试通过：请求被超时取消")
	}

	// ================================================================
	// 测试用例 5：JSON 响应
	// ================================================================

	fmt.Println("\n===== 测试用例 5：JSON 响应 =====")

	server5 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"version": "1.0.0",
		})
	}))
	defer server5.Close()

	result5 := ProbeHTTP(server5.URL, 5*time.Second)
	fmt.Printf("OK: %v, 响应体: %s\n", result5.OK, result5.Body)

	if result5.OK && result5.Body != "" {
		fmt.Println("✅ 测试通过：JSON 响应正确返回")
	}

	// ================================================================
	// 测试用例 6：验证请求头
	// ================================================================

	fmt.Println("\n===== 测试用例 6：验证请求头 =====")

	// 测试客户端是否正确发送了特定的请求头
	var receivedHeaders string
	server6 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Get("X-Custom-Header")
		w.WriteHeader(http.StatusOK)
	}))
	defer server6.Close()

	// 创建一个带自定义请求头的客户端请求
	client6 := &http.Client{Timeout: 5 * time.Second}
	req6, _ := http.NewRequest(http.MethodGet, server6.URL, nil)
	req6.Header.Set("X-Custom-Header", "test-value")
	resp6, _ := client6.Do(req6)
	resp6.Body.Close()

	if receivedHeaders == "test-value" {
		fmt.Println("✅ 测试通过：自定义请求头正确发送")
	} else {
		fmt.Printf("❌ 测试失败：期望 'test-value'，实际 '%s'\n", receivedHeaders)
	}

	// ================================================================
	// 补充：httptest.NewRecorder（测试服务端代码）
	// ================================================================

	fmt.Println("\n===== 补充：httptest.NewRecorder =====")

	// NewRecorder 不需要启动真正的服务器，
	// 它模拟 http.ResponseWriter 的行为，记录写入的响应。
	// 适合测试服务端代码（handler）。

	// 定义一个 handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 1}`))
	})

	// 创建 Recorder
	recorder := httptest.NewRecorder()

	// 创建模拟请求
	req, _ := http.NewRequest(http.MethodPost, "/api/users", nil)

	// 执行 handler（不调用真正的 HTTP 网络）
	handler.ServeHTTP(recorder, req)

	// 检查响应
	fmt.Printf("状态码: %d\n", recorder.Code)
	fmt.Printf("Content-Type: %s\n", recorder.Header().Get("Content-Type"))
	fmt.Printf("响应体: %s\n", recorder.Body.String())

	if recorder.Code == http.StatusCreated {
		fmt.Println("✅ 测试通过")
	}

	fmt.Println("\n===== 总结 =====")
	fmt.Println("httptest.NewServer   → 启动本地服务器，测试客户端")
	fmt.Println("httptest.NewRecorder → 模拟 ResponseWriter，测试服务端")
	fmt.Println("每个测试都独立创建/销毁服务器，互不干扰")
}


package main

// ============================================================
// 第 10 课：错误处理 — 区分网络错误、超时、状态码错误
// ============================================================
//
// 【核心概念】
//   http.Client.Do 返回的 err 有两种可能：
//   1. 网络层错误（err != nil）— 连接失败、超时、DNS 解析失败等
//   2. HTTP 状态码错误（err == nil 但 StatusCode 不是 200）— 404、500 等
//
//   关键：404/500 不算网络错误！err 仍然是 nil！必须手动检查 StatusCode。
//
// 【运行方式】
//   go run 10_error_handling/main.go

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"
)

func main() {
	// 测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		case "/not-found":
			http.NotFound(w, r)
		case "/error":
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("login required"))
		case "/redirect-loop":
			http.Redirect(w, r, "/redirect-loop", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// ================================================================
	// 知识点 1：错误分类
	// ================================================================

	fmt.Println("===== 知识点 1：错误分类 =====")
	fmt.Println("HTTP 请求错误分三种：")
	fmt.Println("  1. 网络错误（err != nil）：超时、连接失败、DNS 失败等")
	fmt.Println("  2. 状态码错误（err == nil + StatusCode 异常）：404、500 等")
	fmt.Println("  3. 响应体读取错误（io.ReadAll 返回 err）：连接中断等")

	// ================================================================
	// 知识点 2：处理网络错误
	// ================================================================

	fmt.Println("\n===== 知识点 2：网络错误 =====")

	// ---- 场景 A：超时 ----
	fmt.Println("\n--- 场景 A：超时 ---")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/ok", nil)

	client := &http.Client{}
	_, err := client.Do(req)
	cancel() // 释放资源

	if err != nil {
		fmt.Println("错误:", err)
		// url.Error 是 http 请求错误的包装类型
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			fmt.Println("  → 这是一个 URL 错误（包装类型）")
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("  → 原因：请求超时了 (context deadline exceeded)")
			}
		}
	}

	// ---- 场景 B：连接被拒绝 ----
	fmt.Println("\n--- 场景 B：连接被拒绝 ---")
	_, err = http.Get("http://127.0.0.1:19999")
	if err != nil {
		fmt.Println("错误:", err)
		var netErr net.Error
		if errors.As(err, &netErr) {
			fmt.Println("  → 这是一个网络错误 (net.Error)")
			if netErr.Timeout() {
				fmt.Println("  → 原因：连接超时")
			} else {
				fmt.Println("  → 原因：连接失败 (connection refused)")
			}
		}
	}

	// ---- 场景 C：重定向次数过多 ----
	fmt.Println("\n--- 场景 C：重定向次数过多 ---")
	_, err = http.Get(server.URL + "/redirect-loop")
	if err != nil {
		fmt.Println("错误:", err)
		if strings.Contains(err.Error(), "stopped after 10 redirects") {
			fmt.Println("  → 原因：重定向超过 10 次")
		}
	}

	// ================================================================
	// 知识点 3：处理 HTTP 状态码错误
	// ================================================================

	fmt.Println("\n===== 知识点 3：HTTP 状态码错误 =====")
	fmt.Println("这些情况 err == nil！必须手动检查 StatusCode！")

	// 404 Not Found
	resp404, _ := http.Get(server.URL + "/not-found")
	defer resp404.Body.Close()
	body404, _ := io.ReadAll(resp404.Body)
	fmt.Printf("404 场景：状态码 %d，响应体: %s（注意 err 是 nil）\n", resp404.StatusCode, string(body404))

	// 500 Internal Server Error
	resp500, _ := http.Get(server.URL + "/error")
	defer resp500.Body.Close()
	body500, _ := io.ReadAll(resp500.Body)
	fmt.Printf("500 场景：状态码 %d，响应体: %s（注意 err 是 nil）\n", resp500.StatusCode, string(body500))

	// 401 Unauthorized
	resp401, _ := http.Get(server.URL + "/unauthorized")
	defer resp401.Body.Close()
	body401, _ := io.ReadAll(resp401.Body)
	fmt.Printf("401 场景：状态码 %d，响应体: %s（注意 err 是 nil）\n", resp401.StatusCode, string(body401))

	// ================================================================
	// 知识点 4：推荐的错误处理模式
	// ================================================================

	fmt.Println("\n===== 知识点 4：推荐的错误处理模式 =====")

	result := doRequest(server.URL + "/ok")
	fmt.Println("正常请求:", result)

	result = doRequest(server.URL + "/not-found")
	fmt.Println("404 请求:", result)

	result = doRequest("http://127.0.0.1:19999/notexist")
	fmt.Println("连接失败:", result)

	fmt.Println("\n===== 总结 =====")
	fmt.Println("err != nil              → 网络层问题（超时、连接失败等）")
	fmt.Println("err == nil + 状态码异常  → HTTP 状态码问题（404、500 等）")
	fmt.Println("io.ReadAll 返回 err     → 响应体读取问题")
	fmt.Println("三种错误都要检查！")
}

// doRequest 演示推荐的错误处理模式
func doRequest(url string) string {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("网络错误: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Sprintf("HTTP 错误: 状态码 %d, 响应体: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("读取响应体失败: %v", err)
	}

	return fmt.Sprintf("成功: %s", string(body))
}

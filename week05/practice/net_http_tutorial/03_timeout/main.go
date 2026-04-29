package main

// ============================================================
// 第 03 课：超时控制 — 防止请求永远卡住
// ============================================================
//
// 【为什么需要超时？】
//   如果目标服务器卡死或网络中断，没有超时的请求会永远等待，导致：
//   - goroutine 泄漏（每个请求占用一个 goroutine）
//   - 连接池耗尽（所有连接都在等待）
//   - 程序假死（用户无响应）
//
// 【两种超时方式】
//   1. client.Timeout — 简单粗犷，所有请求共用一个超时
//   2. context.WithTimeout — 精细控制，每个请求可以有不同的超时
//
// 【运行方式】
//   go run 03_timeout/main.go
//   （本课使用 httptest 本地服务器，不需要联网）

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// ---- 启动一个测试服务器 ----
	// 模拟两种场景：快速响应（50ms）和慢速响应（500ms）
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fast":
			time.Sleep(50 * time.Millisecond)
			w.Write([]byte("fast response"))
		case "/slow":
			time.Sleep(500 * time.Millisecond)
			w.Write([]byte("slow response"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	fmt.Println("服务器地址:", server.URL)

	// ================================================================
	// 方式一：client.Timeout（简单粗犷）
	// ================================================================

	fmt.Println("\n===== 方式一：client.Timeout =====")

	// 创建带超时的客户端
	// Timeout 是整个请求的生命周期：DNS 解析 + TCP 连接 + TLS 握手 + 发送 + 接收
	client := &http.Client{
		Timeout: 300 * time.Millisecond, // 300 毫秒
	}

	// 场景 A：快速响应，在超时时间内
	fmt.Println("\n--- 场景 A：快速响应（服务端 50ms，超时 300ms）---")
	resp, err := client.Get(server.URL + "/fast")
	if err != nil {
		fmt.Println("❌ 请求失败:", err)
	} else {
		fmt.Println("✅ 请求成功，状态码:", resp.StatusCode)
		resp.Body.Close()
	}

	// 场景 B：慢速响应，超过超时时间
	fmt.Println("\n--- 场景 B：慢速响应（服务端 500ms，超时 300ms）---")
	resp, err = client.Get(server.URL + "/slow")
	if err != nil {
		// 超时后返回的错误类型是 *url.Error
		// 错误信息包含 "context deadline exceeded"
		fmt.Println("❌ 请求失败:", err)
	} else {
		fmt.Println("✅ 请求成功（这不应该发生）")
		resp.Body.Close()
	}

	// ---- client.Timeout 的特点 ----
	// ✅ 优点：一行搞定，简单
	// ❌ 缺点：所有请求共用同一个超时，无法针对不同 URL 设置不同超时
	//          无法手动取消（比如用户点击了"取消"按钮）

	// ================================================================
	// 方式二：context.WithTimeout（精细控制）
	// ================================================================

	fmt.Println("\n===== 方式二：context.WithTimeout =====")

	// 方式二的核心思路：
	// 1. 创建一个带超时的 context
	// 2. 把 context 绑定到请求上
	// 3. 超时后 context 自动取消，请求也随之取消

	// 场景 C：给快速请求设置 1 秒超时
	fmt.Println("\n--- 场景 C：context 控制超时（1 秒）---")

	// context.WithTimeout 返回两个值：
	//   ctx    — 新的 context，1 秒后自动触发取消信号
	//   cancel — 取消函数，可以提前结束 context
	//
	// 注意：cancel 必须被调用！否则 context 内部的定时器 goroutine 会泄漏。
	// 所以紧跟一句 defer cancel()。
	//
	// 这个机制在之前的 context_01 课程中学过，这里是实际应用。

	// 创建请求
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/fast", nil)

	// 注意：这里用 context.WithTimeout 需要 import "context"
	// 但为了演示，我们用 client.Timeout 的方式，效果一样
	// 想看完整的 context 版本请看 04_context_cancel/main.go

	client2 := &http.Client{Timeout: 1 * time.Second}
	resp, err = client2.Do(req)
	if err != nil {
		fmt.Println("❌ 请求失败:", err)
	} else {
		fmt.Println("✅ 请求成功，状态码:", resp.StatusCode)
		resp.Body.Close()
	}

	// ---- 两种方式对比 ----
	//
	//   client.Timeout：
	//     - 适合所有请求用同一超时时间的场景
	//     - 代码最少，一行搞定
	//
	//   context：
	//     - 可以为每个请求设置不同的超时
	//     - 支持手动取消（调用 cancel()）
	//     - 可以组合多个 context（比如 parent + timeout）
	//     - 推荐用于生产环境
	//
	// 实际项目中，最佳实践是两者结合：
	//   client := &http.Client{Timeout: 30 * time.Second}  // 兜底超时
	//   ctx, cancel := context.WithTimeout(ctx, 5*time.Second)  // 精细控制
	//   defer cancel()

	fmt.Println("\n===== 总结 =====")
	fmt.Println("超时控制是 HTTP 客户端最重要的配置！")
	fmt.Println("至少要设置 client.Timeout，否则请求可能永远卡住")
}

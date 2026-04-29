package main

// ============================================================
// 第 08 课：context 在 HTTP 请求中的实际应用
// ============================================================
//
// 【context + HTTP 的关系】
//   net/http 标准库深度集成了 context：
//     1. 服务端：每个 HTTP 请求的 r.Context() 返回一个 context
//     2. 客户端：http.NewRequestWithContext 把 context 绑定到请求上
//
// 【服务端】
//   当客户端断开连接时，r.Context() 会被自动取消。
//   这样你的 handler 可以停止正在做的工作，不浪费资源。
//
// 【客户端】
//   用 context 控制请求超时和取消。
//   这是你最常用的模式之一。
//
// 【运行方式】
//   go run 08_http_and_context/main.go

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// ================================================================
	// 服务端：r.Context() 的自动取消
	// ================================================================

	fmt.Println("===== 服务端：r.Context() =====")

	// 当你在服务端编写 handler 时，r.Context() 返回的 context 有以下行为：
	//   - 请求处理完成（handler 返回）时，context 被取消
	//   - 客户端断开连接时，context 被取消
	//   - 服务端关闭时，context 被取消

	// 模拟一个慢 handler：需要 5 秒才能响应
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [服务端] 收到请求，开始处理...")

		// 模拟慢操作：分 5 步，每步 1 秒
		for i := 1; i <= 5; i++ {
			// 每一步都检查 context 是否被取消
			select {
			case <-r.Context().Done():
				fmt.Printf("  [服务端] 客户端断开连接，停止处理（第 %d 步）\n", i)
				return
			case <-time.After(1 * time.Second):
				fmt.Printf("  [服务端] 第 %d/5 步\n", i)
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("处理完成"))
		fmt.Println("  [服务端] 响应完成")
	}))
	defer server.Close()

	// ---- 场景 A：正常请求（客户端等待完成） ----
	fmt.Println("\n--- 场景 A：正常请求（客户端等待 6 秒让服务端完成）---")

	client := &http.Client{Timeout: 6 * time.Second}
	respA, err := client.Get(server.URL)
	if err != nil {
		fmt.Printf("  [客户端] 请求失败: %v\n", err)
	} else {
		defer respA.Body.Close()
		body, _ := io.ReadAll(respA.Body)
		fmt.Printf("  [客户端] 收到响应: %s\n", string(body))
	}

	// ---- 场景 B：客户端提前断开 ----
	fmt.Println("\n--- 场景 B：客户端 2 秒后断开 ---")

	clientB := &http.Client{Timeout: 2 * time.Second}
	respB, err := clientB.Get(server.URL)
	if err != nil {
		fmt.Printf("  [客户端] 超时断开: %v\n", err)
	} else {
		defer respB.Body.Close()
		body, _ := io.ReadAll(respB.Body)
		fmt.Printf("  [客户端] 收到响应: %s\n", string(body))
	}

	// 等服务端清理完成
	time.Sleep(500 * time.Millisecond)

	// ================================================================
	// 客户端：http.NewRequestWithContext
	// ================================================================

	fmt.Println("\n===== 客户端：http.NewRequestWithContext =====")

	// 这是客户端使用 context 的标准方式：
	//
	//   // 第 1 步：创建带超时的 context
	//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//   defer cancel()
	//
	//   // 第 2 步：创建请求，绑定 context
	//   req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	//
	//   // 第 3 步：发送请求
	//   resp, err := client.Do(req)

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.Write([]byte("slow response"))
	}))
	defer server2.Close()

	fmt.Println("  服务端需要 3 秒，客户端设置 1 秒超时")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server2.URL, nil)

	client2 := &http.Client{}
	start := time.Now()
	respC, err := client2.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("  [客户端] 请求失败（%v）: %v\n", elapsed, err)
	} else {
		defer respC.Body.Close()
		body, _ := io.ReadAll(respC.Body)
		fmt.Printf("  [客户端] 收到响应: %s\n", string(body))
	}

	// ================================================================
	// 服务端：向下游传递 context
	// ================================================================

	fmt.Println("\n===== 服务端：向下游服务传递 context =====")

	// 实际项目中，服务端可能还需要调用其他服务。
	// 把 r.Context() 传递给下游，可以实现全链路的取消传播。

	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [下游服务] 收到请求")
		time.Sleep(5 * time.Second)
		w.Write([]byte("downstream done"))
	}))
	defer downstream.Close()

	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [代理服务] 收到请求，转发到下游")

		// 用 r.Context() 创建请求，这样如果原始客户端断开，下游也会被取消
		ctx := r.Context()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, downstream.URL, nil)

		client3 := &http.Client{}
		resp, err := client3.Do(req)
		if err != nil {
			fmt.Printf("  [代理服务] 下游请求失败: %v\n", err)
			http.Error(w, "downstream error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		w.Write(body)
	}))
	defer proxyServer.Close()

	fmt.Println("  客户端设置 1 秒超时 → 代理服务转发 → 下游服务被取消")

	client4 := &http.Client{Timeout: 1 * time.Second}
	respD, err := client4.Get(proxyServer.URL)
	if err != nil {
		fmt.Printf("  [客户端] 超时: %v\n", err)
	} else {
		defer respD.Body.Close()
		io.ReadAll(respD.Body)
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n===== 总结 =====")
	fmt.Println("服务端：r.Context() 在客户端断开时自动取消")
	fmt.Println("客户端：http.NewRequestWithContext 绑定超时控制")
	fmt.Println("传递：把 context 传给下游服务，实现全链路取消")
}

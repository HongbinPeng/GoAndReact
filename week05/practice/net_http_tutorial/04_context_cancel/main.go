package main

// ============================================================
// 第 04 课：context 取消 — 手动控制请求的生命周期
// ============================================================
//
// 【context 能做什么？】
//   1. 超时自动取消 — context.WithTimeout
//   2. 手动立即取消 — context.WithCancel
//   3. 传递请求范围的值 — context.WithValue（不推荐传递关键数据）
//   4. 级联取消 — 父 context 取消时，所有子 context 也会取消
//
// 【典型场景】
//   - 用户点击"取消"按钮，中断正在进行的 HTTP 请求
//   - 程序关闭时，取消所有进行中的请求
//   - 一个请求依赖另一个请求的结果，前一个失败了就取消后一个
//
// 【运行方式】
//   go run 04_context_cancel/main.go

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// 启动一个测试服务器：请求会等待 3 秒才响应
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [服务端] 收到请求，开始等待 3 秒...")
		time.Sleep(3 * time.Second)
		fmt.Println("  [服务端] 等待结束，返回响应")
		w.Write([]byte("done"))
	}))
	defer server.Close()

	// ================================================================
	// 场景 1：超时自动取消（context.WithTimeout）
	// ================================================================

	fmt.Println("===== 场景 1：context.WithTimeout 超时自动取消 =====")

	// 第 1 步：创建一个带超时的 context
	// WithTimeout 返回 (ctx, cancel) 两个值
	// ctx 会在 500ms 后自动触发取消信号
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	// cancel 必须在创建后尽快 defer，确保资源被释放
	defer cancel()

	// 第 2 步：创建请求，绑定 context
	// NewRequestWithContext 把 ctx 和请求绑在一起
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/slow", nil)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}

	// 第 3 步：发送请求
	fmt.Println("  [客户端] 发送请求，超时设置为 500ms")
	start := time.Now()

	client := &http.Client{}
	_, err = client.Do(req)

	elapsed := time.Since(start)
	fmt.Printf("  [客户端] 请求结束，耗时 %v\n", elapsed)

	if err != nil {
		// context.DeadlineExceeded 是超时错误的标准类型
		// 可以用 errors.Is(err, context.DeadlineExceeded) 来精确判断
		fmt.Println("  [客户端] 错误:", err)
	}

	// 观察结果：
	//   - 客户端在 ~500ms 时就返回了错误
	//   - 服务端会继续等待 3 秒（服务端不知道客户端已经放弃了）
	//   - 但客户端的 goroutine 已经释放，不会永远卡住

	// ================================================================
	// 场景 2：手动取消（context.WithCancel）
	// ================================================================

	fmt.Println("\n===== 场景 2：context.WithCancel 手动取消 =====")

	// 模拟用户点击"取消"按钮的场景
	// 我们启动一个 goroutine 发请求，另一个 goroutine 在 200ms 后点击"取消"

	// 第 1 步：创建一个可以手动取消的 context
	ctx2, cancel2 := context.WithCancel(context.Background())

	req2, _ := http.NewRequestWithContext(ctx2, http.MethodGet, server.URL+"/slow", nil)
	client2 := &http.Client{}

	// 第 2 步：启动一个 goroutine 发请求
	done := make(chan bool, 1)
	go func() {
		fmt.Println("  [goroutine] 开始发请求...")
		_, err := client2.Do(req2)
		if err != nil {
			fmt.Println("  [goroutine] 请求被取消:", err)
		} else {
			fmt.Println("  [goroutine] 请求成功")
		}
		done <- true
	}()

	// 第 3 步：模拟用户 200ms 后点击"取消"按钮
	time.Sleep(200 * time.Millisecond)
	fmt.Println("  [主程序] 用户点击了取消按钮，调用 cancel2()")
	cancel2() // 立刻取消！正在进行的请求会被终止

	// 等待 goroutine 结束
	<-done

	// 观察结果：
	//   - 200ms 时调用 cancel2()
	//   - 请求立刻被取消，不会等到 3 秒
	//   - 错误类型是 "context canceled"（注意不是 "deadline exceeded"）

	// ---- context 错误类型对比 ----
	//
	//   context.DeadlineExceeded — 超时导致的取消
	//     错误信息: "context deadline exceeded"
	//     触发方式: WithTimeout 的时间到了
	//
	//   context.Canceled — 手动调用 cancel() 导致的取消
	//     错误信息: "context canceled"
	//     触发方式: 主动调用 cancel()
	//
	// 判断方法：
	//   errors.Is(err, context.DeadlineExceeded)  // 超时？
	//   errors.Is(err, context.Canceled)          // 手动取消？

	// ================================================================
	// 场景 3：context 级联取消
	// ================================================================

	fmt.Println("\n===== 场景 3：context 级联取消 =====")

	// context 是可以组合的：子 context 会继承父 context 的取消信号
	// 当父 context 取消时，所有子 context 也会取消

	// 创建父 context（100ms 后取消）
	parentCtx, parentCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer parentCancel()

	// 创建子 context（继承父 context，另外加 5 秒超时）
	// 实际超时时间取两者中较短的那个：100ms（父）< 5秒（子），所以实际 100ms 取消
	childCtx, childCancel := context.WithTimeout(parentCtx, 5*time.Second)
	defer childCancel()

	req3, _ := http.NewRequestWithContext(childCtx, http.MethodGet, server.URL+"/slow", nil)
	client3 := &http.Client{}

	fmt.Println("  [客户端] 发送请求（子 context 5s 超时，但父 context 100ms 后取消）")
	start3 := time.Now()
	_, err = client3.Do(req3)
	fmt.Printf("  [客户端] 请求结束，耗时 %v，错误: %v\n", time.Since(start3), err)

	// 应用场景：
	//   假设你有一个 Web 服务器，处理一个请求时需要调用 3 个外部 API。
	//   如果用户的 HTTP 请求被断开（浏览器关闭/网络断开），
	//   父 context 会被取消，3 个子请求也会自动取消，不会浪费资源。

	fmt.Println("\n===== 总结 =====")
	fmt.Println("context 是 Go 中控制请求生命周期的核心机制")
	fmt.Println("WithTimeout → 超时自动取消")
	fmt.Println("WithCancel  → 手动立即取消")
	fmt.Println("两者组合  → 灵活的请求控制")
}

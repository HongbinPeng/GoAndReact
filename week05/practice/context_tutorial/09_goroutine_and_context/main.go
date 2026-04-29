package main

// ============================================================
// 第 09 课：context 管理多个 goroutine
// ============================================================
//
// 【本课重点】
//   context 最强大的地方在于：一个取消信号可以控制成百上千个 goroutine。
//   这是实际项目中最常见的用法。
//
// 【运行方式】
//   go run 09_goroutine_and_context/main.go

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// ================================================================
	// 场景 1：用 context 管理 worker pool
	// ================================================================

	fmt.Println("===== 场景 1：Worker Pool 管理 =====")

	// 启动 5 个 worker，共享同一个 context
	ctx, cancel := context.WithCancel(context.Background())

	for i := 1; i <= 5; i++ {
		go func(id int) {
			fmt.Printf("  Worker %d 启动\n", id)
			defer fmt.Printf("  Worker %d 退出\n", id)

			for {
				select {
				case <-ctx.Done():
					fmt.Printf("  Worker %d 收到取消信号: %v\n", id, ctx.Err())
					return
				case <-time.After(100 * time.Millisecond):
					fmt.Printf("  Worker %d 工作了一次\n", id)
				}
			}
		}(i)
	}

	// 运行 500ms 后取消
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  主程序：取消所有 worker")
	cancel()

	time.Sleep(200 * time.Millisecond)

	// ================================================================
	// 场景 2：用 context 控制任务超时
	// ================================================================

	fmt.Println("\n===== 场景 2：任务超时控制 =====")

	// 假设你有 3 个任务要并发执行，但总共只等 1 秒
	ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel2()

	results := make(chan string, 3)

	// 任务 A：300ms 完成（在超时内）
	go func() {
		select {
		case <-ctx2.Done():
			results <- "A: 被取消"
			return
		case <-time.After(300 * time.Millisecond):
			results <- "A: 完成 ✓"
		}
	}()

	// 任务 B：800ms 完成（在超时内）
	go func() {
		select {
		case <-ctx2.Done():
			results <- "B: 被取消"
			return
		case <-time.After(800 * time.Millisecond):
			results <- "B: 完成 ✓"
		}
	}()

	// 任务 C：2 秒完成（会超时）
	go func() {
		select {
		case <-ctx2.Done():
			results <- "C: 被取消 ✗"
			return
		case <-time.After(2 * time.Second):
			results <- "C: 完成 ✓"
		}
	}()

	// 收集结果
	for i := 0; i < 3; i++ {
		r := <-results
		fmt.Printf("  结果: %s\n", r)
	}

	// ================================================================
	// 场景 3：优雅关闭（Graceful Shutdown）
	// ================================================================

	fmt.Println("\n===== 场景 3：优雅关闭 =====")

	// 这是程序退出时的标准模式：
	//   1. 创建一个顶级 context
	//   2. 所有 goroutine 都监听这个 context
	//   3. 程序要退出时，调用 cancel()
	//   4. 等待所有 goroutine 退出
	//   5. 程序退出

	appCtx, appCancel := context.WithCancel(context.Background())

	// 模拟 3 个服务
	services := []string{"HTTP Server", "Database Pool", "Cache Cleaner"}

	for _, name := range services {
		go func(svc string) {
			fmt.Printf("  [%s] 启动\n", svc)
			defer fmt.Printf("  [%s] 已关闭\n", svc)

			// 模拟服务运行
			select {
			case <-appCtx.Done():
				fmt.Printf("  [%s] 收到关闭信号，正在清理...\n", svc)
				time.Sleep(200 * time.Millisecond) // 模拟清理工作
			}
		}(name)
	}

	// 运行一段时间后关闭
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  主程序：发送关闭信号")
	appCancel()

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n===== 总结 =====")
	fmt.Println("一个 context 可以管理任意数量的 goroutine")
	fmt.Println("Worker Pool → select + ctx.Done() 循环检查")
	fmt.Println("任务超时 → WithTimeout + select 并发等待")
	fmt.Println("优雅关闭 → cancel() → 所有 goroutine 收到信号 → 清理 → 退出")
}

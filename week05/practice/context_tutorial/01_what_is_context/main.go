package main

// ============================================================
// 第 01 课：context 是什么？为什么需要它？
// ============================================================
//
// 【context 是什么？】
//   context 是 Go 标准库提供的一个包，核心类型是 context.Context 接口。
//   它用于在 goroutine 之间传递：
//     1. 取消信号（"别干了，停下来！"）
//     2. 超时信息（"你最多只能干 5 秒"）
//     3. 请求范围的值（"这个请求的用户 ID 是 123"）
//
// 【为什么需要 context？】
//
//   场景 1：你启动了一个 goroutine 做耗时工作，但用户取消了操作
//     如果没有 context → goroutine 会继续跑完，浪费 CPU 和内存
//     有了 context → goroutine 收到取消信号，立刻退出
//
//   场景 2：你要调用一个外部 API，但它响应很慢
//     如果没有 context → 你的程序会永远等下去
//     有了 context → 3 秒后自动取消，不会卡死
//
//   场景 3：一个请求需要调用 5 个服务，第 2 个失败了
//     如果没有 context → 剩下的 3 个调用还会继续执行
//     有了 context → 取消信号会传播给所有子调用
//
// 【Context 接口定义了 4 个方法】
//
//   type Context interface {
//       Deadline() (deadline time.Time, ok bool)  // 什么时候超时？
//       Done() <-chan struct{}                     // 取消信号 channel
//       Err() error                                // 为什么被取消？
//       Value(key any) any                         // 附带的值
//   }
//
// 这 4 个方法是理解 context 的关键，后面每课都会用到。
//
// 【运行方式】
//   go run 01_what_is_context/main.go

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("===== context.Context 接口的 4 个方法 =====")

	// Context 是一个接口，所有具体的 context 类型都实现了它。
	// 这 4 个方法看似简单，但组合起来能做很多事情。

	fmt.Println()
	fmt.Println("方法 1：Deadline() → 返回 (时间, 是否有超时)")
	fmt.Println("   作用：告诉调用者，这个 context 什么时候会超时")
	fmt.Println("   返回值：")
	fmt.Println("     deadline time.Time — 超时时间点")
	fmt.Println("     ok bool           — true 表示有超时，false 表示没有")
	fmt.Println("   示例：deadline, ok := ctx.Deadline()")

	fmt.Println()
	fmt.Println("方法 2：Done() → 返回一个只读的 channel")
	fmt.Println("   作用：当 context 被取消时，这个 channel 会被关闭")
	fmt.Println("   监听方式：select { case <-ctx.Done(): ... }")
	fmt.Println("   channel 被关闭后，<-ctx.Done() 会立刻返回（不再阻塞）")
	fmt.Println("   这是 context 最常用的方法！")

	fmt.Println()
	fmt.Println("方法 3：Err() → 返回 error")
	fmt.Println("   作用：context 被取消后，返回错误原因")
	fmt.Println("   可能的错误：")
	fmt.Println("     context.Canceled          → 手动调用 cancel() 取消")
	fmt.Println("     context.DeadlineExceeded  → 超时自动取消")
	fmt.Println("   注意：只有在 Done() 关闭后调用 Err() 才有意义")

	fmt.Println()
	fmt.Println("方法 4：Value(key) → 返回 any")
	fmt.Println("   作用：在 context 中附加一个值，通过 key 来读取")
	fmt.Println("   适用场景：传递请求范围的元数据（用户 ID、追踪 ID 等）")
	fmt.Println("   不建议：传递函数参数（应该用函数参数而不是 context）")

	fmt.Println("\n===== context 的创建方式 =====")

	// context 包提供了 4 个工厂函数来创建 context：
	//
	//   context.Background()    — 最顶层的 context，没有超时、没有值、没有取消信号
	//   context.TODO()          — 和 Background() 一样，语义上表示"还不知道用哪种"
	//   context.WithCancel(parent)        — 可手动取消的 context
	//   context.WithTimeout(parent, d)    — 带超时的 context
	//   context.WithDeadline(parent, t)   — 在指定时刻超时的 context
	//   context.WithValue(parent, k, v)   — 携带值的 context

	fmt.Println("context.Background()  — 所有 context 的根")
	fmt.Println("context.TODO()        — 语义上表示「暂不确定」")
	fmt.Println("WithCancel(parent)    — 手动取消")
	fmt.Println("WithTimeout(parent, d) — 超时取消")
	fmt.Println("WithDeadline(parent, t) — 定时取消")
	fmt.Println("WithValue(parent, k, v) — 携带值")

	fmt.Println("\n===== 用 channel 模拟理解 Done() =====")

	// 在深入 context 之前，先用 channel 理解取消信号的机制：
	done := make(chan struct{}) // Done() 返回的就是这种 channel

	go func() {
		fmt.Println("  goroutine 开始工作...")
		select {
		case <-done:
			// done channel 被关闭时，这个 case 会被选中
			fmt.Println("  goroutine 收到取消信号，退出！")
		case <-time.After(5 * time.Second):
			// 5 秒后才会触发（但如果 done 先关闭，就走不到这里）
			fmt.Println("  goroutine 工作完成")
		}
	}()

	// 1 秒后取消
	time.Sleep(1 * time.Second)
	fmt.Println("  主程序：关闭 done channel")
	close(done) // 关闭 channel，所有监听 done 的 goroutine 都会收到信号

	time.Sleep(100 * time.Millisecond) // 等一下让 goroutine 退出

	fmt.Println("\n===== 总结 =====")
	fmt.Println("context 的核心是「取消信号」机制")
	fmt.Println("Done() 返回的 channel 关闭时 → context 被取消")
	fmt.Println("Err() 告诉你为什么被取消")
	fmt.Println("Deadline() 告诉你什么时候超时")
	fmt.Println("Value() 让你可以附带一些元数据")
}

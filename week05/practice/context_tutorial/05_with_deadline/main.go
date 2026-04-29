package main

// ============================================================
// 第 05 课：WithDeadline — 在指定时刻取消
// ============================================================
//
// 【WithDeadline 做什么？】
//   和 WithTimeout 几乎一样，区别在于：
//     WithTimeout  — 相对时间："5 秒后超时"
//     WithDeadline — 绝对时间："在 2024-01-01 12:00:00 这个时刻超时"
//
// 【函数签名】
//   func WithDeadline(parent Context, deadline time.Time) (ctx Context, cancel CancelFunc)
//
//   参数：
//     deadline — 超时的绝对时间点
//
//   底层关系：
//     WithTimeout(parent, d)  等价于  WithDeadline(parent, time.Now().Add(d))
//
// 【运行方式】
//   go run 05_with_deadline/main.go

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// ================================================================
	// 场景 1：指定一个未来的时间点
	// ================================================================

	fmt.Println("===== 场景 1：指定一个未来的绝对时间 =====")

	// 1.5 秒后超时（用绝对时间表示）
	deadline := time.Now().Add(1500 * time.Millisecond)

	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	fmt.Printf("  设定的超时时刻: %v\n", deadline)
	fmt.Printf("  当前时间:       %v\n", time.Now())

	select {
	case <-time.After(3 * time.Second):
		fmt.Println("  工作完成（不应该走到这里）")
	case <-ctx.Done():
		fmt.Printf("  被取消了！原因: %v\n", ctx.Err())
	}

	// ================================================================
	// 场景 2：Deadline 已经过去 → 立刻取消
	// ================================================================

	fmt.Println("\n===== 场景 2：Deadline 已经过去 = 立刻取消 =====")

	// 如果 deadline 已经过去了（或者设为过去的时间），
	// context 会立刻被取消，Done channel 已经关闭。

	pastDeadline := time.Now().Add(-1 * time.Second) // 1 秒前
	ctx2, cancel2 := context.WithDeadline(context.Background(), pastDeadline)
	defer cancel2()

	// Done channel 已经关闭，所以 <-ctx2.Done() 不会阻塞
	select {
	case <-ctx2.Done():
		fmt.Printf("  立刻被取消！原因: %v\n", ctx2.Err())
	default:
		fmt.Println("  还没有被取消（不应该走到这里）")
	}

	// ================================================================
	// 场景 3：WithDeadline vs WithTimeout 对比
	// ================================================================

	fmt.Println("\n===== 场景 3：WithDeadline vs WithTimeout =====")

	// 这两种写法等价：
	//
	//   ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	//   defer cancel1()
	//
	//   ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	//   defer cancel2()

	// 什么时候用 WithDeadline 而不是 WithTimeout？
	//
	//   场景 A：你有一个固定的截止时间（比如缓存的过期时间）
	//     expiresAt := cacheEntry.ExpiresAt  // time.Time
	//     ctx, cancel := context.WithDeadline(parent, expiresAt)
	//
	//   场景 B：你需要协调多个操作的超时（都必须在某个时刻前完成）
	//     cutoff := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	//     ctx, cancel := context.WithDeadline(parent, cutoff)
	//     // 所有子操作都必须在年底之前完成
	//
	//   场景 C：你只关心"多久之后" → 用 WithTimeout 更直观
	//     ctx, cancel := context.WithTimeout(parent, 5*time.Second)

	fmt.Println("WithDeadline 适合有固定截止时间的场景")
	fmt.Println("WithTimeout 适合'多久之后'超时的场景")
	fmt.Println("两者底层是等价的")

	fmt.Println("\n===== 总结 =====")
	fmt.Println("WithDeadline(parent, t) — 在时间 t 超时")
	fmt.Println("WithTimeout(parent, d)  — 在 d 时长后超时")
	fmt.Println("WithDeadline 已过 → 立刻取消")
	fmt.Println("实际项目中 WithTimeout 用得更多（更直观）")
}

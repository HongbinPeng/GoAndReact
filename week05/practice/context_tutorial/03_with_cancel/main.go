package main

// ============================================================
// 第 03 课：WithCancel — 手动取消 goroutine
// ============================================================
//
// 【WithCancel 做什么？】
//   创建一个可以被手动取消的 context。
//   调用 cancel() 后，所有监听这个 context.Done() 的 goroutine 都会收到信号。
//
// 【函数签名】
//   func WithCancel(parent Context) (ctx Context, cancel CancelFunc)
//
//   返回值：
//     ctx    — 新的 context，它的 Done() channel 会在 cancel() 调用后关闭
//     cancel — 取消函数，调用它就会取消这个 context
//
// 【运行方式】
//   go run 03_with_cancel/main.go

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// ================================================================
	// 场景 1：取消一个 goroutine
	// ================================================================

	fmt.Println("===== 场景 1：取消一个 goroutine =====")

	// 第 1 步：创建可取消的 context
	// parent 通常是 context.Background()
	ctx, cancel := context.WithCancel(context.Background())

	// 启动一个 goroutine，监听取消信号
	go func() {
		for i := 1; i <= 10; i++ {
			// 每次工作前检查是否被取消
			select {
			case <-ctx.Done():
				// Done channel 被关闭了 → context 被取消
				// ctx.Err() 会返回 context.Canceled
				fmt.Printf("  goroutine 被取消，错误: %v\n", ctx.Err())
				return
			default:
				// 没被取消，继续工作
			}

			fmt.Printf("  goroutine 工作 #%d\n", i)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("  goroutine 正常完成")
	}()

	// 第 2 步：主程序等待 500ms 后取消
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  主程序：调用 cancel()")
	cancel() // 调用后 ctx.Done() channel 会被关闭

	// 等一下让 goroutine 退出
	time.Sleep(100 * time.Millisecond)

	// ---- WithCancel 内部做了什么？ ----
	//
	//   func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	//       c := newCancelCtx(parent)          // 创建内部结构
	//       propagateCancel(parent, &c)        // 注册到父 context 的取消链上
	//       return &c, func() { c.cancel(true, Canceled) }
	//   }
	//
	//   当你调用 cancel() 时：
	//   1. 关闭 ctx.Done() channel
	//   2. 设置 ctx.err = context.Canceled
	//   3. 通知所有子 context 也被取消（级联取消）

	// ================================================================
	// 场景 2：取消多个 goroutine
	// ================================================================

	fmt.Println("\n===== 场景 2：一个 cancel 取消多个 goroutine =====")

	ctx2, cancel2 := context.WithCancel(context.Background())

	// 启动 3 个 goroutine，它们共享同一个 context
	for i := 1; i <= 3; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx2.Done():
					fmt.Printf("  goroutine %d 收到取消信号，退出\n", id)
					return
				default:
					fmt.Printf("  goroutine %d 工作中...\n", id)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}(i)
	}

	time.Sleep(300 * time.Millisecond)
	fmt.Println("  主程序：调用 cancel2() 取消所有 goroutine")
	cancel2()

	time.Sleep(200 * time.Millisecond)

	// ================================================================
	// 场景 3：cancel 只能调用一次
	// ================================================================

	fmt.Println("\n===== 场景 3：cancel 多次调用是安全的 =====")

	ctx3, cancel3 := context.WithCancel(context.Background())
	cancel3() // 第一次调用
	cancel3() // 第二次调用（不会 panic，什么都不发生）
	cancel3() // 第三次调用（同样安全）
	fmt.Println("  多次调用 cancel 是安全的，不会 panic")
	fmt.Printf("  ctx.Err() = %v\n", ctx3.Err())

	// ================================================================
	// 场景 4：忘记调用 cancel 会泄漏
	// ================================================================

	fmt.Println("\n===== 场景 4：忘记调用 cancel 的后果 =====")

	// WithCancel 内部创建了一个 timer/goroutine 来管理取消逻辑。
	// 如果不调用 cancel()，这个 goroutine 会一直存在直到父 context 被取消。
	// 虽然最终 GC 会回收，但短期内是泄漏的。

	ctx4, cancel4 := context.WithCancel(context.Background())

	// 演示：检查 Done channel
	if ctx4.Done() == nil {
		fmt.Println("  Done() 返回 nil → 这个 context 不会被取消")
	} else {
		fmt.Println("  Done() 返回 channel → 等待被取消")
	}

	// 正确做法：用完后一定要调用 cancel
	cancel4()
	fmt.Println("  调用了 cancel4()，释放资源")

	fmt.Println("\n===== 总结 =====")
	fmt.Println("WithCancel(parent) → (ctx, cancel)")
	fmt.Println("调用 cancel() → ctx.Done() 关闭 → goroutine 退出")
	fmt.Println("cancel 可以多次调用（安全的）")
	fmt.Println("一定要调用 cancel()，否则内部 goroutine 泄漏")
}

package main

// ============================================================
// 第 04 课：WithTimeout — 超时自动取消
// ============================================================
//
// 【WithTimeout 做什么？】
//   创建一个带超时的 context。超过指定时间后，context 会自动取消。
//   这是实际项目中最常用的 context 创建方式。
//
// 【函数签名】
//   func WithTimeout(parent Context, timeout time.Duration) (ctx Context, cancel CancelFunc)
//
//   参数：
//     parent  — 父 context（通常是 context.Background()）
//     timeout — 多久之后自动取消（比如 5 * time.Second）
//
//   返回值：
//     ctx    — 新的 context，超时后自动取消
//     cancel — 取消函数，可以提前手动取消
//
//   底层等价于：
//     WithDeadline(parent, time.Now().Add(timeout))
//
// 【运行方式】
//   go run 04_with_timeout/main.go

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// ================================================================
	// 场景 1：超时自动取消
	// ================================================================

	fmt.Println("===== 场景 1：超时自动取消 =====")

	// 创建一个 1 秒后超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	// 注意：cancel 必须被调用！即使超时了也要调用释放资源
	defer cancel()

	// 模拟一个需要 5 秒的工作
	fmt.Println("  启动一个需要 5 秒的工作，但 context 1 秒后超时...")

	select {
	case <-time.After(5 * time.Second):
		// 这行永远不会到达，因为超时先发生了
		fmt.Println("  工作完成")
	case <-ctx.Done():
		// 超时后 Done channel 被关闭，这里会被选中
		fmt.Printf("  被取消了！原因: %v\n", ctx.Err())
		// ctx.Err() 会返回 context.DeadlineExceeded（超时）
	}

	// ---- 两种取消错误的区别 ----
	//
	//   context.Canceled          → 手动调用 cancel() 导致的
	//   context.DeadlineExceeded  → 超时自动取消导致的

	// ================================================================
	// 场景 2：在超时前完成工作
	// ================================================================

	fmt.Println("\n===== 场景 2：在超时前完成工作 =====")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	fmt.Println("  启动一个需要 1 秒的工作，context 3 秒后超时...")

	// 模拟工作
	workDone := make(chan struct{})
	go func() {
		time.Sleep(1 * time.Second)
		workDone <- struct{}{}
	}()

	select {
	case <-workDone:
		fmt.Println("  工作在超时前完成了！")
	case <-ctx2.Done():
		// 如果工作超过了 3 秒，会走到这里
		fmt.Printf("  超时了！原因: %v\n", ctx2.Err())
	}

	// ================================================================
	// 场景 3：Deadline() 方法
	// ================================================================

	fmt.Println("\n===== 场景 3：Deadline() 查看超时时间 =====")

	ctx3, cancel3 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel3()

	deadline, ok := ctx3.Deadline()
	if ok {
		fmt.Printf("  超时时间: %v\n", deadline)
		fmt.Printf("  剩余时间: %v\n", time.Until(deadline))
	} else {
		fmt.Println("  没有超时限制")
	}

	// 等一下再检查
	time.Sleep(1 * time.Second)
	deadline2, ok2 := ctx3.Deadline()
	if ok2 {
		fmt.Printf("  1 秒后，剩余时间: %v\n", time.Until(deadline2))
	}

	// ================================================================
	// 场景 4：模拟 HTTP 请求超时（实际项目中最常见的用法）
	// ================================================================

	fmt.Println("\n===== 场景 4：模拟 HTTP 请求超时 =====")

	// 在实际项目中，这是你最常用的模式：
	//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//   defer cancel()
	//   req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	//   resp, err := client.Do(req)
	//
	// 如果 5 秒内服务器没有响应，请求会被自动取消。

	timeout := 500 * time.Millisecond
	ctx4, cancel4 := context.WithTimeout(context.Background(), timeout)
	defer cancel4()

	// 模拟一个慢操作
	slowWorkDone := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second) // 需要 2 秒
		close(slowWorkDone)
	}()

	start := time.Now()
	select {
	case <-slowWorkDone:
		fmt.Println("  慢操作完成")
	case <-ctx4.Done():
		elapsed := time.Since(start)
		fmt.Printf("  操作被取消，耗时 %v（约等于超时时间 %v）\n", elapsed, timeout)
		fmt.Printf("  错误类型: %v\n", ctx4.Err())
	}

	// ================================================================
	// 场景 5：WithTimeout vs 手动 timer 对比
	// ================================================================

	fmt.Println("\n===== 场景 5：为什么不用 time.After 代替 WithTimeout？ =====")

	// 有人可能会想：我直接用 time.After 不就行了？
	//
	//   // ❌ 不好的写法
	//   select {
	//   case <-workDone:
	//       fmt.Println("完成")
	//   case <-time.After(5 * time.Second):
	//       fmt.Println("超时")
	//   }
	//
	// 问题：
	//   1. time.After 只是通知你超时了，但它不能传播给子调用
	//      如果你的 goroutine 内部又调用了其他函数，那些函数不知道超时了
	//
	//   2. 没有标准化的取消信号，每个函数需要自己定义超时逻辑
	//
	//   3. 无法组合（比如一个请求既要超时，又要支持用户手动取消）
	//
	// WithTimeout 的好处：
	//   1. 取消信号可以通过 context 传递到任意深度的调用链
	//   2. 所有标准库和第三方库都支持 context 参数
	//   3. 可以和其他 context 组合（WithValue, WithCancel 等）

	fmt.Println("time.After 只是本地超时通知")
	fmt.Println("WithTimeout 的 context 可以传递到调用链的每一层")
	fmt.Println("所有标准库函数（database/sql, net/http 等）都接受 context")

	fmt.Println("\n===== 总结 =====")
	fmt.Println("WithTimeout(parent, d) → 超时自动取消")
	fmt.Println("超时后 ctx.Done() 关闭，ctx.Err() 返回 DeadlineExceeded")
	fmt.Println("即使超时了，也要调用 cancel() 释放资源")
	fmt.Println("这是实际项目中最常用的 context 创建方式")
}

package main

// ============================================================
// 第 07 课：context 树和级联取消
// ============================================================
//
// 【什么是 context 树？】
//   每个 context 都有一个父 context（除了 Background 和 TODO）。
//   当你创建一个子 context 时，它就挂在父 context 下面，形成一棵树。
//
// 【级联取消】
//   当父 context 被取消时，所有子 context 也会被自动取消。
//   这个效果可以传递任意深度：祖父取消 → 父取消 → 子取消 → 孙取消...
//
// 【运行方式】
//   go run 07_context_tree/main.go

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// ================================================================
	// 场景 1：构建一棵 context 树
	// ================================================================

	fmt.Println("===== 场景 1：context 树的结构 =====")

	// 想象这样一棵树：
	//
	//   root (Background)
	//     ├── cancelCtx  ← WithCancel(root)
	//     │     ├── timeoutCtx1 ← WithTimeout(cancelCtx, 2s)
	//     │     └── timeoutCtx2 ← WithTimeout(cancelCtx, 3s)
	//     └── valueCtx  ← WithValue(root, "key", "val")

	root := context.Background()

	// 第一层：可取消的 context
	cancelCtx, cancelFunc := context.WithCancel(root)

	// 第二层：两个带超时的子 context
	timeoutCtx1, cancel1 := context.WithTimeout(cancelCtx, 2*time.Second)
	defer cancel1()

	timeoutCtx2, cancel2 := context.WithTimeout(cancelCtx, 3*time.Second)
	defer cancel2()

	// 第一层的另一个分支：带值的 context（验证父 context 不受影响）
	_ = context.WithValue(root, "key", "value-in-root")

	fmt.Println("context 树结构：")
	fmt.Println("  root (Background)")
	fmt.Println("    ├── cancelCtx (WithCancel)")
	fmt.Println("    │     ├── timeoutCtx1 (2s)")
	fmt.Println("    │     └── timeoutCtx2 (3s)")
	fmt.Println("    └── valueCtx (WithValue)")

	// ================================================================
	// 场景 2：级联取消演示
	// ================================================================

	fmt.Println("\n===== 场景 2：级联取消 =====")

	// 当调用 cancelFunc() 时：
	//   1. cancelCtx 被取消 → cancelCtx.Done() 关闭
	//   2. timeoutCtx1 被取消（虽然它还有 2s 才超时）→ Done() 关闭
	//   3. timeoutCtx2 被取消（虽然它还有 3s 才超时）→ Done() 关闭
	//   4. valueCtx 不受影响（它是 root 的直接子节点，不是 cancelCtx 的子节点）

	done1 := make(chan struct{})
	done2 := make(chan struct{})

	// goroutine 1：监听 timeoutCtx1
	go func() {
		fmt.Println("  [goroutine 1] 监听 timeoutCtx1（2s 超时）")
		select {
		case <-timeoutCtx1.Done():
			fmt.Printf("  [goroutine 1] 被取消！原因: %v\n", timeoutCtx1.Err())
		}
		close(done1)
	}()

	// goroutine 2：监听 timeoutCtx2
	go func() {
		fmt.Println("  [goroutine 2] 监听 timeoutCtx2（3s 超时）")
		select {
		case <-timeoutCtx2.Done():
			fmt.Printf("  [goroutine 2] 被取消！原因: %v\n", timeoutCtx2.Err())
		}
		close(done2)
	}()

	// 等一下，然后取消父 context
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  [主程序] 调用 cancelFunc() 取消父 context")
	cancelFunc()

	// 等待 goroutine 退出
	<-done1
	<-done2

	// 注意：
	//   timeoutCtx1.Err() 返回的是 context.Canceled（因为是被父取消的，不是超时）
	//   如果是超时导致的取消，会返回 context.DeadlineExceeded

	// ================================================================
	// 场景 3：实际应用场景 — Web 服务器
	// ================================================================

	fmt.Println("\n===== 场景 3：Web 服务器中的 context 树 =====")

	// Web 服务器的 context 树通常长这样：
	//
	//   Background
	//     └── Server context (WithCancel)       ← 服务器启动时创建
	//           └── Request context             ← 每个 HTTP 请求创建一个子 context
	//                 ├── DB query context (WithTimeout)    ← 数据库查询 5s 超时
	//                 ├── Cache context (WithTimeout)       ← 缓存查询 1s 超时
	//                 └── External API context (WithTimeout) ← 外部 API 调用 3s 超时
	//
	// 当用户关闭浏览器（断开连接）时：
	//   Request context 被取消 → DB query、Cache query、External API 全部自动取消！
	//   不会浪费服务器资源做无用功。

	fmt.Println("服务器启动 → 创建 serverContext")
	fmt.Println("  收到请求 → 创建 requestContext（serverContext 的子节点）")
	fmt.Println("    发 DB 查询 → 创建 dbContext（requestContext 的子节点）")
	fmt.Println("    发缓存查询 → 创建 cacheContext（requestContext 的子节点）")
	fmt.Println("  用户断开 → requestContext 取消 → 所有子查询全部自动取消！")

	// ================================================================
	// 场景 4：用 ErrGroup 理解并发 + 级联取消
	// ================================================================

	fmt.Println("\n===== 场景 4：多个子任务，一个失败全部取消 =====")

	// 模拟 3 个子任务，其中第 2 个会失败
	taskCtx, taskCancel := context.WithCancel(context.Background())

	taskDone := make(chan int, 3)

	// 任务 1：需要 3 秒完成
	go func() {
		fmt.Println("  [任务 1] 开始，需要 3 秒")
		for i := 1; i <= 3; i++ {
			select {
			case <-taskCtx.Done():
				fmt.Printf("  [任务 1] 被取消: %v\n", taskCtx.Err())
				taskDone <- 1
				return
			case <-time.After(1 * time.Second):
				fmt.Printf("  [任务 1] 第 %d/3 秒\n", i)
			}
		}
		fmt.Println("  [任务 1] 完成！")
		taskDone <- 1
	}()

	// 任务 2：1 秒后失败
	go func() {
		fmt.Println("  [任务 2] 开始，1 秒后会失败")
		select {
		case <-taskCtx.Done():
			fmt.Printf("  [任务 2] 被取消: %v\n", taskCtx.Err())
			taskDone <- 2
			return
		case <-time.After(1 * time.Second):
			fmt.Println("  [任务 2] ❌ 失败了！取消所有其他任务")
			taskCancel() // 级联取消所有子任务
			taskDone <- 2
			return
		}
	}()

	// 任务 3：需要 5 秒完成
	go func() {
		fmt.Println("  [任务 3] 开始，需要 5 秒")
		for i := 1; i <= 5; i++ {
			select {
			case <-taskCtx.Done():
				fmt.Printf("  [任务 3] 被取消: %v\n", taskCtx.Err())
				taskDone <- 3
				return
			case <-time.After(1 * time.Second):
				fmt.Printf("  [任务 3] 第 %d/5 秒\n", i)
			}
		}
		fmt.Println("  [任务 3] 完成！")
		taskDone <- 3
	}()

	// 等待所有任务结束
	for i := 0; i < 3; i++ {
		<-taskDone
	}

	fmt.Println("\n===== 总结 =====")
	fmt.Println("context 形成一棵树：每个 context 都有一个父节点")
	fmt.Println("级联取消：父节点取消 → 所有子孙节点自动取消")
	fmt.Println("这是 context 最强大的特性：一层取消，层层响应")
}

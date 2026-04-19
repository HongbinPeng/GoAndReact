package main

import (
	"context"
	"fmt"
	"time"
)

// 运行方式：
// go run main.go
//
// 说明：
// context 不是第三方库，而是 Go 标准库中的一个包。
// 它最常见的用途有三类：
// 1. 在函数调用链中传递“取消信号”
// 2. 控制超时时间和截止时间
// 3. 在多个 goroutine 之间协调“任务何时应该停止”
//
// 在“服务健康探测器”这类作业里，context 最重要的用途就是超时控制：
// 如果某个服务一直不返回，我们不希望整个程序一直卡住等待。
//
// 你可以把 context 理解成：
// “一个可以向下游函数广播消息的控制器”
// 这个消息常见是：
// - 超时了，别等了
// - 上层取消任务了，停下来
// - 当前请求已经结束了，相关子任务也该结束了

func main() {
	fmt.Println("========== context 标准库演示 ==========")

	// context.Background() 通常作为根上下文使用。
	// 它本身不会自动取消，也没有超时时间。
	// 一般在 main 函数、初始化逻辑、顶层请求入口里作为“父 context”使用。
	//
	// context.WithTimeout(parent, d) 会基于父 context 创建一个“子 context”。
	// 这个子 context 有两个关键特征：
	// 1. 超过 d 这个时间后会自动取消
	// 2. 会返回一个 cancel 函数，允许我们手动提前取消
	//
	// 这里的意思是：
	// 给这次探测任务设置 150ms 的最大等待时间。
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)

	// 为什么要 defer cancel()？
	// 因为 WithTimeout / WithCancel / WithDeadline 这类函数内部可能会创建定时器等资源。
	// 当任务提前结束时，及时调用 cancel 可以更快释放这些资源。
	//
	// 即使超时后 context 会自动结束，仍然推荐手动 defer cancel()，
	// 这是 Go 中使用 context 的一个很重要的习惯。
	defer cancel()

	// simulateProbe 模拟一个“实际工作需要 300ms 才能完成”的探测任务。
	// 但我们给它的 context 只允许等待 150ms，
	// 所以这个任务应该在超时后提前返回错误，而不是傻等到 300ms。
	err := simulateProbe(ctx, 300*time.Millisecond)
	if err != nil {
		fmt.Println("探测结果：", err)
	}

	fmt.Println("\n再看一个在超时前完成的任务：")

	// 第二次我们把超时时间设置成 500ms，
	// 而模拟任务只需要 80ms 就能完成。
	// 所以这一次任务应该能正常成功返回。
	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	if err := simulateProbe(ctx2, 80*time.Millisecond); err != nil {
		fmt.Println("探测失败：", err)
		return
	}
	fmt.Println("探测成功：任务在超时前完成")
}

// simulateProbe 用来模拟一个“可能会超时的任务”。
//
// 参数说明：
// - ctx：上层传进来的控制信号，里面可能带有超时信息或取消信号
// - workDuration：这个模拟任务本来需要执行多久
//
// 返回值说明：
// - 返回 nil：说明任务在 context 取消之前就完成了
// - 返回 ctx.Err()：说明任务被 context 中断了
//
// 真实项目里，这里通常对应：
// - 一个 HTTP 请求
// - 一个 TCP 拨号
// - 一个数据库查询
// - 一个并发子任务
func simulateProbe(ctx context.Context, workDuration time.Duration) error {
	// select 的作用是“同时等待多个事件，谁先发生就处理谁”。
	//
	// 这里同时等待两个事件：
	// 1. time.After(workDuration)
	//    表示“任务自然完成”
	// 2. ctx.Done()
	//    表示“context 发来了取消信号”
	//
	// 谁先准备好，就进入哪个 case。
	select {
	// time.After 会在指定时间后返回一个 channel，
	// 到时间时这个 channel 会收到一个时间值。
	//
	// 这里我们并不关心收到的具体时间值，
	// 只关心“它代表任务已经耗时 workDuration 并完成了”。
	case <-time.After(workDuration):
		// 走到这里，说明任务先完成了，没有被超时打断。
		return nil

		// ctx.Done() 会返回一个只读 channel。
		// 当 context 被取消、超时、或者父 context 已经结束时，
		// 这个 channel 会被关闭。
		//
		// 在 select 里监听它，是使用 context 最核心的方式之一。
	case <-ctx.Done():
		// ctx.Err() 用来告诉我们“为什么被结束”。
		// 常见返回值有：
		// - context.DeadlineExceeded：超时了
		// - context.Canceled：被主动取消了
		//
		// 这比直接返回一个固定字符串更有价值，
		// 因为调用方可以根据错误类型做进一步处理。
		return ctx.Err()
	}
}

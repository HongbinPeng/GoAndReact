package main

// ============================================================
// 第 10 课：最佳实践和常见陷阱
// ============================================================
//
// 【本课重点】
//   汇总前面 9 课中最重要的规则和常见错误。
//   这课没有新的 API，但内容非常实用。
//
// 【运行方式】
//   go run 10_best_practices/main.go

import (
	"context"
	"fmt"
	"time"
)

func main() {
	// ================================================================
	// 规则 1：永远调用 cancel()
	// ================================================================

	fmt.Println("===== 规则 1：永远调用 cancel() =====")

	// WithCancel / WithTimeout / WithDeadline 都返回一个 cancel 函数。
	// 你必须调用它，否则内部 goroutine 会泄漏。
	//
	// 标准写法：紧跟在创建语句之后，用 defer 调用。
	//
	//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//   defer cancel()  // ← 紧跟这一句！

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // ✅ 正确

	// ❌ 错误示例（不要这样做）：
	//   ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	//   ... 做了很多事情 ...
	//   cancel() // ← 万一中间 return 了，cancel 就不会被调用

	fmt.Println("✅ defer cancel() 紧跟在创建语句之后")
	fmt.Println("即使超时了，cancel 也要调用（释放内部资源）")

	_ = ctx

	// ================================================================
	// 规则 2：不要把 context 放在结构体中
	// ================================================================

	fmt.Println("\n===== 规则 2：不要把 context 放在结构体中 =====")

	// context 应该作为函数的第一个参数传递，而不是存在结构体字段里。
	//
	//   type Server struct {
	//       ctx context.Context  // ❌ 不好！
	//   }
	//
	//   func NewServer(ctx context.Context) *Server {  // ✅ 好！
	//       return &Server{}
	//   }
	//
	//   func (s *Server) DoWork(ctx context.Context) {  // ✅ 好！
	//       // ctx 作为参数传入，而不是用 s.ctx
	//   }

	fmt.Println("❌ 不好: type Server struct { ctx context.Context }")
	fmt.Println("✅ 好:   func (s *Server) DoWork(ctx context.Context)")
	fmt.Println("context 应该是函数的参数，不是结构体的字段")

	// ================================================================
	// 规则 3：context 是不可变的
	// ================================================================

	fmt.Println("\n===== 规则 3：context 是不可变的 =====")

	// 一旦创建了 context，就不能修改它的值、超时或取消函数。
	// 如果需要"修改"，就创建一个新的子 context。
	//
	//   ctx := context.Background()
	//   ctx = context.WithValue(ctx, "key", "value1")  // 创建新的
	//   ctx = context.WithValue(ctx, "key", "value2")  // 再创建新的，覆盖前一个
	//
	//   注意：WithValue 不会修改原来的 ctx，它返回一个新的。
	//   如果你忽略返回值，原来的 ctx 不会有任何变化。

	original := context.WithValue(context.Background(), "key", "original")
	modified := context.WithValue(original, "key", "modified")

	fmt.Printf("original: %v\n", original.Value("key")) // "original"
	fmt.Printf("modified: %v\n", modified.Value("key")) // "modified"

	fmt.Println("context 是不可变的，修改 = 创建新的")

	// ================================================================
	// 规则 4：不要传递 nil context
	// ================================================================

	fmt.Println("\n===== 规则 4：不要传递 nil context =====")

	// 所有接受 context 的函数都期望它不是 nil。
	// 传 nil 会导致 panic。
	//
	//   someFunction(nil)  // ❌ panic！
	//   someFunction(context.Background())  // ✅
	//
	// 如果不确定该用哪个，用 context.TODO()。

	// 如果你的函数接受 context，最好也检查 nil：
	//   func doSomething(ctx context.Context) {
	//       if ctx == nil {
	//           ctx = context.Background()
	//       }
	//       ...
	//   }

	fmt.Println("不要传 nil 给接受 context 的函数")
	fmt.Println("不确定时用 context.TODO() 代替")

	// ================================================================
	// 规则 5：context 只传元数据，不传函数参数
	// ================================================================

	fmt.Println("\n===== 规则 5：context 只传元数据，不传函数参数 =====")

	// ❌ 不好的写法：
	//   ctx := context.WithValue(context.Background(), "db", myDB)
	//   ctx = context.WithValue(ctx, "config", myConfig)
	//   process(ctx)
	//   func process(ctx context.Context) {
	//       db := ctx.Value("db").(*sql.DB)     // 糟糕！
	//       config := ctx.Value("config").(*Config) // 糟糕！
	//   }
	//
	// ✅ 好的写法：
	//   process(myDB, myConfig)
	//   func process(db *sql.DB, config *Config) {
	//       // 直接用参数，清晰明了
	//   }

	fmt.Println("❌ 不要把 db、config 等业务数据放在 context 里")
	fmt.Println("✅ 用函数参数传递业务数据")
	fmt.Println("✅ 用 context 传递请求范围的元数据（trace_id、user_token 等）")

	// ================================================================
	// 规则 6：用 select + Done 检查取消
	// ================================================================

	fmt.Println("\n===== 规则 6：用 select + ctx.Done() 检查取消 =====")

	// 检查 context 是否被取消的标准模式：
	//
	//   for {
	//       select {
	//       case <-ctx.Done():
	//           return  // 被取消，退出
	//       case <-time.After(1 * time.Second):
	//           // 继续工作
	//       case data := <-dataChan:
	//           // 处理数据
	//       }
	//   }
	//
	// 如果是简单的函数调用（不需要阻塞等待），可以用 ctx.Err() 快速检查：
	//
	//   func doWork(ctx context.Context) error {
	//       for i := 0; i < 100; i++ {
	//           if ctx.Err() != nil {  // 快速检查
	//               return ctx.Err()
	//           }
	//           processItem(i)
	//       }
	//       return nil
	//   }

	ctx6, cancel6 := context.WithCancel(context.Background())

	// 演示 ctx.Err() 快速检查
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel6()
	}()

	for i := 0; i < 10; i++ {
		if ctx6.Err() != nil {
			fmt.Printf("  第 %d 次迭代时被取消: %v\n", i, ctx6.Err())
			break
		}
		fmt.Printf("  第 %d 次迭代\n", i)
		time.Sleep(50 * time.Millisecond)
	}

	// ================================================================
	// 规则 7：context 不是同步机制
	// ================================================================

	fmt.Println("\n===== 规则 7：context 不是同步机制 =====")

	// context 不是用来做 goroutine 同步的。它只负责取消和超时。
	//
	// ❌ 不好：用 context 等待结果
	//   ctx, cancel := context.WithCancel(context.Background())
	//   go func() {
	//       result := doWork()
	//       ctx = context.WithValue(ctx, "result", result)  // 不好！
	//       cancel()
	//   }()
	//   <-ctx.Done()
	//   result := ctx.Value("result").(int)
	//
	// ✅ 好：用 channel 传递结果
	//   resultCh := make(chan int)
	//   go func() {
	//       resultCh <- doWork()
	//   }()
	//   result := <-resultCh

	fmt.Println("context 只负责取消和超时")
	fmt.Println("goroutine 同步和结果传递用 channel")

	// ================================================================
	// 速查表
	// ================================================================

	fmt.Println("\n===== context 速查表 =====")

	fmt.Println(`
┌──────────────────────────────────────────────────────────────┐
│ context 速查表                                                 │
├──────────────────────────┬───────────────────────────────────┤
│ 创建根 context            │ context.Background()              │
│ 可手动取消                 │ WithCancel(parent)                │
│ 超时自动取消               │ WithTimeout(parent, duration)     │
│ 指定时刻取消               │ WithDeadline(parent, time.Time)   │
│ 携带值                     │ WithValue(parent, key, value)     │
│ 检查取消                   │ <-ctx.Done() 或 ctx.Err()        │
│ 取消原因                   │ context.Canceled / DeadlineExceeded│
│                                                                  │
│ 规则:                                                              │
│ 1. defer cancel() 紧跟创建语句                                    │
│ 2. context 作为函数参数传递                                       │
│ 3. context 不可变                                                  │
│ 4. 不要传 nil context                                             │
│ 5. 只传元数据，不传业务参数                                       │
│ 6. 用 select + Done 检查取消                                      │
│ 7. context 不是同步机制，同步用 channel                            │
└──────────────────────────────────────────────────────────────┘`)

	fmt.Println("\n===== 总结 =====")
	fmt.Println("context 是 Go 中管理 goroutine 生命周期的核心工具")
	fmt.Println("掌握它，你就能写出更安全、更高效的并发程序")
}

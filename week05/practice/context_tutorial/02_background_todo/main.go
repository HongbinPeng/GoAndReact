package main

// ============================================================
// 第 02 课：context.Background() 和 context.TODO()
// ============================================================
//
// 【这两个是什么？】
//   它们是 context 包的"根"，所有其他 context 都是从它们衍生出来的。
//   它们本质上是同一个类型的两个实例：都不能被取消，都没有超时，都没有值。
//
// 【区别只有语义】
//   context.Background()  — "我就是要从这个开始"（通常用于 main 函数、测试、顶层函数）
//   context.TODO()        — "我还不确定用哪个，先放这里"（表示后续需要修改）
//
//   在代码层面，它们的行为完全一样！区别只是给读者看的。
//
// 【运行方式】
//   go run 02_background_todo/main.go

import (
	"context"
	"fmt"
)

func main() {
	fmt.Println("===== context.Background() =====")

	// Background 返回一个空的 Context，它不能被取消，没有值，没有超时。
	// 它是所有 context 的起点，通常用作最顶层的 context。
	ctx := context.Background()

	// 验证它的 4 个方法
	deadline, ok := ctx.Deadline()
	fmt.Printf("Deadline: %v, ok: %v\n", deadline, ok)
	//   deadline 是零值时间（0001-01-01），ok 是 false → 没有超时

	fmt.Printf("Done channel: %v\n", ctx.Done())
	//   nil → 永远不会被取消

	fmt.Printf("Err: %v\n", ctx.Err())
	//   nil → 没有错误

	fmt.Printf("Value: %v\n", ctx.Value("key"))
	//   nil → 没有值

	fmt.Println("\n===== context.TODO() =====")

	// TODO 和 Background 的行为完全一样，区别只在语义：
	//   - 当你知道这里应该是一个顶层 context → 用 Background()
	//   - 当你还没想好这里该用什么 context → 用 TODO()，提醒自己以后要改

	ctx2 := context.TODO()
	fmt.Printf("TODO 的 Done channel: %v\n", ctx2.Done()) // nil

	fmt.Println("\n===== 典型用法：作为所有 context 的根 =====")

	// 所有其他 context 都需要一个父 context，最终都会追溯到 Background：
	//
	//   ctx := context.Background()           ← 根
	//     └── ctx1 := context.WithCancel(ctx)      ← 第一层：可取消
	//          ├── ctx2 := context.WithTimeout(ctx1, 5s)  ← 第二层：带超时
	//          └── ctx3 := context.WithValue(ctx1, "user", "tom")  ← 第二层：带值

	fmt.Println("context.Background() 是所有 context 的根")
	fmt.Println("其他所有 context 都是从这里衍生出来的")

	fmt.Println("\n===== 什么时候用 Background？什么时候用 TODO？ =====")
	//
	//   用 Background()：
	//     - main 函数中启动程序时
	//     - 测试函数的顶层
	//     - 守护进程的循环中
	//     - 初始化代码中
	//
	//     示例：
	//       func main() {
	//           ctx := context.Background()  // ← 明确的顶层
	//           server.Start(ctx)
	//       }
	//
	//   用 TODO()：
	//     - 重构代码时，还没想好该传哪个 context
	//     - 调用一个需要 context 的函数，但当前代码没有合适的 context
	//
	//     示例：
	//       func (s *Server) HandleRequest() {
	//           // 这里应该从请求的上下文中获取 context，
	//           // 但还没重构好，先用 TODO() 占位
	//           db.Query(context.TODO(), "SELECT ...")
	//           // ↑ TODO 提醒：以后要改成从请求中获取 context
	//       }

	fmt.Println("总结：两者行为完全一样，区别只是语义")
	fmt.Println("Background() → 明确的顶层")
	fmt.Println("TODO() → 暂定，后续要改")
}

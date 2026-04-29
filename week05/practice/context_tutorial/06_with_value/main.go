package main

// ============================================================
// 第 06 课：WithValue — 在 context 中传递数据
// ============================================================
//
// 【WithValue 做什么？】
//   创建一个携带键值对的 context。子 context 可以通过 key 读取父 context 中的值。
//
// 【函数签名】
//   func WithValue(parent Context, key, val any) Context
//
//   注意：WithValue 没有返回 cancel 函数！因为它不可取消。
//
// 【重要警告】
//   context.Value 不是用来传递函数参数的！
//   它只应该传递"请求范围的元数据"，比如：
//     - 请求追踪 ID
//     - 用户 ID / 身份令牌
//     - 日志标签
//
//   不要把业务数据放在 context 中（比如数据库连接、配置参数等）。
//
// 【运行方式】
//   go run 06_with_value/main.go

import (
	"context"
	"fmt"
)

// ---- 最佳实践：用自定义类型作为 key ----
//
// 为什么要用自定义类型？因为如果两个包都用了相同的 key（比如字符串 "user"），
// 后面的会覆盖前面的。用未导出的自定义类型可以避免冲突。

// traceIDKey 是一个未导出的自定义类型，只有本包可以访问
type traceIDKey struct{}

func main() {
	fmt.Println("===== WithValue 基本用法 =====")

	// ---- 创建携带值的 context ----
	// 方式 1：用字符串作为 key（不推荐，容易冲突）
	ctx := context.WithValue(context.Background(), "user", "张三")
	ctx = context.WithValue(ctx, "role", "admin")

	// 读取值
	user := ctx.Value("user")
	role := ctx.Value("role")
	fmt.Printf("user: %v, role: %v\n", user, role)

	// ---- 不存在的 key 返回 nil ----
	missing := ctx.Value("email")
	fmt.Printf("email: %v（不存在的 key 返回 nil）\n", missing)

	fmt.Println("\n===== 最佳实践：用自定义类型作为 key =====")

	// ✅ 推荐做法：用自定义未导出类型作为 key
	ctx2 := context.WithValue(context.Background(), traceIDKey{}, "trace-12345")

	// 定义一个辅助函数来获取追踪 ID
	getTraceID := func(ctx context.Context) string {
		if v := ctx.Value(traceIDKey{}); v != nil {
			return v.(string)
		}
		return "unknown"
	}

	fmt.Printf("traceID: %v\n", getTraceID(ctx2))

	// 如果另一个包也定义了它自己的 keyType，不会和我们的冲突
	// 因为它们是不同的类型（即使底层都是 struct{}）

	fmt.Println("\n===== context 的值传递是单向的 =====")

	// 子 context 可以读父 context 的值，但反过来不行。
	// 同层级的 context 不能互相读取值。

	parentCtx := context.WithValue(context.Background(), traceIDKey{}, "parent-trace")

	// 子 context 可以读到父 context 的值
	childCtx := context.WithValue(parentCtx, "child_key", "child-value")

	fmt.Printf("子 context 读父的值: traceID=%v\n", getTraceID(childCtx))
	fmt.Printf("子 context 读自己的值: child_key=%v\n", childCtx.Value("child_key"))

	// 父 context 读不到子 context 的值
	fmt.Printf("父 context 读子的值: child_key=%v（nil，读不到）\n", parentCtx.Value("child_key"))

	fmt.Println("\n===== 常见陷阱 =====")

	// ---- 陷阱 1：用 context 传递函数参数 ----
	//
	//   // ❌ 不好的写法
	//   func process(ctx context.Context) {
	//       name := ctx.Value("name").(string)  // 不应该这样！
	//   }
	//
	//   // ✅ 好的写法
	//   func process(name string) {
	//       // 用函数参数传递业务数据
	//   }
	//
	// context 不是函数参数的替代品。如果函数需要一个 name 参数，
	// 就把它作为参数传递，不要塞进 context 里。

	// ---- 陷阱 2：key 类型冲突 ----
	//
	//   ctx1 := context.WithValue(ctx, "token", "abc")  // 包 A 用了 "token"
	//   ctx2 := context.WithValue(ctx1, "token", "xyz") // 包 B 也用了 "token" → 冲突！
	//
	//   // 解决办法：用自定义类型
	//   type authToken struct{}
	//   ctx := context.WithValue(parent, authToken{}, "abc")

	// ---- 陷阱 3：context 不适合存储大量数据 ----
	//
	//   context 的设计初衷是传递少量元数据，不是用来替代数据库或缓存的。
	//   如果你需要在 context 中存大量数据，说明设计有问题。

	fmt.Println("陷阱 1：不要用 context 传递函数参数")
	fmt.Println("陷阱 2：用自定义类型作为 key 避免冲突")
	fmt.Println("陷阱 3：context 只适合存少量元数据")

	fmt.Println("\n===== 典型应用场景 =====")

	// 场景 1：请求追踪
	fmt.Println("  场景 1：请求追踪 ID（每个请求带一个唯一 ID，贯穿所有子调用）")

	// 场景 2：认证信息
	fmt.Println("  场景 2：用户身份令牌（中间件解析令牌，下游 handler 直接使用）")

	// 场景 3：日志标签
	fmt.Println("  场景 3：日志标签（在 context 中携带 request_id，每条日志都带上）")

	fmt.Println("\n===== 总结 =====")
	fmt.Println("WithValue(parent, key, val) → 携带值的 context")
	fmt.Println("值传递是单向的：子→父，不能反过来")
	fmt.Println("key 用自定义未导出类型，避免冲突")
	fmt.Println("只传递元数据，不传递函数参数")
}

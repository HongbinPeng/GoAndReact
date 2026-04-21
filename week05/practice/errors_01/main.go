// errors_01 main 包
// 本文件通过代码实例 + 详细注释，讲解 Go 语言 errors 标准库的用法和底层原理
package main

import (
	"errors"
	"fmt"
)

// 运行方式：
// go run main.go

// =============================================================================
// errors 标准库 —— 总览
// =============================================================================
//
// 【这个包是干什么的？】
// errors 包提供了 Go 语言错误处理的核心工具。
//
// Go 的错误处理哲学和 Java/Python 等语言不同：
//   - 没有 try/catch，错误是普通的返回值（error 类型）
//   - 每个可能出错的函数都返回 (result, error)
//   - 调用方负责检查 error 是否为 nil
//
// 【作业里为什么必须学它？】
// 监控器作业中，错误来源非常多：
//   - 网络请求失败（连接超时、DNS 解析失败、TLS 握手失败）
//   - HTTP 状态码不符合预期（404、500）
//   - TCP 连接被拒绝
//   - 配置文件格式错误
//   - JSON 解析失败
//
// 如果不学会"带上下文地包装错误"，你最终只能打印出一个干巴巴的 "timeout"，
// 根本不知道是哪个 URL 超时了。
//
// 【Go 1.13 引入的错误包装机制】
//   - fmt.Errorf("...: %w", err) → 把原始错误包装一层，加入上下文
//   - errors.Is(err, target)    → 检查错误链中是否包含 target（支持包装）
//   - errors.As(err, &target)   → 从错误链中提取特定类型的错误
//   - errors.Unwrap(err)        → 获取被包装的原始错误
//
// 【error 接口的底层】
//   type error interface {
//       Error() string
//   }
//
// 就这么简单——任何实现了 Error() string 方法的类型都是 error。
// errors.New() 返回的是一个实现了这个接口的结构体：
//
//   type errorString struct {
//       s string
//   }
//   func (e *errorString) Error() string { return e.s }
//
// 所以 errors.New("超时") 本质上返回的是 &errorString{s: "超时"}。

// =============================================================================
// 第一部分：创建错误
// =============================================================================

// --- 方式 1：errors.New ---
//
// errors.New 创建一个"纯文本"错误，没有任何额外上下文。
// 适用于：错误本身已经足够描述问题，不需要额外信息。
//
// 底层实现（源码位置：errors/errors.go）：
//
//   func New(text string) error {
//       return &errorString{text}
//   }
//
//   type errorString struct {
//       s string
//   }
//   func (e *errorString) Error() string { return e.s }
//
// 特点：
//   - 轻量，就是分配一个带字符串的小结构体
//   - 没有堆栈信息（Go 的错误不包含堆栈，不像 Java 的 Exception）
//   - 相同的 errors.New("timeout") 调用返回不同的错误对象（地址不同）
//
// 所以你不能这样判断：
//   err1 := errors.New("timeout")
//   err2 := errors.New("timeout")
//   err1 == err2  ← false！虽然文本一样，但地址不同
//
// 这就是为什么下面要用包级变量（见 ErrProbeTimeout）。

// --- 方式 2：fmt.Errorf ---
//
// fmt.Errorf 创建一个带格式化上下文的错误。
// 最常用的形式：fmt.Errorf("描述信息: %w", 原始错误)
//
// %w 是 Go 1.13 引入的特殊格式动词（wrap）：
//   - 它和 %v 一样把错误转成字符串
//   - 但它会在返回的错误对象中"记住"原始错误的引用
//   - 这样 errors.Is 和 errors.As 就能沿着错误链找到原始错误
//
// 底层原理：
//   fmt.Errorf("...: %w", err) 返回的是一个实现了以下接口的结构体：
//
//     type wrapper struct {
//         msg string     // 格式化后的消息
//         err error      // 被包装的原始错误
//     }
//     func (w *wrapper) Error() string { return w.msg }
//     func (w *wrapper) Unwrap() error { return w.err }  ← 关键！
//
//   Unwrap() 方法让 errors.Is 能"剥开"包装层，看到原始错误。

// --- 方式 3：自定义错误类型 ---
//
// 对于复杂场景，可以定义自己的错误类型（实现 error 接口）：
//
//   type TimeoutError struct {
//       Address string
//       Timeout time.Duration
//   }
//   func (e *TimeoutError) Error() string {
//       return fmt.Sprintf("探测 %s 超时（%v）", e.Address, e.Timeout)
//   }
//
// 这种方式可以携带结构化信息（地址、超时时间等），
// 调用方用 errors.As 可以提取出来做针对性处理。
// 作业里如果不需要这么复杂，用 errors.New + fmt.Errorf 就够了。

// =============================================================================
// 包级 sentinel 错误
// =============================================================================

// ErrProbeTimeout 是一个"哨兵错误"（sentinel error）。
//
// 【什么是哨兵错误？】
// 哨兵错误是包级导出的、预定义的错误值。调用方可以用 errors.Is 来精确判断。
//
// 命名惯例：以 Err 开头 + 大驼峰，如 ErrNotFound、ErrTimeout、ErrInvalidInput。
//
// 【为什么必须用包级变量，而不是每次 errors.New？】
//
// ❌ 错误做法：
//
//	func probe() error {
//	    return errors.New("探测超时")  // 每次调用都创建新对象
//	}
//	// 调用方无法用 errors.Is 判断，因为地址每次都不同
//
// ✅ 正确做法：
//
//	var ErrProbeTimeout = errors.New("探测超时")  // 全局唯一对象
//	func probe() error {
//	    return fmt.Errorf("...: %w", ErrProbeTimeout)  // 包装它
//	}
//	// 调用方用 errors.Is(err, ErrProbeTimeout) 就能正确匹配
//
// 【作业中你需要定义哪些哨兵错误？】
//
//	var ErrProbeTimeout    = errors.New("探测超时")      // 网络超时
//	var ErrProbeConnection = errors.New("连接被拒绝")     // TCP 连接失败
//	var ErrStatusCode      = errors.New("HTTP 状态码异常") // 状态码不符合预期
//	var ErrBodyMismatch    = errors.New("响应体不匹配")    // 关键词检查失败
//
// 这样 main 函数里可以针对不同错误做不同处理：
//
//	if errors.Is(err, ErrProbeTimeout) {
//	    // 超时 → 标记为慢服务，可能需要告警
//	} else if errors.Is(err, ErrStatusCode) {
//	    // 状态码异常 → 可能服务挂了
//	}
var ErrProbeTimeout = errors.New("探测超时")

// =============================================================================
// 第二部分：错误包装 —— 加入上下文
// =============================================================================
//
// 【为什么需要包装错误？】
//
// 假设网络层返回了 "dial tcp 127.0.0.1:3306: connection refused"，
// 如果直接把这个错误一路返回给 main 函数，用户看到的就是这串原始信息。
// 但用户想知道的是：**哪个服务**连接被拒绝了？
//
// 包装错误就是在每一层都加上自己的上下文：
//
//   网络层：    "dial tcp 127.0.0.1:3306: connection refused"
//   探测层包装： "探测地址 https://xxx 失败：dial tcp ... connection refused"
//   调度层包装： "目标 [MySQL] 探测失败：探测地址 ... 失败：dial tcp ..."
//
// 最终打印出来是一条完整的错误链，一看就知道哪个服务的什么问题。
//
// 【%w 和 %v 的区别】
//
//   fmt.Errorf("失败: %v", err)  ← 只拼字符串，错误链断裂，errors.Is 无法追踪
//   fmt.Errorf("失败: %w", err)  ← 保留错误引用，错误链完整，errors.Is 可以追踪
//
// 作业中一定要用 %w，不要用 %v。

// =============================================================================
// 第三部分：错误匹配 —— errors.Is 和 errors.As
// =============================================================================
//
// 【errors.Is(err, target)】
//
// 沿着错误链逐层 Unwrap()，检查是否有某一层等于 target。
//
// 伪代码实现：
//
//   func Is(err, target error) bool {
//       for err != nil {
//           if err == target { return true }      // 精确匹配（地址相等）
//           if x, ok := err.(interface{ Unwrap() error }); ok {
//               err = x.Unwrap()                   // 剥开一层包装
//           } else {
//               return false                       // 没有 Unwrap，链断了
//           }
//       }
//       return false
//   }
//
// 所以即使你的错误被包装了三层，errors.Is 也能找到最内层的哨兵错误。
//
// 【errors.As(err, &target)】
//
// 沿着错误链查找特定类型的错误，并把值填到 target 中。
// 适用于自定义错误类型（携带结构化信息的场景）。
//
// 示例：
//
//   type HTTPError struct {
//       StatusCode int
//       URL        string
//   }
//   func (e *HTTPError) Error() string { ... }
//
//   var httpErr *HTTPError
//   if errors.As(err, &httpErr) {
//       fmt.Printf("HTTP 错误，状态码 %d，URL: %s\n", httpErr.StatusCode, httpErr.URL)
//   }
//
// 作业里主要用 errors.Is 判断哨兵错误就足够了。

// =============================================================================
// 演示代码
// =============================================================================

func main() {
	fmt.Println("========== errors 标准库演示 ==========")

	// --- 演示 1：基本错误包装与匹配 ---
	fmt.Println("\n--- 演示 1：错误包装 + errors.Is ---")

	err := probeService("https://httpbin.org/delay/5", true)
	if err != nil {
		// err 的内容：探测地址 https://httpbin.org/delay/5 失败：探测超时
		// 它是一个被 fmt.Errorf("%w") 包装过的错误
		fmt.Println("收到错误：", err)

		// errors.Is 沿着错误链查找，发现内层包装了 ErrProbeTimeout
		// 即使 err 不是直接等于 ErrProbeTimeout（中间隔了一层包装），也能匹配
		if errors.Is(err, ErrProbeTimeout) {
			fmt.Println("判断结果：这是一个超时错误")
		}
	}

	// --- 演示 2：错误链的逐层拆解 ---
	fmt.Println("\n--- 演示 2：错误链拆解 ---")

	err2 := simulateMultiLayerError()
	fmt.Println("完整错误信息：", err2)
	fmt.Println("是超时错误吗？", errors.Is(err2, ErrProbeTimeout))

	// Unwrap 可以手动剥开包装层
	unwrapped := errors.Unwrap(err2)
	fmt.Println("剥开第一层：", unwrapped)

	unwrapped2 := errors.Unwrap(unwrapped)
	fmt.Println("剥开第二层：", unwrapped2)
	fmt.Println("第二层就是 ErrProbeTimeout 本身吗？", unwrapped2 == ErrProbeTimeout)

	// --- 演示 3：errors.Is 的精确性 ---
	fmt.Println("\n--- 演示 3：区分不同错误 ---")

	err3 := probeService("https://example.com", false)
	if err3 != nil {
		fmt.Println("错误：", err3)
	} else {
		fmt.Println("探测成功，err == nil")
	}

	// --- 演示 4：错误比较的常见误区 ---
	fmt.Println("\n--- 演示 4：错误比较的常见误区 ---")

	errA := errors.New("timeout")
	errB := errors.New("timeout")
	fmt.Printf("errA == errB: %v（虽然文本一样，但是两个不同的对象）\n", errA == errB)

	// 正确做法：用 errors.Is 配合哨兵错误
	fmt.Printf("errors.Is(errA, ErrProbeTimeout): %v\n", errors.Is(errA, ErrProbeTimeout))
}

// probeService 模拟一次服务探测。
// 这是作业中 probeHTTP 或 probeTCP 的简化版本。
//
// 返回值 error 的语义：
//   - nil     → 探测成功
//   - 非 nil  → 探测失败，错误信息包含失败原因和上下文
//
// 注意这里的错误处理模式——这是 Go 的标准做法：
// 在出错的位置立即包装错误并返回，不做多余的 if-else 嵌套。
func probeService(address string, simulateTimeout bool) error {
	if simulateTimeout {
		// fmt.Errorf 的 %w 把 ErrProbeTimeout 包装进去，形成错误链：
		//   外层："探测地址 https://httpbin.org/delay/5 失败"
		//   内层：ErrProbeTimeout（"探测超时"）
		//
		// 这样调用方既能看到完整的上下文（哪个地址失败了），
		// 又能用 errors.Is 精确判断错误类型（是超时还是其他）。
		//
		// 作业中你会这样写：
		//   resp, err := client.Do(req)
		//   if err != nil {
		//       return fmt.Errorf("探测地址 %s 失败: %w", target.Address, err)
		//   }
		return fmt.Errorf("探测地址 %s 失败：%w", address, ErrProbeTimeout)
	}
	return nil
}

// simulateMultiLayerError 模拟多层错误包装。
// 展示真实作业中错误从底层到顶层的传递过程。
func simulateMultiLayerError() error {
	// 最底层：网络库返回的原始错误
	networkErr := ErrProbeTimeout

	// 中间层：探测函数包装，加上地址信息
	probeErr := fmt.Errorf("探测地址 %s 失败：%w", "https://example.com", networkErr)

	// 顶层：调度函数包装，加上目标名称
	scheduleErr := fmt.Errorf("目标 [示例服务] 探测失败：%w", probeErr)

	// 最终返回的错误链：
	//   目标 [示例服务] 探测失败：探测地址 https://example.com 失败：探测超时
	//
	// 打印出来一目了然，errors.Is 也能正确匹配到 ErrProbeTimeout
	return scheduleErr
}

// =============================================================================
// 第四部分：作业中错误处理的推荐模式
// =============================================================================

// probeHTTP 作业中 HTTP 探测的错误处理示例。
// 这不是可运行代码，而是展示你应该怎么写。
func probeHTTPExample() {
	// /*
	// func probeHTTP(target Target, timeout time.Duration) ProbeResult {
	//     result := ProbeResult{Name: target.Name}
	//
	//     // 1. 创建带超时的 context
	//     ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//     defer cancel()
	//
	//     // 2. 创建请求
	//     req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.Address, nil)
	//     if err != nil {
	//         result.Error = fmt.Errorf("创建请求失败: %w", err).Error()
	//         return result
	//     }
	//
	//     // 3. 记录开始时间
	//     start := time.Now()
	//
	//     // 4. 发送请求
	//     resp, err := client.Do(req)
	//     result.Latency = time.Since(start).String()
	//
	//     if err != nil {
	//         // 核心：用 %w 包装错误，保留原始错误类型
	//         // context.DeadlineExceeded → 超时
	//         // net.Error.Timeout()     → 超时
	//         // 其他                    → 网络错误
	//         result.Error = fmt.Errorf("请求失败: %w", err).Error()
	//         return result
	//     }
	//     defer resp.Body.Close()
	//
	//     // 5. 检查状态码
	//     if target.ExpectCode > 0 && resp.StatusCode != target.ExpectCode {
	//         result.Error = fmt.Errorf("%w: 期望 %d，实际 %d",
	//             ErrStatusCode, target.ExpectCode, resp.StatusCode).Error()
	//         return result
	//     }
	//
	//     // 6. 检查响应体关键词
	//     if target.Contains != "" {
	//         body, _ := io.ReadAll(resp.Body)
	//         if !strings.Contains(string(body), target.Contains) {
	//             result.Error = fmt.Errorf("%w: 期望包含 %q",
	//                 ErrBodyMismatch, target.Contains).Error()
	//             return result
	//         }
	//     }
	//
	//     result.OK = true
	//     return result
	// }
	// */
}

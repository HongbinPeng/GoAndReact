// net_http_01 main 包
// 本文件通过代码实例 + 详细注释，讲解 Go 语言 net/http 标准库（客户端侧）的用法和底层原理
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// 运行方式：
// go run main.go

// =============================================================================
// net/http 标准库 —— 总览
// =============================================================================
//
// 【这个包是干什么的？】
// net/http 是 Go 标准库提供的 HTTP 协议实现，同时支持客户端和服务端。
// 它覆盖了 HTTP/1.1 协议的完整实现（HTTP/2 也支持，但默认不开启）。
//
// 【这个包分两大块】
//
//   客户端侧（这次作业的重点）：
//     http.Client    — HTTP 客户端，负责发送请求
//     http.Request   — HTTP 请求对象
//     http.Response  — HTTP 响应对象
//     http.Get/Post  — 快捷方法（底层也是用 Client）
//
//   服务端侧（作业暂不需要）：
//     http.Server    — HTTP 服务端
//     http.HandleFunc — 注册路由处理器
//     http.ListenAndServe — 启动服务监听
//
// 【为什么作业里必须学它？】
// 监控器作业的核心就是发 HTTP 请求探测服务是否健康：
//   1. 对每个 URL 发起 GET 或 HEAD 请求
//   2. 检查响应状态码（200 = 正常，404/500 = 异常）
//   3. 读取响应体，检查是否包含指定关键词
//   4. 设置超时，防止某个慢服务卡死整个程序
//   5. 记录耗时，统计哪个服务最慢
//
// 【底层架构简述】
//
//   你写的代码层：
//     client.Do(req)
//           │
//           ▼
//   http.Client（调度层）：
//     - 管理 Transport（连接池）
//     - 处理重定向（默认最多 10 次）
//     - 处理 Cookie（如果启用了 Jar）
//           │
//           ▼
//   http.Transport（传输层）：
//     - DNS 解析
//     - 建立 TCP/TLS 连接
//     - 发送 HTTP 请求字节流
//     - 接收 HTTP 响应字节流
//     - 连接复用（keep-alive）
//           │
//           ▼
//   net.Conn（网络层）：
//     - 底层 TCP 连接
//     - 读写 socket 数据
//
// 【http.Get vs http.Client.Do】
//
//   http.Get(url) 是最简单的用法，但它有一个大问题：
//     - 使用全局默认客户端（http.DefaultClient）
//     - 无法自定义超时时间
//     - 无法自定义 Transport（连接池配置）
//     - 无法绑定 context（精细超时控制）
//
//   所以作业中推荐用 http.Client + client.Do(req)：
//     - 可以设置 client.Timeout
//     - 可以用 http.NewRequestWithContext 绑定 context
//     - 可以自定义 Transport（连接池大小、TLS 配置等）
//     - 方便后续做重试控制

// =============================================================================
// net/http 里的关键结构体
// =============================================================================

// http.Client — HTTP 客户端（调度层）
//
//   type Client struct {
//       Transport   RoundTripper  // 传输层（默认是 http.DefaultTransport）
//       CheckRedirect func(*Request, []*Request) error  // 重定向策略
//       Jar         CookieJar     // Cookie 存储
//       Timeout     time.Duration // 整个请求的超时时间（包含连接+发送+接收+重定向）
//   }
//
// 核心字段解释：
//   Transport   — 负责实际的 HTTP 传输。默认使用 http.DefaultTransport，
//                 它内部有一个连接池，会自动复用 TCP 连接（keep-alive）。
//                 你可以自定义 Transport 来控制：最大空闲连接数、TLS 握手超时等。
//
//   Timeout     — 从请求开始到响应体全部读完的总超时时间。
//                 如果设为 0（默认值），表示永不超时（危险！）。
//                 作业中必须设置，否则一个死掉的服务会永远卡住 goroutine。
//
//   注意：Client 是并发安全的！多个 goroutine 可以同时调用同一个 Client 的 Do 方法。
//         作业中只需创建一个 Client，所有探测任务共用它（复用连接池）。

// http.Request — HTTP 请求
//
//   type Request struct {
//       Method string           // 请求方法："GET", "POST", "HEAD", ...
//       URL    *url.URL         // 目标 URL
//       Header Header           // 请求头（map[string][]string）
//       Body   io.ReadCloser    // 请求体（POST 时才需要）
//       Host   string           // Host 头（通常由 URL 自动填充）
//       // ... 还有十几个字段
//   }
//
// 创建方式：
//   http.NewRequest(method, url, body)           — 基本版
//   http.NewRequestWithContext(ctx, method, url, body) — 带 context 版（推荐）
//
// context 的作用：
//   - 超时取消：超过指定时间自动取消请求（即使连接已建立）
//   - 手动取消：调用 cancel() 可以立即取消进行中的请求
//   - 比 client.Timeout 更精细（可以针对不同请求设置不同超时）

// http.Response — HTTP 响应
//
//   type Response struct {
//       Status     string           // 状态行："200 OK"
//       StatusCode int              // 状态码：200, 404, 500, ...
//       Proto      string           // 协议版本："HTTP/1.1"
//       Header     Header           // 响应头
//       Body       io.ReadCloser    // 响应体（必须 Close！）
//       ContentLength int64         // 响应体长度（-1 表示未知）
//       Request    *Request         // 原始请求（方便追溯）
//       // ...
//   }
//
// 核心注意点：
//   resp.Body 是一个 io.ReadCloser，使用后必须 Close()。
//   不 Close 会导致连接泄漏（TCP 连接无法被连接池回收）。
//   标准写法：defer resp.Body.Close()

// =============================================================================
// 常见 HTTP 状态码（作业中需要判断的）
// =============================================================================
//
//   2xx — 成功
//     200 OK                    — 请求成功
//     201 Created               — 资源已创建
//     204 No Content            — 成功但无响应体
//
//   3xx — 重定向
//     301 Moved Permanently     — 永久重定向（http.Client 默认自动跟随）
//     302 Found                 — 临时重定向
//     304 Not Modified          — 缓存命中（无需重新下载）
//
//   4xx — 客户端错误
//     400 Bad Request           — 请求格式错误
//     401 Unauthorized          — 未认证
//     403 Forbidden             — 无权限
//     404 Not Found             — 资源不存在
//
//   5xx — 服务端错误
//     500 Internal Server Error — 服务器内部错误
//     502 Bad Gateway           — 网关错误
//     503 Service Unavailable   — 服务不可用（可能正在重启）
//     504 Gateway Timeout       — 网关超时（上游服务太慢）
//
// 作业中的判断逻辑：
//   - 200 ~ 299 → 探测成功
//   - 400 ~ 499 → 客户端配置错误（URL 不对、认证失败等）
//   - 500 ~ 599 → 服务端异常（服务可能挂了）
//   - 网络错误（err != nil） → 连接失败/超时

// =============================================================================
// httptest — 测试用的临时 HTTP 服务器
// =============================================================================
//
// 【为什么示例里用 httptest 而不是真实网站？】
//
//   1. 不依赖外网 — 离线也能运行示例
//   2. 完全可控 — 可以模拟各种场景（慢响应、500 错误、超时等）
//   3. 速度快 — 本地回环（localhost）比访问百度快得多
//   4. 测试友好 — 作业写单元测试时也会用到
//
// 【httptest.NewServer 做了什么？】
//
//   1. 在本地随机端口启动一个真实的 HTTP 服务器
//   2. 返回一个 *httptest.Server 对象
//   3. server.URL 是完整地址，如 "http://127.0.0.1:54321"
//   4. defer server.Close() 确保测试结束后释放端口
//
//   你可以在这个服务器里注册任意路由，模拟各种 HTTP 行为：
//     - 正常 200 响应
//     - 慢速响应（time.Sleep）
//     - 500 错误
//     - 大响应体
//     - 特定的响应头
//
// 作业写单元测试时，你会这样用：
//
//   func TestProbeHTTP(t *testing.T) {
//       server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//           w.WriteHeader(200)
//           w.Write([]byte("OK"))
//       }))
//       defer server.Close()
//
//       result := probeHTTP(Target{Address: server.URL})
//       if !result.OK {
//           t.Errorf("期望探测成功，实际失败")
//       }
//   }

// =============================================================================
// 第一部分：HTTP 探测 —— 完整流程拆解
// =============================================================================
//
// 一次完整的 HTTP 探测包含以下步骤：
//
//   步骤 1：创建带超时的 context
//     ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
//     defer cancel()
//     → 1 秒后如果请求还没完成，context 会自动取消
//
//   步骤 2：创建请求对象
//     req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
//     → 把 context 绑定到请求上，超时后请求自动终止
//     → http.MethodGet 是常量 "GET"，比手写字符串安全
//
//   步骤 3：发送请求
//     resp, err := client.Do(req)
//     → 这一步可能阻塞（等待 DNS 解析、TCP 握手、服务器响应）
//     → err != nil 表示网络层出错（超时、连接被拒绝、DNS 失败等）
//
//   步骤 4：检查 HTTP 状态码
//     if resp.StatusCode != 200 {
//         // 404、500 等 — 请求发出去了，但服务器返回了错误状态
//     }
//     → 注意：err == nil 不代表业务成功！404/500 时 err 也是 nil
//
//   步骤 5：读取响应体
//     body, err := io.ReadAll(resp.Body)
//     → io.ReadAll 一次性读完整个响应体
//     → 如果响应体很大（几 MB），可以用 io.LimitReader 限制读取量
//
//   步骤 6：关闭响应体
//     defer resp.Body.Close()
//     → 必须做！否则连接泄漏
//
//   步骤 7：处理响应内容
//     strings.Contains(string(body), "Go")
//     → 作业中用来检查响应体是否包含指定关键词

// =============================================================================
// 第二部分：超时控制的两种方式
// =============================================================================
//
// 方式一：client.Timeout（粗粒度）
//
//   client := &http.Client{Timeout: 5 * time.Second}
//   resp, err := client.Get(url)
//
//   特点：
//     - 简单，一行搞定
//     - 超时时间对所有请求统一生效
//     - 超时后返回的 error 类型是 *url.Error，可以用 errors.Is 判断
//
// 方式二：context.WithTimeout（细粒度）
//
//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//   defer cancel()
//   req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
//   resp, err := client.Do(req)
//
//   特点：
//     - 可以为每个请求设置不同的超时时间
//     - 支持手动取消（调用 cancel() 立即终止请求）
//     - 超时后返回的 error 是 context.DeadlineExceeded
//
// 作业中的推荐做法：两者结合
//
//   client := &http.Client{Timeout: 10 * time.Second}  // 兜底超时
//   ctx, cancel := context.WithTimeout(ctx, 3*time.Second)  // 精细控制
//   defer cancel()
//   req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
//
//   这样既有 per-request 的精细控制，又有 client 级别的兜底保护。

// =============================================================================
// 第三部分：连接复用与性能优化
// =============================================================================
//
// 【http.Transport 的连接池】
//
// http.Client 默认的 Transport 会自动复用 TCP 连接（HTTP keep-alive）：
//
//   client := &http.Client{}
//   // 第一次请求 → 建立新 TCP 连接
//   client.Get("http://example.com/api/1")
//   // 第二次请求 → 复用同一个 TCP 连接（省去了 TCP 握手 + TLS 握手）
//   client.Get("http://example.com/api/2")
//
// 如果你每个探测任务都创建一个新 Client：
//
//   for _, url := range urls {
//       client := &http.Client{}  // ❌ 每次都新建连接，性能差
//       client.Get(url)
//   }
//
// 正确做法：
//
//   client := &http.Client{}  // ✅ 只创建一个，所有请求共用
//   for _, url := range urls {
//       client.Get(url)
//   }
//
// 【自定义 Transport 的高级配置】
//
//   transport := &http.Transport{
//       MaxIdleConns:        100,              // 最大空闲连接数
//       MaxIdleConnsPerHost: 10,               // 每个 host 最大空闲连接数
//       IdleConnTimeout:     90 * time.Second, // 空闲连接多久后关闭
//       TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时
//       DisableKeepAlives:   false,            // 是否禁用连接复用
//   }
//   client := &http.Client{Transport: transport}
//
// 作业里用默认 Transport 就够了，不需要自定义。

// =============================================================================
// 第四部分：io 包在 HTTP 中的角色
// =============================================================================
//
// 【io.ReadAll】
//
//   body, err := io.ReadAll(resp.Body)
//
//   一次性读取 resp.Body 的全部内容到 []byte。
//   内部实现：先分配一个初始缓冲区，不够就扩容（类似 slice append）。
//
//   适用场景：响应体不大（几 KB ~ 几 MB）。
//
//   不适用场景：
//     - 响应体非常大（几百 MB）→ 内存爆炸
//     - 只需要检查前几个字节 → 浪费
//
// 【io.LimitReader】
//
//   limited := io.LimitReader(resp.Body, 1024)  // 最多读 1KB
//   body, _ := io.ReadAll(limited)
//
//   限制读取的最大字节数，防止恶意服务器返回超大响应体耗尽内存。
//   作业中如果只需要检查关键词，可以限制读取量。
//
// 【io.NopCloser】
//
//   把 io.Reader 包装成 io.ReadCloser（Close 方法什么都不做）。
//   主要用于测试场景，构造假的 resp.Body。

// =============================================================================
// 演示代码
// =============================================================================

func main() {
	// ------------------------------------------------------------------------
	// 用 httptest 启动一个本地临时 HTTP 服务器
	// ------------------------------------------------------------------------
	//
	// http.HandlerFunc 是一个适配器：把普通函数转成 http.Handler 接口。
	// 函数签名：func(http.ResponseWriter, *http.Request)
	//
	// 这个服务器模拟了三种场景：
	//   /ok      → 正常 200 响应，返回 "Go monitor is healthy"
	//   /slow    → 延迟 200ms 后返回（用于测试超时控制）
	//   其他路径 → 404 Not Found
	//
	// server.URL 类似 "http://127.0.0.1:54321"，每次运行端口可能不同。
	// defer server.Close() 确保 main 函数退出时释放端口。

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// w http.ResponseWriter — 写入响应的接口
		// r *http.Request       — 读取请求信息的结构体
		//
		// r.URL.Path 是请求路径，如 "/ok"、"/slow"、"/not-found"
		//
		// http.ResponseWriter 的两个核心方法：
		//   w.WriteHeader(code)  — 写入 HTTP 状态码（必须在 Write 之前调用）
		//   w.Write(data)        — 写入响应体字节
		//   w.Header()           — 获取/设置响应头（必须在 WriteHeader 之前设置）

		switch r.URL.Path {
		case "/ok":
			// 正常响应：200 OK + 响应体
			// WriteHeader 是可选的——如果不写，第一次 Write 时会自动写入 200
			// 但显式写出更清晰
			w.WriteHeader(http.StatusOK) // 200
			_, _ = w.Write([]byte("Go monitor is healthy"))
			// http.StatusOK 是常量 200，比手写数字安全
			// w.Write 返回 (n int, err error)，n 是实际写入的字节数

		case "/slow":
			// 模拟慢响应：延迟 200ms 后再返回
			// 用于测试客户端的超时控制是否生效
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("slow but ok"))
			// 注意：如果客户端超时时间 < 200ms，
			// 客户端会收到 context deadline exceeded 错误，
			// 但服务端（这里）仍然会继续执行完 Sleep + Write。
			// 只是客户端不再等待响应了。

		default:
			// 其他路径 → 404 Not Found
			// http.NotFound 是标准库提供的便捷函数：
			//   func NotFound(w ResponseWriter, r *Request) {
			//       w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			//       w.Header().Set("X-Content-Type-Options", "nosniff")
			//       w.WriteHeader(StatusNotFound)
			//       fmt.Fprintln(w, "404 page not found")
			//   }
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	fmt.Println("========== net/http 标准库演示 ==========")

	// 场景 1：探测正常服务（200 OK）
	probeHTTP(server.URL + "/ok")

	// 场景 2：探测不存在的路径（404 Not Found）
	probeHTTP(server.URL + "/not-found")

	// 场景 3：探测慢速服务（超时控制）
	probeSlowEndpoint(server.URL + "/slow")
}

// probeHTTP 模拟一次完整的 HTTP 探测。
// 这是作业中 probeHTTP 函数的简化版本。
//
// 参数 url：目标地址，如 "http://127.0.0.1:54321/ok"
//
// 返回值：无（实际作业中会返回 ProbeResult 结构体）
func probeHTTP(url string) {
	fmt.Printf("\n探测地址：%s\n", url)

	// ------------------------------------------------------------------------
	// 步骤 1：创建带超时的 context
	// ------------------------------------------------------------------------
	//
	// context.WithTimeout 返回两个值：
	//   ctx     — 新的 context，1 秒后自动触发 Done 信号
	//   cancel  — 取消函数，提前结束 context（必须 defer 调用！）
	//
	// context.Background() 是最顶层的 context，没有父级，没有超时，没有值。
	// 通常作为所有 context 树的根。
	//
	// cancel 必须调用！否则 context 内部的定时器 goroutine 会泄漏，
	// 直到超时时间到了才会被 GC 回收。
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// ------------------------------------------------------------------------
	// 步骤 2：创建 HTTP 客户端
	// ------------------------------------------------------------------------
	//
	// &http.Client{} 使用默认 Transport（带连接池）。
	// 没有设置 Timeout 字段 → 依赖 context 控制超时。
	//
	// 作业中你可以这样写：
	//   client := &http.Client{Timeout: 5 * time.Second}  // 兜底超时
	// 同时配合 context 做精细控制。
	client := &http.Client{}

	// ------------------------------------------------------------------------
	// 步骤 3：创建请求对象
	// ------------------------------------------------------------------------
	//
	// http.NewRequestWithContext 的参数：
	//   ctx    — 绑定超时控制的 context
	//   method — 请求方法，http.MethodGet = "GET"
	//   url    — 目标 URL
	//   body   — 请求体，GET 请求传 nil
	//
	// 可能的错误：
	//   - URL 格式不对（如缺少协议头 "http://"）
	//   - method 是空字符串
	//
	// http.MethodGet 是 Go 1.7+ 引入的常量，避免手写字符串出错。
	// 其他常量：http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodHead
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("创建请求失败：", err)
		return
	}

	// ------------------------------------------------------------------------
	// 步骤 4：记录开始时间（用于计算耗时）
	// ------------------------------------------------------------------------
	start := time.Now()

	// ------------------------------------------------------------------------
	// 步骤 5：发送请求
	// ------------------------------------------------------------------------
	//
	// client.Do(req) 是 HTTP 客户端的核心方法。
	// 它做了以下事情：
	//
	//   1. 检查 req.Context 是否已取消 → 是则立即返回错误
	//   2. 如果 URL 没有 Host，从 URL 提取
	//   3. 通过 Transport 获取/建立连接：
	//      a. 检查连接池中是否有可用的空闲连接
	//      b. 没有 → 新建 TCP 连接（含 DNS 解析 + 三次握手 + TLS 握手）
	//   4. 将 HTTP 请求写入连接
	//   5. 从连接读取 HTTP 响应
	//   6. 如果状态码是 3xx 且 CheckRedirect 允许 → 自动跟随重定向
	//   7. 返回 *http.Response
	//
	// 返回值：
	//   resp — HTTP 响应对象（即使 404/500 也会返回，err == nil）
	//   err  — 网络层错误（超时、连接被拒绝、TLS 失败、context 取消等）
	//          注意：HTTP 状态码错误（404/500）不算 err！
	resp, err := client.Do(req)
	if err != nil {
		// 常见网络错误：
		//   - context deadline exceeded  → 超时
		//   - dial tcp ... connection refused  → 端口没监听
		//   - dial tcp ... i/o timeout  → 连接超时
		//   - no such host  → DNS 解析失败
		//   - tls: ...  → TLS 握手失败
		fmt.Println("请求失败：", err)
		return
	}
	// 注意：走到这里 err == nil，但 resp.StatusCode 可能是 404/500！
	// 必须检查状态码才能知道业务是否成功。

	// ------------------------------------------------------------------------
	// 步骤 6：确保响应体被关闭
	// ------------------------------------------------------------------------
	//
	// resp.Body 实现了 io.ReadCloser 接口。
	// Close() 的作用是：
	//   - 释放 TCP 连接回连接池（如果不是 keep-alive 则关闭连接）
	//   - 释放内部缓冲区
	//
	// 不 Close 的后果：
	//   - 连接泄漏：连接池耗尽后，新请求无法建立连接
	//   - goroutine 泄漏：底层有一个 goroutine 在等待读取剩余数据
	//
	// 标准写法：defer resp.Body.Close()
	// 放在 err 检查之后、读取 Body 之前。
	defer resp.Body.Close()

	// ------------------------------------------------------------------------
	// 步骤 7：读取响应体
	// ------------------------------------------------------------------------
	//
	// io.ReadAll 一次性读完 resp.Body 的全部内容。
	// 内部实现大致是：
	//
	//   func ReadAll(r io.Reader) ([]byte, error) {
	//       b := make([]byte, 0, 512)  // 初始 512 字节
	//       for {
	//           n, err := r.Read(b[len(b):cap(b)])
	//           b = b[:len(b)+n]
	//           if err != nil {
	//               if err == io.EOF { return b, nil }
	//               return b, err
	//           }
	//           if len(b) == cap(b) {   // 容量不够，扩容
	//               b = append(b, 0)[:len(b)]
	//           }
	//       }
	//   }
	//
	// 作业中如果只需要检查关键词，可以用 io.LimitReader 限制读取量：
	//   limited := io.LimitReader(resp.Body, 10*1024)  // 最多读 10KB
	//   body, _ := io.ReadAll(limited)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应体失败：", err)
		return
	}

	// ------------------------------------------------------------------------
	// 步骤 8：处理响应信息
	// ------------------------------------------------------------------------
	//
	// resp.StatusCode 是 int 类型的 HTTP 状态码。
	// 作业中你需要判断：
	//   200 ~ 299 → 成功
	//   其他     → 失败（记录具体状态码）

	fmt.Println("状态码：", resp.StatusCode)
	fmt.Println("耗时：", time.Since(start))
	fmt.Println("响应体：", string(body))

	// strings.Contains 检查响应体是否包含指定关键词。
	// 作业中对应 Target.Contains 字段：
	//   如果配置了 contains，就需要检查响应体里有没有这个词。
	//   比如 contains: "Go" → 检查响应体里有没有 "Go"。
	fmt.Println("是否包含 Go：", strings.Contains(string(body), "Go"))

	// ------------------------------------------------------------------------
	// 作业中你会这样封装：
	// ------------------------------------------------------------------------
	//
	//   func probeHTTP(target Target, client *http.Client, timeout time.Duration) ProbeResult {
	//       result := ProbeResult{Name: target.Name}
	//       start := time.Now()
	//
	//       ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//       defer cancel()
	//
	//       req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.Address, nil)
	//       if err != nil {
	//           result.Error = fmt.Sprintf("创建请求失败: %v", err)
	//           result.Latency = time.Since(start).String()
	//           return result
	//       }
	//
	//       resp, err := client.Do(req)
	//       result.Latency = time.Since(start).String()
	//
	//       if err != nil {
	//           result.Error = fmt.Sprintf("请求失败: %v", err)
	//           return result
	//       }
	//       defer resp.Body.Close()
	//
	//       // 检查状态码
	//       if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	//           result.Error = fmt.Sprintf("状态码异常: %d", resp.StatusCode)
	//           return result
	//       }
	//
	//       // 检查关键词
	//       if target.Contains != "" {
	//           body, _ := io.ReadAll(resp.Body)
	//           if !strings.Contains(string(body), target.Contains) {
	//               result.Error = fmt.Sprintf("响应体不包含关键词 %q", target.Contains)
	//               return result
	//           }
	//       }
	//
	//       result.OK = true
	//       return result
	//   }
}

// probeSlowEndpoint 演示超时控制的实际效果。
//
// 这里的超时时间是 100ms，但服务端 /slow 路径会延迟 200ms 才返回。
// 所以客户端会在 100ms 时收到 context 超时错误，不会等到 200ms。
func probeSlowEndpoint(url string) {
	fmt.Printf("\n超时探测地址：%s\n", url)

	// 注意：超时时间设为 100ms，小于服务端的 200ms 延迟。
	// 这会导致 client.Do(req) 在 ~100ms 时返回错误。
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("创建慢请求失败：", err)
		return
	}

	// 这里会阻塞约 100ms（而不是服务端的 200ms），
	// 因为 context 超时后会立即返回，不等服务端响应。
	_, err = client.Do(req)
	if err != nil {
		// 输出类似：因为超时导致的请求失败：Get "http://127.0.0.1:xxx/slow": context deadline exceeded
		//
		// context deadline exceeded 是 context 包返回的标准超时错误。
		// 可以用 errors.Is(err, context.DeadlineExceeded) 来精确判断。
		fmt.Println("因为超时导致的请求失败：", err)
		return
	}
}
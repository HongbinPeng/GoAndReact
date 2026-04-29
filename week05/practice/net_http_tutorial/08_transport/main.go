package main

// ============================================================
// 第 08 课：Transport — 控制底层连接行为
// ============================================================
//
// 【Transport 是什么？】
//   Transport 是 http.Client 的"传输层"，负责：
//   - 建立 TCP 连接
//   - 发送 HTTP 请求
//   - 接收 HTTP 响应
//   - 连接池管理（keep-alive）
//   - TLS 握手
//
// 默认情况下，http.Client 使用 http.DefaultTransport，你不需要关心它。
// 但有时候你需要自定义连接池大小、禁用 keep-alive 等。
//
// 【运行方式】
//   go run 08_transport/main.go

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

func main() {
	// 测试服务器：返回请求的编号
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Write([]byte(fmt.Sprintf("request #%d", requestCount)))
	}))
	defer server.Close()

	// ================================================================
	// 知识点 1：默认 Transport 的连接池
	// ================================================================

	fmt.Println("===== 知识点 1：默认 Transport（连接复用）=====")

	// 默认的 http.Client 使用 DefaultTransport，它会自动复用 TCP 连接。
	// 这意味着：
	//   第 1 次请求 → 建立 TCP 连接（需要 DNS 解析 + 三次握手）
	//   第 2 次请求 → 复用第 1 次的连接（省去握手时间，更快）
	//   第 3 次请求 → 继续复用...
	//
	// 这就是所谓的 "HTTP Keep-Alive" 或 "持久连接"。

	client := &http.Client{}

	// 连续发 3 个请求，后两个会复用连接
	for i := 0; i < 3; i++ {
		start := time.Now()
		resp, err := client.Get(server.URL)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("请求 %d 失败: %v\n", i+1, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("请求 %d 完成，耗时 %v\n", i+1, elapsed)
	}

	fmt.Println("注意：第一个请求通常最慢（需要建立连接），后面的更快（复用连接）")

	// ================================================================
	// 知识点 2：自定义 Transport
	// ================================================================

	fmt.Println("\n===== 知识点 2：自定义 Transport =====")

	// http.Transport 结构体有很多字段，可以精细控制连接行为：
	//
	//   type Transport struct {
	//       // 连接池控制
	//       Proxy                 func(*Request) (*url.URL, error)  // 代理设置
	//       DialContext           func(context.Context, string, string) (net.Conn, error)
	//       DialTLSContext        func(context.Context, string, string) (net.Conn, error)
	//       TLSHandshakeTimeout   time.Duration  // TLS 握手超时
	//       DisableKeepAlives     bool           // 是否禁用连接复用
	//       DisableCompression    bool           // 是否禁用 gzip 压缩
	//       MaxIdleConns          int            // 最大空闲连接总数
	//       MaxIdleConnsPerHost   int            // 每个 host 的最大空闲连接数
	//       MaxConnsPerHost       int            // 每个 host 的最大连接数（含活跃）
	//       IdleConnTimeout       time.Duration  // 空闲连接的过期时间
	//       ResponseHeaderTimeout time.Duration  // 等待响应头的超时时间
	//       ExpectContinueTimeout time.Duration  // 100-continue 的超时时间
	//       TLSClientConfig       *tls.Config    // TLS 配置（跳过证书验证等）
	//   }

	// 常见自定义场景：

	// ---- 场景 A：禁用连接复用（每次请求都建新连接） ----
	// 适用场景：测试环境、调试、每次请求需要独立连接
	transportNoKeepAlive := &http.Transport{
		DisableKeepAlives: true, // 每次请求都新建 TCP 连接
	}
	clientA := &http.Client{
		Transport: transportNoKeepAlive,
		Timeout:   5 * time.Second,
	}

	fmt.Println("\n--- 场景 A：禁用 Keep-Alive ---")
	for i := 0; i < 2; i++ {
		start := time.Now()
		resp, err := clientA.Get(server.URL)
		elapsed := time.Since(start)
		if err != nil {
			fmt.Printf("请求 %d 失败: %v\n", i+1, err)
			continue
		}
		resp.Body.Close()
		fmt.Printf("请求 %d 完成，耗时 %v（每次都建新连接）\n", i+1, elapsed)
	}

	// ---- 场景 B：限制并发连接数 ----
	// 适用场景：防止对同一个目标服务器并发过多连接，被限流
	transportLimited := &http.Transport{
		MaxIdleConns:        10,               // 最大空闲连接数
		MaxIdleConnsPerHost: 5,                // 每个 host 最大空闲连接
		MaxConnsPerHost:     10,               // 每个 host 最大连接数（含活跃）
		IdleConnTimeout:     30 * time.Second, // 空闲连接 30 秒后关闭
	}
	clientB := &http.Client{
		Transport: transportLimited,
		Timeout:   5 * time.Second,
	}

	respB, _ := clientB.Get(server.URL)
	respB.Body.Close()
	fmt.Println("场景 B：限制了连接池大小（适合高并发场景）")

	// ---- 场景 C：跳过 TLS 证书验证（仅用于测试！） ----
	// 注意：生产环境不要这样做，会有中间人攻击风险！
	// import "crypto/tls"
	// transportInsecure := &http.Transport{
	//     TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// clientC := &http.Client{Transport: transportInsecure}

	// ================================================================
	// 知识点 3：Transport 的生命周期管理
	// ================================================================

	fmt.Println("\n===== 知识点 3：Transport 的生命周期 =====")

	// 重要：Transport 内部有连接池和后台 goroutine。
	// 如果你创建了一个自定义 Transport，使用完毕后应该调用 CloseIdleConnections()
	// 释放空闲连接。

	transport := &http.Transport{}
	clientTemp := &http.Client{Transport: transport}

	respT, _ := clientTemp.Get(server.URL)
	respT.Body.Close()

	// 释放空闲连接
	transport.CloseIdleConnections()
	fmt.Println("调用了 transport.CloseIdleConnections()，释放空闲连接")

	// ---- 最佳实践 ----
	// 1. 一个程序中通常只需要一个 http.Client（共享连接池）
	// 2. 如果需要不同的 Transport 配置，可以为每种配置创建一个 Client
	// 3. 不要在循环中创建新的 Transport（每次都会新建连接池）
	// 4. 程序退出前调用 transport.CloseIdleConnections() 释放资源

	fmt.Println("\n===== 总结 =====")
	fmt.Println("默认 Transport 够用：自动复用连接，不需要手动配置")
	fmt.Println("需要自定义时：创建 Transport 并传给 Client.Transport")
	fmt.Println("程序退出前：调用 transport.CloseIdleConnections()")
}

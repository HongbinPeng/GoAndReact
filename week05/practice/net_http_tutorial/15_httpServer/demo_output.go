package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// ====================================================================
	// 启动测试服务器
	// ====================================================================

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from server!"))
	})

	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("slow response"))
	})

	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"name":"Tom","age":20}`))
	})

	server := &http.Server{
		Addr:         ":8899",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		server.ListenAndServe()
	}()

	// 等服务器启动
	time.Sleep(200 * time.Millisecond)

	// ====================================================================
	// 演示代码和输出捕获
	// ====================================================================

	fmt.Println("====================================================================")
	fmt.Println("五、http.Server — HTTP 服务器实例")
	fmt.Println("====================================================================")

	// 演示 Server 结构体字段
	server2 := &http.Server{
		Addr:         ":8080",
		Handler:      nil, // nil 表示使用 http.DefaultServeMux
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	fmt.Printf("Server.Addr:            %s\n", server2.Addr)
	fmt.Printf("Server.Handler:         %v\n", server2.Handler)
	fmt.Printf("Server.ReadTimeout:     %s\n", server2.ReadTimeout)
	fmt.Printf("Server.WriteTimeout:    %s\n", server2.WriteTimeout)
	fmt.Printf("Server.IdleTimeout:     %s\n", server2.IdleTimeout)
	fmt.Printf("Server.MaxHeaderBytes:  %d 字节\n", server2.MaxHeaderBytes)

	fmt.Println("\n--- 启动方式 ---")
	fmt.Println("  server.ListenAndServe()              → 阻塞启动 HTTP")
	fmt.Println("  server.ListenAndServeTLS(cert,key)   → 阻塞启动 HTTPS")
	fmt.Println("  go server.ListenAndServe()           → 非阻塞启动（goroutine）")

	fmt.Println("\n--- 停止方式 ---")
	fmt.Println("  server.Close()       → 立即关闭，正在处理的请求被中断")
	fmt.Println("  server.Shutdown(ctx) → 优雅关闭，等正在处理的请求完成")

	fmt.Println("\n--- 安全要点 ---")
	fmt.Println("  ReadTimeout  → 防止慢速客户端攻击（Slowloris）")
	fmt.Println("  WriteTimeout → 防止服务端处理太慢")
	fmt.Println("  IdleTimeout  → 空闲连接过期，释放资源")

	fmt.Println()
	fmt.Println("====================================================================")
	fmt.Println("六、http.ServeMux — 路由多路复用器")
	fmt.Println("====================================================================")

	fmt.Println("--- Go 1.22+ 新路由语法 ---")
	mux2 := http.NewServeMux()

	// 精确匹配
	mux2.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("user list"))
	})

	// 前缀匹配（必须以 / 结尾）
	mux2.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static file"))
	})

	// Go 1.22+ 方法限定：GET /api/users
	mux2.HandleFunc("GET /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		w.Write([]byte("user id: " + id))
	})

	// 通配符匹配（Go 1.22+）
	mux2.HandleFunc("/api/{rest...}", func(w http.ResponseWriter, r *http.Request) {
		rest := r.PathValue("rest")
		w.Write([]byte("wildcard: " + rest))
	})

	fmt.Println("  精确匹配:   mux.HandleFunc(\"/api/users\", handler)")
	fmt.Println("  前缀匹配:   mux.HandleFunc(\"/static/\", handler)")
	fmt.Println("  方法+路径:  mux.HandleFunc(\"GET /users/{id}\", handler)")
	fmt.Println("  通配符:     mux.HandleFunc(\"/api/{rest...}\", handler)")
	fmt.Println()
	fmt.Println("  优先级: 方法+路径 > 精确路径 > 前缀路径 > 通配符")

	fmt.Println("\n--- 路由匹配规则 ---")
	fmt.Println("  1. 先匹配方法（GET/POST），不匹配则返回 405 Method Not Allowed")
	fmt.Println("  2. 再匹配路径（精确 > 前缀 > 通配符）")
	fmt.Println("  3. 没有匹配则返回 404 Not Found")
	fmt.Println("  4. 多个 ServeMux 互不干扰，可以嵌套")

	fmt.Println()
	fmt.Println("====================================================================")
	fmt.Println("七、http.Client — HTTP 客户端")
	fmt.Println("====================================================================")

	fmt.Println("--- http.Client 结构体字段 ---")
	fmt.Println("  Transport       RoundTripper  — 传输层（连接池）")
	fmt.Println("  CheckRedirect   func(...)     — 重定向策略")
	fmt.Println("  Jar             CookieJar     — Cookie 存储")
	fmt.Println("  Timeout         time.Duration — 整个请求的超时时间")

	fmt.Println("\n--- 方式一：http.Get（最简单）---")
	resp1, err := http.Get("http://localhost:8899/hello")
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		defer resp1.Body.Close()
		body, _ := io.ReadAll(resp1.Body)
		fmt.Println("  resp.StatusCode:", resp1.StatusCode)
		fmt.Println("  resp.Status:    ", resp1.Status)
		fmt.Println("  resp.Proto:     ", resp1.Proto)
		fmt.Println("  resp.Body:      ", string(body))
	}

	fmt.Println("\n--- 方式二：http.Client.Do（推荐）---")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8899/json", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "MyClient/1.0")

	resp2, err := client.Do(req)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		defer resp2.Body.Close()
		body, _ := io.ReadAll(resp2.Body)
		fmt.Println("  resp.StatusCode:", resp2.StatusCode)
		fmt.Println("  resp.Header[Content-Type]:", resp2.Header.Get("Content-Type"))
		fmt.Println("  resp.Body:      ", string(body))
		fmt.Println()
		fmt.Println("  → 可自定义超时、请求头、Transport")
	}

	fmt.Println("\n--- 方式三：超时控制 ---")
	fmt.Println("  方式 A: client.Timeout（粗粒度）")
	_ = &http.Client{Timeout: 1 * time.Second}
	fmt.Println("    → 所有请求统一 1 秒超时")

	fmt.Println("  方式 B: context.WithTimeout（细粒度）")
	fmt.Println("    ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)")
	fmt.Println("    defer cancel()")
	fmt.Println("    req, _ := http.NewRequestWithContext(ctx, \"GET\", url, nil)")
	fmt.Println("    → 每个请求独立超时，支持手动取消")

	fmt.Println()
	fmt.Println("====================================================================")
	fmt.Println("八、http.Transport — 传输层（连接池管理）")
	fmt.Println("====================================================================")

	fmt.Println("--- http.Transport 核心字段 ---")

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   false,
		DisableCompression:  false,
	}

	fmt.Printf("  MaxIdleConns:        %d  （所有 host 的最大空闲连接总数）\n", transport.MaxIdleConns)
	fmt.Printf("  MaxIdleConnsPerHost: %d  （每个 host 的最大空闲连接数）\n", transport.MaxIdleConnsPerHost)
	fmt.Printf("  MaxConnsPerHost:     %d  （每个 host 的最大连接数，含活跃）\n", transport.MaxConnsPerHost)
	fmt.Printf("  IdleConnTimeout:     %s （空闲连接过期时间）\n", transport.IdleConnTimeout)
	fmt.Printf("  TLSHandshakeTimeout: %s （TLS 握手超时）\n", transport.TLSHandshakeTimeout)
	fmt.Printf("  DisableKeepAlives:   %v  （false=启用连接复用）\n", transport.DisableKeepAlives)
	fmt.Printf("  DisableCompression:  %v  （false=启用 gzip 压缩）\n", transport.DisableCompression)

	fmt.Println("\n--- Transport 的生命周期 ---")

	clientT := &http.Client{Transport: transport, Timeout: 5 * time.Second}

	respT, _ := clientT.Get("http://localhost:8899/hello")
	bodyT, _ := io.ReadAll(respT.Body)
	respT.Body.Close()
	fmt.Printf("  第 1 次请求 → 建立连接 → 响应: %s\n", string(bodyT))

	respT2, _ := clientT.Get("http://localhost:8899/hello")
	bodyT2, _ := io.ReadAll(respT2.Body)
	respT2.Body.Close()
	fmt.Printf("  第 2 次请求 → 复用连接 → 响应: %s\n", string(bodyT2))

	transport.CloseIdleConnections()
	fmt.Println("  transport.CloseIdleConnections() → 释放所有空闲连接")

	fmt.Println("\n--- Transport 与 Client 的关系 ---")
	fmt.Println("  一个 Client 包含一个 Transport")
	fmt.Println("  多个 goroutine 共享同一个 Client = 共享同一个连接池（推荐）")
	fmt.Println("  每个请求都新建 Client = 每次都新建连接池（性能差，不要这样）")

	fmt.Println()
	fmt.Println("====================================================================")
	fmt.Println("九、完整架构层次")
	fmt.Println("====================================================================")
	fmt.Println()
	fmt.Println("  你的代码层:")
	fmt.Println("    client.Do(req)           ← 发送请求")
	fmt.Println("    handler(w, r)            ← 处理请求")
	fmt.Println("          │")
	fmt.Println("          ▼")
	fmt.Println("  http.Client / http.Server  ← 调度层")
	fmt.Println("    - Client: 管理 Transport、重定向、Cookie、超时")
	fmt.Println("    - Server: 管理 ServeMux、读写超时、连接管理")
	fmt.Println("          │")
	fmt.Println("          ▼")
	fmt.Println("  http.Transport              ← 传输层（仅客户端）")
	fmt.Println("    - 连接池（keep-alive）")
	fmt.Println("    - DNS 解析、TCP 连接、TLS 握手")
	fmt.Println("    - gzip 压缩/解压")
	fmt.Println("          │")
	fmt.Println("          ▼")
	fmt.Println("  net.Conn                    ← 网络层")
	fmt.Println("    - 底层 TCP socket")
	fmt.Println("    - 读写字节流")

	fmt.Println()
	fmt.Println("====================================================================")
	fmt.Println("十、对象关系总结")
	fmt.Println("====================================================================")
	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────────────┐")
	fmt.Println("  │                     客户端发送请求                     │")
	fmt.Println("  ├──────────┬───────────┬──────────┬───────────────────┤")
	fmt.Println("  │ Client   │ Transport │ Request  │ Response（收到）   │")
	fmt.Println("  │ 调度请求  │ 连接池     │ 请求信息  │ 响应信息           │")
	fmt.Println("  └─────────────────────┴──────────┴───────────────────┘")
	fmt.Println()
	fmt.Println("  ┌──────────────────────────────────────────────────────┐")
	fmt.Println("  │                     服务端接收请求                     │")
	fmt.Println("  ├──────────┬───────────┬──────────┬───────────────────┤")
	fmt.Println("  │ Server   │ ServeMux  │ Request  │ ResponseWriter     │")
	fmt.Println("  │ 监听端口  │ 路由匹配   │ 请求信息  │ 构建响应           │")
	fmt.Println("  └──────────┴───────────┴──────────┴───────────────────┘")
	fmt.Println()
	fmt.Println("  关键记忆:")
	fmt.Println("  - 客户端: Client → Transport → Request → Response")
	fmt.Println("  - 服务端: Server → ServeMux → Request → ResponseWriter")
	fmt.Println("  - Request 是共用的（客户端和服务端都有）")
	fmt.Println("  - 客户端收到 Response，服务端写入 ResponseWriter")

	// 关闭服务器
	server.Close()
}
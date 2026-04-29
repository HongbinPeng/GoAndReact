package main

// ============================================================
// 服务端高并发限流演示
// ============================================================
//
// 服务端限流的三种方式：
//   1. 连接层限流 — 控制同时建立的 TCP 连接数
//   2. Handler 层限流 — 控制同时处理的请求数
//   3. 超时控制 — 防止慢连接耗尽资源
//
// 【运行方式】
//   go run main.go
//   然后在另一个终端并发请求：
//     for i in {1..20}; do curl http://localhost:8080/ & done

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// ================================================================
// 方式 1：Handler 层限流（最常用）
// ================================================================

// 用信号量模式限制同时处理的请求数
type LimitingHandler struct {
	handler   http.Handler
	semaphore chan struct{} // 信号量：容量 = 最大并发数
	active    int64         // 当前活跃请求数（用于展示）
}

func NewLimitingHandler(handler http.Handler, maxConcurrent int) *LimitingHandler {
	return &LimitingHandler{
		handler:   handler,
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

func (h *LimitingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 尝试获取信号量（拿到一个空位）
	select {
	case h.semaphore <- struct{}{}:
		// 拿到了，放行
		defer func() { <-h.semaphore }() // 处理完释放
	default:
		// 没拿到，已满，拒绝
		http.Error(w, "服务繁忙，请稍后重试", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&h.active, 1)
	defer atomic.AddInt64(&h.active, -1)

	fmt.Printf("  [限流器] 活跃请求: %d / 上限: %d\n",
		atomic.LoadInt64(&h.active), cap(h.semaphore))

	h.handler.ServeHTTP(w, r)
}

// ================================================================
// 方式 2：连接层限流（控制 TCP 连接数）
// ================================================================

// 用自定义 Listener 限制同时建立的连接数
type LimitingListener struct {
	net.Listener
	semaphore chan struct{}
}

func NewLimitingListener(addr string, maxConns int) (*LimitingListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &LimitingListener{
		Listener:  l,
		semaphore: make(chan struct{}, maxConns),
	}, nil
}

func (l *LimitingListener) Accept() (net.Conn, error) {
	// 获取信号量（没拿到就阻塞，不继续 Accept 新连接）
	l.semaphore <- struct{}{}

	conn, err := l.Listener.Accept()
	if err != nil {
		<-l.semaphore // 接受失败，释放信号量
		return nil, err
	}

	// 包装 conn，在 Close 时释放信号量
	return &limitingConn{Conn: conn, semaphore: l.semaphore}, nil
}

type limitingConn struct {
	net.Conn
	semaphore chan struct{}
	closed    bool
}

func (c *limitingConn) Close() error {
	if !c.closed {
		c.closed = true
		<-c.semaphore // 释放信号量
	}
	return c.Conn.Close()
}

// ================================================================
// 方式 3：超时控制（防止慢连接耗尽资源）
// ================================================================

func main() {
	// 模拟一个慢 handler
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [Handler] 开始处理请求，需要 3 秒...")
		time.Sleep(3 * time.Second)
		fmt.Println("  [Handler] 处理完成")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("处理完成"))
	})

	// ---- 方式 1：Handler 层限流（推荐） ----
	fmt.Println("===== 方式 1：Handler 层限流（最多 3 个并发）=====")
	fmt.Println("第 4 个请求会被直接拒绝")

	limitedHandler := NewLimitingHandler(slowHandler, 3)

	// ---- 方式 3：超时控制（必须加！） ----
	// 即使不限流，也要设置超时，防止一个慢连接永远占着资源
	server := &http.Server{
		Handler:      limitedHandler,
		ReadTimeout:  5 * time.Second,  // 客户端发送请求的超时
		WriteTimeout: 10 * time.Second, // 服务端返回响应的超时
		IdleTimeout:  60 * time.Second, // 空闲连接保持时间
	}

	// 启动服务器（在 goroutine 中，让 main 继续执行）
	go func() {
		fmt.Println("  服务器启动在 :8080")
		fmt.Println("  在另一个终端运行：for i in {1..5}; do curl http://localhost:8080/ & done")
		fmt.Println("  观察前 3 个请求正常处理，第 4、5 个被拒绝")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("  服务器错误:", err)
		}
	}()

	// 等待 15 秒后关闭（演示用）
	time.Sleep(15 * time.Second)
	fmt.Println("\n演示结束，关闭服务器")
	server.Close()

	// ---- 方式 2：连接层限流（示例代码） ----
	// 这种方式更底层，限制的是 TCP 连接数而不是请求数
	//
	// limitingListener, _ := NewLimitingListener(":8081", 100)
	// server2 := &http.Server{
	//     Handler: slowHandler,
	//     ReadTimeout:  5 * time.Second,
	//     WriteTimeout: 10 * time.Second,
	// }
	// server2.Serve(limitingListener)
	//
	// 超过 100 个连接时，新的 TCP 连接建立请求会被阻塞（三次握手完成但不会被 Accept）
	// 客户端表现为"连接中..."一直转圈

	fmt.Println("\n===== 三种方式对比 =====")
	fmt.Println("方式 1（Handler 限流）：限制同时处理的请求数 → 推荐")
	fmt.Println("方式 2（连接限流）：限制同时建立的 TCP 连接数 → 更底层")
	fmt.Println("方式 3（超时控制）：防止慢连接耗尽资源 → 必须加")
}
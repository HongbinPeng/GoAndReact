package main

// ============================================================
// 让服务端一直监听（优雅启动 + 优雅关闭）
// ============================================================
//
// 要点：
//   1. ListenAndServe 是阻塞调用，不会返回 → 天然"一直监听"
//   2. 用 os.Signal 监听退出信号（Ctrl+C）
//   3. 用 server.Shutdown 优雅关闭（等正在处理的请求完成）
//
// 【运行方式】
//   go run main.go
//   然后用浏览器访问 http://localhost:8080/
//   按 Ctrl+C 优雅退出

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// ================================================================
	// 第 1 步：注册路由
	// ================================================================

	mux := http.NewServeMux()

	// 正常请求
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [请求] 收到请求:", r.URL.Path)
		w.Write([]byte("Hello, World!"))
	})

	// 模拟慢请求（用来测试优雅关闭）
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("  [请求] 收到慢请求，需要 5 秒...")
		time.Sleep(5 * time.Second)
		fmt.Println("  [请求] 慢请求处理完成")
		w.Write([]byte("慢请求完成"))
	})

	// ================================================================
	// 第 2 步：创建服务器（必须设置超时！）
	// ================================================================

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,   // 防止慢速攻击
		WriteTimeout: 10 * time.Second,  // 防止响应太慢
		IdleTimeout:  120 * time.Second, // 空闲连接超时
	}

	fmt.Println("=============================================")
	fmt.Println("  服务器启动在 :8080")
	fmt.Println("  访问 http://localhost:8080/")
	fmt.Println("  访问 http://localhost:8080/slow （测试优雅关闭）")
	fmt.Println("  按 Ctrl+C 优雅退出")
	fmt.Println("=============================================")

	// ================================================================
	// 第 3 步：在 goroutine 中启动服务器
	// ================================================================

	go func() {
		// ListenAndServe 是阻塞调用！
		// 它会一直监听，直到服务器被关闭或出错
		// 这就是"一直监听"的关键
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("服务器启动失败:", err)
			os.Exit(1)
		}
	}()

	// ================================================================
	// 第 4 步：监听退出信号（Ctrl+C）
	// ================================================================

	// os.Signal 通道，用来接收系统信号
	quit := make(chan os.Signal, 1)
	// 注册要监听的信号：
	//   SIGINT  — Ctrl+C
	//   SIGTERM — kill 命令
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞等待信号
	<-quit
	fmt.Println("\n收到退出信号，开始优雅关闭...")

	// ================================================================
	// 第 5 步：优雅关闭
	// ================================================================

	// 创建一个 5 秒的超时 context
	// 如果 5 秒内正在处理的请求还没完成，强制关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// server.Shutdown 会：
	//   1. 停止接受新连接
	//   2. 等待正在处理的请求完成
	//   3. 超过超时时间则强制关闭
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Println("强制关闭:", err)
	}

	fmt.Println("服务器已关闭")
}

// ================================================================
// 对比：简单粗暴的方式（不推荐）
// ================================================================

// // ❌ 方式 1：直接 os.Exit(1)
// // 问题：正在处理的请求被强制中断，客户端收到连接重置
//
// // ❌ 方式 2：server.Close()
// // 问题：和 os.Exit 类似，不等待正在处理的请求
//
// // ✅ 方式 3：server.Shutdown(ctx)
// // 优点：等待正在处理的请求完成后再关闭
// //       客户端能收到完整响应

// ================================================================
// 对比：最简单的"一直监听"写法（不带优雅关闭）
// ================================================================

// func main() {
//     http.HandleFunc("/", handler)
//     // ListenAndServe 会阻塞，一直监听
//     // 按 Ctrl+C 会直接终止进程（不优雅，但能用）
//     http.ListenAndServe(":8080", nil)
// }
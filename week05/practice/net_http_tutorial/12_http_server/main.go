package main

// ============================================================
// 第 12 课：HTTP 服务端 — 搭建你自己的 Web API
// ============================================================
//
// 【客户端 vs 服务端】
//   前面 11 课都在讲客户端（发请求），这课讲服务端（收请求）。
//   服务端的核心概念：
//   1. http.Server — 服务器实例
//   2. http.HandleFunc — 注册路由
//   3. http.ResponseWriter — 写响应
//   4. http.ListenAndServe — 启动监听
//
// 【运行方式】
//   go run 12_http_server/main.go
//   程序会启动服务器在 :8080 端口，然后用浏览器或 curl 访问：
//     curl http://localhost:8080/
//     curl http://localhost:8080/hello?name=Tom
//     curl -X POST http://localhost:8080/api/users -d '{"name":"Tom"}'
//     curl http://localhost:8080/unknown  （测试 404）
//
//   程序启动后 10 秒会自动关闭（为了演示，不会一直挂着）

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	// ================================================================
	// 知识点 1：注册路由
	// ================================================================

	// ---- 什么是路由？ ----
	// 路由就是：当用户访问某个 URL 路径时，执行对应的函数。
	//
	//   用户访问 /hello   → 执行 helloHandler
	//   用户访问 /api/xxx  → 执行 apiHandler
	//   用户访问 /  → 执行 rootHandler
	//
	// Go 标准库提供了三种注册路由的方式：

	// ---- 方式 1：http.HandleFunc（最常用） ----
	// 签名：func HandleFunc(pattern string, handler func(ResponseWriter, *Request))
	// 把一个函数注册为某个路径的处理器
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/api/users", apiUsersHandler)

	// ---- 方式 2：http.Handle（用 Handler 接口） ----
	// 签名：func Handle(pattern string, handler Handler)
	// Handler 是一个接口：type Handler interface { ServeHTTP(ResponseWriter, *Request) }
	// 适合有状态的处理器（比如需要保存数据库连接等）
	http.Handle("/status", http.HandlerFunc(statusHandler))

	// ---- 方式 3：http.HandleFunc 带路径前缀匹配 ----
	// http.HandleFunc("/api/", apiPrefixHandler)
	// 访问 /api/anything 都会匹配到这个 handler

	// ================================================================
	// 知识点 2：启动服务器
	// ================================================================

	// http.ListenAndServe(addr, handler) — 启动 HTTP 服务器
	//   addr    — 监听地址，":8080" 表示监听所有网卡 8080 端口
	//   handler — 路由处理器，传 nil 表示使用 DefaultServeMux（就是我们上面注册的那些路由）
	//
	// 这个函数会阻塞（一直运行），直到服务器关闭或出错。
	// 为了让教程能自动退出，我们用 goroutine 启动，然后 10 秒后关闭。

	fmt.Println("🚀 服务器启动中...")
	fmt.Println("可用路由：")
	fmt.Println("  GET  /           — 首页")
	fmt.Println("  GET  /hello?name=xxx  — 带参数的问候")
	fmt.Println("  POST /api/users  — 创建用户（接收 JSON）")
	fmt.Println("  GET  /status     — 服务器状态")
	fmt.Println()
	fmt.Println("请在浏览器打开：http://localhost:8080/")
	fmt.Println("（10 秒后服务器自动关闭）")

	// 启动服务器
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Printf("服务器错误: %v", err)
		}
	}()

	// 10 秒后自动退出（教程演示用）
	time.Sleep(10 * time.Second)
	fmt.Println("\n演示结束，服务器关闭。")
}

// ================================================================
// 知识点 3：Handler 函数 — 处理 HTTP 请求
// ================================================================

// Handler 函数的标准签名：
//   func handler(w http.ResponseWriter, r *http.Request)
//
// 参数说明：
//   w — http.ResponseWriter 接口，用来写入 HTTP 响应
//     w.WriteHeader(code)  — 设置状态码（必须在 Write 之前调用）
//     w.Write(data)        — 写入响应体
//     w.Header()           — 获取响应头（必须在 WriteHeader 之前设置）
//
//   r — *http.Request 结构体，包含请求的所有信息
//     r.Method             — 请求方法（GET, POST 等）
//     r.URL                — URL 对象
//     r.URL.Path           — 请求路径，如 "/hello"
//     r.URL.Query()        — URL 查询参数，如 ?name=Tom
//     r.Header             — 请求头
//     r.Body               — 请求体（POST/PUT 时有数据）

// rootHandler — 首页
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// 只处理根路径 /
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 设置响应头（必须在 WriteHeader 或 Write 之前）
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 写入响应
	// 注意：如果不显式调用 WriteHeader，第一次 Write 时会自动发送 200
	fmt.Fprintln(w, "<h1>欢迎！</h1>")
	fmt.Fprintln(w, "<ul>")
	fmt.Fprintln(w, "<li><a href='/hello?name=Tom'>问候页面</a></li>")
	fmt.Fprintln(w, "<li><a href='/status'>服务器状态</a></li>")
	fmt.Fprintln(w, "</ul>")
}

// helloHandler — 带 URL 参数的问候
func helloHandler(w http.ResponseWriter, r *http.Request) {
	// ---- 获取 URL 查询参数 ----
	// r.URL.Query() 返回 url.Values 类型（底层是 map[string][]string）
	// .Get("name") 获取 "name" 参数的第一个值
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "陌生人"
	}

	// 设置 Content-Type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 写入响应
	fmt.Fprintf(w, "<h1>你好, %s!</h1>", name)
	fmt.Fprintf(w, "<p>请求方法: %s</p>", r.Method)
	fmt.Fprintf(w, "<p>完整 URL: %s</p>", r.URL.String())
	fmt.Fprintf(w, "<p>远程地址: %s</p>", r.RemoteAddr)
}

// apiUsersHandler — 接收 JSON 并返回
func apiUsersHandler(w http.ResponseWriter, r *http.Request) {
	// ---- 检查请求方法 ----
	// API 通常要区分 GET（查询）和 POST（创建）
	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 方法", http.StatusMethodNotAllowed)
		return
	}

	// ---- 读取请求体 ----
	// r.Body 是 io.ReadCloser 类型，用 io.ReadAll 读完
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求体失败", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// ---- 解析 JSON ----
	var user map[string]any
	err = json.Unmarshal(body, &user)
	if err != nil {
		// http.Error 是一个便捷函数：
		//   1. 设置 Content-Type: text/plain
		//   2. 设置状态码
		//   3. 写入错误信息到响应体
		http.Error(w, "JSON 格式错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	// ---- 返回 JSON 响应 ----
	// 模拟创建成功，返回创建的用户信息
	response := map[string]any{
		"success": true,
		"user":    user,
		"id":      42, // 模拟 ID
	}

	// 设置响应 Content-Type
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created

	// 用 json.NewEncoder 直接写入 ResponseWriter（不需要中间的 []byte）
	json.NewEncoder(w).Encode(response)
}

// statusHandler — 服务器状态
func statusHandler(w http.ResponseWriter, r *http.Request) {
	// ---- 用结构体返回状态信息 ----
	status := map[string]any{
		"status":    "running",
		"time":      time.Now().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
		"user_agent": r.UserAgent(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// ================================================================
// 知识点 4：生产环境的服务端写法
// ================================================================

// 上面的写法用的是全局 DefaultServeMux，实际项目中推荐这样做：

// func main() {
//     // 创建独立的路由器（不污染全局的 DefaultServeMux）
//     mux := http.NewServeMux()
//     mux.HandleFunc("/", rootHandler)
//     mux.HandleFunc("/hello", helloHandler)
//     mux.HandleFunc("/api/users", apiUsersHandler)
//     mux.HandleFunc("/status", statusHandler)
//
//     // 创建自定义服务器（可以配置超时等）
//     srv := &http.Server{
//         Addr:         ":8080",
//         Handler:      mux,                  // 使用自定义路由器
//         ReadTimeout:  5 * time.Second,      // 读取请求超时
//         WriteTimeout: 10 * time.Second,     // 写入响应超时
//         IdleTimeout:  120 * time.Second,    // 空闲连接超时
//     }
//
//     // 启动服务器
//     log.Println("服务器启动:", srv.Addr)
//     if err := srv.ListenAndServe(); err != nil {
//         log.Fatal("服务器启动失败:", err)
//     }
// }

// ---- 服务端安全要点 ----
//
// 1. 设置超时 — 防止慢连接耗尽资源
//    ReadTimeout:  客户端发送请求的超时时间
//    WriteTimeout: 服务端返回响应的超时时间
//    IdleTimeout:  空闲连接的保持时间（keep-alive 连接）
//
// 2. 限制请求体大小 — 防止超大请求耗尽内存
//    r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 最大 1MB
//
// 3. 使用独立的 mux — 不要污染全局 DefaultServeMux
//    mux := http.NewServeMux()
//
// 4. 优雅关闭 — 用 context 让服务器平滑退出
//    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//    defer cancel()
//    srv.Shutdown(ctx)


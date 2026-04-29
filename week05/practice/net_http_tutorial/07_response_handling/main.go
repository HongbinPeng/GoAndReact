package main

// ============================================================
// 第 07 课：处理响应 — 状态码、响应头、流式读取
// ============================================================
//
// 【本课重点】
//   1. 如何判断请求是否成功（状态码范围）
//   2. 如何读取响应头
//   3. 如何安全地读取响应体（防止恶意服务器返回超大内容）
//   4. 如何流式读取（边读边处理，不需要全部加载到内存）
//
// 【运行方式】
//   go run 07_response_handling/main.go

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func main() {
	// 启动测试服务器，返回不同内容
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"name":"Tom","age":20,"city":"Beijing"}`))
		case "/large":
			// 返回 1 MB 的数据
			for i := 0; i < 1024; i++ {
				w.Write([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789\n"))
			}
		case "/stream":
			// 模拟流式响应（通常用于 SSE 或长连接）
			w.Header().Set("Content-Type", "text/event-stream")
			for i := 0; i < 5; i++ {
				w.Write([]byte(fmt.Sprintf("chunk %d\n", i)))
				w.(http.Flusher).Flush()
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// ================================================================
	// 知识点 1：判断请求成功还是失败
	// ================================================================

	fmt.Println("===== 知识点 1：状态码判断 =====")

	resp, _ := http.Get(server.URL + "/json")
	defer resp.Body.Close()

	// 方法 1：直接比较状态码
	if resp.StatusCode == 200 {
		fmt.Println("方法1：请求成功（== 200）")
	}

	// 方法 2：检查 2xx 范围（推荐！）
	// StatusCode >= 200 && StatusCode < 300
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("方法2：请求成功（2xx 范围）")
	}

	// 方法 3：用 http.StatusText 获取状态描述
	// http.StatusText(200) → "OK"
	// http.StatusText(404) → "Not Found"
	// http.StatusText(500) → "Internal Server Error"
	fmt.Println("状态描述:", http.StatusText(resp.StatusCode))

	// ---- 状态码分类 ----
	//
	//   100 ~ 199: 信息性状态码（继续等待）
	//   200 ~ 299: 成功（请求被成功处理）
	//     200 OK            — 成功
	//     201 Created       — 资源已创建（POST 常用）
	//     204 No Content    — 成功但无返回内容（DELETE 常用）
	//
	//   300 ~ 399: 重定向（需要进一步操作）
	//     301 Moved Permanently — 永久重定向
	//     302 Found             — 临时重定向
	//     304 Not Modified      — 缓存命中，不用重新下载
	//
	//   400 ~ 499: 客户端错误（你的请求有问题）
	//     400 Bad Request       — 请求格式错误
	//     401 Unauthorized      — 未认证（需要登录）
	//     403 Forbidden         — 无权限（已登录但无权访问）
	//     404 Not Found         — 资源不存在
	//
	//   500 ~ 599: 服务端错误（服务器出问题了）
	//     500 Internal Server Error — 服务器内部错误
	//     502 Bad Gateway           — 网关错误
	//     503 Service Unavailable   — 服务不可用（可能在维护）

	// ================================================================
	// 知识点 2：解析 JSON 响应
	// ================================================================

	fmt.Println("\n===== 知识点 2：解析 JSON 响应 =====")

	// 服务端返回 JSON：{"name":"Tom","age":20,"city":"Beijing"}

	// 方法 1：直接读取字符串
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("原始响应:", string(body))

	// 方法 2：反序列化为结构体
	// 先定义结构体，字段名首字母大写（可导出），带 json 标签
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
		City string `json:"city"`
	}

	var user User
	err := json.Unmarshal(body, &user) // 注意要传指针！
	if err != nil {
		fmt.Println("JSON 解析失败:", err)
		return
	}
	fmt.Printf("解析结果: Name=%s, Age=%d, City=%s\n", user.Name, user.Age, user.City)

	// 方法 3：解析到 map（适合不知道 JSON 结构的场景）
	var data map[string]any
	json.Unmarshal(body, &data)
	fmt.Printf("map 解析: Name=%v, Age=%v\n", data["name"], data["age"])

	// ================================================================
	// 知识点 3：安全读取 — 限制最大读取量
	// ================================================================

	fmt.Println("\n===== 知识点 3：安全读取（io.LimitReader）=====")

	// 如果一个恶意服务器返回 1 GB 的数据，io.ReadAll 会让你的程序内存溢出！
	// 解决办法：用 io.LimitReader 限制最多读取多少字节。

	resp2, _ := http.Get(server.URL + "/large")
	defer resp2.Body.Close()

	// 最多只读 1 KB（1024 字节）
	const maxBytes = 1024
	limitedReader := io.LimitReader(resp2.Body, maxBytes)
	safeBody, _ := io.ReadAll(limitedReader)

	fmt.Printf("限制读取 1KB，实际读取了 %d 字节:\n", len(safeBody))
	if len(safeBody) > 50 {
		fmt.Println(string(safeBody[:50]), "...（已截断）")
	}

	// ---- io.LimitReader 原理 ----
	//
	//   func LimitReader(r Reader, n int64) Reader {
	//       return &LimitedReader{R: r, N: n}
	//   }
	//
	//   type LimitedReader struct {
	//       R Reader  // 底层的 Reader
	//       N int64   // 还剩多少字节可以读
	//   }
	//
	//   func (l *LimitedReader) Read(p []byte) (n int, err error) {
	//       if l.N <= 0 {
	//           return 0, io.EOF  // 已读满，返回 EOF
	//       }
	//       // ... 最多读 N 字节
	//   }
	//
	// 当读完限制后，会返回 io.EOF，io.ReadAll 遇到 EOF 就停止。

	// ================================================================
	// 知识点 4：流式读取（边读边处理）
	// ================================================================

	fmt.Println("\n===== 知识点 4：流式读取 =====")

	resp3, _ := http.Get(server.URL + "/stream")
	defer resp3.Body.Close()

	// 流式读取：不需要一次性读完，可以边读边处理。
	// 适用场景：
	//   - 数据量很大，不想全部加载到内存
	//   - 服务端是流式返回（SSE、gRPC 等）
	//   - 需要逐行/逐块处理数据

	// 用 bufio.Scanner 逐行读取
	// scanner := bufio.NewScanner(resp3.Body)
	// for scanner.Scan() {
	//     line := scanner.Text()
	//     fmt.Println("收到行:", line)
	// }

	// 这里用 io.ReadAll 演示（因为 httptest 的流式返回在本地不是真正的流式）
	streamBody, _ := io.ReadAll(resp3.Body)
	fmt.Println("流式响应内容:", string(streamBody))

	fmt.Println("\n===== 总结 =====")
	fmt.Println("状态码判断：resp.StatusCode >= 200 && < 300 → 成功")
	fmt.Println("JSON 解析：json.Unmarshal(body, &struct)")
	fmt.Println("安全读取：io.LimitReader 防止超大响应体")
	fmt.Println("流式读取：边读边处理，不需要全部加载到内存")
}

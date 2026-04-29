package main

// ============================================================
// 第 06 课：设置请求头 — User-Agent、认证等
// ============================================================
//
// 【请求头是什么？】
//   请求头是 HTTP 请求中的元数据，告诉服务器一些额外信息：
//   - 你是谁（User-Agent）
//   - 你想要什么格式的数据（Accept）
//   - 你的身份凭证（Authorization）
//   - 请求体的类型（Content-Type）
//
// 【运行方式】
//   go run 06_request_headers/main.go

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func main() {
	// 测试服务器：返回收到的所有请求头
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprintf(w, "{\n")
		fmt.Fprintf(w, "  \"method\": %q,\n", r.Method)
		fmt.Fprintf(w, "  \"headers\": {\n")
		i := 0
		for key, values := range r.Header {
			if i > 0 {
				fmt.Fprintf(w, ",\n")
			}
			fmt.Fprintf(w, "    %q: %q", key, values[0])
			i++
		}
		fmt.Fprintf(w, "\n  }\n}")
	}))
	defer server.Close()

	// ================================================================
	// 知识点 1：http.Request.Header 的类型
	// ================================================================

	fmt.Println("===== 知识点 1：Header 的类型 =====")

	// http.Header 底层是 map[string][]string
	// type Header map[string][]string
	//
	// 为什么值是切片？因为一个头可以有多个值：
	//   Accept: text/html
	//   Accept: application/json
	//
	// 常用方法：
	//   header.Set(key, value)     — 设置（覆盖已有值）
	//   header.Get(key)            — 获取第一个值
	//   header.Add(key, value)     — 追加（不覆盖）
	//   header.Del(key)            — 删除
	//   header.Values(key)         — 获取所有值（切片）

	// ================================================================
	// 知识点 2：常用请求头
	// ================================================================

	fmt.Println("\n===== 知识点 2：常用请求头 =====")

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	// ---- User-Agent：标识客户端身份 ----
	//
	// 很多网站会检查 User-Agent，拒绝没有 User-Agent 的请求（比如爬虫）。
	// 默认的 Go 客户端 User-Agent 是 "Go-http-client/1.1"。
	req.Header.Set("User-Agent", "MyApp/1.0 (Go Tutorial)")
	// User-Agent 是 HTTP 请求头中的一个标准字段，用来告诉服务器"我是谁"。
	// 具体作用
	// 标识客户端身份：告诉服务器你用什么浏览器/客户端、什么操作系统来访问
	// 内容协商：服务器根据 User-Agent 返回不同格式的内容（比如给手机返回移动端页面，给桌面端返回桌面版页面）
	// 统计与分析：网站统计访客用的是 Chrome 还是 Safari，是 iOS 还是 Android
	// 反爬虫：很多网站会检查这个字段，拒绝没有或异常的 User-Agent
	// 它唯一吗？
	// 完全不唯一。 它只是一个普通字符串，没有任何唯一性保证：
	// 所有用 Go http.Get 发请求的客户端默认都是 "Go-http-client/1.1"
	// 所有用 Chrome 浏览器的人都有一样的 User-Agent 前缀
	// 任何人都可以随便改它，没有校验机制

	// ---- Accept：告诉服务器你想要什么格式的数据 ----
	//
	//   application/json  — 我想要 JSON 格式
	//   text/html         — 我想要 HTML 格式
	//   text/plain        — 我想要纯文本
	//   */*               — 任何格式都可以
	req.Header.Set("Accept", "application/json")

	// ---- Authorization：身份认证 ----
	//
	// 最常见的认证方式是 Bearer Token（JWT）：
	//   Authorization: Bearer <token>
	//
	// 另一种是 Basic Auth（用户名:密码 的 Base64 编码）：
	//   Authorization: Basic <base64编码>
	//
	// 项目中更推荐用 HTTP 客户端内置的认证方式：
	//   req.SetBasicAuth("username", "password")
	// 它会自动设置 Authorization: Basic xxx
	req.Header.Set("Authorization", "Bearer my-secret-token-123")

	// ---- Content-Type：请求体的格式 ----
	// POST/PUT 请求时必须设置，告诉服务器怎么解析请求体
	//   application/json              — JSON 格式
	//   application/x-www-form-urlencoded — 表单格式
	//   multipart/form-data           — 文件上传
	//   text/plain                    — 纯文本
	// 这课是 GET 请求，没有请求体，所以不需要设置。

	// ---- Cookie：会话凭证 ----
	//
	// Cookie 头可以手动设置，但更推荐用 http.Client 的 Jar（自动管理）：
	//   req.Header.Set("Cookie", "session=abc123; user=tom")
	//
	// 更优雅的方式：
	//   req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("服务端收到的请求头:")
	fmt.Println(string(body))

	// ================================================================
	// 知识点 3：Add vs Set 的区别
	// ================================================================

	fmt.Println("\n===== 知识点 3：Add vs Set =====")

	req2, _ := http.NewRequest(http.MethodGet, server.URL, nil)

	// Set：覆盖之前的值（同一个 key 只保留最后一个）
	req2.Header.Set("Accept", "text/html")
	req2.Header.Set("Accept", "application/json")
	// 最终 Accept 只有一个值: "application/json"

	// Add：追加（同一个 key 可以有多个值）
	req2.Header.Add("Accept", "text/html")
	req2.Header.Add("Accept", "application/json")
	req2.Header.Add("Accept", "text/plain")
	// 最终 Accept 有三个值

	// SetBasicAuth：快捷设置 Basic 认证
	req2.SetBasicAuth("admin", "password123")

	resp2, _ := http.DefaultClient.Do(req2)
	if resp2 != nil {
		defer resp2.Body.Close()
		body2, _ := io.ReadAll(resp2.Body)
		fmt.Println("服务端收到的请求头:")
		fmt.Println(string(body2))
	}

	// ================================================================
	// 知识点 4：响应头也同理
	// ================================================================

	fmt.Println("\n===== 知识点 4：读取响应头 =====")

	// 服务端返回的响应头也可以用类似的方式读取
	// resp.Header.Get("Content-Type")  — 获取单个值
	// resp.Header.Values("Set-Cookie") — 获取所有 Set-Cookie
	//
	// 常见的响应头：
	//   Content-Type    — 响应体的 MIME 类型
	//   Content-Length  — 响应体的字节数
	//   Set-Cookie      — 设置 Cookie
	//   X-Request-Id    — 请求追踪 ID（很多 API 会返回）
	//   RateLimit-Limit — 速率限制信息

	fmt.Println("\n===== 总结 =====")
	fmt.Println("Header.Set(key, value)  → 覆盖设置")
	fmt.Println("Header.Add(key, value)  → 追加设置")
	fmt.Println("Header.Get(key)         → 获取值")
	fmt.Println("req.SetBasicAuth(u, p)  → Basic 认证快捷方式")
}

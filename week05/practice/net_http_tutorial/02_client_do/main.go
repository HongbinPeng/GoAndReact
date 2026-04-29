package main

// ============================================================
// 第 02 课：http.Client + client.Do — 推荐的 HTTP 请求方式
// ============================================================
//
// 【为什么不用 http.Get？】
//   http.Get 是最简单的，但它用全局默认客户端，无法自定义超时、连接池等。
//   实际项目中 99% 的场景都用 http.Client + client.Do。
//
// 【核心概念】
//   1. http.Client — HTTP 客户端，负责发送请求和管理连接
//   2. http.Request — 请求对象，包含方法、URL、请求头、请求体
//   3. client.Do(req) — 发送请求并返回响应
//
// 【运行方式】
//   go run 02_client_do/main.go

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	fmt.Println("===== 知识点 1：创建 http.Client =====")

	// ---- http.Client 结构体 ----
	//
	// type Client struct {
	//     Transport   RoundTripper  // 传输层（连接池），默认是 http.DefaultTransport
	//     CheckRedirect func(*Request, []*Request) error  // 重定向策略
	//     Jar         CookieJar     // Cookie 存储（一般用不到）
	//     Timeout     time.Duration // 整个请求的超时时间
	// }
	//
	// 创建方式 1：空客户端（使用默认 Transport，无超时）
	client := &http.Client{}
	fmt.Println("方式1：空客户端创建成功")

	// 创建方式 2：带超时的客户端（推荐！）
	// Timeout 是整个请求的生命周期：连接 + 发送 + 接收 + 重定向
	// 超过这个时间，请求会被自动取消
	// clientWithTimeout := &http.Client{
	//     Timeout: 10 * time.Second,
	// }

	fmt.Println("\n===== 知识点 2：创建 http.Request =====")

	// ---- 创建请求的两种方式 ----

	// 方式 1：http.NewRequest（基本版）
	// 签名：func NewRequest(method, url string, body io.Reader) (*Request, error)
	// 第三个参数 body 是 io.Reader 接口，GET 请求传 nil，POST 请求传数据源
	req1, err := http.NewRequest(http.MethodGet, "https://www.baidu.com", nil)
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}
	fmt.Println("方式1：请求创建成功，方法:", req1.Method, "URL:", req1.URL)

	// 方式 2：http.NewRequestWithContext（推荐！）
	// 签名：func NewRequestWithContext(ctx context.Context, method, url string, body io.Reader) (*Request, error)
	// 可以绑定 context，实现超时取消等高级控制
	// 这个在下一课详细讲，这里先知道有这个函数
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	// ---- 请求方法常量 ----
	// Go 提供了常量，比手写字符串更安全（拼写错误编译器会报错）：
	//   http.MethodGet    = "GET"
	//   http.MethodPost   = "POST"
	//   http.MethodPut    = "PUT"
	//   http.MethodDelete = "DELETE"
	//   http.MethodHead   = "HEAD"
	//   http.MethodPatch  = "PATCH"
	//   http.MethodOptions = "OPTIONS"

	fmt.Println("\n===== 知识点 3：发送请求 client.Do(req) =====")

	// ---- client.Do 做了什么？ ----
	//
	// 1. 检查请求的 context 是否已取消 → 已取消则立即返回错误
	// 2. 从连接池获取 TCP 连接（如果没有就新建一个）
	// 3. 把 HTTP 请求写入连接
	// 4. 从连接读取 HTTP 响应
	// 5. 如果状态码是 3xx，自动跟随重定向（最多 10 次）
	// 6. 返回 *http.Response
	//
	// 重要：即使返回 404/500，err 也是 nil！只有网络层出错时 err 才非 nil。

	resp, err := client.Do(req1)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("状态码:", resp.StatusCode)
	fmt.Println("状态描述:", resp.Status) // 如 "200 OK"
	fmt.Println("协议:", resp.Proto)
	fmt.Println("ETag:", resp.Header.Get("ETag"))                   // 获取响应头中的 ETag 字段
	fmt.Println("Last-Modified:", resp.Header.Get("Last-Modified")) // 获取响应头中的 Last-Modified 字段

	fmt.Println("\n===== 知识点 4：读取响应体 =====")

	// io.ReadAll 一次性读完响应体，返回 []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应体失败:", err)
		return
	}

	// 转成字符串，打印前 150 个字符
	content := string(body)
	if len(content) > 150 {
		content = content[:150] + "...（已截断）"
	}
	fmt.Println("响应体:", content)

	fmt.Println("\n===== 知识点 5：对比 http.Get 和 client.Do =====")

	// http.Get(url) 内部实现（简化版）：
	//   func Get(url string) (resp *Response, err error) {
	//       req, err := NewRequest("GET", url, nil)
	//       if err != nil {
	//           return nil, err
	//       }
	//       return DefaultClient.Do(req)  // ← 用的全局默认客户端！
	//   }
	//
	// 所以 http.Get 本质上也是 Client.Do，只是你没法自定义 Client。

	// 等价写法对比：
	//
	// ❌ http.Get — 不能自定义，不推荐用于生产
	// resp, err := http.Get("https://www.baidu.com")
	//
	// ✅ http.Client.Do — 可以自定义超时、连接池等
	// client := &http.Client{Timeout: 10 * time.Second}
	// req, _ := http.NewRequest("GET", "https://www.baidu.com", nil)
	// resp, err := client.Do(req)

	fmt.Println("总结：推荐用 http.Client + client.Do，功能更强大、更安全")
}

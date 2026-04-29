package main

// ============================================================
// 第 01 课：http.Get — 最简单的 HTTP 请求
// ============================================================
//
// 【你要知道什么】
//   http.Get 是 net/http 包里最简单的用法，一行代码就能发起 HTTP 请求。
//   适合快速测试、简单脚本，但生产环境不推荐（后面会讲为什么）。
//
// 【核心概念】
//   1. http.Get(url) 会帮你完成：创建请求 → 发送请求 → 等待响应 → 返回结果
//   2. 它使用的是全局默认客户端 http.DefaultClient
//   3. 返回两个值：resp（响应）和 err（错误）
//
// 【运行方式】
//   cd 到本目录，执行：go run 01_http_get/main.go

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	// ---- 知识点 1：发起一个 GET 请求 ----
	//
	// http.Get 的函数签名：
	//   func Get(url string) (resp *Response, err error)
	//
	// 参数：
	//   url — 必须是完整地址，包含协议头（http:// 或 https://）
	//       常见错误：忘记写 http:// 会导致 "unsupported protocol scheme"
	//
	// 返回值：
	//   resp — *http.Response 指针，包含状态码、响应头、响应体等
	//   err  — 网络层错误（连接失败、超时、DNS 解析失败等）
	//        注意：404、500 等状态码错误不算 err！err 仍然是 nil！

	fmt.Println("===== 知识点 1：用 http.Get 获取百度首页 =====")

	resp, err := http.Get("https://www.baidu.com")
	if err != nil {
		// 走到这里说明网络层出了问题（断网、DNS 失败等）
		fmt.Println("请求失败:", err)
		return
	}
	// 走到这里说明请求成功发送了，但还要看状态码才知道服务器怎么响应

	// ---- 知识点 2：必须关闭 resp.Body ----
	//
	// resp.Body 是一个 io.ReadCloser，底层持有 TCP 连接。
	// 如果不关闭，会导致连接泄漏（TCP 连接无法被回收）。
	// 标准写法：收到 resp 后立刻 defer Close。
	defer resp.Body.Close()

	// ---- 知识点 3：检查状态码 ----
	//
	// resp.StatusCode 是 int 类型：
	//   200 = 成功
	//   301/302 = 重定向
	//   404 = 页面不存在
	//   500 = 服务器内部错误
	//
	// http.StatusOK 是常量，值等于 200。
	// 推荐用常量而不是直接写数字 200，更安全（写错数字编译器会报错）。
	fmt.Println("状态码:", resp.StatusCode)
	fmt.Println("协议版本:", resp.Proto) // HTTP/1.1 或 HTTP/2.0

	// ---- 知识点 4：读取响应体 ----
	//
	// resp.Body 是 io.ReadCloser 接口，不能直接当字符串用。
	// 需要借助 io.ReadAll 一次性读完，返回 []byte。
	//
	// io.ReadAll 的内部逻辑（简化版）：
	//   func ReadAll(r io.Reader) ([]byte, error) {
	//       var buf []byte
	//       for {
	//           // 每次读一些数据，追加到 buf
	//           n, err := r.Read(buf[...])
	//           if err == io.EOF { // 读完了
	//               return buf, nil
	//           }
	//       }
	//   }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应体失败:", err)
		return
	}

	// body 是 []byte 类型，用 string() 转成字符串打印
	// 注意：网页内容可能很长，这里只打印前 200 个字符
	content := string(body)
	if len(content) > 200 {
		content = content[:200] + "...（已截断）"
	}
	fmt.Println("响应体（前 200 字符）:", content)

	// ---- 知识点 5：读取响应头 ----
	//
	// resp.Header 的类型是 http.Header，底层是 map[string][]string。
	// 一个响应头可以有多个值（比如多个 Set-Cookie），所以值是切片。
	//
	// 常用方法：
	//   resp.Header.Get("Content-Type")     — 获取单个值（取第一个）
	//   resp.Header.Values("Set-Cookie")    — 获取所有值（切片）
	//   resp.Header["Content-Type"]         — 直接用 map 语法
	fmt.Println("\nContent-Type:", resp.Header.Get("Content-Type"))
	fmt.Println("Server:", resp.Header.Get("Server"))
	fmt.Println("Set-Cookie:", resp.Header.Values("Set-Cookie")) // 可能有多个 Set-Cookie

	// ---- 知识点 6：http.Get 的局限性（为什么不推荐用于生产） ----
	//
	// http.Get 底层用的是 http.DefaultClient：
	//
	//   var DefaultClient = &Client{}
	//
	// 默认 Client 的 Timeout 是 0（永不超时）！
	// 这意味着如果对方服务器卡死，你的程序会永远等下去。
	//
	// 而且你没法自定义：
	//   - 超时时间
	//   - 连接池大小
	//   - TLS 配置
	//   - 请求头（比如 User-Agent）
	//
	// 所以后面我们会用 http.Client + client.Do() 的方式，功能更强大。

	fmt.Println("\n===== 总结 =====")
	fmt.Println("http.Get 适合快速测试，但生产环境应该用 http.Client.Do")
	fmt.Println("核心三件事：发请求 → 检查状态码 → 读响应体（记得 Close！）")
}

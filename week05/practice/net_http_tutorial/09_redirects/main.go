package main

// ============================================================
// 第 09 课：重定向处理 — 301、302 和自动跟随
// ============================================================
//
// 【重定向是什么？】
//   服务器告诉客户端："你要找的资源不在这里，去另一个地址吧"。
//   常见场景：
//   - HTTP → HTTPS 升级（301）
//   - 短链接展开（302）
//   - 登录后跳转到原页面（302）
//
// 【http.Client 的默认行为】
//   http.Client 会自动跟随重定向，最多 10 次。
//   你不需要手动处理 301/302，除非你想拦截或控制重定向。
//
// 【运行方式】
//   go run 09_redirects/main.go

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func main() {
	// 启动测试服务器，模拟重定向
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/redirect-301":
			// 301 永久重定向
			// 浏览器会缓存这个结果，下次直接访问新地址
			http.Redirect(w, r, "/final", http.StatusMovedPermanently)
		case "/redirect-302":
			// 302 临时重定向
			// 浏览器每次都会先访问原地址
			http.Redirect(w, r, "/final", http.StatusFound)

		case "/redirect-chain":
			// 链式重定向：A → B → C → final
			http.Redirect(w, r, "/step2", http.StatusFound)

		case "/step2":
			http.Redirect(w, r, "/step3", http.StatusFound)

		case "/step3":
			http.Redirect(w, r, "/final", http.StatusFound)

		case "/final":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("你到达了终点！"))

		case "/no-redirect":
			// 自定义重定向策略：不允许重定向
			w.WriteHeader(http.StatusFound)
			w.Write([]byte("手动处理重定向"))

		default:
			fmt.Printf("未找到路径: %s\n", r.URL.Path)
			for k, v := range r.Header {
				for id, vv := range v {
					fmt.Printf("Header[%d]: %s: %s\n", id, k, vv)
				}
			}
			w.Write([]byte(fmt.Sprintf("你好啊 %s", r.RemoteAddr+r.URL.String())))
			http.NotFound(w, r)
		}
	}))
	fmt.Println("测试服务器启动在:", server.URL)
	ch := make(chan struct{})
	<-ch
	// defer server.Close()
	// ================================================================
	// 知识点 1：自动跟随重定向
	// ================================================================

	fmt.Println("===== 知识点 1：自动跟随重定向 =====")

	// http.Client 默认会自动跟随重定向（最多 10 次）
	// 你收到的是最终的结果，中间的重定向过程你感知不到

	// 301 重定向：自动跟随
	resp1, err := http.Get(server.URL + "/redirect-301")
	if err != nil {
		fmt.Println("请求失败:", err)
	} else {
		defer resp1.Body.Close()
		body, _ := io.ReadAll(resp1.Body)
		// 注意：虽然经过了重定向，但最终拿到的状态码是最终页面的状态码
		fmt.Println("301 重定向后，状态码:", resp1.StatusCode)
		fmt.Println("响应体:", string(body))
		// resp.Request.URL 是最终请求的 URL（重定向后的）
		fmt.Println("最终 URL:", resp1.Request.URL.Path)
	}

	// 302 重定向：同样自动跟随
	resp2, _ := http.Get(server.URL + "/redirect-302")
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	fmt.Println("\n302 重定向后，状态码:", resp2.StatusCode)
	fmt.Println("响应体:", string(body2))

	// 链式重定向：A → B → C → final（3 次重定向）
	resp3, _ := http.Get(server.URL + "/redirect-chain")
	defer resp3.Body.Close()
	body3, _ := io.ReadAll(resp3.Body)
	fmt.Println("\n链式重定向后，状态码:", resp3.StatusCode)
	fmt.Println("响应体:", string(body3))

	// ---- http.Redirect 函数 ----
	//
	//   func Redirect(w ResponseWriter, r *Request, url string, code int)
	//
	// 它是一个便捷函数，会设置：
	//   Location: <url>  响应头
	//   状态码: <code>   （301, 302, 307, 308 等）
	//
	// 常见重定向状态码：
	//   301 Moved Permanently    — 永久重定向（搜索引擎会更新索引）
	//   302 Found                — 临时重定向
	//   307 Temporary Redirect   — 临时重定向，保持请求方法不变
	//   308 Permanent Redirect   — 永久重定向，保持请求方法不变

	// ================================================================
	// 知识点 2：自定义重定向策略
	// ================================================================

	fmt.Println("\n===== 知识点 2：自定义重定向策略 =====")

	// CheckRedirect 是 http.Client 的一个字段，类型是：
	//   func(req *Request, via []*Request) error
	//
	// 参数：
	//   req  — 即将跟随的重定向请求
	//   via  — 已经经过的请求历史
	//
	// 返回值：
	//   nil     — 允许跟随这个重定向
	//   error   — 阻止跟随，返回这个错误
	//
	// 默认行为是：
	//   func checkRedirect(req *Request, via []*Request) error {
	//       if len(via) >= 10 {
	//           return errors.New("stopped after 10 redirects")
	//       }
	//       return nil
	//   }

	// ---- 场景 A：完全禁止重定向 ----

	clientNoRedirect := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// 返回错误，阻止任何重定向
			return http.ErrUseLastResponse
		},
	}

	respA, err := clientNoRedirect.Get(server.URL + "/redirect-302")
	if err != nil {
		// http.ErrUseLastResponse 是一个特殊错误，表示"使用最后一次响应"
		// client 会停止跟随重定向，但不会把它当作错误返回给你
		// 所以这里 err == nil，respA 是重定向响应
		if err != http.ErrUseLastResponse {
			fmt.Println("请求失败:", err)
		}
	}
	if respA != nil {
		defer respA.Body.Close()
		fmt.Println("禁止重定向后，状态码:", respA.StatusCode, "(302)")
		// Location 头告诉你重定向的目标
		fmt.Println("Location:", respA.Header.Get("Location"))
	}

	// ---- 场景 B：只允许最多 2 次重定向 ----

	clientLimitedRedirect := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 2 {
				return fmt.Errorf("超过 2 次重定向限制")
			}
			return nil
		},
	}

	fmt.Println("\n--- 最多 2 次重定向 ---")
	respB, err := clientLimitedRedirect.Get(server.URL + "/redirect-chain")
	// redirect-chain 需要 3 次重定向，但我们限制了最多 2 次
	if err != nil {
		fmt.Println("被阻止:", err)
	} else {
		defer respB.Body.Close()
		fmt.Println("状态码:", respB.StatusCode)
	}

	fmt.Println("\n===== 总结 =====")
	fmt.Println("http.Client 默认自动跟随重定向（最多 10 次）")
	fmt.Println("301 → 永久，302 → 临时")
	fmt.Println("CheckRedirect 可以自定义重定向策略")
	fmt.Println("http.ErrUseLastResponse 可以阻止重定向但保留响应")
}

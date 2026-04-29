package main

// ============================================================
// 第 05 课：POST 请求 — 提交表单和 JSON 数据
// ============================================================
//
// 【GET vs POST】
//   GET  — 从服务器获取数据，没有请求体（body 为 nil）
//   POST — 向服务器提交数据，数据放在请求体中
//
// 【POST 的三种常见方式】
//   1. http.PostForm — 提交表单数据（application/x-www-form-urlencoded）
//   2. http.Post + strings.NewReader — 提交 JSON 字符串
//   3. client.Do + bytes.NewBuffer — 最灵活的方式
//
// 【运行方式】
//   go run 05_post_request/main.go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func main() {
	// 启动一个测试服务器，接收 POST 请求并返回收到的数据
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 读取请求体
		body, _ := io.ReadAll(r.Body)
		defer r.Body.Close()

		// 返回收到的内容
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"method":"%s","content_type":"%s","body":"%s"}`,
			r.Method,
			r.Header.Get("Content-Type"),
			string(body),
		)))
	}))
	defer server.Close()

	// ================================================================
	// 方式 1：http.PostForm — 提交表单数据
	// ================================================================

	fmt.Println("===== 方式 1：http.PostForm（表单提交）=====")

	// 表单数据的格式是 application/x-www-form-urlencoded
	// 类似浏览器提交表单：name=张三&age=20&email=zhangsan@example.com
	//
	// url.Values 底层是 map[string][]string
	// 用 Add 方法添加键值对，自动处理 URL 编码（比如中文会被编码成 %E5%BC%A0%E4%B8%89）

	formData := map[string]string{
		"name":  "张三",
		"age":   "20",
		"email": "zhangsan@example.com",
	}

	// 转换为 url.Values
	values := make(map[string][]string)
	for k, v := range formData {
		values[k] = []string{v}
	}

	// http.PostForm 的签名：
	//   func PostForm(url string, data url.Values) (resp *Response, err error)
	//
	// 它会自动设置 Content-Type: application/x-www-form-urlencoded
	resp, err := http.PostForm(server.URL, values)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("服务端返回:", string(body))

	// ---- http.PostForm 内部做了什么？ ----
	//
	//   func PostForm(url string, data url.Values) (resp *Response, err error) {
	//       // 1. 把 map 编码成 "name=张三&age=20&email=zhangsan%40example.com"
	//       encodedData := data.Encode()
	//       // 2. 创建 POST 请求，body 是编码后的字符串
	//       req, _ := NewRequest("POST", url, strings.NewReader(encodedData))
	//       // 3. 设置 Content-Type
	//       req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//       // 4. 发送请求
	//       return DefaultClient.Do(req)
	//   }

	// ================================================================
	// 方式 2：http.Post — 提交任意数据
	// ================================================================

	fmt.Println("\n===== 方式 2：http.Post（直接提交字符串）=====")

	// http.Post 的签名：
	//   func Post(url, contentType string, body io.Reader) (resp *Response, err error)
	//
	// 第二个参数 contentType 要手动指定，比如 "application/json"
	// 第三个参数是 io.Reader 接口，任何实现了 Read 方法的类型都可以

	// 提交纯文本
	resp, err = http.Post(server.URL, "text/plain", strings.NewReader("Hello, 世界!"))
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Println("服务端返回:", string(body))

	// ================================================================
	// 方式 3：client.Do + JSON — 提交 JSON 数据（最常用！）
	// ================================================================

	fmt.Println("\n===== 方式 3：client.Do + JSON（推荐）=====")

	// 这是实际项目中最常用的方式：
	// 1. 定义结构体
	// 2. 用 json.Marshal 序列化为 []byte
	// 3. 用 bytes.NewBuffer 包装成 io.Reader
	// 4. 设置 Content-Type: application/json
	// 5. 用 client.Do 发送

	// 第 1 步：定义要发送的数据
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	reqBody := LoginRequest{
		Username: "admin",
		Password: "123456",
	}

	// 第 2 步：序列化为 JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("JSON 序列化失败:", err)
		return
	}
	fmt.Println("JSON 数据:", string(jsonData))

	// 第 3 步：创建 POST 请求
	req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("创建请求失败:", err)
		return
	}

	// 第 4 步：设置 Content-Type
	// 这很重要！服务端需要知道请求体的格式才能正确解析
	req.Header.Set("Content-Type", "application/json")

	// 第 5 步：发送请求
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, _ = io.ReadAll(resp.Body)
	fmt.Println("服务端返回:", string(body))

	// ---- 简化写法 ----
	// 也可以用 strings.NewReader 代替 bytes.NewBuffer，效果一样：
	//   req, _ := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))

	// ================================================================
	// 方式 4：提交文件 / 大文件
	// ================================================================

	fmt.Println("\n===== 方式 4：提交大文件（流式发送）=====")

	// 如果数据很大（比如上传文件），不要一次性加载到内存。
	// 直接用实现了 io.Reader 的文件对象作为 body。

	// 模拟大文件数据（实际项目中用 os.Open 打开文件）
	// largeData := strings.Repeat("x", 1024*1024*10)  // 10 MB
	//
	// req, _ := http.NewRequest("POST", url, strings.NewReader(largeData))
	// req.Header.Set("Content-Type", "application/octet-stream")
	// resp, err := client.Do(req)
	//
	// 这种方式不会把 10MB 全部加载到内存再发送，而是边读边发，节省内存。

	fmt.Println("流式发送：用 io.Reader 作为 body，边读边发，不需要全部加载到内存")

	fmt.Println("\n===== 总结 =====")
	fmt.Println("表单提交  → http.PostForm")
	fmt.Println("简单字符串 → http.Post")
	fmt.Println("JSON 数据  → http.NewRequest + json.Marshal + client.Do")
	fmt.Println("大文件    → io.Reader 作为 body，流式发送")
}

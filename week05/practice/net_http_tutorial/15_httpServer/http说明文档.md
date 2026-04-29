# net/http 核心类型速查手册

> Go 1.22+ 标准库 | `net/http` 包

---

## 一、`*http.Request` — 请求对象

服务端收到的每一个 HTTP 请求都会被封装成一个 `*http.Request` 对象。

### 字段（属性）

| 字段 | 类型 | 含义 | 示例值 |
|---|---|---|---|
| `Method` | `string` | HTTP 请求方法 | `"GET"`、`"POST"`、`"PUT"`、`"DELETE"` |
| `URL` | `*url.URL` | 请求的 URL 对象（见下方子字段） | — |
| `URL.Path` | `string` | URL 路径部分 | `"/users/123"` |
| `URL.RawQuery` | `string` | URL 查询参数字符串 | `"page=1&size=10"` |
| `URL.Scheme` | `string` | 协议，服务端通常为 `""`（空） | `"http"`（仅客户端创建时有值） |
| `URL.Host` | `string` | 主机，服务端通常为 `""`（空） | `"localhost:8080"`（仅客户端创建时有值） |
| `URL.Opaque` | `string` | 非标准格式 URL 的原始字符串 | 极少用到 |
| `Proto` | `string` | HTTP 协议版本 | `"HTTP/1.1"` |
| `ProtoMajor` | `int` | 协议主版本号 | `1`（HTTP/1.x）或 `2`（HTTP/2） |
| `ProtoMinor` | `int` | 协议次版本号 | `1` |
| `Header` | `http.Header` | 请求头，底层是 `map[string][]string` | — |
| `Body` | `io.ReadCloser` | 请求体（POST/PUT 时有数据，用完后需 Close） | — |
| `ContentLength` | `int64` | 请求体字节数（-1 表示未知） | `256` |
| `TransferEncoding` | `[]string` | 传输编码，如 `["chunked"]` | — |
| `Host` | `string` | 请求的 Host 头值 | `"localhost:8080"` |
| `RemoteAddr` | `string` | 客户端 IP:端口（TCP 层记录，不可伪造） | `"127.0.0.1:62345"` |
| `RequestURI` | `string` | 原始请求行中的 URI（未经解析） | `"/search?q=go&page=1"` |
| `Trailer` | `http.Header` | Trailer 请求头（HTTP/1.1 chunked 尾部头） | — |
| `TLS` | `*tls.ConnectionState` | TLS 连接信息（HTTPS 请求时非 nil） | — |
| `Cancel` | `<-chan struct{}` | **已废弃**，用 Context 代替 | — |
| `Response` | `*Response` | 客户端请求时才会填充，服务端为 nil | — |
| `Pattern` | `string` | 匹配此请求的路由模式（Go 1.22+） | `"/users/{id}"` |

### 方法

| 方法 | 返回 | 含义 |
|---|---|---|
| `FormValue(key string)` | `string` | 取表单/URL 参数的值。自动调用 `ParseForm()`，URL 参数和表单数据都会查 |
| `PostFormValue(key string)` | `string` | 只取 POST 请求体中的表单值，不查 URL 参数 |
| `FormFile(key string)` | `multipart.File, *multipart.FileHeader, error` | 获取上传的文件（multipart/form-data） |
| `MultipartReader()` | `*multipart.Reader, error` | 获取 multipart 读取器（用于大文件流式读取） |
| `ParseForm()` | `error` | 解析 URL 查询参数和请求体表单数据，填充 `r.Form` 和 `r.PostForm` |
| `ParseMultipartForm(maxMemory int64)` | `error` | 解析 multipart 表单，最多缓存 maxMemory 字节到内存，超出部分存临时文件 |
| `Cookie(name string)` | `*http.Cookie, error` | 获取指定名称的 Cookie |
| `Cookies()` | `[]*http.Cookie` | 获取所有 Cookie |
| `UserAgent()` | `string` | 快捷方法：`r.Header.Get("User-Agent")` |
| `Referer()` | `string` | 快捷方法：`r.Header.Get("Referer")` |
| `WithContext(ctx context.Context)` | `*Request` | 返回一个带新 context 的浅拷贝请求（原请求不变） |
| `PathValue(name string)` | `string` | 获取路径参数（Go 1.22+），路由为 `/users/{id}` 时取 `id` 的值 |
| `SetPathValue(name, value string)` | — | 设置路径参数值（Go 1.22+，较少用） |
| `BasicAuth()` | `username, password string, ok bool` | 解析 Basic 认证头，返回用户名和密码 |

### 使用示例

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // === 基本信息 ===
    fmt.Println(r.Method)           // "GET"
    fmt.Println(r.URL.Path)         // "/users"
    fmt.Println(r.URL.RawQuery)     // "page=1&size=10"
    fmt.Println(r.RemoteAddr)       // "127.0.0.1:62345"
    fmt.Println(r.Host)             // "localhost:8080"
    fmt.Println(r.Proto)            // "HTTP/1.1"
    fmt.Println(r.RequestURI)       // "/users?page=1&size=10"

    // === 路径参数（Go 1.22+）===
    id := r.PathValue("id")         // 路由 /users/{id} → "123"
    fmt.Println(r.Pattern)          // "/users/{id}"

    // === 查询参数 ===
    page := r.FormValue("page")     // "1"
    size := r.URL.Query().Get("size") // "10"

    // === 请求头 ===
    ua := r.UserAgent()             // "Mozilla/5.0 ..."
    ref := r.Referer()              // "http://example.com/"
    contentType := r.Header.Get("Content-Type")

    // === Cookie ===
    cookie, err := r.Cookie("session")
    cookies := r.Cookies()          // 所有 Cookie

    // === Basic 认证 ===
    user, pass, ok := r.BasicAuth() // 解析 "Authorization: Basic xxx"

    // === 表单数据 ===
    username := r.FormValue("username")
    password := r.PostFormValue("password")

    // === 文件上传 ===
    file, header, err := r.FormFile("avatar")
    defer file.Close()

    // === 读取请求体（JSON）===
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()
}
```

---

## 二、`http.ResponseWriter` — 响应写入器

`http.ResponseWriter` 是一个**接口**（不是结构体），由 Go 的 HTTP 服务器在调用 handler 时自动传入。

### 接口方法

| 方法 | 签名 | 含义 |
|---|---|---|
| `Header()` | `Header()` | 返回响应头 `http.Header`，可设置 Content-Type、Set-Cookie 等 |
| `Write()` | `Write([]byte) (int, error)` | 写入响应体数据 |
| `WriteHeader()` | `WriteHeader(statusCode int)` | 设置 HTTP 状态码（如 200、404、500） |

### `Header()` 的常用方法

`r.Header()` 返回 `http.Header`（`map[string][]string`），常用操作：

| 方法 | 含义 | 示例 |
|---|---|---|
| `Set(key, value)` | 设置一个响应头（覆盖已有值） | `w.Header().Set("Content-Type", "application/json")` |
| `Add(key, value)` | 追加一个响应头（不覆盖） | `w.Header().Add("Set-Cookie", "a=1")` |
| `Get(key)` | 获取第一个值 | `w.Header().Get("Content-Type")` |
| `Del(key)` | 删除一个响应头 | `w.Header().Del("X-Powered-By")` |

### 关键规则

1. **`Header()` 必须在 `WriteHeader()` 和 `Write()` 之前调用**
   ```go
   w.Header().Set("Content-Type", "application/json")  // ✅ 先设置头
   w.WriteHeader(http.StatusOK)                         // ✅ 再写状态码
   w.Write([]byte(`{"msg":"ok"}`))                      // ✅ 最后写响应体
   ```

2. **如果不显式调用 `WriteHeader()`**，第一次 `Write()` 时会自动发送 `200 OK`

3. **`WriteHeader()` 只能调用一次**，多次调用无效（不会报错，但只有第一次生效）

4. **`WriteHeader()` 之后再调用 `Header()` 无效**，响应头已经发送出去了

### 常用响应头

| 响应头 | 含义 | 示例 |
|---|---|---|
| `Content-Type` | 响应体的 MIME 类型 | `application/json`、`text/html; charset=utf-8` |
| `Content-Length` | 响应体字节数（通常自动设置） | `256` |
| `Set-Cookie` | 设置 Cookie | `session=abc123; Path=/; HttpOnly` |
| `Location` | 重定向地址（配合 301/302） | `https://example.com/` |
| `Cache-Control` | 缓存控制 | `no-cache`、`max-age=3600` |
| `Access-Control-Allow-Origin` | CORS 跨域 | `*` 或具体域名 |

### 使用示例

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // === 设置响应头（必须在 WriteHeader 之前）===
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("X-Request-Id", "abc-123")
    w.Header().Add("Set-Cookie", "session=xyz; Path=/; HttpOnly")

    // === 设置状态码（必须在 Write 之前，只能调用一次）===
    w.WriteHeader(http.StatusOK)  // 200
    // w.WriteHeader(http.StatusCreated)  // 201
    // w.WriteHeader(http.StatusNotFound)  // 404

    // === 写入响应体 ===
    w.Write([]byte(`{"message": "success"}`))

    // === 便捷写法 ===
    // 返回 JSON
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"msg": "ok"})

    // === 快捷函数（内部已设置好状态码和头）===
    http.Error(w, "Not Found", http.StatusNotFound)      // 404 + 纯文本
    http.Redirect(w, r, "/new-url", http.StatusFound)    // 302 重定向
    http.ServeFile(w, r, "./index.html")                  // 返回文件
}
```

---

## 三、`http.Response` — 客户端收到的响应

这是**客户端**发请求后收到的响应对象（`*http.Response`），与服务端的 `ResponseWriter` 不同。

| 字段 | 类型 | 含义 |
|---|---|---|
| `Status` | `string` | 状态行，如 `"200 OK"` |
| `StatusCode` | `int` | 状态码，如 `200`、`404`、`500` |
| `Proto` | `string` | 协议版本，如 `"HTTP/1.1"` |
| `Header` | `http.Header` | 响应头 |
| `Body` | `io.ReadCloser` | 响应体（必须 Close！） |
| `ContentLength` | `int64` | 响应体长度（-1 表示未知） |
| `Request` | `*Request` | 原始请求对象（方便追溯） |
| `TLS` | `*tls.ConnectionState` | TLS 连接信息（HTTPS 时非 nil） |

### 客户端使用示例

```go
resp, err := http.Get("https://www.baidu.com")
if err != nil {
    // 网络错误
    log.Fatal(err)
}
defer resp.Body.Close()  // 必须关闭！

fmt.Println(resp.StatusCode)         // 200
fmt.Println(resp.Status)             // "200 OK"
fmt.Println(resp.Header.Get("Content-Type"))  // "text/html"

body, _ := io.ReadAll(resp.Body)
fmt.Println(string(body))
```

---

## 四、速查对比

| | `*http.Request` | `http.ResponseWriter` | `*http.Response` |
|---|---|---|---|
| 谁持有 | 服务端 | 服务端 | 客户端 |
| 类型 | 结构体指针 | 接口 | 结构体指针 |
| 核心用途 | 读请求信息 | 写响应信息 | 读响应信息 |
| 改头 | `r.Header.Set()`（发请求时） | `w.Header().Set()` | 只读 |
| 改状态码 | — | `w.WriteHeader(code)` | 只读 |
| Body | 读请求体 `r.Body` | 写响应体 `w.Write()` | 读响应体 `resp.Body` |

---

## 五、`http.Server` — HTTP 服务器实例

`http.Server` 是一个结构体，代表一个完整的 HTTP 服务器。它封装了监听、路由分发、连接管理和生命周期控制。

### 核心字段

| 字段 | 类型 | 含义 | 示例值 |
|---|---|---|---|
| `Addr` | `string` | 监听地址，格式 `"host:port"`，端口为空则使用 `:http`（80） | `":8080"` |
| `Handler` | `http.Handler` | 请求处理器，`nil` 时使用 `http.DefaultServeMux` | `mux` |
| `ReadTimeout` | `time.Duration` | 从连接建立到读完请求头的最大时间 | `5 * time.Second` |
| `WriteTimeout` | `time.Duration` | 从请求处理完毕到写完响应的最大时间 | `10 * time.Second` |
| `IdleTimeout` | `time.Duration` | keep-alive 连接的最大空闲时间 | `120 * time.Second` |
| `MaxHeaderBytes` | `int` | 请求头的最大字节数（超过返回 431） | `1 << 20`（1MB） |
| `TLSConfig` | `*tls.Config` | TLS 配置，用于 HTTPS | — |
| `TLSNextProto` | `map[string]func(...)` | 协议升级回调（如 h2） | — |
| `ConnState` | `func(net.Conn, ConnState)` | 连接状态变化回调 | — |
| `ErrorLog` | `*log.Logger` | 错误日志，默认 `log.Printf` | — |
| `BaseContext` | `func(net.Listener) context.Context` | 自定义 server 基础 context | — |
| `ConnContext` | `func(context.Context, net.Conn) context.Context` | 为每个连接自定义 context | — |

### 核心方法

| 方法 | 返回 | 含义 |
|---|---|---|
| `ListenAndServe()` | `error` | 监听 TCP 地址并处理 HTTP 请求，阻塞调用 |
| `ListenAndServeTLS(certFile, keyFile string)` | `error` | 同上，但启用 HTTPS |
| `Serve(l net.Listener)` | `error` | 在已有 Listener 上服务 HTTP 请求 |
| `Shutdown(ctx context.Context)` | `error` | 优雅关闭：停止接受新连接，等待已有连接处理完毕 |
| `Close()` | `error` | 立即关闭所有连接 |

### 使用示例

```go
server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
    MaxHeaderBytes: 1 << 20, // 1MB
}

// 启动（阻塞）
go func() {
    if err := server.ListenAndServe(); err != http.ErrServerClosed {
        log.Fatal(err)
    }
}()

// 优雅关闭
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(ctx)
```

### 运行结果

```
Server.Addr:            :8080
Server.Handler:         <nil>
Server.ReadTimeout:     5s
Server.WriteTimeout:    10s
Server.IdleTimeout:     2m0s
Server.MaxHeaderBytes:  1048576 字节

--- 启动方式 ---
  server.ListenAndServe()              → 阻塞启动 HTTP
  server.ListenAndServeTLS(cert,key)   → 阻塞启动 HTTPS
  go server.ListenAndServe()           → 非阻塞启动（goroutine）

--- 停止方式 ---
  server.Close()       → 立即关闭，正在处理的请求被中断
  server.Shutdown(ctx) → 优雅关闭，等正在处理的请求完成

--- 安全要点 ---
  ReadTimeout  → 防止慢速客户端攻击（Slowloris）
  WriteTimeout → 防止服务端处理太慢
  IdleTimeout  → 空闲连接过期，释放资源
```

---

## 六、`http.ServeMux` — 路由多路复用器

`http.ServeMux` 是 Go 标准库内置的 HTTP 路由器（多路复用器），负责将请求分发给对应的 handler。

### 核心方法

| 方法 | 签名 | 含义 |
|---|---|---|
| `HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))` | — | 注册函数型 handler |
| `Handle(pattern string, handler http.Handler)` | — | 注册 `http.Handler` 接口型 handler |
| `Handler(r *http.Request)` | `(h Handler, pattern string)` | 根据请求匹配路由，返回匹配的 handler 和模式 |
| `ServeHTTP(w, r)` | — | 实现 `http.Handler` 接口，使 ServeMux 本身可作为 handler |

### Go 1.22+ 路由模式语法

| 模式类型 | 示例 | 说明 |
|---|---|---|
| 精确匹配 | `/api/users` | 仅匹配 `GET /api/users`（Go 1.22+ 默认带方法限定） |
| 前缀匹配 | `/static/` | 匹配 `/static/` 开头的所有路径（必须以 `/` 结尾） |
| 方法+路径 | `GET /users/{id}` | 限定 HTTP 方法 + 路径参数 |
| 通配符 | `/api/{rest...}` | 匹配 `/api/` 下任意深层路径 |
| 主机匹配 | `example.com/` | 匹配指定 Host 的请求 |

### 路径参数提取

```go
mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")  // 提取路径参数
    w.Write([]byte("user id: " + id))
})

mux.HandleFunc("/api/{rest...}", func(w http.ResponseWriter, r *http.Request) {
    rest := r.PathValue("rest")  // 通配符捕获剩余路径
    w.Write([]byte("wildcard: " + rest))
})
```

### 优先级规则

1. 方法+路径 > 精确路径 > 前缀路径 > 通配符
2. 先匹配方法（GET/POST），不匹配则返回 `405 Method Not Allowed`
3. 再匹配路径（精确 > 前缀 > 通配符）
4. 没有匹配则返回 `404 Not Found`

### 运行结果

```
--- Go 1.22+ 新路由语法 ---
  精确匹配:   mux.HandleFunc("/api/users", handler)
  前缀匹配:   mux.HandleFunc("/static/", handler)
  方法+路径:  mux.HandleFunc("GET /users/{id}", handler)
  通配符:     mux.HandleFunc("/api/{rest...}", handler)

  优先级: 方法+路径 > 精确路径 > 前缀路径 > 通配符

--- 路由匹配规则 ---
  1. 先匹配方法（GET/POST），不匹配则返回 405 Method Not Allowed
  2. 再匹配路径（精确 > 前缀 > 通配符）
  3. 没有匹配则返回 404 Not Found
  4. 多个 ServeMux 互不干扰，可以嵌套
```

---

## 七、`http.Client` — HTTP 客户端

`http.Client` 是 Go 标准库的 HTTP 客户端，负责发起 HTTP 请求并获取响应。

### 核心字段

| 字段 | 类型 | 含义 |
|---|---|---|
| `Transport` | `RoundTripper` | 传输层实现（连接池管理），默认使用 `http.DefaultTransport` |
| `CheckRedirect` | `func(req *Request, via []*Request) error` | 重定向策略函数，返回 error 可阻止重定向 |
| `Jar` | `CookieJar` | Cookie 管理器，`nil` 表示不自动处理 Cookie |
| `Timeout` | `time.Duration` | 整个请求（含重定向）的超时时间，0 表示不超时 |

### 核心方法

| 方法 | 返回 | 含义 |
|---|---|---|
| `Get(url string)` | `(*Response, error)` | 发送 GET 请求（快捷方法） |
| `Post(url, contentType string, body io.Reader)` | `(*Response, error)` | 发送 POST 请求（快捷方法） |
| `PostForm(url string, data url.Values)` | `(*Response, error)` | 发送表单 POST 请求 |
| `Do(req *http.Request)` | `(*Response, error)` | 发送自定义请求（**推荐使用**） |
| `CloseIdleConnections()` | — | 关闭所有空闲连接 |

### 三种请求方式

#### 方式一：`http.Get`（最简单）

```go
resp, err := http.Get("http://localhost:8080/hello")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
```

#### 方式二：`http.Client.Do`（推荐，可自定义一切）

```go
client := &http.Client{Timeout: 5 * time.Second}

req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/json", nil)
req.Header.Set("Accept", "application/json")
req.Header.Set("User-Agent", "MyClient/1.0")

resp, err := client.Do(req)
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()
```

#### 方式三：超时控制

```go
// 粗粒度：Client 级别超时
client := &http.Client{Timeout: 1 * time.Second}

// 细粒度：每个请求独立超时（推荐）
ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
defer cancel()
req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### 运行结果

```
--- http.Client 结构体字段 ---
  Transport       RoundTripper  — 传输层（连接池）
  CheckRedirect   func(...)     — 重定向策略
  Jar             CookieJar     — Cookie 存储
  Timeout         time.Duration — 整个请求的超时时间

--- 方式一：http.Get（最简单）---
  resp.StatusCode: 200
  resp.Status:     200 OK
  resp.Proto:      HTTP/1.1
  resp.Body:       Hello from server!

--- 方式二：http.Client.Do（推荐）---
  resp.StatusCode: 200
  resp.Header[Content-Type]: application/json
  resp.Body:       {"name":"Tom","age":20}

  → 可自定义超时、请求头、Transport

--- 方式三：超时控制 ---
  方式 A: client.Timeout（粗粒度）
    → 所有请求统一 1 秒超时
  方式 B: context.WithTimeout（细粒度）
    ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    → 每个请求独立超时，支持手动取消
```

---

## 八、`http.Transport` — 传输层（连接池管理）

`http.Transport` 是 `http.Client` 的默认 `RoundTripper` 实现，负责管理底层 TCP 连接池、TLS 握手、代理、gzip 压缩等传输细节。

### 核心字段

| 字段 | 类型 | 含义 |
|---|---|---|
| `Proxy` | `func(*http.Request) (*url.URL, error)` | 代理函数，返回代理 URL，`nil` 表示不使用代理 |
| `DialContext` | `func(ctx, network, addr string) (net.Conn, error)` | 自定义 TCP 拨号逻辑 |
| `DialTLSContext` | `func(ctx, network, addr string) (net.Conn, error)` | 自定义 TLS 拨号逻辑 |
| `TLSClientConfig` | `*tls.Config` | TLS 客户端配置（证书验证等） |
| `MaxIdleConns` | `int` | 所有 host 的最大空闲连接总数，0 表示无限制 |
| `MaxIdleConnsPerHost` | `int` | 每个 host 的最大空闲连接数，默认 `2` |
| `MaxConnsPerHost` | `int` | 每个 host 的最大连接数（含活跃连接），0 表示无限制 |
| `IdleConnTimeout` | `time.Duration` | 空闲连接过期时间，0 表示永不过期 |
| `ResponseHeaderTimeout` | `time.Duration` | 等待响应头的超时时间 |
| `TLSHandshakeTimeout` | `time.Duration` | TLS 握手超时时间 |
| `DisableKeepAlives` | `bool` | `true` 表示禁用连接复用（每个请求新建连接） |
| `DisableCompression` | `bool` | `true` 表示禁用自动 gzip 压缩 |
| `ExpectContinueTimeout` | `time.Duration` | `Expect: 100-continue` 等待确认的超时 |

### 核心方法

| 方法 | 返回 | 含义 |
|---|---|---|
| `RoundTrip(req *http.Request)` | `(*http.Response, error)` | 实现 `RoundTripper` 接口，执行单次 HTTP 事务 |
| `CloseIdleConnections()` | — | 关闭所有空闲连接 |
| `Clone()` | `*Transport` | 创建浅拷贝（安全修改配置用） |
| `RegisterProtocol(scheme string, rt RoundTripper)` | — | 注册自定义协议处理器 |

### 运行结果

```
--- http.Transport 核心字段 ---
  MaxIdleConns:        100  （所有 host 的最大空闲连接总数）
  MaxIdleConnsPerHost: 10  （每个 host 的最大空闲连接数）
  MaxConnsPerHost:     50  （每个 host 的最大连接数，含活跃）
  IdleConnTimeout:     1m30s （空闲连接过期时间）
  TLSHandshakeTimeout: 10s （TLS 握手超时）
  DisableKeepAlives:   false  （false=启用连接复用）
  DisableCompression:  false  （false=启用 gzip 压缩）

--- Transport 的生命周期 ---
  第 1 次请求 → 建立连接 → 响应: Hello from server!
  第 2 次请求 → 复用连接 → 响应: Hello from server!
  transport.CloseIdleConnections() → 释放所有空闲连接

--- Transport 与 Client 的关系 ---
  一个 Client 包含一个 Transport
  多个 goroutine 共享同一个 Client = 共享同一个连接池（推荐）
  每个请求都新建 Client = 每次都新建连接池（性能差，不要这样）
```

---

## 九、完整架构层次

```
  你的代码层:
    client.Do(req)           ← 发送请求
    handler(w, r)            ← 处理请求
          │
          ▼
  http.Client / http.Server  ← 调度层
    - Client: 管理 Transport、重定向、Cookie、超时
    - Server: 管理 ServeMux、读写超时、连接管理
          │
          ▼
  http.Transport              ← 传输层（仅客户端）
    - 连接池（keep-alive）
    - DNS 解析、TCP 连接、TLS 握手
    - gzip 压缩/解压
          │
          ▼
  net.Conn                    ← 网络层
    - 底层 TCP socket
    - 读写字节流
```

---

## 十、对象关系总结

### 客户端发送请求

| 对象 | 职责 | 关键操作 |
|---|---|---|
| `Client` | 调度请求 | 管理超时、重定向、Cookie |
| `Transport` | 连接池 | 管理 TCP 连接复用 |
| `Request` | 请求信息 | 携带方法、URL、请求头、请求体 |
| `Response` | 响应信息 | 携带状态码、响应头、响应体 |

### 服务端接收请求

| 对象 | 职责 | 关键操作 |
|---|---|---|
| `Server` | 监听端口 | 管理连接、超时、TLS |
| `ServeMux` | 路由匹配 | 将请求分发给对应 handler |
| `Request` | 请求信息 | 携带方法、URL、请求头、请求体 |
| `ResponseWriter` | 构建响应 | 设置状态码、响应头、写入响应体 |

### 关键对应关系

| 客户端 | ↔ | 服务端 |
|---|---|---|
| `Client` | ↔ | `Server` |
| `Transport` | ↔ | （无对应，传输层由 OS 处理） |
| `Request`（发出） | ↔ | `Request`（收到） |
| `Response`（收到） | ↔ | `ResponseWriter`（写入） |

### 一句话记忆

- **客户端**: `Client` → `Transport` → `Request` → 收到 `Response`
- **服务端**: `Server` → `ServeMux` → `Request` → 写入 `ResponseWriter`
- `Request` 是共用的（客户端创建、服务端接收）
- 客户端收到 `Response`，服务端通过 `ResponseWriter` 构建 `Response`

# 服务健康探测器：我为完成作业需要学习的标准库

## 先说结论

这次作业不是只学一个 `net/http` 就能完成的。  
按照题目要求，我至少需要把下面这些标准库分层掌握：

### 第一层：必须掌握

- `encoding/json`
- `flag`
- `os`
- `time`
- `net/http`
- `net`
- `sync`
- `io`
- `strings`
- `fmt`

### 第二层：强烈建议掌握

- `context`
- `sort`
- `text/tabwriter`
- `errors`
- `path/filepath`

### 第三层：写单元测试时必须掌握

- `testing`
- `net/http/httptest`

> 说明：  
> `goroutine`、`channel`、`defer`、`struct tag`、`error 返回值` 这些不是标准库，但它们是完成本次作业的语言基础，也必须配套掌握。

---

## 一、先把题目拆成“功能模块”

根据作业截图，这次“服务健康探测器”至少包含下面几个模块：

1. 读取外部 JSON 配置文件。
2. 根据配置识别探测类型，例如 HTTP、TCP。
3. 支持命令行参数，例如：
   - `--config`
   - `--timeout`
   - `-v`
4. 并发启动多个探测任务。
5. 每个任务支持超时控制。
6. 每个任务支持可选的重试次数 `retry_count`。
7. 对 HTTP 服务做状态码检查，必要时检查响应体是否包含指定内容。
8. 对 TCP 服务做连通性检查。
9. 安全汇总所有结果，不能有数据竞争。
10. 统计成功率、失败率、响应时间分布、最慢服务。
11. 在当前目录生成一份日志或汇总文件。
12. 编写单元测试。

把题目拆完之后，就能反推出到底该学哪些标准库。

---

## 二、标准库与作业功能的对应关系

| 作业功能 | 主要标准库 |
| --- | --- |
| 读取 JSON 配置 | `os`、`encoding/json` |
| 解析命令行参数 | `flag` |
| HTTP 健康探测 | `net/http`、`context`、`time`、`io` |
| TCP 连通性探测 | `net`、`context`、`time` |
| 并发执行探测任务 | `sync` |
| 结果汇总与统计 | `sync`、`sort`、`fmt` |
| 关键词匹配 | `strings` |
| 生成日志文件 | `os`、`fmt`、`time`、可选 `path/filepath` |
| 单元测试 | `testing`、`net/http/httptest`、`net` |
| CLI 表格化输出 | `fmt`、可选 `text/tabwriter` |

---

## 三、必须掌握的标准库

## 1. `encoding/json`

### 为什么必须学

题目已经明确要求通过外部配置文件 `config.json` 定义多个探测目标，所以你一定要会：

- 把 JSON 文件读进来。
- 反序列化成 Go 结构体。
- 处理可选字段，例如 `retry_count`。
- 如果后面要把结果写成 JSON，也要知道如何序列化。

### 这次作业会用到的核心能力

- `json.Unmarshal`
- 结构体标签 `json:"field_name"`
- 可选字段
- 嵌套结构体或结构体切片

### 重点 API

| API | 作用 |
| --- | --- |
| `json.Unmarshal(data, &v)` | 把 JSON 字节反序列化到结构体中 |
| `json.Marshal(v)` | 把结构体序列化为 JSON |
| `json.MarshalIndent(v, "", "  ")` | 生成带缩进的 JSON |

### 这次作业里你大概率会写的结构

```go
type Config struct {
    Targets []Target `json:"targets"`
}

type Target struct {
    Name       string `json:"name"`
    Protocol   string `json:"protocol"`
    Address    string `json:"address"`
    RetryCount int    `json:"retry_count,omitempty"`
    ExpectCode int    `json:"expect_code,omitempty"`
    Contains   string `json:"contains,omitempty"`
}
```

### 必须理解的坑

- 结构体字段如果首字母小写，`json.Unmarshal` 不会赋值。
- `json` 标签名写错，会导致字段读不到。
- JSON 中没有 `retry_count` 时，Go 里的 `int` 会自动变成 `0`。
- 如果你要区分“没写 retry_count”和“明确写了 0”，就不能只用 `int`，而要考虑 `*int`。

### 学完标准

你应该能独立写出：

- `loadConfig(path string) (Config, error)`

---

## 2. `flag`

### 为什么必须学

题目明确要求支持命令行参数：

- `--config`
- `--timeout`
- `-v`

这就是 `flag` 标准库的典型使用场景。

### 重点 API

| API | 作用 |
| --- | --- |
| `flag.String(name, default, usage)` | 定义字符串参数 |
| `flag.Int(name, default, usage)` | 定义整型参数 |
| `flag.Bool(name, default, usage)` | 定义布尔参数 |
| `flag.Parse()` | 真正解析命令行参数 |
| `flag.Args()` | 读取剩余位置参数 |

### 这次作业里典型写法

```go
configPath := flag.String("config", "config.json", "配置文件路径")
timeoutSec := flag.Int("timeout", 3, "单次探测超时时间（秒）")
verbose := flag.Bool("v", false, "是否开启详细模式")
flag.Parse()
```

### 必须理解的坑

- 一定要先 `flag.Parse()`，再使用参数值。
- `flag.Int()` 返回的是 `*int`，不是 `int`。
- `timeout` 是秒，最终你需要转成 `time.Duration`。

### 学完标准

你应该能独立写出：

- `parseFlags() (configPath string, timeout time.Duration, verbose bool)`

---

## 3. `os`

### 为什么必须学

`os` 几乎贯穿整个作业：

- 读取配置文件。
- 创建输出日志文件。
- 可能要检查文件是否存在。
- 必要时退出程序。

### 重点 API

| API | 作用 |
| --- | --- |
| `os.ReadFile(path)` | 读取整个文件内容 |
| `os.Create(path)` | 创建文件 |
| `file.Close()` | 关闭文件 |
| `os.WriteFile(path, data, perm)` | 直接把内容写进文件 |
| `os.Stat(path)` | 查看文件状态 |
| `os.Exit(code)` | 退出程序 |

### 这次作业里最常见的用途

#### 1. 读取配置文件

```go
data, err := os.ReadFile(configPath)
```

#### 2. 生成日志文件

```go
file, err := os.Create(logFileName)
if err != nil {
    return err
}
defer file.Close()
```

### 必须理解的坑

- 创建了文件一定要 `Close()`。
- 读取的是相对路径时，要清楚当前工作目录在哪里。
- 日志文件名不要写死，题目要求里明显暗示了时间戳命名方式。

### 学完标准

你应该能独立写出：

- `loadConfigFile(path string) ([]byte, error)`
- `saveReport(path string, content string) error`

---

## 4. `time`

### 为什么必须学

这个作业里的超时、耗时、日志文件名，几乎都离不开 `time`。

### 这次作业会用到的能力

- 设置超时时间。
- 计算单次探测耗时。
- 生成带时间戳的日志文件名。
- 做重试间隔时也可能会用到。

### 重点 API

| API | 作用 |
| --- | --- |
| `time.Now()` | 获取当前时间 |
| `time.Since(start)` | 计算从开始到现在的耗时 |
| `time.Duration` | 表示时间间隔 |
| `time.Second` | 秒级时间单位 |
| `time.Millisecond` | 毫秒级时间单位 |
| `time.Now().Format(layout)` | 格式化时间字符串 |

### 题目里最关键的两个用法

#### 1. 计算响应时间

```go
start := time.Now()
// 发起探测
latency := time.Since(start)
```

#### 2. 生成日志文件名

```go
name := "monitor-log-" + time.Now().Format("20060102150405") + ".log"
```

### 必须理解的坑

- Go 的时间格式不是 `YYYY-MM-DD` 这种写法，而是固定布局：
  - `2006-01-02 15:04:05`
  - `20060102150405`
- `flag` 里传入的是整数秒，最终要转成：
  - `time.Duration(timeoutSec) * time.Second`

### 学完标准

你应该能独立写出：

- `measureLatency(func() error) time.Duration`
- `buildLogFileName() string`

---

## 5. `net/http`

### 为什么必须学

题目里的大部分目标都是 HTTP 地址，例如：

- 百度
- 网易
- GitHub
- B 站 API
- Raw 文本文件
- 慢速响应模拟

所以 `net/http` 是本次作业的核心标准库之一。

### 这次作业会用到的能力

- 创建 HTTP 客户端。
- 发起 GET 请求。
- 设置请求超时。
- 读取响应状态码。
- 读取响应体内容。
- 判断响应体是否包含指定字符串。

### 重点 API

| API | 作用 |
| --- | --- |
| `http.Client` | 自定义 HTTP 客户端 |
| `client.Do(req)` | 发送请求 |
| `http.NewRequestWithContext()` | 绑定超时控制的请求 |
| `resp.StatusCode` | HTTP 状态码 |
| `resp.Body` | 响应体 |
| `resp.Body.Close()` | 释放连接资源 |

### 推荐写法

```go
client := &http.Client{}
req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.Address, nil)
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
```

### 为什么不能只会 `http.Get`

`http.Get` 虽然简单，但这次作业里你需要：

- 自定义超时
- 绑定 `context`
- 未来可能支持不同方法
- 更方便做重试控制

所以更推荐学会：

- `http.NewRequestWithContext`
- `client.Do`

### 必须理解的坑

- `resp.Body` 用完一定要关，否则连接泄漏。
- 只判断 `err == nil` 不够，还要检查 `StatusCode`。
- 检查文本内容时要先读取 `Body`。
- 如果响应体很大，可以配合 `io.LimitReader` 限制读取量。

### 学完标准

你应该能独立写出：

- `probeHTTP(target Target, timeout time.Duration) Result`

---

## 6. `net`

### 为什么必须学

题目不仅要探测 HTTP，还要探测 TCP，例如：

- `localhost:3306`
- `localhost:22`

这类目标不是 HTTP 请求，而是 TCP 连通性检测。

### 这次作业会用到的能力

- 尝试建立 TCP 连接。
- 限制连接超时。
- 判断连接是否成功。

### 重点 API

| API | 作用 |
| --- | --- |
| `net.DialTimeout(network, address, timeout)` | 在指定超时下尝试建立网络连接 |
| `conn.Close()` | 关闭连接 |
| `net.SplitHostPort(addr)` | 拆分主机和端口 |
| `net.JoinHostPort(host, port)` | 安全拼接主机和端口 |

### 典型写法

```go
conn, err := net.DialTimeout("tcp", "localhost:3306", 3*time.Second)
if err != nil {
    return err
}
defer conn.Close()
```

### 必须理解的坑

- TCP 探测成功，不代表应用层协议一定可用，只代表“端口可连接”。
- 拨号成功后也要 `Close()`。
- 地址字符串格式不对时，`DialTimeout` 会报错。

### 学完标准

你应该能独立写出：

- `probeTCP(target Target, timeout time.Duration) Result`

---

## 7. `sync`

### 为什么必须学

题目有明确的并发要求：

- 要“瞬间”并发启动所有探测任务。
- 某个目标阻塞不能拖死整个程序。
- 要安全汇总结果。

这说明你必须掌握并发同步。

### 这次作业会用到的能力

- 用 `WaitGroup` 等待所有任务完成。
- 用 channel 或互斥锁安全收集结果。
- 保证最终报告只在全部任务结束后生成。

### 重点 API

| API | 作用 |
| --- | --- |
| `sync.WaitGroup` | 等待一组 goroutine 完成 |
| `wg.Add(n)` | 增加计数 |
| `wg.Done()` | 减少计数 |
| `wg.Wait()` | 等待所有 goroutine 完成 |
| `sync.Mutex` | 保护共享数据 |

### 这次作业里最常见的两种结果收集方式

#### 方式一：每个 goroutine 往 channel 发结果

优点：

- 结构清晰
- 不容易写出数据竞争

#### 方式二：多个 goroutine 直接 append 到共享切片

这时必须加 `Mutex`。

### 你当前已经有的基础

从 [goroutine_01/main.go](../practice/goroutine_01/main.go) 看，你已经练过：

- `goroutine`
- `sync.WaitGroup`
- `channel`
- `time.Since`

所以这一块你不是从零开始，重点是把它们真正用到业务场景里。

### 必须理解的坑

- `wg.Add(1)` 要在启动 goroutine 前调用。
- 忘了 `wg.Done()` 会导致主协程永远卡住。
- 多 goroutine 写共享切片不加锁会有数据竞争。
- 如果关闭 channel 的时机不对，会 `panic`。

### 学完标准

你应该能独立写出：

- `runAllTargets(targets []Target) []Result`

---

## 8. `io`

### 为什么必须学

题目里有一个 Raw 文本文件探测目标，并要求判断是否包含某个字符串，例如 `Contains "Go"`。  
这意味着你需要读取 HTTP 响应体内容。

### 重点 API

| API | 作用 |
| --- | --- |
| `io.ReadAll(r)` | 读取全部内容 |
| `io.LimitReader(r, n)` | 限制最多读取 n 字节 |

### 典型写法

```go
body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
if err != nil {
    return err
}
```

### 为什么推荐 `LimitReader`

如果对方响应体特别大，而你只是为了检查里面是否包含一个关键词，直接全部读取会浪费时间和内存。

### 学完标准

你应该能独立写出：

- `readBodyContains(resp *http.Response, keyword string) (bool, error)`

---

## 9. `strings`

### 为什么必须学

这个包虽然简单，但实际使用频率会很高：

- 判断协议类型
- 去除配置字符串空格
- 检查响应体是否包含目标关键词
- 规范化用户输入

### 重点 API

| API | 作用 |
| --- | --- |
| `strings.TrimSpace(s)` | 去掉首尾空白 |
| `strings.ToLower(s)` | 统一转小写 |
| `strings.Contains(s, sub)` | 判断是否包含子串 |
| `strings.HasPrefix(s, prefix)` | 判断前缀 |

### 学完标准

你应该能独立写出：

- `normalizeProtocol(s string) string`
- `matchContains(body string, keyword string) bool`

---

## 10. `fmt`

### 为什么必须学

最后的结果展示一定离不开格式化输出：

- 详细模式下实时打印探测状态
- 最终汇总报告
- 日志文件写入
- 错误信息包装

### 重点 API

| API | 作用 |
| --- | --- |
| `fmt.Printf` | 格式化输出到终端 |
| `fmt.Sprintf` | 生成格式化字符串 |
| `fmt.Fprintf` | 格式化写入文件 |
| `fmt.Errorf` | 生成带上下文的错误信息 |

### 学完标准

你应该能独立写出：

- `printVerbose(result Result)`
- `renderSummary(results []Result) string`

---

## 四、强烈建议掌握的标准库

## 11. `context`

### 为什么非常建议学

虽然 HTTP 可以直接依赖 `http.Client.Timeout`，TCP 可以依赖 `net.DialTimeout`，  
但如果你想让整个探测逻辑更统一、更工程化，`context` 会非常有帮助。

### 这次作业里的主要作用

- 为每次探测创建统一超时上下文。
- 请求超时后自动取消。
- HTTP 和 TCP 的调用风格更统一。

### 重点 API

| API | 作用 |
| --- | --- |
| `context.Background()` | 根上下文 |
| `context.WithTimeout(parent, d)` | 创建带超时的上下文 |
| `cancel()` | 释放资源 |

### 学完标准

你应该至少知道：

- `ctx, cancel := context.WithTimeout(context.Background(), timeout)`

---

## 12. `sort`

### 为什么建议学

题目要求：

- 统计成功/失败比例
- 分析响应时间
- 筛选表现最差（响应最慢）的服务

只要涉及“最慢”“排序”“Top N”，基本就要用 `sort`。

### 重点 API

| API | 作用 |
| --- | --- |
| `sort.Slice(data, lessFunc)` | 按自定义规则排序切片 |

### 典型用途

- 按 `Latency` 从大到小排序，找最慢服务
- 把结果按服务名、协议、是否成功分组前先排序

### 学完标准

你应该能独立写出：

- `findSlowest(results []Result) []Result`

---

## 13. `text/tabwriter`

### 为什么建议学

题目要求输出一份“具有工程参考价值的 CLI 汇总报告”。  
如果你只用最普通的 `fmt.Printf`，也能输出；但想让终端表格更整齐，`text/tabwriter` 很有用。

### 重点 API

| API | 作用 |
| --- | --- |
| `tabwriter.NewWriter(...)` | 创建对齐输出器 |
| `writer.Flush()` | 刷新缓冲区 |

### 典型用途

把这样的内容整齐输出成列：

- 服务名
- 协议
- 地址
- 状态
- 状态码
- 耗时
- 错误信息

### 学完标准

你至少应该知道：

- 这是一个“让 CLI 表格更好看”的标准库。

---

## 14. `errors`

### 为什么建议学

本次作业会产生很多不同错误：

- 配置文件读取失败
- JSON 解析失败
- HTTP 请求失败
- TCP 连接失败
- 状态码不符合预期
- 响应体不包含关键词

`fmt.Errorf` 已经够你完成大多数场景，但 `errors` 能帮你更清晰地处理错误语义。

### 重点 API

| API | 作用 |
| --- | --- |
| `errors.New(msg)` | 创建普通错误 |
| `errors.Is(err, target)` | 判断错误类型 |

### 学完标准

你至少应该知道：

- 什么时候直接返回错误。
- 什么时候给错误包一层上下文。

---

## 15. `path/filepath`

### 为什么建议学

如果你后面想把日志文件路径、配置文件路径处理得更稳妥一些，这个包会很好用。

### 重点 API

| API | 作用 |
| --- | --- |
| `filepath.Join(elem...)` | 按系统规则拼接路径 |
| `filepath.Base(path)` | 取文件名 |
| `filepath.Abs(path)` | 转绝对路径 |

### 学完标准

你至少应该知道：

- 路径拼接不要手写字符串加斜杠。

---

## 五、单元测试必须掌握的标准库

## 16. `testing`

### 为什么必须学

题目明确要求写单元测试。

### 这次作业里你应该测什么

- 配置文件解析是否正确。
- HTTP 探测成功/失败场景。
- TCP 探测成功/失败场景。
- 重试次数是否按预期执行。
- 汇总统计是否正确。
- 最慢服务筛选是否正确。

### 重点 API

| API | 作用 |
| --- | --- |
| `func TestXxx(t *testing.T)` | 测试函数规范 |
| `t.Fatalf(...)` | 失败即停止当前测试 |
| `t.Run(name, func(t *testing.T){})` | 子测试 |
| `t.TempDir()` | 创建临时目录 |

### 学完标准

你应该能独立写出：

- `TestLoadConfig`
- `TestProbeHTTP`
- `TestProbeTCP`
- `TestCollectStats`

---

## 17. `net/http/httptest`

### 为什么必须学

如果你直接拿真实网站做测试：

- 不稳定
- 速度慢
- 依赖外网
- 很难构造各种异常场景

`httptest` 可以在本地临时起一个 HTTP 测试服务器，让你稳定验证逻辑。

### 重点 API

| API | 作用 |
| --- | --- |
| `httptest.NewServer(handler)` | 启动本地测试 HTTP 服务 |
| `server.URL` | 获取测试地址 |
| `server.Close()` | 关闭测试服务 |

### 这次作业里能测什么

- 返回 `200 OK`
- 返回 `500`
- 返回慢响应
- 返回指定文本内容

### 学完标准

你应该能独立写出：

- 一个返回 `"Go"` 的测试服务器
- 一个故意 `Sleep` 很久的慢响应服务器

---

## 六、虽然不是标准库，但同样必须配套掌握

这部分我单独列出来，因为它们不是“包”，但本次作业离不开：

### 1. goroutine

- 并发启动所有探测任务。

### 2. channel

- 收集各 goroutine 返回的结果。

### 3. `defer`

- 关闭文件
- 关闭 HTTP Body
- 关闭 TCP 连接
- `wg.Done()`

### 4. 结构体与 tag

- 配置结构体
- 结果结构体
- JSON 标签

### 5. `error` 返回值风格

- Go 最基础的错误处理方式。

---

## 七、按作业阶段给出学习顺序

如果我想高效完成这个作业，建议按下面顺序学，而不是乱学：

### 第一步：先把“配置 + 参数”打通

先学：

- `flag`
- `os`
- `encoding/json`

完成目标：

- 程序能读取 `--config`
- 程序能读取 `--timeout`
- 程序能读取 `-v`
- 程序能把 `config.json` 解析成结构体

### 第二步：先写“单个 HTTP 探测”

再学：

- `net/http`
- `time`
- `io`
- `strings`

完成目标：

- 能探测一个 HTTP 地址
- 能拿到状态码
- 能计算响应耗时
- 能检查响应体是否包含某关键词

### 第三步：补上 TCP 探测

再学：

- `net`
- `time`

完成目标：

- 能判断 `localhost:3306` 是否能连通

### 第四步：把所有目标并发跑起来

再学：

- `sync`
- goroutine
- channel

完成目标：

- 所有目标并发探测
- 程序不会因为某个目标阻塞而卡死
- 所有结果都能安全收集

### 第五步：加入重试和超时控制

再学：

- `context`
- `time`

完成目标：

- 每个目标有超时
- 每个目标按 `retry_count` 重试

### 第六步：做汇总分析和日志输出

再学：

- `sort`
- `fmt`
- `os`
- 可选 `text/tabwriter`

完成目标：

- 统计成功数、失败数、成功率
- 找到最慢服务
- 生成日志文件
- 输出整齐的 CLI 报告

### 第七步：补测试

最后学：

- `testing`
- `net/http/httptest`
- `net`

完成目标：

- 关键逻辑都有单测

---

## 八、这次作业里最小可交付的标准库清单

如果按“最低可完成要求”来算，我至少要真正会用下面这些：

- [ ] `encoding/json`
- [ ] `flag`
- [ ] `os`
- [ ] `time`
- [ ] `fmt`
- [ ] `net/http`
- [ ] `net`
- [ ] `sync`
- [ ] `io`
- [ ] `strings`
- [ ] `testing`
- [ ] `net/http/httptest`

如果想写得更规范、更像一个完整的小工程，最好再补上：

- [ ] `context`
- [ ] `sort`
- [ ] `text/tabwriter`
- [ ] `errors`
- [ ] `path/filepath`

---

## 九、对我当前基础的判断

结合你目前的练习代码，我可以这样判断：

### 已经有基础的部分

- `goroutine`
- `channel`
- `sync.WaitGroup`
- `time.Since`
- 基础的切片和结构体使用

### 这次作业最需要补的部分

- `encoding/json`
- `flag`
- `net/http`
- `net`
- `io`
- `context`
- `testing`
- `net/http/httptest`
- `sort`

也就是说，这次作业真正的新东西，不是并发本身，而是：

- 标准库驱动的工程化实现
- 多协议探测
- 配置解析
- CLI 参数处理
- 测试与报告输出

---

## 十、最后的学习重点结论

如果我只抓最关键的学习重点，那么本次作业最应该优先突破的是：

1. `encoding/json`：因为配置读取是入口。
2. `flag`：因为命令行参数是入口。
3. `net/http`：因为 HTTP 探测是核心功能。
4. `net`：因为题目明确包含 TCP 探测。
5. `sync`：因为题目明确要求高并发执行。
6. `time`：因为超时、耗时、日志命名都依赖它。
7. `testing` + `httptest`：因为题目明确要求单元测试。

如果这 7 个部分真正掌握了，这个作业就已经具备完成条件了。

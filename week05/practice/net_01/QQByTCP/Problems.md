# QQByTCP Problems

这份文档基于当前目录下的 `server.go` 和 `client.go` 做静态检查整理，重点关注：

- 是否有会直接影响运行的错误
- 并发和资源释放是否安全
- TCP 使用方式是否合理
- 交互协议是否稳定
- 后续可以怎么优化

## 1. 当前状态概览

- `go build server.go`：可以通过
- `go build client.go`：可以通过
- `go build .`：不能通过，因为同一目录里有两个 `main` 函数

报错表现：

```text
.\server.go:16:6: main redeclared in this block
    .\client.go:11:6: other declaration of main
```

这不影响你分别用 `go run server.go` 和 `go run client.go` 学习，但会影响把这个目录当成一个完整包来构建。后续如果想更规范，可以拆成：

- `cmd/server/main.go`
- `cmd/client/main.go`

## 2. 会直接影响运行的高优先级问题

### 2.1 `ListenTCP` 失败后没有停止

位置：

- `server.go:20-23`

当前逻辑：

```go
tcplistener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: 8000})
if err != nil {
    fmt.Println("发生错误：", err)
}
```

问题：

- 如果监听失败，`tcplistener` 可能是 `nil`
- 后面继续执行 `tcplistener.Accept()`，会直接出问题

优化方向：

- 监听失败后立刻 `return`
- 同时建议 `defer tcplistener.Close()`

### 2.2 `handleConn` 有多个提前 `return`，但清理逻辑不在 `defer`

位置：

- 提前返回：`server.go:81-83`、`server.go:90-92`、`server.go:107-109`
- 清理逻辑：`server.go:113-118`

问题：

- 当前连接关闭、用户下线、释放 `connNum` 这三件事只在函数走到最后时才执行
- 但函数中间有多个 `return`
- 一旦提前返回，就会出现资源没释放的情况

可能后果：

- `conn` 没关闭
- `user2user` 里残留“假在线用户”
- `connNum` 的名额没归还，时间久了以后即使客户端都断开了，也可能出现“连接数已满”

优化方向：

- 一进入 `handleConn` 就写一个统一的 `defer`
- 在 `defer` 里做：
  - `conn.Close()`
  - 从 `user2user` 删除当前用户
  - `<-connNum`

### 2.3 客户端在注册阶段断开时，服务端可能进入死循环

位置：

- `server.go:53-59`

当前逻辑：

```go
n, err := conn.Read(message)
if err != nil {
    fmt.Println("发生错误：", err)
    continue
}
```

问题：

- 如果客户端已经断开，`Read` 很可能会一直报错
- 这里不是 `return`，而是 `continue`
- 于是这个 goroutine 可能一直循环打印错误并不断重试

可能后果：

- goroutine 泄漏
- 连接名额不释放
- 日志刷屏

优化方向：

- 对 `Read` 错误直接结束当前连接处理
- 不要在网络连接已经异常时继续读

### 2.4 同一个用户名虽然已经改成“检查+写入同锁”，但锁内还做了网络 IO

位置：

- `server.go:66-74`

当前逻辑已经比之前安全，因为：

- “检查用户名是否存在”
- “写入 `user2user`”

这两个动作都放在同一把 `Lock` 里了。

但问题还在于：

```go
userMu.Lock()
if user2user[userName] != nil {
    conn.Write(...)
    userMu.Unlock()
    continue
}
user2user[userName] = conn
userMu.Unlock()
```

其中失败分支里的 `conn.Write(...)` 发生在持锁期间。

问题：

- 网络 IO 可能阻塞
- 一旦阻塞，整个 `userMu` 会长时间不释放
- 其他 goroutine 就无法注册、删除、检查在线用户

优化方向：

- 锁里只做内存操作
- 先记下结果，解锁后再 `conn.Write`

## 3. 并发和锁相关问题

### 3.1 `PrintAllOnlineUser` 在持有 `RLock` 时直接写网络

位置：

- `server.go:41-48`

问题：

- `conn.Write` 是阻塞 IO
- 读锁虽然允许多个读者同时进入，但会阻塞写锁
- 如果这里卡住，其他 goroutine 想注册用户、删除用户都会被卡住

优化方向：

- 先在锁里把用户名列表拷贝出来
- 解锁后再拼接字符串并写给客户端

### 3.2 多个 goroutine 会同时向同一个客户端连接写数据

涉及位置：

- 当前用户自己的提示输出：`server.go:54`、`server.go:95`、`server.go:111`
- 其他用户给他发消息：`server.go:106`

问题：

- 一个用户自己的 `handleConn` goroutine 会向自己的连接写提示
- 同时，别人的 `handleConn` goroutine 也会通过 `targetUserConn.Write(...)` 向这个连接写聊天消息
- 虽然 `net.Conn` 可以被多个 goroutine 同时调用方法，但 TCP 是字节流，没有消息边界
- 多个 goroutine 并发 `Write` 时，应用层看到的内容可能前后交织，不容易区分

可能表现：

- 聊天消息和系统提示粘在一起
- 客户端屏幕上出现难以阅读的混合输出

优化方向：

- 给每个用户设计一个专门的“发送协程”
- 其他地方不要直接写 `conn`
- 统一往这个用户的消息队列里投递内容，再由单一 writer goroutine 顺序发送

### 3.3 `checkOnline` 和真正写入之间不是同一个原子动作

位置：

- `server.go:101-106`

问题：

- 先检查目标用户在线
- 然后再调用 `targetUserConn.Write(...)`
- 在这两个动作之间，对方可能已经断开

说明：

- 这不是 map 并发安全问题，而是业务上“状态可能变化”
- 这个问题很常见，不一定要完全消除，但要接受它存在

优化方向：

- 保留现在这种“写失败就处理错误”的思路即可
- 如果后续要更严谨，可以把用户对象设计得更完整，而不是只存 `net.Conn`

## 4. TCP 使用方式上的问题

### 4.1 当前协议把一次 `Read` 当成一条完整消息，存在粘包/拆包问题

位置：

- `server.go:56`
- `server.go:96`
- `server.go:134`
- `client.go:36`
- `client.go:51`

问题本质：

- TCP 是字节流，不保证“一次 `Write` 对应一次 `Read`”
- 可能出现：
  - 一条消息分两次读到（拆包）
  - 两条消息一次读到（粘包）

当前代码里的风险：

- 用户名可能被多读或少读
- 目标用户名和聊天消息可能被拼在一起
- 聊天消息可能只读到一半

优化方向：

- 最简单的方式：约定每条消息以 `\n` 结尾
- 服务端用 `bufio.Reader.ReadString('\n')` 或 `Scanner` 按行读
- 客户端发送时补上换行符，例如 `conn.Write([]byte(text + "\n"))`

### 4.2 服务端和客户端缓冲区大小不匹配

位置：

- 服务端读取缓冲区：`server.go:52`、`server.go:130`，大小 1024
- 客户端 `Scanner` 最大输入：`client.go:28`，最大 1MB

问题：

- 客户端允许输入很长
- 但服务端一次只按 1024 字节读
- 超长输入在 TCP 下会被拆成多段，当前逻辑无法正确组装

优化方向：

- 如果改成按行协议，这个问题会自然缓解
- 同时可以明确限制最大消息长度

### 4.3 服务端转发给目标用户的消息没有统一格式

位置：

- `server.go:106`

问题：

- 当前直接把原始消息写给目标用户
- 收到消息的人看不到是谁发来的
- 也没有固定前缀或换行，和提示信息容易混在一起

优化方向：

- 转发时带上发送者用户名
- 统一格式，例如：

```text
[alice] hello
```

## 5. 错误处理和资源管理问题

### 5.1 大量 `conn.Write` 没有检查错误

位置：

- `server.go:42`
- `server.go:46`
- `server.go:48`
- `server.go:55`
- `server.go:63`
- `server.go:68`
- `server.go:74`
- `server.go:82`
- `server.go:91`
- `server.go:95`
- `server.go:98`
- `server.go:103`
- `server.go:108`
- `server.go:111`
- `server.go:133`
- `server.go:140`
- `server.go:145`

问题：

- 一旦对端断开，很多写操作都会失败
- 如果错误被忽略，定位问题会很麻烦
- 某些分支可能在连接已经异常时还继续往下执行

优化方向：

- 对关键的 `Write` 至少检查错误
- 如果给当前连接写提示都失败了，通常可以直接结束这个连接处理

### 5.2 服务端没有设置读写超时

位置：

- 整个 `server.go`

问题：

- 客户端连接后如果不发数据，`Read` 会一直阻塞
- 客户端如果不读数据，`Write` 也可能阻塞

可能后果：

- goroutine 长时间挂住
- 连接名额被长期占用
- 体验上像“假死”

优化方向：

- 使用：
  - `conn.SetReadDeadline(...)`
  - `conn.SetWriteDeadline(...)`

### 5.3 客户端读服务端失败时直接 `os.Exit(0)`

位置：

- `client.go:52-56`

问题：

- `os.Exit` 会直接终止进程
- 所有 `defer` 都不会执行
- 虽然这里影响不算特别大，但退出方式比较生硬

优化方向：

- 用 channel 通知主 goroutine 退出
- 让 `createClient()` 正常 `return`

### 5.4 客户端 `Write` 也没有设置超时

位置：

- `client.go:36`

问题：

- 如果服务端长时间不读，客户端发送可能阻塞

优化方向：

- 给客户端连接也加写超时

## 6. 输入校验和交互问题

### 6.1 用户输入没有做 `TrimSpace`

位置：

- `server.go:61`
- `server.go:138`
- `client.go:30`

问题：

- 用户名如果包含前后空格，服务端会把它当成真实用户名
- 纯空格字符串不会被识别成空用户名

优化方向：

- 对输入做 `strings.TrimSpace`

### 6.2 `client.go` 没有发送换行符

位置：

- `client.go:36`

问题：

- 当前代码发送的是原始文本，不带 `\n`
- 在“当前服务端按裸字节读取”的前提下暂时能工作
- 但一旦协议切到“按行读取”，客户端必须带换行

优化方向：

- 如果后续把服务端改成行协议，这里同步改成：

```go
conn.Write([]byte(text + "\n"))
```

### 6.3 `exit` 这个单词被客户端保留为退出命令

位置：

- `client.go:31`

问题：

- 如果用户真想发送文本 `"exit"`，现在发不出去

优化方向：

- 影响不大，学习阶段可以接受
- 如果要更自然，可以用 `/exit` 这种命令风格

## 7. 结构设计上的改进方向

### 7.1 目前 `user2user` 只保存 `net.Conn`，信息太少

位置：

- `server.go:12`

问题：

- 现在用户数据只有连接对象
- 如果以后要扩展以下能力，会比较别扭：
  - 用户状态
  - 最后活跃时间
  - 消息队列
  - 统一发送协程

优化方向：

- 后续可以改成：

```go
type User struct {
    Name string
    Conn net.Conn
}
```

再进一步可以加入：

- `Send chan string`

### 7.2 全局变量较多，后续维护成本会升高

位置：

- `server.go:10-13`

问题：

- 当前项目小，使用全局变量没问题
- 但逻辑一多，`connNum`、`user2user`、`userMu` 分散在全局，函数之间耦合会比较明显

优化方向：

- 封装一个 `Server` 结构体，把共享状态放进去

### 7.3 功能上目前更像“固定目标单聊”，不是完整聊天室

当前行为：

- 用户登录
- 选择一个目标用户
- 后面一直给这个目标发消息

限制：

- 不能动态切换目标
- 不能群发
- 没有系统消息
- 没有上下线广播

优化方向：

- 可以逐步支持：
  - `/list`
  - `/to 用户名`
  - `/quit`
  - 广播消息

## 8. 推荐的优化顺序

如果按“收益最大、最适合当前学习阶段”的顺序，建议先做这几步：

1. 把 `handleConn` 的清理逻辑改成统一 `defer`
2. `ListenTCP` 失败后立刻返回
3. 所有关键 `Read` / `Write` 失败时及时结束连接
4. 不要在持锁期间做 `conn.Write`
5. 把协议改成“按行读取”
6. 给连接设置读写超时
7. 设计单独的用户发送协程，避免多 goroutine 直接写同一连接

## 9. 总结

这套代码已经具备了一个“最小可运行聊天模型”的核心结构：

- 服务端监听
- 用户注册
- 在线用户选择
- 私聊转发

它现在最大的问题不是“不会跑”，而是：

- 资源释放路径不统一
- TCP 消息边界没定义
- 锁和网络 IO 混在一起
- 多 goroutine 写同一连接会让输出变乱

如果先把这几类问题解决掉，这份练习代码会一下子稳很多，也更接近真实网络程序的写法。

// net_01 main 包
// 本文件通过代码实例 + 详细注释，讲解 Go 语言 net 标准库的 TCP 相关用法和底层原理
package main

import (
	"fmt"
	"net"
	"time"
)

// 运行方式：
// go run main.go

// =============================================================================
// net 标准库 —— 总览
// =============================================================================
//
// 【这个包是干什么的？】
// net 包是 Go 标准库中所有网络 I/O 的基础设施。它提供了：
//   - TCP/UDP/IP 网络通信（这次作业的重点）
//   - DNS 解析
//   - 网络地址解析和格式化
//   - 网络接口查询
//
// 【net 包的结构】
//
//   net 包内部大致分为几层：
//
//     用户 API 层（你直接调用的）：
//       net.Dial / net.DialTimeout  — 客户端拨号
//       net.Listen / net.ListenTCP  — 服务端监听
//       net.SplitHostPort           — 地址拆分
//       net.JoinHostPort            — 地址拼接
//       net.LookupHost              — DNS 解析
//
//     协议实现层（net 包内部）：
//       TCPConn / UDPConn / IPConn  — 各种协议连接类型
//       TCPListener / UDPConn       — 各种协议监听器
//
//     系统调用层（syscall 包）：
//       socket() / connect() / bind() / listen() / accept()
//       这些是操作系统的底层系统调用
//
// 【为什么作业里必须学它？】
// 作业要求支持 TCP 探测（不只是 HTTP），例如：
//   - localhost:3306  → MySQL 数据库端口
//   - localhost:22    → SSH 服务端口
//   - localhost:6379  → Redis 缓存端口
//
// 这些不是 HTTP 服务，不能用 http.Get 探测，只能用 TCP 拨号检测端口是否开放。
//
// 【TCP 探测和 HTTP 探测的区别】
//
//   HTTP 探测（net/http）：
//     - 应用层协议（HTTP/1.1 或 HTTP/2）
//     - 需要完整的请求/响应交互
//     - 可以检查状态码、响应体内容
//     - 失败原因更精细（404、500、超时等）
//
//   TCP 探测（net）：
//     - 传输层协议（TCP）
//     - 只检查"端口能否建立连接"
//     - 不关心应用层协议（不知道是 MySQL 还是 Redis）
//     - 失败原因较粗（连接被拒绝、超时、网络不可达）
//
//   TCP 探测成功 ≠ 服务一定正常！
//     - 端口开放只代表"有程序在监听这个端口"
//     - 不代表应用层协议能正常工作
//     - 比如 MySQL 进程还在，但内部死锁了，TCP 探测仍然成功
//
// 【net 包的命名惯例】
//
//   Dial   — 客户端主动连接（类似打电话，你拨号给别人）
//   Listen — 服务端被动等待（类似总机，等别人打进来）
//   Accept — 服务端接受一个 incoming 连接（总机接起一个来电）
//
//   这些术语来自 Berkeley Sockets API（C 语言网络编程的标准接口），
//   Go 沿用了这套命名。

// =============================================================================
// TCP 连接建立过程 —— 三次握手
// =============================================================================
//
// net.DialTimeout("tcp", addr, timeout) 内部发生了什么？
//
//   客户端                                    服务端
//     │                                        │
//     │  SYN (seq=x)                          │  监听中
//     │ ─────────────────────────────────────> │
//     │                                        │  收到 SYN → 分配资源
//     │                                        │
//     │              SYN-ACK (seq=y, ack=x+1) │
//     │ <───────────────────────────────────── │
//     │  收到 SYN-ACK                          │
//     │                                        │
//     │  ACK (ack=y+1)                         │
//     │ ─────────────────────────────────────> │  握手完成，连接建立
//     │                                        │
//     │  连接已建立，可以发送数据               │
//
// net.DialTimeout 的工作流程：
//   1. DNS 解析（如果 address 是域名而不是 IP）
//   2. 创建 socket（系统调用 socket()）
//   3. 发起 TCP 三次握手（系统调用 connect()）
//   4. 握手成功 → 返回 *net.TCPConn
//   5. 握手失败（超时、被拒绝、网络不可达） → 返回 error
//
// timeout 的作用：限制第 1-3 步的总耗时。
//   - DNS 解析太慢 → 超时
//   - 握手被防火墙丢弃（没有 SYN-ACK 返回） → 超时
//   - 端口没监听（收到 RST 重置包） → 立即返回 "connection refused"
//
// 【为什么 DialTimeout 比 client.Timeout 更底层？】
//
//   http.Client.Timeout 内部最终也是调用了 net.DialTimeout（或等价的带超时 dial）。
//   所以：
//     net.DialTimeout  → 传输层（TCP）
//     http.Client      → 应用层（HTTP，底层也用到 TCP）

// =============================================================================
// net 包的关键类型
// =============================================================================

// net.Listener — 服务端监听器接口
//
//   type Listener interface {
//       Accept() (Conn, error)    // 接受一个新连接
//       Close() error             // 关闭监听器
//       Addr() Addr               // 返回监听的地址
//   }
//
// net.Listen("tcp", "127.0.0.1:0") 创建一个 TCP 监听器。
// "0" 表示让操作系统自动分配一个可用端口（随机端口）。
// 返回的 Listener.Addr().String() 类似 "127.0.0.1:54321"。
//
// 作业中不需要自己启动 TCP 服务，只需要做客户端拨号（Dial）。
// 这里的 net.Listen 只是为了在示例中模拟一个可连接的目标。

// net.Conn — 网络连接接口
//
//   type Conn interface {
//       Read(b []byte) (n int, err error)     // 读取数据
//       Write(b []byte) (n int, err error)    // 写入数据
//       Close() error                          // 关闭连接
//       LocalAddr() Addr                       // 本地地址
//       RemoteAddr() Addr                      // 远端地址
//       SetDeadline(t time.Time) error         // 设置读写超时
//       SetReadDeadline(t time.Time) error     // 只设置读超时
//       SetWriteDeadline(t time.Time) error    // 只设置写超时
//   }
//
// net.DialTimeout 返回的就是 net.Conn（具体类型是 *net.TCPConn）。
// 作业中 TCP 探测只需要：
//   1. DialTimeout → 拿到 Conn
//   2. Conn.Close() → 关闭连接
//   不需要 Read/Write（因为只检测连通性，不发送应用层数据）

// net.Addr — 网络地址接口
//
//   type Addr interface {
//       Network() string  // 网络类型："tcp"、"udp"、"ip"
//       String() string   // 地址字符串："127.0.0.1:3306"
//   }
//
// Listener.Addr() 返回服务端监听地址。
// Conn.RemoteAddr() 返回远端地址。
// Conn.LocalAddr() 返回本地地址。

// =============================================================================
// 地址处理函数
// =============================================================================
//
// 【net.SplitHostPort】
//
//   host, port, err := net.SplitHostPort("127.0.0.1:3306")
//   // host = "127.0.0.1"
//   // port = "3306"
//
//   为什么不用 strings.Split(s, ":") 自己拆分？
//   因为 IPv6 地址是这样的："[::1]:3306"
//   strings.Split("[::1]:3306", ":") 会拆成 ["[", "", "1]", "3306"] → 完全错误！
//
//   net.SplitHostPort 正确处理所有情况：
//     "127.0.0.1:3306"  → host="127.0.0.1", port="3306"
//     "[::1]:3306"      → host="::1",     port="3306"
//     "example.com:80"  → host="example.com", port="80"
//     "bad:address"     → 返回错误
//
// 【net.JoinHostPort】
//
//   address := net.JoinHostPort("127.0.0.1", "3306")
//   // address = "127.0.0.1:3306"
//
//   对于 IPv6，它会自动加方括号：
//     net.JoinHostPort("::1", "3306") → "[::1]:3306"
//
//   为什么不用 fmt.Sprintf("%s:%s", host, port)？
//   因为 IPv6 地址需要方括号，手动拼接容易出错。
//
// 【作业中的应用】
//   config.json 里的地址可能是 "localhost:3306"，
//   你可以直接用 net.DialTimeout("tcp", "localhost:3306", timeout)。
//   但如果配置里 host 和 port 是分开的字段，就用 JoinHostPort 拼接。

// =============================================================================
// 网络错误类型判断
// =============================================================================
//
// net.DialTimeout 返回的错误可能是多种类型，作业中需要区分：
//
//   1. 连接超时
//      err.Error() 包含 "i/o timeout"
//      或 errors.Is(err, os.ErrDeadlineExceeded)
//      → 目标服务器太慢，防火墙丢弃了包
//
//   2. 连接被拒绝
//      err.Error() 包含 "connection refused"
//      → 端口上没有程序在监听
//
//   3. DNS 解析失败
//      err.Error() 包含 "no such host"
//      → 域名不存在或 DNS 服务器不可用
//
//   4. 网络不可达
//      err.Error() 包含 "network is unreachable"
//      → 没有路由到目标网络
//
// 作业中可以用 net.Error 接口判断是否是超时：
//
//   if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
//       // 超时错误
//   }
//
// 更简单的方式是直接检查错误信息：
//
//   if strings.Contains(err.Error(), "connection refused") {
//       // 端口没开
//   }

// =============================================================================
// 演示代码
// =============================================================================

func main() {
	fmt.Println("========== net 标准库演示 ==========")

	// ------------------------------------------------------------------------
	// 步骤 1：用 net.Listen 启动一个本地 TCP 监听器（模拟服务端）
	// ------------------------------------------------------------------------
	//
	// net.Listen 的参数：
	//   "tcp"              — 网络类型，表示 TCP 协议
	//   "127.0.0.1:0"      — 监听地址
	//     - 127.0.0.1      — 只监听本地回环（其他机器无法连接）
	//     - :0             — 端口设为 0，让操作系统自动分配一个空闲端口
	//
	// 为什么用 "127.0.0.1:0" 而不是具体端口？
	//   - 避免端口冲突（如果 3306 被 MySQL 占用了会报错）
	//   - 自动分配的端口每次运行都不同，适合测试
	//
	// 返回的 listener 实现了 net.Listener 接口：
	//   - Accept()  → 等待并接受一个 incoming 连接
	//   - Close()   → 停止监听，释放端口
	//   - Addr()    → 返回实际监听的地址（如 "127.0.0.1:54321"）
	//
	// 作业中你只需要做客户端拨号（Dial），不需要启动服务端（Listen）。
	// 这里的 Listen 只是为了创建一个可连接的目标，方便演示。

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("启动本地 TCP 服务失败：", err)
		return
	}
	defer listener.Close()

	// ------------------------------------------------------------------------
	// 步骤 2：在后台 goroutine 中接受连接（模拟服务端响应）
	// ------------------------------------------------------------------------
	//
	// listener.Accept() 是阻塞调用——它会一直等待，直到有客户端连接进来。
	// 所以我们把它放在 goroutine 里，避免阻塞 main 函数。
	//
	// 这个 goroutine 做的事情很简单：
	//   1. Accept() 等待客户端连接
	//   2. 连接建立后，立即 Close() 关闭它
	//   3. 不发送任何数据（TCP 探测只需要建立连接，不需要数据交换）
	//
	// 注意：这里用了 blank identifier _ 忽略 conn.Close() 的返回值，
	// 因为关闭一个已经建立的 TCP 连接几乎不会失败。
	//
	// 作业中不需要写这个——这是为了模拟一个"活着的"TCP 服务。

	go func() {
		// Accept() 阻塞等待，直到有客户端调用 Dial 连接进来。
		// 三次握手完成后，Accept() 返回一个新的 Conn 对象。
		conn, err := listener.Accept()
		if err == nil {
			// 连接成功建立，立即关闭。
			// TCP 探测的场景中，服务端不需要发送数据，
			// 客户端只要能连上就说明端口是开放的。
			_ = conn.Close()
		}
	}()

	// ------------------------------------------------------------------------
	// 步骤 3：从监听器地址中拆分 host 和 port
	// ------------------------------------------------------------------------
	//
	// listener.Addr() 返回 net.Addr 接口，String() 方法返回 "127.0.0.1:随机端口"。
	// 例如："127.0.0.1:54321"
	//
	// net.SplitHostPort 把这个字符串拆成 host 和 port 两部分。
	// 这是 net 包提供的安全拆分方式，正确处理 IPv4、IPv6、域名等各种格式。

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		fmt.Println("拆分 host:port 失败：", err)
		return
	}
	fmt.Println("host：", host) // 127.0.0.1
	fmt.Println("port：", port) // 随机端口号，如 54321

	// ------------------------------------------------------------------------
	// 步骤 4：重新拼接地址
	// ------------------------------------------------------------------------
	//
	// net.JoinHostPort 把 host 和 port 重新拼成 "host:port" 格式。
	// 对于 IPv6 地址会自动加方括号：JoinHostPort("::1", "80") → "[::1]:80"
	//
	// 作业中如果你从配置文件分别读到 host 和 port，就用这个函数拼接。
	// 如果配置文件里已经是 "localhost:3306" 这种完整地址，可以直接用，不需要拼接。

	address := net.JoinHostPort(host, port)
	fmt.Println("重新拼接后的地址：", address)

	// ------------------------------------------------------------------------
	// 步骤 5：TCP 拨号（核心操作）
	// ------------------------------------------------------------------------
	//
	// net.DialTimeout 的参数：
	//   "tcp"               — 网络类型，可以是 "tcp"、"tcp4"（仅 IPv4）、"tcp6"（仅 IPv6）
	//   address             — 目标地址，如 "127.0.0.1:54321" 或 "google.com:443"
	//   2*time.Second       — 超时时间，包含 DNS 解析 + TCP 握手的总耗时
	//
	// 内部流程：
	//   1. 解析 address（如果是域名则做 DNS 查询）
	//   2. 创建 socket 文件描述符（syscall.Socket）
	//   3. 设置 socket 为非阻塞模式（syscall.SetsockoptInt）
	//   4. 发起 connect() 系统调用（发送 SYN 包）
	//   5. 用 select/poll/epoll 等待连接完成或超时
	//   6. 握手成功 → 返回 *net.TCPConn
	//   7. 超时/失败 → 返回 error
	//
	// 可能的错误：
	//   - "dial tcp 127.0.0.1:xxx: connect: connection refused"
	//     → 目标端口没有程序监听（TCP 收到了 RST 重置包）
	//
	//   - "dial tcp 127.0.0.1:xxx: i/o timeout"
	//     → 超时（防火墙丢弃了 SYN 包，没有 SYN-ACK 返回）
	//
	//   - "dial tcp: lookup xxx: no such host"
	//     → DNS 解析失败（域名不存在）
	//
	//   - "dial tcp: connect: network is unreachable"
	//     → 没有路由到目标网络
	//
	// 作业中的典型用法：
	//   conn, err := net.DialTimeout("tcp", target.Address, timeout)
	//   if err != nil {
	//       result.Error = fmt.Sprintf("TCP 连接失败: %v", err)
	//       return result
	//   }
	//   defer conn.Close()
	//   result.OK = true
	//   return result

	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		fmt.Println("DialTimeout 失败：", err)
		return
	}
	// 走到这里说明 TCP 三次握手成功，端口是开放的。

	// ------------------------------------------------------------------------
	// 步骤 6：关闭连接
	// ------------------------------------------------------------------------
	//
	// conn.Close() 发送 TCP FIN 包，优雅地关闭连接。
	// 不 Close 会导致：
	//   - 文件描述符泄漏（每个连接占用一个 fd）
	//   - 连接停留在 TIME_WAIT 状态，占用系统资源
	//
	// defer 确保即使后面代码 panic 也会执行 Close。这个close的意义是：
	defer conn.Close()

	fmt.Println("连接成功，耗时：", time.Since(start))

	// ------------------------------------------------------------------------
	// 知识点总结
	// ------------------------------------------------------------------------
	//
	// 这些是 TCP 探测中最容易踩的坑。

	fmt.Println("\n常见知识点：")
	fmt.Println("1. net.DialTimeout 是 TCP 健康探测里最常用的 API。")
	fmt.Println("   它内部做了 DNS 解析 + TCP 三次握手，超时则返回错误。")

	fmt.Println("2. TCP 可连接只代表端口打开，不代表应用协议就一定正常。")
	fmt.Println("   例如：MySQL 端口通了，但 MySQL 可能已经死锁无法处理查询。")
	fmt.Println("   TCP 探测 = 传输层检查，HTTP 探测 = 应用层检查。")

	fmt.Println("3. Dial 成功后要记得 Close，否则文件描述符泄漏。")
	fmt.Println("   用 defer conn.Close() 确保连接一定被关闭。")

	fmt.Println("4. SplitHostPort / JoinHostPort 可以安全处理 IPv4/IPv6/域名。")
	fmt.Println("   不要用 strings.Split(addr, \":\") 自己拆分，IPv6 会出错。")

	// ------------------------------------------------------------------------
	// 额外演示：网络错误类型判断
	// ------------------------------------------------------------------------
	fmt.Println("\n========== 错误类型判断演示 ==========")

	// 模拟连接到一个不存在的端口
	_, err = net.DialTimeout("tcp", "127.0.0.1:59999", 1*time.Second)
	if err != nil {
		// 输出类似：dial tcp 127.0.0.1:59999: connect: connection refused
		fmt.Println("连接不存在端口的错误：", err)

		// 判断错误类型的方式 1：用 net.Error 接口
		if netErr, ok := err.(net.Error); ok {
			fmt.Printf("  是网络错误吗？%v\n", true)
			fmt.Printf("  是超时错误吗？%v\n", netErr.Timeout())
			// Timeout() 返回 true 表示超时，false 表示其他网络错误
		}

		// 判断错误类型的方式 2：检查错误信息（简单粗暴但有效）
		errStr := err.Error()
		if contains(errStr, "connection refused") {
			fmt.Println("  → 连接被拒绝：端口上没有程序在监听")
		} else if contains(errStr, "i/o timeout") {
			fmt.Println("  → 连接超时：服务器没响应或防火墙丢弃了包")
		} else if contains(errStr, "no such host") {
			fmt.Println("  → DNS 解析失败：域名不存在")
		}
	}

	// ------------------------------------------------------------------------
	// 额外演示：DNS 解析
	// ------------------------------------------------------------------------
	fmt.Println("\n========== DNS 解析演示 ==========")

	// net.LookupHost 查询域名对应的 IP 地址。
	// 返回一个 IP 地址列表（一个域名可能有多个 IP）。
	ips, err := net.LookupHost("localhost")
	if err != nil {
		fmt.Println("DNS 查询失败：", err)
	} else {
		fmt.Printf("localhost 的 IP 地址：%v\n", ips)
		// 输出类似：[::1 127.0.0.1]（IPv6 + IPv4）
	}

	// net.LookupPort 查询服务名对应的端口号。
	// 使用 /etc/services 文件中的服务名映射。
	tcpPort, err := net.LookupPort("tcp", "http")
	if err != nil {
		fmt.Println("端口查询失败：", err)
	} else {
		fmt.Printf("http 服务的 TCP 端口：%d\n", tcpPort) // 80
	}
}

// contains 简单的字符串包含检查（避免在 main 里 import strings）。
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// 作业中 probeTCP 的完整写法参考
// =============================================================================
//
// 下面展示作业中 TCP 探测函数应该怎么写。
// 这不是可运行代码，而是结构参考。
//
// func probeTCP(target Target, timeout time.Duration) ProbeResult {
//     result := ProbeResult{Name: target.Name}
//     start := time.Now()
//
//     // 1. 发起 TCP 拨号
//     conn, err := net.DialTimeout("tcp", target.Address, timeout)
//     result.Latency = time.Since(start).String()
//
//     if err != nil {
//         // 连接失败，用 %w 包装错误以便上层用 errors.Is 判断
//         result.Error = fmt.Errorf("TCP 连接失败: %w", err).Error()
//         return result
//     }
//
//     // 2. 连接成功，立即关闭（TCP 探测只需要检查连通性）
//     conn.Close()
//
//     // 3. 标记成功
//     result.OK = true
//     return result
// }
//
// 注意：
//   - 如果 target.Address 是 "localhost:3306" 这种完整地址，直接传即可
//   - 如果 host 和 port 是分开的，用 net.JoinHostPort(host, port) 拼接
//   - 不需要 Read/Write，只要 DialTimeout 不返回 error 就说明端口开放
//   - 必须 Close() 连接，否则文件描述符泄漏

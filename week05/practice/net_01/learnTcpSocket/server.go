package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	defer listener.Close()
	fmt.Printf("服务端正在监听8000端口----------\n")
	for {
		conn, err := listener.Accept() //服务端代码在这里阻塞，不往下执行，只有当客户端连接到服务端时，才会往下执行 ，返回conn
		if err != nil {
			fmt.Print("服务端连接客户端失败", err.Error())
			return
		}
		go handlerRequst(conn) //启动一个goroutine，处理这个客户端的请求
	}
}

// 没有解决的问题示例
func handlerRequst(conn net.Conn) {
	/*
		这个代码还有很多问题，比如：
		1、由于服务端没有设置读超时时间，连接过期时间，当有恶意的请求大量的连接时，服务端就会存在很多个goroutine
		导致服务端的性能下降
		2、由于服务端没有设置写超时时间，当客户端没有读时，服务端的写缓冲区可能被占满，所有的goroutine都会被阻塞在Write操作上
		会导致服务端的内存泄露
		3、由于服务端没有设置最大的客户端连接数，当有恶意的请求大量的连接时，服务端就会存在很多个goroutine，导致服务端内存泄露，文件描述符耗尽
		因此在实际应用中，服务端需要设置最大的客户端连接数，来防止服务端内存泄露，文件描述符耗尽。
	*/

	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer) //这个goroutine在这里阻塞，不往下执行，只有当客户端发送消息时，才会往下执行 ，返回n,这里的n是读取到的字节数量
	if err != nil {
		fmt.Println("Error Reading:", err.Error())
		return
	}
	fmt.Printf("服务端接受到:%s\n", string(buffer[:n])) //这里的n是读取到的字节数量，所以要使用buffer[:n]来获取读取到的字符串
	conn.Write([]byte("服务端已经接受到客户端消息"))
}

// ============================================================================
// 【正例】安全的生产级服务端（新增）
// 改进点：限制最大并发连接数、读写超时控制、支持循环读取、优雅释放资源
// 使用方法：在 main 函数中把 net.Listen 那部分注释掉，改为 runSecureServer() 即可
// ============================================================================

const (
	maxConnections = 100         // 最大并发连接数
	readTimeout    = 10 * time.Second // 读超时：客户端 10 秒不发数据则断开
	writeTimeout   = 5 * time.Second  // 写超时：客户端 5 秒不收数据则断开
)

// connSem 信号量：用带缓冲 channel 控制并发连接数
var connSem = make(chan struct{}, maxConnections)

// runSecureServer 启动安全服务端（替换原 main 中的监听逻辑）
func runSecureServer() {
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		fmt.Println("监听失败:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("[正例] 安全服务端启动 | 最大连接数: %d | 读超时: %v\n", maxConnections, readTimeout)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("接收连接失败:", err)
			continue
		}

		// 尝试获取信号量
		select {
		case connSem <- struct{}{}:
			// 有空闲名额，启动 goroutine 处理
			go handleSecureConn(conn)
		default:
			// 名额已满，拒绝连接
			fmt.Println("连接数已满，拒绝新连接")
			conn.Close()
		}
	}
}

// handleSecureConn 处理单个安全连接
func handleSecureConn(conn net.Conn) {
	// 确保连接关闭后释放信号量
	defer func() {
		conn.Close()
		<-connSem
	}()

	// 设置读写超时（Deadline 是绝对时间点）
	conn.SetReadDeadline(time.Now().Add(readTimeout))
	conn.SetWriteDeadline(time.Now().Add(writeTimeout))

	buffer := make([]byte, 1024)
	for {
		// 每次循环刷新读超时，防止连接永久挂起
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		n, err := conn.Read(buffer)
		if err != nil {
			// 区分超时错误和普通错误
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("客户端读超时，断开连接")
			} else {
				fmt.Println("读取错误或客户端断开:", err)
			}
			return
		}

		// 处理数据
		msg := string(buffer[:n])
		fmt.Printf("收到数据: %s\n", msg)

		// 发送响应前刷新写超时
		conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		_, err = conn.Write([]byte("服务端已确认: " + msg))
		if err != nil {
			fmt.Println("发送数据失败:", err)
			return
		}
	}
}

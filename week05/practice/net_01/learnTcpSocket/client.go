package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	lockByRead()
	// conn, err := net.Dial("tcp", "localhost:8000")
	// if err != nil {
	// 	fmt.Print("出现错误", err.Error())
	// 	return
	// }
	// defer conn.Close()
	// conn.Write([]byte("客户端消息")) //客户端发送消息给服务端
	// buffer := make([]byte, 1024)
	// n, err := conn.Read(buffer) //客户端在这里阻塞，不往下执行，只有当服务端发送消息时，才会往下执行 ，返回n,这里的n是读取到的字节数量，所以这里的n是读取到的字节数量
	// if err != nil {
	// 	fmt.Println("Error Reading:", err.Error())
	// 	return
	// }
	// fmt.Printf("客户端接受到:%s\n", string(buffer[:n])) //这里的n是读取到的字节数量，所以要使用buffer[:n]来获取读取到的字符串

}
func lockByRead() {
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	defer conn.Close()
	// 这里模拟一个读操作，模拟死锁情况
	b := make([]byte, 1024)
	n, err := conn.Read(b)
	if err != nil {
		fmt.Println("Error Reading:", err.Error())
		return
	}
	remoteAddr := conn.RemoteAddr().String()
	fmt.Printf("服务端地址是: %s\n", remoteAddr)
	fmt.Printf("我是客户端，我的地址是: %s\n", conn.LocalAddr())
	err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		fmt.Println("Error Setting Read Deadline:", err.Error())
		return
	}
	fmt.Printf("客户端接受到:%s\n", string(b[:n])) //这里的n是读取到的字节数量，所以要使用b[:n]来获取读取到的字符串
}

/*
值得注意的是，Read操作是阻塞的，而Write操作是非阻塞的，Write操作会立即返回执行，那么如果服务端也先Read，客户端也先Read，
那么客户端会阻塞在Read操作上，服务端的goroutine也会阻塞在Read操作上,造成死锁，两个goroutine会一直阻塞，直到有一个goroutine被取消
另外，TCP通信是全双工的，客户端可以同时发送消息给服务端，服务端也可以同时发送消息给客户端，发的消息都存在对方的缓冲区中，也就是双方都可以同时通信，互不干扰，
两者都有读缓冲区和写缓冲区，所以客户端和服务端都可以同时发送消息给对方，互不干扰，
*/

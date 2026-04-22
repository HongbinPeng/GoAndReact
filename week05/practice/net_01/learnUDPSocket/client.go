package main

import (
	"fmt"
	"net"
)

func main() {
	createClient()
}

func createClient() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:8000") //这里只是创建地址的结构体，并不是真正的使用
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	fmt.Printf("UDP地址: %v\n", addr.AddrPort())
	conn, err := net.DialUDP("udp", nil, addr)
	/*这里的nil是客户端的默认地址，addr是服务端的地址
		这里返回的是UDPConn结构体的指针，这里面具有UDP协议特有的接受方法如ReadFromUDP和WriteToUDP
		同时后面就不用每次都使用UDP的特有的conn.ReadFromUDP和conn.WriteToUDP方法了，因为这里默认绑定了服务端地址
		是addr
	同时如果中间的参数为nil，就代表由系统默认的分配客户端的地址端口，随机的分配一个端口
	*/
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	defer conn.Close()
	// for {
	// 	n, err := conn.Write([]byte("你好啊，服务端"))
	// 	if err != nil {
	// 		fmt.Print("出现错误", err.Error())
	// 		return
	// 	}
	// 	fmt.Printf("客户端发送消息:%s,发送的字节数为%d\n", string([]byte("你好啊，服务端")), n)
	// }
	/*
		这里模拟客户端无限的发送消息给服务端，会导致服务端的不停的读，当这样的而客户端多了以后
		会导致服务端的接受区满了，造成丢包
	*/
	n, err := conn.Write([]byte("你好啊，服务端"))
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	fmt.Println("客户端发送消息:", string([]byte("你好啊，服务端")))
	// 读取服务端回复
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	fmt.Println("服务端回复消息:", string(buf[:n]))
}

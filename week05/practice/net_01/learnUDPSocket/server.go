package main

import (
	"fmt"
	"net"
)

func main() {
	// createServer()
}
func createServer() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:8000") //这里实际上就仅仅是解析了地址，没有绑定地址
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	fmt.Printf("UDP地址: %v\n", addr.AddrPort())
	conn, err := net.ListenUDP("udp", addr) //这里进行了监听，返回了一个UDPConn结构体的指针，这里面具有UDP协议特有的接受方法如ReadFromUDP和WriteToUDP
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024)
	if err != nil {
		fmt.Print("出现错误", err.Error())
		return
	}
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Print("出现错误", err.Error())
			return
		}
		fmt.Println("这个客户端地址为:", clientAddr)
		fmt.Println("服务端收到消息:", string(buf[:n]))
		n, err = conn.WriteToUDP([]byte("你好啊，"+clientAddr.String()), clientAddr)
		if err != nil {
			fmt.Print("出现错误", err.Error())
			return
		}
		fmt.Println("服务端回复消息:", string([]byte("你好啊，"+clientAddr.String())))
	}
}
func creatSafeServer() {

}

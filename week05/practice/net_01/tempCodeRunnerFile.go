package main

import (
	"fmt"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("启动本地 TCP 服务失败：", err)
		return
	}
	defer listener.Close()
}

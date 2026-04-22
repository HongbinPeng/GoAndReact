package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	createClient()
}

func createClient() {
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println("连接服务端失败：", err)
		return
	}
	defer conn.Close()

	fmt.Println("已连接到服务端，按照提示输入用户名、目标用户和消息即可。输入 exit 退出。")

	go readFromServer(conn)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "exit" {
			fmt.Println("客户端退出")
			return
		}

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("发送消息失败：", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("读取输入失败：", err)
	}
}

func readFromServer(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("读取服务端消息失败：", err)
			}
			os.Exit(0)
		}

		fmt.Print(string(buf[:n]))
	}
}

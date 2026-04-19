package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// 运行方式：
// go run main.go
//
// 在监控器作业里，io 主要负责：
// 1. 读取 HTTP 响应体
// 2. 限制最多读取多少内容
// 3. 把一个 Reader 的内容复制到另一个 Writer

func main() {
	fmt.Println("========== io 标准库演示 ==========")
	demonstrateReadAll()
	demonstrateLimitReader()
	demonstrateCopy()
}

func demonstrateReadAll() {
	fmt.Println("\n--- 1. io.ReadAll ---")
	reader := strings.NewReader("Go monitor is healthy")
	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Println("ReadAll 失败：", err)
		return
	}
	fmt.Println("读取结果：", string(data))
}

func demonstrateLimitReader() {
	fmt.Println("\n--- 2. io.LimitReader ---")
	reader := strings.NewReader("这是一段很长的内容，用于演示只读取前几个字节。")
	limited := io.LimitReader(reader, 12)
	data, err := io.ReadAll(limited)
	if err != nil {
		fmt.Println("LimitReader + ReadAll 失败：", err)
		return
	}
	fmt.Println("只读取前 12 个字节后的内容：", string(data))
}

func demonstrateCopy() {
	fmt.Println("\n--- 3. io.Copy ---")
	reader := strings.NewReader("把这段内容复制进缓冲区。")
	var buffer bytes.Buffer

	written, err := io.Copy(&buffer, reader)
	if err != nil {
		fmt.Println("io.Copy 失败：", err)
		return
	}

	fmt.Printf("复制了 %d 个字节\n", written)
	fmt.Println("缓冲区内容：", buffer.String())

	fmt.Println("\n常见知识点：")
	fmt.Println("1. io.ReadAll 最常用于一次性读取响应体。")
	fmt.Println("2. io.LimitReader 可以避免把超大响应全部读进内存。")
	fmt.Println("3. io.Copy 常用于在 Reader 和 Writer 之间搬运数据。")
}

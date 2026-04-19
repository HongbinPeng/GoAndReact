package main

import (
	"fmt"
	"strings"
)

// 运行方式：
// go run main.go
//
// 在监控器作业里，strings 主要负责：
// 1. 规范化 protocol 字段
// 2. 判断响应体是否包含指定关键字
// 3. 清理配置文件里的空白字符

func main() {
	fmt.Println("========== strings 标准库演示 ==========")

	rawProtocol := "  HTTP  "
	normalized := strings.ToLower(strings.TrimSpace(rawProtocol))
	fmt.Printf("原始协议：%q -> 规范化后：%q\n", rawProtocol, normalized)

	body := "Go monitor is healthy"
	fmt.Println("响应体是否包含 Go：", strings.Contains(body, "Go"))

	url := "https://api.bilibili.com/x/web-interface/nav"
	fmt.Println("是否以 https:// 开头：", strings.HasPrefix(url, "https://"))

	parts := strings.Split("localhost:3306", ":")
	fmt.Println("Split 结果：", parts)

	joined := strings.Join([]string{"monitor", "log", "20260419"}, "-")
	fmt.Println("Join 结果：", joined)

	fmt.Println("\n常见知识点：")
	fmt.Println("1. TrimSpace 很适合清理配置字符串。")
	fmt.Println("2. ToLower 常用于把 protocol 统一成 http / tcp。")
	fmt.Println("3. Contains 常用于“文本里是否包含某关键词”的判断。")
	fmt.Println("4. HasPrefix 适合判断 URL 是否以 http 开头。")
}

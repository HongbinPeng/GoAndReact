package main

import (
	"fmt"
	"path/filepath"
)

// 运行方式：
// go run main.go
//
// filepath 主要用于跨平台地处理文件路径。
// 在监控器作业里，它很适合处理：
// 1. 配置文件路径
// 2. 日志文件输出路径

func main() {
	fmt.Println("========== path/filepath 标准库演示 ==========")

	logPath := filepath.Join("week05", "homework", "monitor", "monitor-log-20260419120000.log")
	fmt.Println("Join 结果：", logPath)

	fmt.Println("Base：", filepath.Base(logPath))
	fmt.Println("Dir：", filepath.Dir(logPath))
	fmt.Println("Ext：", filepath.Ext(logPath))

	absPath, err := filepath.Abs(logPath)
	if err != nil {
		fmt.Println("Abs 失败：", err)
		return
	}
	fmt.Println("绝对路径：", absPath)

	fmt.Println("\n常见知识点：")
	fmt.Println("1. filepath.Join 会自动按当前系统规则拼接路径。")
	fmt.Println("2. 不要手写字符串拼接路径，尤其不要硬编码斜杠。")
	fmt.Println("3. Base / Dir / Ext 都是处理文件名时很常用的工具。")
}

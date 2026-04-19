package main

import (
	"fmt"
	"time"
)

// 运行方式：
// go run main.go
//
// 这份示例用于练习 time 标准库。
// 在监控器作业里，time 主要用于：
// 1. 控制超时时间
// 2. 计算请求耗时
// 3. 生成日志文件名时间戳

func main() {
	fmt.Println("========== time 标准库演示 ==========")

	now := time.Now() //这个返回类型是time.Time，包含当前时间的所有信息
	fmt.Println("当前时间：", now)
	fmt.Println("格式化输出：", now.Format("2006-01-02 15:04:05"))
	fmt.Println("日志文件时间戳格式：", now.Format("20060102150405"))

	fmt.Println("\n演示耗时统计：")
	start := time.Now()
	time.Sleep(120 * time.Millisecond)
	elapsed := time.Since(start)
	fmt.Println("模拟任务耗时：", elapsed)

	fmt.Println("\n演示 Duration：")
	timeout := 3 * time.Second
	retryInterval := 500 * time.Millisecond
	fmt.Println("超时时间：", timeout)
	fmt.Println("重试间隔：", retryInterval)

	fmt.Println("\n常见知识点：")
	fmt.Println("1. time.Now() 用于记录开始时刻。")
	fmt.Println("2. time.Since(start) 常用于统计响应时间。")
	fmt.Println("3. time.Duration 是时间间隔类型。")
	fmt.Println("4. Go 的时间格式必须用固定模板 2006-01-02 15:04:05。")
	fmt.Println("5. 在命令行读取 timeout=3 之后，通常要转成 3 * time.Second。")
}

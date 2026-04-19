package main

import (
	"fmt"
	"os"
	"strings"
)

// 运行方式：
// go run main.go
//
// fmt 在作业里主要负责：
// 1. 终端打印实时探测状态
// 2. 终端打印汇总报告
// 3. 生成格式化字符串
// 4. 向日志文件写入内容
// 5. 包装错误信息

func main() {
	fmt.Println("========== fmt 标准库演示 ==========")

	serviceName := "GitHub"
	statusCode := 200
	latency := "86ms"

	fmt.Printf("Printf：服务=%s 状态码=%d 耗时=%s\n", serviceName, statusCode, latency)

	line := fmt.Sprintf("Sprintf：服务=%s 状态码=%d 耗时=%s", serviceName, statusCode, latency)
	fmt.Println(line)

	var builder strings.Builder
	if _, err := fmt.Fprintf(&builder, "Fprintf 写入 strings.Builder：服务=%s 状态码=%d", serviceName, statusCode); err != nil {
		fmt.Println("Fprintf 到 Builder 失败：", err)
		return
	}
	fmt.Println(builder.String())

	file, err := os.Create("fmt_demo.log")
	if err != nil {
		fmt.Println("创建日志文件失败：", err)
		return
	}
	defer os.Remove("fmt_demo.log")
	defer file.Close()

	if _, err := fmt.Fprintf(file, "日志示例：服务=%s 状态码=%d\n", serviceName, statusCode); err != nil {
		fmt.Println("写日志文件失败：", err)
		return
	}

	wrappedErr := fmt.Errorf("探测服务 %s 失败：%w", "not-exist-domain-123.com", os.ErrNotExist)
	fmt.Println("包装错误：", wrappedErr)

	fmt.Println("\n常见知识点：")
	fmt.Println("1. Printf 直接输出到终端。")
	fmt.Println("2. Sprintf 返回格式化后的字符串。")
	fmt.Println("3. Fprintf 可以把内容写到文件、缓冲区、网络连接。")
	fmt.Println("4. Errorf 可以顺手拼出一条带上下文的错误信息。")
	fmt.Println("5. 汇总报告和 verbose 日志基本都离不开 fmt。")
}

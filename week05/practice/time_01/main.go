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

	now := time.Now() //这个返回类型是time.Time结构体，但是由于now重写了String方法，所以可以直接打印
	// 格式说明
	/*
			2026-04-21	年-月-日
		23:44:54.1062314	时:分:秒.纳秒
		+0800	时区偏移量（东八区）
		CST	时区名称（中国标准时间）
		m=+0.000000001	单调时钟信息（自程序启动后的纳秒数）
	*/
	fmt.Println("当前时间：", now)                               //包含当前时间的所有信息如：2026-04-21 23:44:54.1062314 +0800 CST m=+0.000000001
	fmt.Println("格式化输出：", now.Format(time.DateTime))        //这个时间戳不能随便写
	fmt.Println("日志文件时间戳格式：", now.Format("20060102150405")) //这个时间戳不能随便写

	fmt.Println("\n演示耗时统计：")
	start := time.Now()
	time.Sleep(120 * time.Millisecond)
	elapsed := time.Since(start) //这里返回的是time.Duration类型,时间的单位是纳秒
	/*

		特性	time.Now()					time.Duration
		类型	函数，返回 time.Time 类型	类型（int64 的别名），表示时间段
		含义	获取当前时间点（时刻）		表示两个时间点之间的时间间隔
		单位	无（表示具体时刻）			纳秒（底层存储为纳秒数）
		示例	2026-04-21 23:44:54 +0800 CST	5s（5秒）、100ms（100毫秒）
	*/
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

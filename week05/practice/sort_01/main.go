package main

import (
	"fmt"
	"sort"
	"time"
)

// 运行方式：
// go run main.go
//
// 在监控器作业里，sort 主要用于：
// 1. 找出响应最慢的服务
// 2. 按耗时排序输出结果

type ProbeResult struct {
	Name    string
	Latency time.Duration
}

func main() {
	fmt.Println("========== sort 标准库演示 ==========")

	results := []ProbeResult{
		{Name: "百度", Latency: 80 * time.Millisecond},
		{Name: "GitHub", Latency: 120 * time.Millisecond},
		{Name: "B站API", Latency: 60 * time.Millisecond},
		{Name: "慢速响应模拟", Latency: 5 * time.Second},
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Latency > results[j].Latency
	})

	fmt.Println("按耗时从大到小排序后的结果：")
	for _, result := range results {
		fmt.Printf("服务=%s 耗时=%v\n", result.Name, result.Latency)
	}

	fmt.Printf("\n最慢服务：%s\n", results[0].Name)
}

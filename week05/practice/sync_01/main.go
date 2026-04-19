package main

import (
	"fmt"
	"sync"
	"time"
)

// 运行方式：
// go run main.go
//
// 这份示例专门练习 sync 包中最贴合作业的两块：
// 1. sync.WaitGroup：等待所有探测任务结束
// 2. sync.Mutex：保护共享结果切片或共享计数器

type ProbeResult struct {
	Name string
	Cost time.Duration
}

func main() {
	fmt.Println("========== sync 标准库演示 ==========")
	demonstrateWaitGroupAndMutex()
}

func demonstrateWaitGroupAndMutex() {
	targets := []string{"百度", "GitHub", "B站API", "Raw文本"}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []ProbeResult
	)

	for _, target := range targets {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			start := time.Now()
			time.Sleep(50 * time.Millisecond)
			result := ProbeResult{
				Name: name,
				Cost: time.Since(start),
			}

			// 多个 goroutine 同时 append 同一个切片时，必须加锁。
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(target)
	}

	// Wait 会阻塞，直到所有 goroutine 都调用了 Done。
	wg.Wait()

	fmt.Println("所有探测任务已经完成，结果如下：")
	for _, result := range results {
		fmt.Printf("服务=%s 耗时=%v\n", result.Name, result.Cost)
	}

	fmt.Println("\n常见知识点：")
	fmt.Println("1. WaitGroup 负责“等大家都做完”。")
	fmt.Println("2. Mutex 负责“同一时刻只允许一个 goroutine 修改共享数据”。")
	fmt.Println("3. wg.Add(1) 一定要写在启动 goroutine 之前。")
	fmt.Println("4. goroutine 里常用 defer wg.Done() 保证不会漏写。")
	fmt.Println("5. 作业里如果你用 channel 汇总结果，也可以少用一部分锁。")
}

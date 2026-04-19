package main

import (
	"fmt"
	"sync"
	"time"
)

func doSquare(origin []int, start int, end int, result chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	var ans int
	for i := start; i < end; i++ {
		ans += origin[i] * origin[i]
		time.Sleep(time.Millisecond * 5)
	}
	result <- ans
}
func task1() {
	fmt.Println("========== 任务1：10000个goroutine ==========")
	start := time.Now()
	num := make([]int, 10000)
	for i := 0; i < 10000; i++ {
		num[i] = i
	}
	result := make(chan int, 10000)
	var wg sync.WaitGroup
	for j := 0; j < 10000; j++ {
		wg.Add(1)
		go doSquare(num, j, j+1, result, &wg) // 每个goroutine只算1个元素
	}
	var total int
	for i := 0; i < 10000; i++ {
		total += <-result
	}
	wg.Wait()
	close(result)
	elapsed := time.Since(start)
	fmt.Printf("累加结果为: %d\n", total)
	fmt.Printf("执行耗时: %v\n\n", elapsed)
}

func task2() {
	fmt.Println("========== 任务2：10个goroutine ==========")
	start := time.Now()
	num := make([]int, 10000)
	for i := 0; i < 10000; i++ {
		num[i] = i
	}
	result := make(chan int, 10)
	var wg sync.WaitGroup
	chunkSize := 10000 / 10 // 每个goroutine处理1000个元素
	for j := 0; j < 10; j++ {
		wg.Add(1)
		begin := j * chunkSize
		end := begin + chunkSize
		go doSquare(num, begin, end, result, &wg)
	}
	var total int
	for i := 0; i < 10; i++ {
		total += <-result
	}
	wg.Wait()
	close(result)
	elapsed := time.Since(start)
	fmt.Printf("累加结果为: %d\n", total)
	fmt.Printf("执行耗时: %v\n\n", elapsed)
}
func main() {
	task1()
	task2()
}

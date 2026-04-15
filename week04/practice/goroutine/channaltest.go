package main

import (
	"fmt"
	"time"
)

func worker(ch, workerfirst chan int) {
	fmt.Println("Worker: Waiting for data...")
	workerfirst <- 1
	data := <-ch
	// workerfirst <- 1 //这样会发生死锁
	fmt.Println("Worker: Received data:", data)
}

func worker2(id int, jobs chan int, results chan int) {
	for job := range jobs {
		fmt.Printf("Worker %d: Processing job %d\n", id, job)
		time.Sleep(2 * time.Second) // 模拟处理时间
		results <- job * 2          // 将结果发送回主函数
	}
}
func dojob() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	for i := 1; i <= 2; i++ {
		go worker2(i, jobs, results)
	}
	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	close(jobs)
	// 收集结果
	for a := 1; a <= 5; a++ {
		<-results
	}
}
func main() {
	// ch := make(chan int)
	// workerfirst := make(chan int)
	// go worker(ch, workerfirst)
	// <-workerfirst //确保Worker先打印"Worker: Waiting for data..."，再发送数据
	// fmt.Println("Main: Sending data...")
	// ch <- 42
	// fmt.Println("Main: Data sent.")
	dojob()

}
func main2() {
	ch := make(chan int, 2)
	ch <- 1
	ch <- 2
	fmt.Println(<-ch)
	fmt.Println(<-ch)
}

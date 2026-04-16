package main

import (
	"fmt"
	"time"
)

func producer(ch chan<- int) {
	for i := 0; i < 5; i++ {
		fmt.Println("Producing:", i)
		time.Sleep(1 * time.Second) // 模拟生产时间
		ch <- i
	}
	close(ch)
}
func consummer(ch <-chan int) {
	for data := range ch {
		fmt.Println("Consuming:", data)
		time.Sleep(2 * time.Second) // 模拟消费时间
	}
}

func main() {
	ch := make(chan int, 10)
	go producer(ch)
	consummer(ch)
	//超时控制
	chan1 := make(chan string)

	go func() {
		time.Sleep(3 * time.Second)
		chan1 <- "Operation completed"
	}()

	select {
	case msg := <-chan1:
		fmt.Println(msg)
	case <-time.After(2 * time.Second):
		fmt.Println("Operation timed out")
	}
}

package main

import (
	"fmt"
	"time"
)

func printNumbers() {
	for i := 1; i <= 5; i++ {
		fmt.Println("我是print1", i)
		time.Sleep(500 * time.Millisecond)
	}
}
func printNumbersSyncByChanl(ch chan bool) {
	for i := 1; i <= 5; i++ {
		fmt.Println("我是print2", i)
		time.Sleep(500 * time.Millisecond)
	}
	ch <- true // 发送信号表示goroutine完成
}

func main() {
	go printNumbers() // 创建一个goroutine
	// time.Sleep(3 * time.Second) // 等待goroutine执行完毕
	/**如果将上面的sleep时间改为1秒，可能会看到部分数字输出，
	因为主函数可能在goroutine完成之前就结束了。**/
	// 或者可以通过channal实现goroutine的同步，确保主函数在goroutine完成之前不会退出。
	ch := make(chan bool)
	go printNumbersSyncByChanl(ch)
	<-ch // 等待goroutine完成
	fmt.Println("Main function finished")
	main2()
}

// 使用goroputine下载多个文件
func downloadFile(url string, ch chan string) {
	time.Sleep(2 * time.Second) // 模拟下载时间
	ch <- fmt.Sprintf("下载完成: %s", url)
}
func main2() {
	urls := []string{
		"https://example.com/file1",
		"https://example.com/file2",
		"https://example.com/file3",
	}
	ch := make(chan string)
	for _, url := range urls {
		go downloadFile(url, ch)
	}
	for range urls {
		fmt.Println(<-ch)
	}
}

package main

import (
	"fmt"
	"sync"
	"time"
)

func MutexDemo() {
	var count int = 0
	var mutex sync.Mutex
	var wait sync.WaitGroup
	for i := 0; i < 5; i++ {
		wait.Add(1)
		go func() {
			for j := 0; j < 3; j++ {
				mutex.Lock()
				fmt.Printf("我是第%d个goroutine,此时count为：%d\n", i, count)
				count++
				mutex.Unlock()
			}
			wait.Done()
		}()
	}
	wait.Wait()
	fmt.Printf("最终count的值为：%d\n", count)
}
func RwMutexExample() {
	var count int = 0
	var rwlock sync.RWMutex
	var wait sync.WaitGroup
	var writeDone sync.WaitGroup
	writeDone.Add(1)
	go func() {
		rwlock.Lock()
		time.Sleep(500 * time.Millisecond) //模拟写操作的耗时
		count++
		fmt.Printf("我是写操作的goroutine,此时count为：%d\n", count)
		rwlock.Unlock()
		writeDone.Done()
	}()
	for i := 0; i < 3; i++ {
		wait.Add(1)
		go func() {
			writeDone.Wait() //确保写操作完成后再进行读操作
			rwlock.RLock()
			time.Sleep(100 * time.Millisecond) //模拟读操作的耗时
			fmt.Printf("我是第%d个读goroutine,此时count为：%d\n", i, count)
			rwlock.RUnlock()
			wait.Done()
		}()
	}
	wait.Wait()
}
func main() {
	MutexDemo()
	RwMutexExample()

}

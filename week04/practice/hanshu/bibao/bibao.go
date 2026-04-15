package main

import "time"

// 练习使用闭包实现一个函数的延迟执行，要求在调用该函数时，传入一个函数作为参数，该函数将在2秒后被执行。
func deferdemobybibao(a func()) func() {
	return func() {
		time.Sleep(2 * time.Second)
		a()
	}
}

//回调函数
//回调函数是指在其他函数中作为参数传递的函数，当其他函数需要执行某个操作时，会调用回调函数。
func callbackfunc(num []int, a func(int)) {
	for _, v := range num {
		a(v)
	}
}
func main() {
	// 练习使用闭包实现一个函数的延迟执行，要求在调用该函数时，传入一个函数作为参数，该函数将在2秒后被执行。
	deferde := deferdemobybibao(func() {
		println("这是被延迟执行的函数")
	})
	deferde()
	callbackfunc([]int{1, 2, 3, 4, 5}, func(num int) {
		println("这是回调函数的参数函数:", num)
	})
}

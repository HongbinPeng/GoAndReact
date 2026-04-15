package main

import "time"

type callback func(string)

func processData(msg string) {
	println("处理数据:", msg)
}
func getData(url string, cal callback) {
	time.Sleep(2 * time.Second)
	// 假设获取到的数据
	data := "Data from " + url
	println("获取到的数据:", data)
	// 调用回调函数处理数据
	cal(data)
}
func main() {
	getData("https://www.baidu.com", processData)
}

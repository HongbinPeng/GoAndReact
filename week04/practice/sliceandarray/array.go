package main

import "fmt"

func findmax(arr []int) int {
	max := arr[0]
	for _, value := range arr {
		if value > max {
			max = value
		}
	}
	return max
}
func main() {
	// 数组的初始化
	var a [3]int
	fmt.Println(a)
	var b = [5]int{1, 2, 3, 4}
	fmt.Println(b)
	c := [5]int{1, 2}
	fmt.Println(c)
	d := [...]int{1, 2, 3, 4, 5}
	fmt.Println(d)
	for index, value := range c {
		fmt.Printf("index: %d, value: %d\n", index, value)
	}
	var e = [...]int{1, 2, 3, 4, 5}
	var suma int
	for _, value := range e {
		suma += value
	}
	fmt.Println("数组的和为：", suma)
	fmt.Println("数组的最大值为：", findmax(e[:]))
}

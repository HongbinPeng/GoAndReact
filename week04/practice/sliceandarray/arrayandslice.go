package main

import "fmt"

func main() {
	// 数组
	var arr1 [3]int            // 零值：[0 0 0]
	fmt.Println("arr1:", arr1) // 输出 [0 0 0]
	arr2 := [5]int{1, 2, 3}    // 显式指定长度
	fmt.Println("arr2:", arr2) // 输出 [1 2 3 0 0]

	arr3 := [...]int{4, 5, 6}  // 自动推断长度
	fmt.Println("arr3:", arr3) // 输出 [4 5 6]

	// 切片
	var s1 []int            // 零值：nil
	fmt.Println("s1:", s1)  // 输出 []
	s2 := []int{1, 2, 3}    // 字面量初始化
	fmt.Println("s2:", s2)  // 输出 [1 2 3]
	s3 := make([]int, 2, 5) // make初始化（len=2, cap=5）
	fmt.Println("s3:", s3)  // 输出 []
	s4 := arr2[1:4]         // 从数组截取（左闭右开）
	fmt.Println("s4:", s4)  // 输出 [2 3 0]

	// 数组：值传递，拷贝整个数组
	arrA := [3]int{1, 2, 3}
	arrB := arrA
	arrB[0] = 99
	fmt.Println(arrA) // 输出 [1 2 3]（原数组不受影响）
	fmt.Println(arrB) // 输出 [99 2 3]

	// 切片：结构体值传递，共享底层数组
	sA := []int{1, 2, 3}
	sB := sA
	sB[0] = 99
	fmt.Println(sA) // 输出 [99 2 3]（原切片受影响！）
	fmt.Println(sB) // 输出 [99 2 3]
}

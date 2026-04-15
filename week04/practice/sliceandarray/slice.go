package main

import "fmt"

func printSlice(s []int) { //s是一个切片参数，类型为[]int
	for _, v := range s {
		fmt.Printf("%v ", v)
	}
	fmt.Println()
}
func printSlice2(s ...int) { //s是一个可变参数，类型为...int，实际上也是一个切片
	for _, v := range s {
		fmt.Printf("%v ", v)
	}
	fmt.Println()
}
func returnallnumgreaterthan10(num []int) []int {
	var result []int
	for _, v := range num {
		if v > 10 {
			result = append(result, v)
		}
	}
	return result
}
func main() {
	//切片的初始化
	s1 := []int{1, 2, 3}
	fmt.Printf("s1:%v\n 地址：%p\n", s1, &s1) // 输出 [1 2 3]
	s2 := make([]int, 2, 5)                // make初始化（len=2, cap=5）
	fmt.Println("s2:", s2)                 // 输出 [0 0]
	//len和cap, len是切片的元素个数, cap是切片的底层数组的容量
	fmt.Println("len(s1):", len(s1)) // 输出 3
	fmt.Println("len(s2):", len(s2)) // 输出 2
	fmt.Println("cap(s2):", cap(s2)) // 输出 5
	s1 = append(s1, 4, 5, 6, 7, 8)   // append函数用于向切片末尾添加元素，扩容后返回新的切片
	//不扩容就不改变底层数组的容量和地址
	fmt.Printf("append s1:%v,地址：%p\n", s1, &s1) // 输出 [1 2 3 4 5]
	// 调用函数
	printSlice(s1)
	printSlice2(s1...)         //切片必须展开，否则会报错
	printSlice2(9, 8, 7, 6, 5) //直接传入可变参数

	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	/**
	copy函数的行为取决于切片元素的类型：
	1. 值类型（如int, float64等）：copy函数会复制元素的值，而不是指针地址
	2. 指针类型（如*int, *float64等）：copy函数会复制指针地址，而不是指针指向的底层值
	因此，对于值类型的切片，修改dst的元素不会影响src；而对于指针类型的切片，修改dst的元素指向的底层值会影响src，因为它们共享同一个底层值。
	copy函数只会复制目标切片的长度范围内的元素，不会改变目标切片的容量。
	**/
	// 复制元素
	copy(dst, src) //copy函数用于将src切片的元素复制到dst切片中，因此类似于深拷贝
	fmt.Println("复制后初始状态：")
	fmt.Println("src:", src) // 输出 [1 2 3]
	fmt.Println("dst:", dst) // 输出 [1 2 3]
	// 修改 dst 的元素，src 不受影响（因为元素是值类型，复制的是值）
	dst[0] = 99
	fmt.Println("\n修改 dst 后：")
	fmt.Println("src:", src) // 输出 [1 2 3]（不变）
	fmt.Println("dst:", dst) // 输出 [99 2 3]（变了）

	// 定义一个元素是指针的切片
	a, b := 1, 2
	src2 := []*int{&a, &b}
	dst2 := make([]*int, len(src))
	// 复制元素（复制的是指针地址）//由于拷贝的是指针元素，所以修改dst2的元素指向的底层值，src2也会受影响
	copy(dst2, src2)
	fmt.Println("复制后初始状态：")
	fmt.Println("src 元素指向的值：", *src2[0], *src2[1]) // 输出 1 2
	fmt.Println("dst 元素指向的值：", *dst2[0], *dst2[1]) // 输出 1 2
	// 修改 dst 元素指向的底层值，src 也会受影响！
	*dst2[0] = 99
	fmt.Println("\n修改 dst 指向的底层值后：")
	fmt.Println("src 元素指向的值：", *src2[0], *src2[1]) // 输出 99 2（src 也变了！）
	fmt.Println("dst 元素指向的值：", *dst2[0], *dst2[1]) // 输出 99 2

	/**
	截取操作不会创建新的底层数组，而是共享原切片的底层数组。因此，修改子切片会影响原切片。
	**/
	original := []int{1, 2, 3, 4, 5}
	sub := original[1:4]               // 截取 original 的一部分，得到 sub 切片
	fmt.Println("sub:", sub)           // 输出 [2 3 4]
	fmt.Println("original:", original) // 输出 [1 2 3 4 5]
	// 修改 sub 的元素，original 也会受影响！
	sub[0] = 99
	fmt.Println("\n修改 sub 后：")
	fmt.Println("sub:", sub)           // 输出 [99 3 4]
	fmt.Println("original:", original) // 输出 [1 99 3 4 5]
	//切片没有删除的概念，截取操作只是创建了一个新的切片头部，指向原切片的底层数组，因此修改子切片会影响原切片。
	//如果要删除子切片，需要使用append函数删除子切片的元素，然后重新截取原切片。
	nums := []int{1, 2, 3, 4, 5, 90, 23, 12, 3, 43}
	result := returnallnumgreaterthan10(nums)
	fmt.Println("大于10的数有：", result)
}

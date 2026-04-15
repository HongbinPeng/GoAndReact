package main

import (
	"fmt"
	"hanshu/calculate"
)

func main() {

	/**
	学习包名、模块名的使用，
	函数的定义和调用，函数的参数传递
	函数的返回值，函数的异常处理
	函数的变量作用域，函数的内存地址等知识点。
	**/
	var a float64
	var b float64
	var result float64
	// var err error
	fmt.Printf("result的地址是 :%p\n", &result) // 打印result的地址，验证函数内部是否为同一个result变量
	fmt.Println("请输入第一个数字：")
	fmt.Scanln(&a)
	fmt.Println("请输入你要实现的运算操作：(+, -, *, /)")
	var operator string
	fmt.Scanln(&operator)
	fmt.Println("请输入第二个数字：")
	fmt.Scanln(&b)
	if opfunc, ok := calculate.OperatorFuncs[operator]; ok { // 从calculate包中获取包级的全局变量，但是不推荐这么做，因为这样会增加代码的耦合度，建议在main函数中定义一个局部的运算符函数映射，或者直接调用calculate包中的函数。
		result, err := opfunc(a, b)
		fmt.Printf("result的地址是 ：%p\n", &result) // 打印result的地址，这时不是同一个变量
		// 处理异常
		if err != nil {
			fmt.Printf("发生错误：%s\n", err)
			return
		}
		fmt.Printf("结果是：%.2f\n", result)
	} else {
		fmt.Println("无效的运算符")
	}

}

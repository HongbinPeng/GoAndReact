package main

import (
	"errors"
	"fmt"
)

func add(a, b float64) (float64, error) {
	return a + b, nil
}
func sub(a, b float64) (float64, error) {
	return a - b, nil
}
func mul(a, b float64) (float64, error) {
	return a * b, nil
}
func div(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为零")
	}
	return a / b, nil
}
func main() {
	var a float64
	var b float64
	var result float64
	fmt.Println("请输入第一个数字：")
	fmt.Scanln(&a)
	fmt.Println("请输入你要实现的运算操作：(+, -, *, /)")
	var operator string
	fmt.Scanln(&operator)
	fmt.Println("请输入第二个数字：")
	fmt.Scanln(&b)
	operatorfunc := map[string]func(float64, float64) (float64, error){
		"+": add,
		"-": sub,
		"*": mul,
		"/": div,
	}
	opfunc, ok := operatorfunc[operator]
	if !ok {
		fmt.Println("无效的运算符")
		return
	}
	result, err := opfunc(a, b)
	// 处理异常
	if err != nil {
		fmt.Printf("发生错误：%s\n", err)
		return
	}
	fmt.Printf("结果是：%.2f\n", result)
}

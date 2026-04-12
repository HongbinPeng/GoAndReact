package main

import "fmt"

func add(a, b float64) float64 {
	return a + b
}
func sub(a, b float64) float64 {
	return a - b
}
func mul(a, b float64) float64 {
	return a * b
}
func div(a, b float64) float64 {

	return a / b
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
	switch operator {
	case "+":
		result = add(a, b)
	case "-":
		result = sub(a, b)
	case "*":
		result = mul(a, b)
	case "/":
		result = div(a, b)
	default:
		fmt.Println("请输入正确的运算符")
		return
	}
	fmt.Printf("结果是：%.2f\n", result)
}

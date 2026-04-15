package calculate

import "errors"

//加法
//接受两个参数，返回一个结果和一个错误，错误在除法运算中可能会发生，其他运算不会发生错误，所以返回值中的错误可以为nil。
func Add(a, b float64) (float64, error) {
	return a + b, nil
}

//减法
// 函数的命名要符合Go语言的命名规范，函数名要以大写字母开头，这样才能在其他包中被调用，如果函数名以小写字母开头，那么这个函数只能在当前包中被调用。
func Sub(a, b float64) (float64, error) {
	return a - b, nil
}

//乘法运算
func Mul(a, b float64) (float64, error) {
	return a * b, nil
}

//除法运算，除数不能为零，如果除数为零，返回一个错误，否则返回结果和nil。
func Div(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除数不能为零")
	}
	return a / b, nil
}

// 定义一个全局的运算符函数映射，方便在其他地方调用
var OperatorFuncs = map[string]func(float64, float64) (float64, error){
	"+": Add,
	"-": Sub,
	"*": Mul,
	"/": Div,
}

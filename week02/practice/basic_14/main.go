package main

import "fmt"

func Huiwen(x int64) bool {
	if x < 0 {
		return false
	}
	var reversed int64
	var original int64 = x
	for x != 0 {
		reversed = reversed*10 + x%10
		x /= 10
	}
	return original == reversed
}
func main() {
	fmt.Println("121是否是回文数：", Huiwen(121))
	fmt.Println("(-121)是否是回文数：", Huiwen(-121))
	fmt.Println("10是否是回文数：", Huiwen(10))
}

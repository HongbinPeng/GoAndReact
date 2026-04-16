package rune

import (
	"fmt"
	"strings"
)

func TestString() {
	fmt.Print(
		`
	这是一个字符串
	这是第二行
	`)
}
func Reverse(s string) {
	fmt.Printf("original string: %s\n", s)
	resouce := []rune(s)
	l, r := 0, len(resouce)-1
	for l <= r {
		resouce[l], resouce[r] = resouce[r], resouce[l]
		l++
		r--
	}
	fmt.Printf("reversed string: %s\n", string(resouce))
	strings.Split(s, " ")

}

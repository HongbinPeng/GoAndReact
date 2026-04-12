package main

import (
	"fmt"
	"hello/a"
	"hello/b"
)

func init() {
	fmt.Println("hello initialized")
}

func main() {
	fmt.Println("main go")
	a.SayHelloFromA()
	b.SayHelloFromB()
}

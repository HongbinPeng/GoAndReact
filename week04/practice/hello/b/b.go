package b

import "fmt"

func init() {
	fmt.Println("b initialized")
}

func SayHelloFromB() {
	fmt.Println("hello from package b")
}

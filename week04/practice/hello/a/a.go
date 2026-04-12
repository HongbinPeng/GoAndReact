package a

import (
	"fmt"
)

func init() {
	fmt.Println("a initialized")
}

func SayHelloFromA() {
	fmt.Println("hello from package a")
}

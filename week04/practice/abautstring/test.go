package main

import (
	"fmt"

	"aboutstring/rune"
)

func main() {
	rune.TestString()
	fmt.Println("Hello, 世界")
	rune.Reverse("Hello, 世界")
}

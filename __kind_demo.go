package main
import (
    "fmt"
    "reflect"
    "sync"
)
func main() {
    var ch chan int
    var wg sync.WaitGroup
    fmt.Println("ch kind:", reflect.TypeOf(ch).Kind())
    fmt.Println("wg kind:", reflect.TypeOf(wg).Kind())
}

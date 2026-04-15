package main
import (
	"fmt"
)
func main(){
	mapa:=make(map[string]int){
		"apple":  5,
		"orange": 10,
		"banana": 20,
	}
	fmt.Println(mapa)
	// 遍历map
	for k,v:=range mapa{
		fmt.Printf(k,v)
	}
	// 删除map中的元素
	delete(mapa,"orange")
	fmt.Println(mapa)
	// 清空map
	mapa=make(mapa)
	fmt.Println(mapa)
	// 判断map中是否存在某个键
	_,ok:=mapa["apple"]
	fmt.Println(ok)
}
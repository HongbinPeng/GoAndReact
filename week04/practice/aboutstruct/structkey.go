package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

//可以为结构体设置函数,函数的接收者可以是值接收者也可以是指针接收者
// 值接收者：操作副本，不会影响原对象
// 指针接收者：操作原对象，会影响原对象,实际就是指针传递
// 值接收者：操作副本
func (p *Person) GrowUpValue() {
	p.Age++ // 修改的是副本的 Age
	fmt.Println("值接收者内部 Age:", p.Age)
}

// 指针接收者：操作原对象
func (p *Person) GrowUpPointer() {
	p.Age++ // 修改的是原对象的 Age
	fmt.Println("指针接收者内部 Age:", p.Age)
}

type User struct {
	Name    string
	Age     int
	Address Address
}
type Address struct {
	City     string
	Province string
	Contry   string
}

//可以为结构体设置函数

func main() {
	bob := Person{Name: "Bob", Age: 20}
	fmt.Println("初始 Age:", bob.Age) // 输出：20
	bob.GrowUpValue()
	fmt.Println("值接收者后 Age:", bob.Age) // 输出：20（原对象没变！）
	bob.GrowUpPointer()
	fmt.Println("指针接收者后 Age:", bob.Age) // 输出：21（原对象变了！）
	/**
	结构体的初始化
	之所以值和指针都能用 . 访问字段，
	是因为 Go 对结构体指针支持自动解引用，
	所以 p.Name 实际上可以等价理解为 (*p).Name。
	也就是u := &User{Name: "Alice"}
	fmt.Println(u.Name)
	等价于
	fmt.Println((*u).Name)
	输出：Alice
	**/
	u := User{ //结构体字面量初始化
		Name: "Alice",
		Age:  30,
		Address: Address{
			City:     "Beijing",
			Province: "Beijing",
			Contry:   "China",
		},
	}

	fmt.Println(u)              //输出：{Alice 30 {Beijing Beijing China}}
	fmt.Println(u.Address)      //输出：{Beijing Beijing China}
	fmt.Println(u.Address.City) //输出：Beijing
	u2 := new(User)             //结构体指针初始化
	u2.Name = "Bob"
	u2.Age = 25
	u2.Address = Address{
		City:     "Shanghai",
		Province: "Shanghai",
		Contry:   "China",
	}
	fmt.Println(u2)
	fmt.Println(u2.Address)
	fmt.Println(u2.Address.City)
	var u3 User = User{
		Name: "Charlie",
		Age:  40,
		Address: Address{
			City:     "Guangzhou",
			Province: "Guangdong",
			Contry:   "China",
		},
	}
	fmt.Println(u3)
	fmt.Println(u3.Address)
	fmt.Println(u3.Address.City)
	u4 := &User{
		Name: "David",
		Age:  35,
		Address: Address{
			City:     "chengdu",
			Province: "Sichuan",
			Contry:   "China",
		},
	}
	fmt.Println(u4)
	fmt.Println(u4.Address)
	fmt.Println(u4.Address.City)

}

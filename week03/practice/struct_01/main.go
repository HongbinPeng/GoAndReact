package main

import (
	"encoding/json"
	"fmt"
)

type Person struct {
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

func NewPerson(name string, age int, email string) Person {
	return Person{
		Name:  name,
		Age:   age,
		Email: email,
	}
}
func PrintPerson(p Person) {
	fmt.Printf("姓名: %s\n年龄: %d\n邮箱: %s\n", p.Name, p.Age, p.Email)
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		fmt.Printf("JSON 转换失败: %v\n", err)
		return
	}
	fmt.Printf("JSON: %s\n", string(jsonBytes))
}
func main() {
	p := NewPerson("张三", 25, "zhangsan@example.com")
	PrintPerson(p)
}

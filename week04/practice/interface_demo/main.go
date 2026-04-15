package main

import "fmt"

// Speaker 接口定义了一种“会说话”的行为。
// 只要某个类型实现了 Speak() 方法，它就自动实现了这个接口，
// 不需要像其他语言那样显式写 implements。
type Speaker interface {
	Speak() string
}

// Walker 接口定义“会走路”的行为。
type Walker interface {
	Walk() string
}

// Pet 是接口组合。
// 这表示：想成为 Pet，必须同时实现 Speak 和 Walk。
type Pet interface {
	Speaker
	Walker
}

type Dog struct {
	Name string
}

func (d Dog) Speak() string {
	return d.Name + "：汪汪汪"
}

func (d Dog) Walk() string {
	return d.Name + " 正在开心地散步"
}

type Robot struct {
	ID string
}

// Robot 只实现了 Speak，没有实现 Walk。
// 所以 Robot 是 Speaker，但不是 Pet。
func (r Robot) Speak() string {
	return "机器人 " + r.ID + "：你好，我可以陪你聊天"
}

// talkTo 只关心“你会不会说话”，并不关心传进来的是 Dog 还是 Robot。
// 这就是接口的意义：面向行为编程，而不是只面向具体类型编程。
func talkTo(s Speaker) {
	fmt.Printf("类型：%T，内容：%s\n", s, s.Speak())
}

// playWith 要求参数同时具备 Speak 和 Walk 两种能力。
func playWith(p Pet) {
	fmt.Println("和宠物互动：")
	fmt.Println(p.Speak())
	fmt.Println(p.Walk())
}

// printValue 演示 any 的用途。
// any 是 interface{} 的别名，表示“可以接收任意类型的值”。
func printValue(v any) {
	switch value := v.(type) {
	case int:
		fmt.Printf("这是 int 类型，值是 %d\n", value)
	case string:
		fmt.Printf("这是 string 类型，值是 %q\n", value)
	case Dog:
		fmt.Printf("这是 Dog 类型，它会说：%s\n", value.Speak())
	default:
		fmt.Printf("未知类型：%T，值：%v\n", value, value)
	}
}

func main() {
	dog := Dog{Name: "旺财"}
	robot := Robot{ID: "R2D2"}

	fmt.Println("示例 1：接口的隐式实现")
	talkTo(dog)
	talkTo(robot)

	fmt.Println()
	fmt.Println("示例 2：接口组合")
	playWith(dog)
	// playWith(robot)
	// 上面这行如果取消注释，会编译报错。
	// 因为 Robot 没有 Walk() 方法，不满足 Pet 接口。

	fmt.Println()
	fmt.Println("示例 3：any + type switch")
	printValue(18)
	printValue("hello interface")
	printValue(dog)
	printValue(3.14)
}

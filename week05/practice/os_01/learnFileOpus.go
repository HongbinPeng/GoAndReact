package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("========== 文件操作 = ==========")
	// file, err := os.Create("os_create_demo.txt")
	//如果当前的工作目录在C:\AllCode\VueProject\penghongbin，那么创建的文件
	//会在C:\AllCode\VueProject\penghongbin目录下面，因为这个是工作目录+传入的相对路径构造的
	//Create()方法底层调用了os.OpenFile()方法,并且默认权限是0666
	// 做了三件事：
	// O_RDWR — 读写模式打开
	// O_CREATE — 文件不存在则创建，但是前提是目录存在，否则会报错
	// O_TRUNC — 文件存在则截断清空，就是存在则清空，与之相反的是O_APPEND，就是存在则追加
	// 权限 0666（实际会受 umask 影响，最终变为 0644
	//如果目录存在，那么会返回一个错误,
	//关于打开文件的flag问题，有以下几个参数需要记着：
	// O_RDONLY — 读模式打开
	// O_WRONLY — 只写模式打开
	// O_RDWR — 读写模式打开
	// O_APPEND — 追加模式打开
	// O_CREATE — 文件不存在则创建，但是前提是目录存在，否则会报错
	// O_TRUNC — 文件存在则截断清空，就是存在则清空，与之相反的是O_APPEND，就是存在则追加
	// 权限 0666（实际会受 umask 影响，最终变为 0644
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// defer file.Close()
	// log.Println("文件创建成功")
	// readFile, err := os.OpenFile("os_create_demo.txt", os.O_RDONLY, 0666)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }
	// defer readFile.Close()
	// log.Println("文件读取成功")
	// n, err := readFile.Write([]byte("hello world")) //这里就会失败，因为是只读模式打开的，不能写入
	// if err != nil {
	// 	log.Println(err)
	// }
	/*
			┌──────────────────┬─────────────────────────────────────────────┐
		│ 组合              │ 效果                                        │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_RDONLY          │ 只读打开（对应 os.Open）                    │
		│                  │ 等价于：OpenFile(name, O_RDONLY, 0)          │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_WRONLY          │ 只写打开，文件必须存在                        │
		│                  │ 文件不存在 → 报错                             │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_RDWR            │ 读写打开，文件必须存在                        │
		│                  │ 文件不存在 → 报错                             │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_WRONLY|O_CREATE │ 只写打开，不存在则创建                       │
		│                  │ 存在 → 从开头覆盖写入                         │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_WRONLY|O_CREATE │ 写打开，不存在则创建，存在则追加              │
		│     |O_APPEND     │ （日志文件的标准写法）                       │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_WRONLY|O_CREATE │ 只写打开，不存在则创建，存在则清空            │
		│     |O_TRUNC      │ （覆盖文件的写法）                           │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_RDWR|O_CREATE   │ 读写打开，不存在则创建                       │
		│     |O_TRUNC      │ 存在则清空（对应 os.Create）                 │
		│                  │ 等价于：Create(name)                         │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_WRONLY|O_CREATE │ 独占创建：文件必须不存在                      │
		│     |O_EXCL       │ 存在就报错，避免覆盖已有文件                  │
		├──────────────────┼─────────────────────────────────────────────┤
		│ O_WRONLY|O_SYNC   │ 每次写入都立即刷盘                           │
		│                  │ 适合写重要数据（如数据库 WAL）                 │
		└──────────────────┴─────────────────────────────────────────────┘

	*/
	appendFile, err := os.OpenFile("os_create_demo.txt", os.O_APPEND, 0666) //OpengFile是底层的最佳实践
	if err != nil {
		log.Println(err)
	}
	defer appendFile.Close()
	n, err := appendFile.Write([]byte("\nhello world"))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("文件追加成功")
	log.Println(n)
}

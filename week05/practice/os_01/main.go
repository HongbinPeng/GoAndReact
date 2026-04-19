package main

import (
	"fmt"
	"os"
)

// 运行方式：
// go run main.go
//
// 这份示例用于练习 os 标准库。
// 在作业里，os 主要负责：
// 1. 读取 config.json
// 2. 创建监控报告文件
// 3. 查看文件是否存在
// 4. 获取当前工作目录

func main() {
	fmt.Println("========== os 标准库演示 ==========")

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("获取当前目录失败：", err)
		return
	}
	fmt.Println("当前工作目录：", currentDir)

	fileName := "os_demo.txt"
	content := []byte("这是 os.WriteFile 写入的示例内容。\n")

	if err := os.WriteFile(fileName, content, 0644); err != nil {
		fmt.Println("写文件失败：", err)
		return
	}
	defer os.Remove(fileName)

	fmt.Printf("已写入文件：%s\n", fileName)

	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("读文件失败：", err)
		return
	}
	fmt.Printf("读取到的内容：%s", string(data))

	info, err := os.Stat(fileName)
	if err != nil {
		fmt.Println("Stat 失败：", err)
		return
	}
	fmt.Printf("文件大小：%d 字节\n", info.Size())
	fmt.Printf("是否目录：%v\n", info.IsDir())

	f, err := os.Create("os_create_demo.txt")
	if err != nil {
		fmt.Println("Create 失败：", err)
		return
	}
	if _, err := f.WriteString("这是通过 os.Create 创建并写入的内容。\n"); err != nil {
		fmt.Println("写入 Create 文件失败：", err)
		f.Close()
		return
	}
	f.Close()
	defer os.Remove("os_create_demo.txt")

	fmt.Println("\n常见知识点：")
	fmt.Println("1. os.ReadFile 适合直接读完整个小文件。")
	fmt.Println("2. os.Create 会创建或截断一个文件。")
	fmt.Println("3. 文件创建后要记得 Close。")
	fmt.Println("4. os.Stat 常用于检查文件是否存在、大小多少、是不是目录。")
	fmt.Println("5. 作业里读取 config.json 最常用的就是 os.ReadFile。")
}

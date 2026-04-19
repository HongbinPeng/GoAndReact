package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// 运行方式：
// go run main.go
//
// tabwriter 不是“必须才能交作业”的标准库，
// 但如果你想让命令行报告更像工程里的 CLI 表格，它非常值得学。

func main() {
	fmt.Println("========== text/tabwriter 标准库演示 ==========")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "服务名\t协议\t状态\t耗时")
	fmt.Fprintln(w, "百度\tHTTP\t200 OK\t80ms")
	fmt.Fprintln(w, "GitHub\tHTTP\t200 OK\t120ms")
	fmt.Fprintln(w, "MySQL\tTCP\tConnected\t3ms")
	fmt.Fprintln(w, "无效服务\tHTTP\tFailed\t3s")
	w.Flush()

	fmt.Println("\n常见知识点：")
	fmt.Println("1. 列之间用 \\t 分隔。")
	fmt.Println("2. tabwriter 会自动把各列对齐。")
	fmt.Println("3. Flush 之后内容才会真正输出。")
}

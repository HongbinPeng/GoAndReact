package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 存储命令行参数
type Config struct {
	File      string
	Operation string
	Output    string
}

// parseFlags 解析命令行参数
func parseFlags() Config {
	file := flag.String("file", "", "指定要处理的文件路径") //默认文件路径是相对路径，不指定就为空字符串
	operation := flag.String("operation", "", "指定操作类型（count、convert、upper）")
	output := flag.String("output", "", "指定输出文件路径")
	flag.Parse() //解析命令行参数并将结果存储在对应的变量中

	return Config{
		File:      *file,
		Operation: *operation,
		Output:    *output,
	}
}

// validateFlags 校验必填参数
func validateFlags(cfg Config) {
	if cfg.File == "" || cfg.Operation == "" {
		fmt.Println("用法: main -file=<文件路径> -operation=<操作类型> [-output=<输出路径>]")
		fmt.Println("操作类型: count(统计字符数)  convert(转数字格式)  upper(转大写)")
		os.Exit(1)
	}
}

// readFile读取文件内容
func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		os.Exit(1)
	}
	return string(data)
}

// countChars统计文件字符数
func countChars(content string) string {
	return fmt.Sprintf("字符数: %d\n", len([]rune(content)))
}

// extractNumbers提取内容中的数字行
func extractNumbers(content string) string {
	var nums []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if _, err := strconv.Atoi(line); err == nil {
			nums = append(nums, line)
		}
	}
	return strings.Join(nums, "\n")
}

// toUpper将文本转为大写
func toUpper(content string) string {
	return strings.ToUpper(content)
}

// writeOutput将结果输出到文件或终端
func writeOutput(result, outputPath string) {
	if outputPath != "" {
		err := os.WriteFile(outputPath, []byte(result), 0644)
		if err != nil {
			fmt.Printf("写入文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("结果已写入文件: %s\n", outputPath)
	} else {
		fmt.Print(result)
	}
}
func main() {
	cfg := parseFlags()
	validateFlags(cfg)
	content := readFile(cfg.File)
	var result string
	switch cfg.Operation {
	case "count":
		result = countChars(content)
	case "convert":
		result = extractNumbers(content)
	case "upper":
		result = toUpper(content)
	default:
		fmt.Printf("不支持的操作类型: %s\n", cfg.Operation)
		os.Exit(1)
	}
	writeOutput(result, cfg.Output)
}

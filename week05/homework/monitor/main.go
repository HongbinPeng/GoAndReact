package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	options, err := parseOptions(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "参数解析失败：", err)
		os.Exit(1)
	}

	cfg, err := loadConfig(options.ConfigPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "配置加载失败：", err)
		os.Exit(1)
	}

	monitor := NewMonitor(options.Timeout, options.Verbose)
	results := monitor.Run(cfg.Targets)

	generatedAt := time.Now()
	summary := buildSummary(results)
	report := renderReport(results, summary, options, generatedAt)

	fmt.Println(report)

	fileName, err := writeReportFile(report, generatedAt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "报告写入失败：", err)
		os.Exit(1)
	}

	fmt.Printf("报告已生成：%s\n", fileName)
}

func parseOptions(args []string) (Options, error) {
	fs := flag.NewFlagSet("monitor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := fs.String("config", "config.json", "配置文件路径")
	timeoutSeconds := fs.Int("timeout", 3, "单次探测超时时间（秒）")
	verbose := fs.Bool("v", false, "开启详细模式")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if *timeoutSeconds <= 0 {
		return Options{}, fmt.Errorf("timeout 必须大于 0")
	}

	return Options{
		ConfigPath: *configPath,
		Timeout:    time.Duration(*timeoutSeconds) * time.Second,
		Verbose:    *verbose,
	}, nil
}

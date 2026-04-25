package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
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

	monitor := NewMonitor(options.Timeout, options.Verbose, options.Proxy)
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
	proxy := fs.String("proxy", "", "HTTP/HTTPS/SOCKS5 代理地址，例如 http://127.0.0.1:7890")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}
	if *timeoutSeconds <= 0 {
		return Options{}, fmt.Errorf("timeout 必须大于 0")
	}

	normalizedProxy, err := normalizeProxyAddress(*proxy)
	if err != nil {
		return Options{}, err
	}

	return Options{
		ConfigPath: *configPath,
		Timeout:    time.Duration(*timeoutSeconds) * time.Second,
		Verbose:    *verbose,
		Proxy:      normalizedProxy,
	}, nil
}

func normalizeProxyAddress(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	parsedURL, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("proxy 格式不合法：%w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	switch scheme {
	case "http", "https", "socks5", "socks5h":
	default:
		return "", fmt.Errorf("proxy 协议只支持 http、https、socks5 或 socks5h")
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("proxy 必须包含主机和端口")
	}

	return parsedURL.String(), nil
}

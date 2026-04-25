package main

import (
	"testing"
	"time"
)

func TestParseOptionsAcceptsProxy(t *testing.T) {
	options, err := parseOptions([]string{
		"--config", "custom.json",
		"--timeout", "5",
		"-v",
		"--proxy", "http://127.0.0.1:7890",
	})
	if err != nil {
		t.Fatalf("parseOptions 返回了非预期错误：%v", err)
	}

	if options.ConfigPath != "custom.json" {
		t.Fatalf("ConfigPath = %q，期望 custom.json", options.ConfigPath)
	}
	if options.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %v，期望 5s", options.Timeout)
	}
	if !options.Verbose {
		t.Fatalf("Verbose = false，期望 true")
	}
	if options.Proxy != "http://127.0.0.1:7890" {
		t.Fatalf("Proxy = %q，期望 http://127.0.0.1:7890", options.Proxy)
	}
}

func TestParseOptionsRejectsInvalidProxy(t *testing.T) {
	if _, err := parseOptions([]string{"--proxy", "127.0.0.1:7890"}); err == nil {
		t.Fatalf("期望 parseOptions 因非法 proxy 失败，但实际成功")
	}
}

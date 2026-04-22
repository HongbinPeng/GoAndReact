package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigSuccess(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "targets": [
    {
      "name": "测试 HTTP",
      "protocol": " HTTP ",
      "address": " https://example.com ",
      "expect": {
        "status_code": 200
      }
    },
    {
      "name": "测试 TCP",
      "protocol": "tcp",
      "address": "127.0.0.1:3306",
      "retry_count": 2
    }
  ]
}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试配置文件失败：%v", err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig 返回了非预期错误：%v", err)
	}

	if len(cfg.Targets) != 2 {
		t.Fatalf("目标数量 = %d，期望 2", len(cfg.Targets))
	}

	if cfg.Targets[0].Protocol != "http" {
		t.Fatalf("HTTP 协议规范化失败，实际为 %q", cfg.Targets[0].Protocol)
	}

	if cfg.Targets[0].Address != "https://example.com" {
		t.Fatalf("HTTP 地址规范化失败，实际为 %q", cfg.Targets[0].Address)
	}

	if cfg.Targets[1].RetryCount != 2 {
		t.Fatalf("retry_count = %d，期望 2", cfg.Targets[1].RetryCount)
	}
}

func TestLoadConfigRejectsRetryCountAboveThree(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "targets": [
    {
      "name": "错误目标",
      "protocol": "http",
      "address": "https://example.com",
      "retry_count": 4
    }
  ]
}`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("写入测试配置文件失败：%v", err)
	}

	if _, err := loadConfig(configPath); err == nil {
		t.Fatalf("期望 loadConfig 因 retry_count > 3 失败，但实际成功")
	}
}

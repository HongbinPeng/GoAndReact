package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("解析配置文件失败: %w", err)
	}

	if err := ensureSingleJSONDocument(decoder); err != nil {
		return Config{}, err
	}

	if err := validateAndNormalizeConfig(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ensureSingleJSONDocument(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return errors.New("配置文件中存在多余的 JSON 内容")
		}
		return fmt.Errorf("检查配置文件尾部内容失败: %w", err)
	}
	return nil
}

func validateAndNormalizeConfig(cfg *Config) error {
	if len(cfg.Targets) == 0 {
		return errors.New("配置文件中至少需要一个探测目标")
	}

	seenNames := make(map[string]struct{}, len(cfg.Targets))

	for i := range cfg.Targets {
		target := &cfg.Targets[i]
		target.Index = i
		target.Name = strings.TrimSpace(target.Name)
		target.Protocol = strings.ToLower(strings.TrimSpace(target.Protocol))
		target.Address = strings.TrimSpace(target.Address)
		target.Expect.Contains = strings.TrimSpace(target.Expect.Contains)

		if target.Name == "" {
			return fmt.Errorf("第 %d 个目标缺少 name", i+1)
		}
		if _, exists := seenNames[target.Name]; exists {
			return fmt.Errorf("目标名称重复: %s", target.Name)
		}
		seenNames[target.Name] = struct{}{}

		if target.Protocol == "" {
			return fmt.Errorf("目标 %s 缺少 protocol", target.Name)
		}
		if target.Address == "" {
			return fmt.Errorf("目标 %s 缺少 address", target.Name)
		}
		if target.RetryCount < 0 || target.RetryCount > 3 {
			return fmt.Errorf("目标 %s 的 retry_count 必须在 0 到 3 之间", target.Name)
		}

		switch target.Protocol {
		case "http":
			if err := validateHTTPConfig(*target); err != nil {
				return err
			}
		case "tcp":
			if err := validateTCPConfig(*target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("目标 %s 使用了不支持的 protocol: %s", target.Name, target.Protocol)
		}
	}

	return nil
}

func validateHTTPConfig(target Target) error {
	parsedURL, err := url.ParseRequestURI(target.Address)
	if err != nil {
		return fmt.Errorf("HTTP 目标 %s 的地址不合法: %w", target.Name, err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("HTTP 目标 %s 只能使用 http 或 https 协议", target.Name)
	}
	if parsedURL.Host == "" {
		return fmt.Errorf("HTTP 目标 %s 缺少主机名", target.Name)
	}
	if target.Expect.Connected != nil {
		return fmt.Errorf("HTTP 目标 %s 不能配置 expect.connected", target.Name)
	}
	if target.Expect.StatusCode != nil {
		if *target.Expect.StatusCode < 100 || *target.Expect.StatusCode > 599 {
			return fmt.Errorf("HTTP 目标 %s 的 expect.status_code 必须在 100 到 599 之间", target.Name)
		}
	}
	return nil
}

func validateTCPConfig(target Target) error {
	if _, _, err := net.SplitHostPort(target.Address); err != nil {
		return fmt.Errorf("TCP 目标 %s 的地址必须是 host:port 格式: %w", target.Name, err)
	}
	if target.Expect.StatusCode != nil {
		return fmt.Errorf("TCP 目标 %s 不能配置 expect.status_code", target.Name)
	}
	if target.Expect.Contains != "" {
		return fmt.Errorf("TCP 目标 %s 不能配置 expect.contains", target.Name)
	}
	if target.Expect.Connected != nil && !*target.Expect.Connected {
		return fmt.Errorf("TCP 目标 %s 的 expect.connected 只能为 true", target.Name)
	}
	return nil
}

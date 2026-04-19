package main

import (
	"strings"
	"testing"
)

// 运行方式：
// go test -v
//
// 这份示例用于学习 testing 标准库最常用的写法：
// 1. TestXxx 函数命名
// 2. 表驱动测试
// 3. t.Run 子测试
// 4. t.Fatalf 失败即停止

func normalizeProtocol(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func TestNormalizeProtocol(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "大写 HTTP", input: "HTTP", want: "http"},
		{name: "前后有空格", input: "  TCP  ", want: "tcp"},
		{name: "已经是小写", input: "http", want: "http"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeProtocol(tc.input)
			if got != tc.want {
				t.Fatalf("normalizeProtocol(%q) = %q, 期望 %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRetryCountDefaultMeaning(t *testing.T) {
	retryCount := 0
	if retryCount != 0 {
		t.Fatalf("默认 retryCount 应该是 0，实际得到 %d", retryCount)
	}
}

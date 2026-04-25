package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Config 表示整个配置文件。
type Config struct {
	Targets []Target `json:"targets"`
}

/*
在这里我们要注意零值，当用户的配置文件中没有设置某个字段时，反序列化是就会有一个默认的零值，
下面是Go语言中常见类型的零值：
Go 字段类型	JSON 中省略该字段 → 反序列化后的值
string	""
int / float64	0
bool	false
struct（值类型）	结构体内部所有字段均为各自的零值
[]T（切片）	nil
map[K]V	nil
*T（指针）	nil
*/
// Expectation 表示某个探测目标的预期结果。
type Expectation struct {
	StatusCode *int   `json:"status_code,omitempty"` //omitempty表示如果是nil或0值则不输出，这里设置为指针的原因是为了区分用户是否设置了这个字段，如果是0值但用户没有设置，我们就认为没有设置这个字段。
	Contains   string `json:"contains,omitempty"`
	Connected  *bool  `json:"connected,omitempty"` //omitempty表示如果是nil或0值则不输出，这里设置为指针的原因是为了区分用户是否设置了这个字段，如果是0值但用户没有设置，我们就认为没有设置这个字段。
}

// Target 表示一个探测目标。
type Target struct {
	Name       string      `json:"name"`                  // 目标名称
	Protocol   string      `json:"protocol"`              // 协议，tcp/http
	Address    string      `json:"address"`               // 地址，例如 tcp://localhost:8080, http://localhost:8080
	Expect     Expectation `json:"expect"`                // 预期结果
	RetryCount int         `json:"retry_count,omitempty"` // 重试次数，默认3次
	Index      int         `json:"-"`                     // 配置文件中的索引，这里的json:"-"表示这个字段不会被序列化或反序列化，因为它是我们在程序中动态设置的，不需要从配置文件中读取或写入。
}

// Options 表示命令行参数解析后的结果。
type Options struct {
	ConfigPath string
	Timeout    time.Duration
	Verbose    bool
	Proxy      string
}

// ProbeResult 表示单个目标的最终探测结果。
type ProbeResult struct {
	Index    int
	Name     string
	Protocol string
	Address  string
	Success  bool
	Expected string
	Observed string
	Error    string
	Latency  time.Duration
	Attempts int
}

// Summary 表示最终汇总统计。
type Summary struct {
	Total          int
	SuccessCount   int
	FailureCount   int
	SuccessRate    float64
	FailureRate    float64
	AverageLatency time.Duration
	LatencyBuckets map[string]int
	Slowest        []ProbeResult
}

// expectedDescription 返回目标的预期结果描述。
func (t Target) expectedDescription() string {
	switch t.Protocol {
	case "tcp":
		return "Connected"
	case "http":
		parts := make([]string, 0, 2)
		if t.Expect.StatusCode != nil {
			parts = append(parts, fmt.Sprintf("%d %s", *t.Expect.StatusCode, http.StatusText(*t.Expect.StatusCode)))
		} else {
			parts = append(parts, "2xx status")
		}
		if t.Expect.Contains != "" {
			parts = append(parts, fmt.Sprintf("Contains %q", t.Expect.Contains))
		}
		return strings.Join(parts, " + ")
	default:
		return "N/A"
	}
}

func (r ProbeResult) statusLabel() string {
	if r.Success {
		return "OK"
	}
	return "FAIL"
}

package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"
)

func buildSummary(results []ProbeResult) Summary {
	summary := Summary{
		Total: len(results),
		LatencyBuckets: map[string]int{
			"<100ms":      0,
			"100ms~500ms": 0,
			"500ms~1s":    0,
			">1s":         0,
		},
	}

	var totalLatency time.Duration

	for _, result := range results {
		if result.Success {
			summary.SuccessCount++
		} else {
			summary.FailureCount++
		}

		totalLatency += result.Latency
		summary.LatencyBuckets[classifyLatency(result.Latency)]++
	}

	if summary.Total > 0 {
		summary.SuccessRate = float64(summary.SuccessCount) / float64(summary.Total) * 100
		summary.FailureRate = float64(summary.FailureCount) / float64(summary.Total) * 100
		summary.AverageLatency = time.Duration(int64(totalLatency) / int64(summary.Total))
	}

	slowest := append([]ProbeResult(nil), results...)
	sort.Slice(slowest, func(i, j int) bool {
		return slowest[i].Latency > slowest[j].Latency
	})
	if len(slowest) > 3 {
		slowest = slowest[:3]
	}
	summary.Slowest = slowest

	return summary
}

func classifyLatency(latency time.Duration) string {
	switch {
	case latency < 100*time.Millisecond:
		return "<100ms"
	case latency < 500*time.Millisecond:
		return "100ms~500ms"
	case latency < time.Second:
		return "500ms~1s"
	default:
		return ">1s"
	}
}

func renderReport(results []ProbeResult, summary Summary, options Options, generatedAt time.Time) string {
	var builder strings.Builder

	fmt.Fprintln(&builder, "========== 服务健康探测报告 ==========")
	fmt.Fprintf(&builder, "生成时间：%s\n", generatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&builder, "配置文件：%s\n", options.ConfigPath)
	fmt.Fprintf(&builder, "单次探测超时：%v\n", options.Timeout)
	fmt.Fprintf(&builder, "总目标数：%d\n", summary.Total)
	fmt.Fprintf(&builder, "成功数量：%d\n", summary.SuccessCount)
	fmt.Fprintf(&builder, "失败数量：%d\n", summary.FailureCount)
	fmt.Fprintf(&builder, "成功率：%.2f%%\n", summary.SuccessRate)
	fmt.Fprintf(&builder, "失败率：%.2f%%\n", summary.FailureRate)
	fmt.Fprintf(&builder, "平均耗时：%s\n", summary.AverageLatency.Truncate(time.Millisecond))

	fmt.Fprintln(&builder, "\n响应时间分布：")
	buckets := []string{"<100ms", "100ms~500ms", "500ms~1s", ">1s"}
	for _, bucket := range buckets {
		fmt.Fprintf(&builder, "- %-12s %d\n", bucket+":", summary.LatencyBuckets[bucket])
	}

	fmt.Fprintln(&builder, "\n最慢服务 TOP 3：")
	if len(summary.Slowest) == 0 {
		fmt.Fprintln(&builder, "- 无数据")
	} else {
		for i, result := range summary.Slowest {
			fmt.Fprintf(
				&builder,
				"%d. %s (%s, %s) - 耗时=%s - 实际结果=%s\n",
				i+1,
				result.Name,
				strings.ToUpper(result.Protocol),
				result.Address,
				result.Latency.Truncate(time.Millisecond),
				sanitizeText(result.Observed),
			)
		}
	}

	fmt.Fprintln(&builder, "\n详细结果：")
	tw := tabwriter.NewWriter(&builder, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "序号\t服务名\t协议\t地址\t结果\t预期\t实际\t耗时\t尝试次数\t错误")
	for i, result := range results {
		errorText := sanitizeText(result.Error)
		if errorText == "" {
			errorText = "-"
		}
		fmt.Fprintf(
			tw,
			"%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			i+1,
			result.Name,
			strings.ToUpper(result.Protocol),
			result.Address,
			result.statusLabel(),
			sanitizeText(result.Expected),
			sanitizeText(result.Observed),
			result.Latency.Truncate(time.Millisecond),
			result.Attempts,
			errorText,
		)
	}
	tw.Flush()

	return builder.String()
}

func sanitizeText(input string) string {
	replacer := strings.NewReplacer("\t", " ", "\n", " ", "\r", " ")
	return replacer.Replace(input)
}

func writeReportFile(content string, generatedAt time.Time) (string, error) {
	fileName := fmt.Sprintf("monitor-log-%s.log", generatedAt.Format("20060102150405"))
	if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入报告文件失败: %w", err)
	}
	return fileName, nil
}

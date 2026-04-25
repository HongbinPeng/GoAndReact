package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
	"unicode"
)

/*
这个结构体用于我计算格式化输出报告中proberesult的每列最大宽度,
解决中文导致的每列宽度不一致问题
*/
type ColumnWidth struct {
	Index    int
	Name     int
	Protocol int
	Address  int
	Result   int
	Expected int
	Observed int
	Latency  int
	Attempts int
	Error    int
}

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
	/*
		这里使用string.Builder来构建报告字符串，而不是直接拼接字符串。
		因为报告的内容太多，选择拼接字符串会导致内存分配和复制操作，影响性能。
		使用string.Builder可以避免这个问题，同时提供更方便的API来构建字符串。
		go中report += "xxx"都不是在原地修改，而是：
			1、创建一个新字符串
			2、把旧内容复制过去
			3、再把新内容接上
		同时这里使用
	*/
	fmt.Fprintln(&builder, "========== 服务健康探测报告 ==========")
	fmt.Fprintf(&builder, "%-12s%20s\n", "生成时间：", generatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&builder, "%-12s%20s\n", "配置文件：", options.ConfigPath)
	fmt.Fprintf(&builder, "%-12s%20s\n", "单次探测超时：", options.Timeout.String())
	fmt.Fprintf(&builder, "%-12s%20d\n", "总目标数：", summary.Total)
	fmt.Fprintf(&builder, "%-12s%20d\n", "成功数量：", summary.SuccessCount)
	fmt.Fprintf(&builder, "%-12s%20d\n", "失败数量：", summary.FailureCount)
	fmt.Fprintf(&builder, "%-12s%20.2f%%\n", "成功率：", summary.SuccessRate)
	fmt.Fprintf(&builder, "%-12s%20.2f%%\n", "失败率：", summary.FailureRate)
	fmt.Fprintf(&builder, "%-12s%20s\n", "平均耗时：", summary.AverageLatency.Truncate(time.Millisecond))

	fmt.Fprintln(&builder, "\n响应时间分布：")
	buckets := []string{"<100ms", "100ms~500ms", "500ms~1s", ">1s"}
	/*
		这里使用%-12s来格式化桶的名称输出，他的规则是：
			%-12s 表示左对齐，宽度为12个字符
			如果字符串长度 小于 12，右边补空格到 12 宽
			如果字符串长度 等于 12，刚好输出
			如果字符串长度 大于 12，就直接完整输出，不会裁剪
		fmt.Printf("|%-12.12s|\n", "ThisIsALongString")这个意思是：
			最多取 12 个字符
			宽度至少 12
			左对齐
	*/
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
	/*
		这里有问题，我查询AI得知，/t以及右对齐，左对齐前面的宽度都是基于runa的个数进行计算的，
		但是在实际的终端或者文本编辑器中，中文字符通常占用两个字符宽度，而英文字符占用一个字符宽度。
		还有全角冒号、特殊Unicode字符等都不一定只占用一个字符宽度，这就导致了使用tabwriter时，表格对齐可能会出现问题。
		因此在这里我计划重新编写一个函数用于输出这个表格。
	*/
	widths := findMaxWidth(results) //首先计算探测结果中每列的最大宽度，方便后续格式化输出
	fmt.Fprintf(&builder, "%s%s%s%s%s%s%s%s%s%s\n",
		modifyStringToWidth("序号", widths.Index),
		modifyStringToWidth("服务名", widths.Name),
		modifyStringToWidth("协议", widths.Protocol),
		modifyStringToWidth("地址", widths.Address),
		modifyStringToWidth("结果", widths.Result),
		modifyStringToWidth("预期", widths.Expected),
		modifyStringToWidth("实际", widths.Observed),
		modifyStringToWidth("耗时", widths.Latency),
		modifyStringToWidth("尝试次数", widths.Attempts),
		modifyStringToWidth("错误", widths.Error))
	// "序号\t服务名\t协议\t地址\t结果\t预期\t实际\t耗时\t尝试次数\t错误"
	for i, result := range results {
		fmt.Fprintf(&builder, "%s%s%s%s%s%s%s%s%s%s\n",
			modifyStringToWidth(fmt.Sprintf("%d", i+1), widths.Index),
			modifyStringToWidth(result.Name, widths.Name),
			modifyStringToWidth(result.Protocol, widths.Protocol),
			modifyStringToWidth(result.Address, widths.Address),
			modifyStringToWidth(result.statusLabel(), widths.Result),
			modifyStringToWidth(sanitizeText(result.Expected), widths.Expected),
			modifyStringToWidth(sanitizeText(result.Observed), widths.Observed),
			modifyStringToWidth(result.Latency.Truncate(time.Millisecond).String(), widths.Latency),
			modifyStringToWidth(fmt.Sprintf("%d", result.Attempts), widths.Attempts),
			modifyStringToWidth(sanitizeText(result.Error), widths.Error),
		)
	}
	return builder.String()
}

func sanitizeText(input string) string {
	replacer := strings.NewReplacer("\t", " ", "\n", " ", "\r", " ") //这里将特殊字符替换为空格
	return replacer.Replace(input)
}
func writeReportFile(content string, generatedAt time.Time) (string, error) {
	fileName := fmt.Sprintf("monitor-log-%s.log", generatedAt.Format("20060102150405"))
	if err := os.WriteFile(fileName, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入报告文件失败: %w", err)
	}
	return fileName, nil
}
func findMaxWidth(results []ProbeResult) ColumnWidth {
	widths := ColumnWidth{
		Index:    calculateWidth("序号") + 3,
		Name:     calculateWidth("服务名") + 3,
		Protocol: calculateWidth("协议") + 3,
		Address:  calculateWidth("地址") + 3,
		Result:   calculateWidth("结果") + 3,
		Expected: calculateWidth("预期") + 3,
		Observed: calculateWidth("实际") + 3,
		Latency:  calculateWidth("耗时") + 3,
		Attempts: calculateWidth("尝试次数") + 3,
		Error:    calculateWidth("错误") + 3,
	}
	for i, result := range results {
		widths.Index = max(widths.Index, calculateWidth(fmt.Sprintf("%d", i+1))+3)
		widths.Name = max(widths.Name, calculateWidth(result.Name)+3)
		widths.Protocol = max(widths.Protocol, calculateWidth(result.Protocol)+3)
		widths.Address = max(widths.Address, calculateWidth(result.Address)+3)
		widths.Result = max(widths.Result, calculateWidth(result.statusLabel())+3)
		widths.Expected = max(widths.Expected, calculateWidth(sanitizeText(result.Expected))+3)
		widths.Observed = max(widths.Observed, calculateWidth(sanitizeText(result.Observed))+3)
		widths.Latency = max(widths.Latency, calculateWidth(result.Latency.Truncate(time.Millisecond).String())+3)
		widths.Attempts = max(widths.Attempts, calculateWidth(fmt.Sprintf("%d", result.Attempts))+3)
		errorText := sanitizeText(result.Error)
		if errorText == "" {
			errorText = "-"
		}
		widths.Error = max(widths.Error, calculateWidth(errorText))
	}

	return widths
}
func modifyStringToWidth(s string, maxWidth int) string {
	var str strings.Builder
	if calculateWidth(s) <= maxWidth {
		str.WriteString(s)
		str.WriteString(strings.Repeat(" ", maxWidth-calculateWidth(s)))
		return str.String()
	}
	return s
}
func calculateWidth(s string) int {
	width := 0
	for _, r := range s {
		if isWideRuna(r) {
			width += 2
		} else {
			width += 1
		}
	}
	return width
}
func isWideRuna(r rune) bool { //在这里我只筛选了中文
	return unicode.In(r,
		unicode.Han, //中文码点
	)
}

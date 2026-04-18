package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	BillingRuleVersion = "1.0.0"

	// 阶梯分界点
	Tier1Limit float64 = 200
	Tier2Limit float64 = 400

	// 各阶梯单价
	Tier1Price float64 = 0.5
	Tier2Price float64 = 0.8
	Tier3Price float64 = 1.2

	// 峰时段范围：[8:00, 22:00)
	PeakStartHour = 8
	PeakEndHour   = 22

	PeakTimeFactor   float64 = 1.10
	ValleyTimeFactor float64 = 0.80
)

var systemInitializedAt string

func init() {
	systemInitializedAt = time.Now().Format("2006-01-02 15:04:05")

	fmt.Printf("计费规则版本号：%s\n", BillingRuleVersion)
	fmt.Printf("系统初始化时间：%s\n", systemInitializedAt)
}

// 读取用电量和用电时段。
func readUserInput() (float64, string, error) {
	var usage float64
	fmt.Print("请输入用电量：")
	if _, err := fmt.Scanln(&usage); err != nil {
		return 0, "", fmt.Errorf("读取用电量失败：%w", err)
	}

	var timeText string
	fmt.Print("请输入用电时段（格式 HH:MM）：")
	if _, err := fmt.Scanln(&timeText); err != nil {
		return 0, "", fmt.Errorf("读取用电时段失败：%w", err)
	}

	return usage, strings.TrimSpace(timeText), nil
}

func validateUsage(usage float64) error {
	if usage < 0 {
		return errors.New("用电量不能为负数")
	}

	return nil
}

func validateTime(hour, minute int) error {
	if hour < 0 || hour > 23 {
		return errors.New("小时必须在 0 到 23 之间")
	}

	if minute < 0 || minute > 59 {
		return errors.New("分钟必须在 0 到 59 之间")
	}

	return nil
}

// 负责把 HH:MM 形式的字符串解析为小时和分钟。
func ParseTime(timeText string) (int, int, error) {
	timeText = strings.TrimSpace(timeText)
	parts := strings.Split(timeText, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return 0, 0, errors.New("用电时段格式应为 HH:MM")
	}

	hour, hourErr := strconv.Atoi(parts[0])
	if hourErr != nil {
		return 0, 0, errors.New("小时必须是数字")
	}

	minute, minuteErr := strconv.Atoi(parts[1])
	if minuteErr != nil {
		return 0, 0, errors.New("分钟必须是数字")
	}

	if err := validateTime(hour, minute); err != nil {
		return 0, 0, err
	}

	return hour, minute, nil
}

// 只计算阶梯电价本身，不包含峰谷倍率。
func calculateTierCost(usage float64) float64 {
	switch {
	case usage <= Tier1Limit:
		return usage * Tier1Price
	case usage <= Tier2Limit:
		return Tier1Limit*Tier1Price + (usage-Tier1Limit)*Tier2Price
	default:
		return Tier1Limit*Tier1Price +
			(Tier2Limit-Tier1Limit)*Tier2Price +
			(usage-Tier2Limit)*Tier3Price
	}
}

func isPeakTime(hour int) bool {
	return hour > PeakStartHour && hour <= PeakEndHour // 注意：这里的峰时段定义为 (8:00, 22:00]，即不包含 8:00，但包含 22:00。
}

// 根据时段对基础电费做峰谷调整。
func applyTimeFactor(baseCost float64, hour int) float64 {
	if isPeakTime(hour) {
		return baseCost * PeakTimeFactor
	}

	return baseCost * ValleyTimeFactor
}

func CalculateBill(usage float64, hour, minute int) (float64, error) {
	if err := validateUsage(usage); err != nil {
		return 0, err
	}

	if err := validateTime(hour, minute); err != nil {
		return 0, err
	}

	baseCost := calculateTierCost(usage)
	return applyTimeFactor(baseCost, hour), nil
}

// 统一生成账单输出内容，方便主流程复用和测试。
func formatBill(usage float64, hour, minute int, totalCost float64) string {
	return fmt.Sprintf(
		"--- 账单明细 ---\n当前用电：%.2f 度\n当前时段：%02d:%02d 点\n最终电费：%.2f 元\n",
		usage,
		hour,
		minute,
		totalCost,
	)
}

func main() {
	usage, timeText, err := readUserInput()
	if err != nil {
		fmt.Println(err)
		return
	}

	hour, minute, err := ParseTime(timeText)
	if err != nil {
		fmt.Println(err)
		return
	}

	totalCost, err := CalculateBill(usage, hour, minute)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print(formatBill(usage, hour, minute, totalCost))
}

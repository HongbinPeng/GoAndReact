package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	Tier1Max  float64 = 200 // 第一阶：0-200 度
	Tier1Rate float64 = 0.5 // 第一阶单价

	Tier2Max  float64 = 400 // 第二阶：200-400 度
	Tier2Rate float64 = 0.8 // 第二阶单价

	Tier3Rate float64 = 1.2 // 第三阶：400 以上，单价 1.2
)

type TimeRateConfig struct {
	Version                   string
	PeekTimeRate, LowTimeRate float64
	PeekStart, PeekEnd        int
}
type BillData struct {
	Usage        float64
	hour, minute int
}
type BillResult struct {
	Usage     float64
	TimeStr   string
	TotalCost float64
}
type Printer interface {
	Print()
}

func (billdata *BillData) Print() {
	fmt.Printf("用电量：%.2f 度\n用电时间：%02d:%02d\n", billdata.Usage, billdata.hour, billdata.minute)
}
func (billresult *BillResult) Print() {
	fmt.Printf(`---账单明细---
当前用电：%.2f 度
当前时段：%s
最终电费：%.2f 元`, billresult.Usage, billresult.TimeStr, billresult.TotalCost)
}

var timeRateConfig TimeRateConfig

func init() {
	timeRateConfig = TimeRateConfig{
		PeekTimeRate: 1.1,
		LowTimeRate:  0.8,
		PeekStart:    8,
		PeekEnd:      22,
		Version:      "1.0.0",
	}
	fmt.Printf("计费规则版本号:%s\n系统初始化时间：%s\n", timeRateConfig.Version, time.Now().Format("2006-01-02 15:04:05"))
}
func UserDirector() (BillData, error) {
	fmt.Println("请输入用电量：")
	var usage float64
	fmt.Scanln(&usage)
	fmt.Println("请输入用电时间（24小时制）：")
	var timeStr string
	fmt.Scanln(&timeStr)
	timeStr = strings.Trim(timeStr, " ")
	hourAndMinute := strings.Split(timeStr, ":")
	if len(hourAndMinute) != 2 {
		return BillData{}, fmt.Errorf("输入的时间格式不正确，请重新输入")
	}
	hour, minutes := hourAndMinute[0], hourAndMinute[1]
	hourInt, herr := strconv.Atoi(hour)
	minuteInt, merr := strconv.Atoi(minutes)
	if herr != nil || merr != nil || hourInt < 0 || hourInt > 23 || minuteInt < 0 || minuteInt > 59 {
		return BillData{}, fmt.Errorf("输入的时间格式不正确，请重新输入")
	}
	billData := BillData{
		Usage:  usage,
		hour:   hourInt,
		minute: minuteInt,
	}
	return billData, nil
}
func CalculateBill(billData BillData) BillResult {
	var totalCost float64
	if billData.Usage <= Tier1Max {
		totalCost = billData.Usage * Tier1Rate
	} else if billData.Usage <= Tier2Max {
		totalCost = 200*Tier1Rate + (billData.Usage-Tier1Max)*Tier2Rate
	} else {
		totalCost = 200*Tier1Rate + 200*Tier2Rate + (billData.Usage-Tier2Max)*Tier3Rate
	}
	if billData.hour >= timeRateConfig.PeekStart && billData.hour < timeRateConfig.PeekEnd {
		totalCost *= timeRateConfig.PeekTimeRate
	} else {
		totalCost *= timeRateConfig.LowTimeRate
	}
	billResult := BillResult{
		Usage:     billData.Usage,
		TimeStr:   fmt.Sprintf("%02d:%02d", billData.hour, billData.minute),
		TotalCost: totalCost,
	}
	return billResult
}
func main() {
	billdata, error := UserDirector()
	if error != nil {
		fmt.Println(error)
		return
	} else {
		billresult := CalculateBill(billdata)
		billresult.Print()
	}
}

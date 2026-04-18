package main

import (
	"math"
	"testing"
)

const floatTolerance = 1e-9

func almostEqual(got, want float64) bool {
	return math.Abs(got-want) <= floatTolerance
}

// 采用表驱动测试，集中覆盖时间解析的正常与异常场景。
func TestParseTime(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		wantHour   int
		wantMinute int
		wantErr    bool
	}{
		{
			name:       "正常时间",
			input:      "14:00",
			wantHour:   14,
			wantMinute: 0,
		},
		{
			name:       "带空格的时间",
			input:      " 08:30 ",
			wantHour:   8,
			wantMinute: 30,
		},
		{
			name:    "缺少冒号",
			input:   "1400",
			wantErr: true,
		},
		{
			name:    "小时超出范围",
			input:   "24:00",
			wantErr: true,
		},
		{
			name:    "分钟超出范围",
			input:   "09:60",
			wantErr: true,
		},
		{
			name:    "时间包含非数字字符",
			input:   "ab:cd",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotHour, gotMinute, err := ParseTime(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseTime(%q) 期望返回错误，但实际为 nil", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseTime(%q) 返回了非预期错误: %v", tc.input, err)
			}

			if gotHour != tc.wantHour || gotMinute != tc.wantMinute {
				t.Fatalf(
					"ParseTime(%q) = (%d, %d), 期望 (%d, %d)",
					tc.input,
					gotHour,
					gotMinute,
					tc.wantHour,
					tc.wantMinute,
				)
			}
		})
	}
}

// 重点验证三档阶梯电价在边界值附近的计算是否正确。
func TestCalculateTierCost(t *testing.T) {
	testCases := []struct {
		name  string
		usage float64
		want  float64
	}{
		{
			name:  "零用电量",
			usage: 0,
			want:  0,
		},
		{
			name:  "第一档上边界",
			usage: 200,
			want:  100,
		},
		{
			name:  "第二档中间值",
			usage: 350,
			want:  220,
		},
		{
			name:  "第二档上边界",
			usage: 400,
			want:  260,
		},
		{
			name:  "第三档用电量",
			usage: 500,
			want:  380,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := calculateTierCost(tc.usage)
			if !almostEqual(got, tc.want) {
				t.Fatalf("calculateTierCost(%.2f) = %.2f, 期望 %.2f", tc.usage, got, tc.want)
			}
		})
	}
}

// 同时覆盖峰谷边界和非法输入，确认最终电费计算可靠。
func TestCalculateBill(t *testing.T) {
	testCases := []struct {
		name    string
		usage   float64
		hour    int
		minute  int
		want    float64
		wantErr bool
	}{
		{
			name:   "早上八点前属于谷时段",
			usage:  100,
			hour:   7,
			minute: 59,
			want:   40,
		},
		{
			name:   "早上八点开始属于谷时段",
			usage:  100,
			hour:   8,
			minute: 0,
			want:   40,
		},
		{
			name:   "晚上十点前仍属于峰时段",
			usage:  400,
			hour:   21,
			minute: 59,
			want:   286,
		},
		{
			name:   "晚上十点开始属于峰时段",
			usage:  450,
			hour:   22,
			minute: 0,
			want:   352,
		},
		{
			name:    "负数用电量",
			usage:   -1,
			hour:    10,
			minute:  0,
			wantErr: true,
		},
		{
			name:    "小时非法",
			usage:   100,
			hour:    24,
			minute:  0,
			wantErr: true,
		},
		{
			name:    "分钟非法",
			usage:   100,
			hour:    10,
			minute:  60,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CalculateBill(tc.usage, tc.hour, tc.minute)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("CalculateBill 期望返回错误，但实际为 nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CalculateBill 返回了非预期错误: %v", err)
			}

			if !almostEqual(got, tc.want) {
				t.Fatalf(
					"CalculateBill(%.2f, %d, %d) = %.2f, 期望 %.2f",
					tc.usage,
					tc.hour,
					tc.minute,
					got,
					tc.want,
				)
			}
		})
	}
}

func TestFormatBill(t *testing.T) {
	got := formatBill(400, 14, 0, 286)
	want := "--- 账单明细 ---\n当前用电：400.00 度\n当前时段：14:00 点\n最终电费：286.00 元\n"

	if got != want {
		t.Fatalf("formatBill() = %q, 期望 %q", got, want)
	}
}

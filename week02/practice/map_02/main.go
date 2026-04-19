package main

import "fmt"

func countMostFrequent(s string) rune {
	counts := make(map[rune]int)
	// 统计每个字符出现的次数
	for _, ch := range s {
		counts[ch]++
	}
	// 找出出现次数最多的字符
	var maxCh rune
	maxCount := 0
	for ch, count := range counts {
		if count > maxCount {
			maxCount = count
			maxCh = ch
		}
	}
	return maxCh
}
func main() {
	s := "abcabccccc1231111"
	mostFrequent := countMostFrequent(s)
	fmt.Printf("出现次数最多的字符是: %c\n", mostFrequent)
}

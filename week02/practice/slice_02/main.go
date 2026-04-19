package main

import "fmt"

func main() {
	var slice1 []int = []int{1, 2, 3, 4}
	var slice2 []int = []int{3, 4, 5, 6}
	combinedSlice := append(slice1, slice2...)
	uniqueSlice := make([]int, 0, len(combinedSlice))
	seen := make(map[int]bool)
	for _, num := range combinedSlice {
		if _, ok := seen[num]; !ok {
			uniqueSlice = append(uniqueSlice, num)
			seen[num] = true
		}
	}
	fmt.Println("Combined Slice:", combinedSlice)
	fmt.Println("Unique Slice:", uniqueSlice)
}

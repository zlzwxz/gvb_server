package testdata

import (
	"fmt"
	"testing"
)

func TestFanZhuan(t *testing.T) {
	// 使用示例
	nums := []int{10, 20, 30}
	fmt.Println(nums)
	Reverse(nums) // 原地反转
	fmt.Println(nums)
}

func Reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

package utils

import (
	"strings"
)

// SplitTrim 分割字符串并去掉空格
func SplitTrim(input, sep string) []string {
	slc := strings.Split(input, sep)
	for i := range slc {
		slc[i] = strings.TrimSpace(slc[i])
	}
	return slc
}

// SplitToSlice 分割字符串，去掉空格并遍历
func SplitToSlice[T any](input, sep string, fn func(item string) T) []T {
	slc := strings.Split(input, sep)
	var res []T
	for i := range slc {
		res = append(res, fn(strings.TrimSpace(slc[i])))
	}
	return res
}

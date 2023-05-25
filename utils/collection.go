package utils

import (
	"fmt"
	"strings"
)

func ContainsString(collection []string, element string) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}

func ContainsInt64(collection []int64, element int64) bool {
	for _, item := range collection {
		if item == element {
			return true
		}
	}
	return false
}

func JoinInt64(arr []int64, sep string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(arr)), sep), "[]")
}

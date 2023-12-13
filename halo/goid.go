package halo

import "github.com/petermattis/goid"

// GetGoID 返回协程ID
func GetGoID() int64 {
	return goid.Get()
}

package utils

import (
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomInt64(min int64, max int64) int64 {
	return r.Int63n(max-min+1) + min
}

func RandomInt(min int, max int) int {
	return r.Intn(max-min+1) + min
}

func RandomInt32(min int32, max int32) int32 {
	return r.Int31n(max-min+1) + min
}

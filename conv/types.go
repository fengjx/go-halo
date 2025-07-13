package conv

import (
	"reflect"
	"time"
)

// enumerates all types
var (
	intType   = reflect.TypeOf((*int)(nil)).Elem()
	int8Type  = reflect.TypeOf((*int8)(nil)).Elem()
	int16Type = reflect.TypeOf((*int16)(nil)).Elem()
	int32Type = reflect.TypeOf((*int32)(nil)).Elem()
	int64Type = reflect.TypeOf((*int64)(nil)).Elem()

	uintType   = reflect.TypeOf((*uint)(nil)).Elem()
	uint8Type  = reflect.TypeOf((*uint8)(nil)).Elem()
	uint16Type = reflect.TypeOf((*uint16)(nil)).Elem()
	uint32Type = reflect.TypeOf((*uint32)(nil)).Elem()
	uint64Type = reflect.TypeOf((*uint64)(nil)).Elem()

	float32Type = reflect.TypeOf((*float32)(nil)).Elem()
	float64Type = reflect.TypeOf((*float64)(nil)).Elem()

	complex64Type  = reflect.TypeOf((*complex64)(nil)).Elem()
	complex128Type = reflect.TypeOf((*complex128)(nil)).Elem()

	stringType = reflect.TypeOf((*string)(nil)).Elem()
	boolType   = reflect.TypeOf((*bool)(nil)).Elem()
	byteType   = reflect.TypeOf((*byte)(nil)).Elem()
	bytesType  = reflect.SliceOf(byteType)

	timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
)

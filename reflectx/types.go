package reflectx

import (
	"reflect"
	"time"
)

// 预定义的类型常量
var (
	IntType        = reflect.TypeOf((*int)(nil)).Elem()
	Int8Type       = reflect.TypeOf((*int8)(nil)).Elem()
	Int16Type      = reflect.TypeOf((*int16)(nil)).Elem()
	Int32Type      = reflect.TypeOf((*int32)(nil)).Elem()
	Int64Type      = reflect.TypeOf((*int64)(nil)).Elem()
	UintType       = reflect.TypeOf((*uint)(nil)).Elem()
	Uint8Type      = reflect.TypeOf((*uint8)(nil)).Elem()
	Uint16Type     = reflect.TypeOf((*uint16)(nil)).Elem()
	Uint32Type     = reflect.TypeOf((*uint32)(nil)).Elem()
	Uint64Type     = reflect.TypeOf((*uint64)(nil)).Elem()
	Float32Type    = reflect.TypeOf((*float32)(nil)).Elem()
	Float64Type    = reflect.TypeOf((*float64)(nil)).Elem()
	Complex64Type  = reflect.TypeOf((*complex64)(nil)).Elem()
	Complex128Type = reflect.TypeOf((*complex128)(nil)).Elem()
	StringType     = reflect.TypeOf((*string)(nil)).Elem()
	BoolType       = reflect.TypeOf((*bool)(nil)).Elem()
	ByteType       = reflect.TypeOf((*byte)(nil)).Elem()
	BytesType      = reflect.SliceOf(ByteType)
	TimeType       = reflect.TypeOf((*time.Time)(nil)).Elem()
)

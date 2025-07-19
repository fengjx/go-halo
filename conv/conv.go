package conv

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/fengjx/go-halo/reflectx"
)

var (
	mu sync.Mutex
	// registry converter
	registry = make(map[tKey]any)

	// mapperCache 用于缓存不同 tagName 的 reflectx.Mapper
	mapperCache sync.Map // map[string]*reflectx.Mapper

	// 预定义的类型常量
	intType        = reflect.TypeOf((*int)(nil)).Elem()
	int8Type       = reflect.TypeOf((*int8)(nil)).Elem()
	int16Type      = reflect.TypeOf((*int16)(nil)).Elem()
	int32Type      = reflect.TypeOf((*int32)(nil)).Elem()
	int64Type      = reflect.TypeOf((*int64)(nil)).Elem()
	uintType       = reflect.TypeOf((*uint)(nil)).Elem()
	uint8Type      = reflect.TypeOf((*uint8)(nil)).Elem()
	uint16Type     = reflect.TypeOf((*uint16)(nil)).Elem()
	uint32Type     = reflect.TypeOf((*uint32)(nil)).Elem()
	uint64Type     = reflect.TypeOf((*uint64)(nil)).Elem()
	float32Type    = reflect.TypeOf((*float32)(nil)).Elem()
	float64Type    = reflect.TypeOf((*float64)(nil)).Elem()
	complex64Type  = reflect.TypeOf((*complex64)(nil)).Elem()
	complex128Type = reflect.TypeOf((*complex128)(nil)).Elem()
	stringType     = reflect.TypeOf((*string)(nil)).Elem()
	boolType       = reflect.TypeOf((*bool)(nil)).Elem()
	byteType       = reflect.TypeOf((*byte)(nil)).Elem()
	bytesType      = reflect.SliceOf(byteType)
	timeType       = reflect.TypeOf((*time.Time)(nil)).Elem()
)

var (
	ErrMapKeyNotMatch = errors.New("map key type not match")
	ErrNotSupportType = errors.New("not support type, src and dst must be struct or ptr to struct")
)

const (
	Second      TimeUnit = "s"
	Millisecond TimeUnit = "ms"
	Microsecond TimeUnit = "us"
	Nanosecond  TimeUnit = "ns"
)

type TimeUnit string

// Option 用于自定义转换行为的函数类型
type Option func(*options)

type options struct {
	tagName  string   // 结构体字段映射时使用的 tag 名称
	timeUnit TimeUnit // 时间戳单位，ms/us/ns/s，默认s
}

// WithTag 设置结构体 tag 名称（如 "json"、"db" 等）
func WithTag(tag string) Option {
	return func(o *options) {
		o.tagName = tag
	}
}

// WithTimeUnit 设置时间戳单位
func WithTimeUnit(unit TimeUnit) Option {
	return func(o *options) {
		o.timeUnit = unit
	}
}

// Result 封装类型转换结果和错误，支持链式调用
type Result[SRC, DIST any] struct {
	value DIST
	err   error
}

// Val 返回转换结果
func (r Result[SRC, DIST]) Val() DIST {
	return r.value
}

// Error 返回转换错误
func (r Result[SRC, DIST]) Error() error {
	return r.err
}

// OnError 链式处理错误
func (r Result[SRC, DIST]) OnError(f func(error)) Result[SRC, DIST] {
	if r.err != nil {
		f(r.err)
	}
	return r
}

// Then 链式处理成功结果
func (r Result[SRC, DIST]) Then(f func(DIST)) Result[SRC, DIST] {
	if r.err == nil {
		f(r.value)
	}
	return r
}

// Result 展开 Result 对象
func (r Result[SRC, DIST]) Result() (DIST, error) {
	return r.value, r.err
}

// Convert 执行类型转换，支持 Option
func Convert[SRC, DIST any](from SRC, opts ...Option) Result[SRC, DIST] {
	var zero DIST
	if isNil(any(from)) {
		return Result[SRC, DIST]{value: zero, err: nil}
	}

	opt := &options{
		tagName:  "json",
		timeUnit: "s",
	}
	for _, o := range opts {
		o(opt)
	}

	var s SRC
	var d DIST
	srcType := reflect.TypeOf(s)
	dstType := reflect.TypeOf(d)

	fn := getConverter(srcType, dstType, opt)
	var to DIST
	err := fn(from, &to)
	return Result[SRC, DIST]{value: to, err: err}
}

type converter func(src, dst any) error

func getConverter(srcType, dstType reflect.Type, opt *options) converter {
	// 获取实际的类型（去除指针）
	srcType, _ = indirectType(srcType)
	dstType, _ = indirectType(dstType)

	if srcType.Kind() != reflect.Struct {
		return notSupportConvert
	}
	if dstType.Kind() != reflect.Struct {
		return notSupportConvert
	}

	key := typeKey(srcType, dstType, opt.tagName)
	fn, ok := registry[key]
	if ok {
		return fn.(converter)
	}

	mu.Lock()
	// double check
	fn, ok = registry[key]
	if ok {
		mu.Unlock()
		return fn.(converter)
	}

	fn = genConverter(srcType, dstType, opt)
	registry[key] = fn
	mu.Unlock()
	return fn.(converter)
}

// genConverter 支持递归结构体赋值
func genConverter(srcType, dstType reflect.Type, opt *options) converter {
	mapper := getCachedMapper(opt.tagName)
	dstFields := mapper.TypeMap(dstType)
	srcFields := mapper.TypeMap(srcType)

	setValueFuncMap := make(map[string]setValueFunc)

	for name, dstField := range dstFields.Names {
		srcField, ok := srcFields.Names[name]
		if !ok {
			continue
		}

		srcFieldType := srcField.Field.Type
		dstFieldType := dstField.Field.Type

		// 直接赋值（类型可赋值）
		if srcFieldType.AssignableTo(dstFieldType) {
			setValueFuncMap[name] = setAssignableTo
			continue
		}

		if srcFieldType.ConvertibleTo(dstFieldType) {
			setValueFuncMap[name] = setConvertibleTo
			continue
		}

		// 时间戳转换
		if srcFieldType == int64Type && dstFieldType == timeType {
			setValueFuncMap[name] = setInt64ToTime
			continue
		}

		if srcFieldType == timeType && dstFieldType == int64Type {
			setValueFuncMap[name] = setTimeToInt64
			continue
		}

		// 递归处理结构体
		if srcFieldType.Kind() == reflect.Struct && dstFieldType.Kind() == reflect.Struct {
			subConv := genConverter(srcFieldType, dstFieldType, opt)
			setValueFuncMap[name] = func(dst, src reflect.Value, opt *options) {
				_ = subConv(src.Addr().Interface(), dst.Addr().Interface())
			}
			continue
		}
	}

	return func(from any, to any) error {
		// 处理源值
		fromVal := reflect.ValueOf(from)
		if fromVal.Kind() == reflect.Ptr {
			if fromVal.IsNil() {
				return nil
			}
			fromVal = fromVal.Elem()
		}

		// 处理目标值
		toVal := reflect.ValueOf(to)
		if toVal.Kind() == reflect.Ptr {
			if toVal.IsNil() {
				// 创建新的目标对象
				toVal.Set(reflect.New(toVal.Type().Elem()))
			}
			toVal = toVal.Elem()
		}

		// 字段赋值
		for name, fn := range setValueFuncMap {
			srcFieldVal := reflectx.FieldByIndexes(fromVal, srcFields.Names[name].Index)
			dstFieldVal := reflectx.FieldByIndexes(toVal, dstFields.Names[name].Index)

			if !srcFieldVal.IsValid() || !dstFieldVal.CanSet() {
				continue
			}

			fn(dstFieldVal, srcFieldVal, opt)
		}
		return nil
	}
}

func notSupportConvert(src, dst any) error {
	return ErrNotSupportType
}

type setValueFunc func(dst, src reflect.Value, opt *options)

var setAssignableTo = func(dst, src reflect.Value, opt *options) {
	dst.Set(src)
}

var setConvertibleTo = func(dst, src reflect.Value, opt *options) {
	dst.Set(src.Convert(dst.Type()))
}

var setTimeToInt64 = func(dst, src reflect.Value, opt *options) {
	switch opt.timeUnit {
	case Nanosecond:
		dst.Set(reflect.ValueOf(src.Interface().(time.Time).UnixNano()))
	case Microsecond:
		dst.Set(reflect.ValueOf(src.Interface().(time.Time).UnixNano() / 1e3))
	case Millisecond:
		dst.Set(reflect.ValueOf(src.Interface().(time.Time).UnixNano() / 1e6))
	case Second:
		dst.Set(reflect.ValueOf(src.Interface().(time.Time).Unix()))
	default:
		dst.Set(reflect.ValueOf(src.Interface().(time.Time).Unix()))
	}
}

var setInt64ToTime = func(dst, src reflect.Value, opt *options) {
	switch opt.timeUnit {
	case Nanosecond:
		dst.Set(reflect.ValueOf(time.Unix(0, src.Interface().(int64))))
	case Microsecond:
		dst.Set(reflect.ValueOf(time.Unix(0, src.Interface().(int64)*1e3)))
	case Millisecond:
		dst.Set(reflect.ValueOf(time.Unix(0, src.Interface().(int64)*1e6)))
	case Second:
		dst.Set(reflect.ValueOf(time.Unix(src.Interface().(int64), 0)))
	default:
		dst.Set(reflect.ValueOf(time.Unix(src.Interface().(int64), 0)))
	}
}

// getCachedMapper 获取或创建指定 tagName 的 reflectx.Mapper
func getCachedMapper(tagName string) *reflectx.Mapper {
	v, ok := mapperCache.Load(tagName)
	if ok {
		return v.(*reflectx.Mapper)
	}
	mapper := reflectx.NewMapper(tagName)
	mapperCache.Store(tagName, mapper)
	return mapper
}

type tKey struct {
	src reflect.Type
	dst reflect.Type
	tag string
}

// typeKey 生成类型唯一 key，用于注册表索引
func typeKey(srcType, dstType reflect.Type, tagName string) tKey {
	return tKey{
		src: srcType,
		dst: dstType,
		tag: tagName,
	}
}

// isNil 判断接口值是否为 nil
func isNil(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// indirectType 获取类型的间接类型和是否为指针
func indirectType(reflectType reflect.Type) (_ reflect.Type, isPtr bool) {
	for reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
		isPtr = true
	}
	return reflectType, isPtr
}

package conv

import (
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/fengjx/go-halo/reflectx"
)

var (
	// registry converter
	registry sync.Map // map[tKey]converter

	// mapperCache 用于缓存不同 tagName 的 reflectx.Mapper
	mapperCache sync.Map // map[string]*reflectx.Mapper

	// customConverterRegistry 自定义类型转换器注册表
	customConverterRegistry sync.Map // map[typePair]SetValueFunc

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
type Option func(*Options)

type Options struct {
	TagName  string   // 结构体字段映射时使用的 tag 名称
	TimeUnit TimeUnit // 时间戳单位，ms/us/ns/s，默认s
}

// WithTag 设置结构体 tag 名称（如 "json"、"db" 等）
func WithTag(tag string) Option {
	return func(o *Options) {
		o.TagName = tag
	}
}

// WithTimeUnit 设置时间戳单位
func WithTimeUnit(unit TimeUnit) Option {
	return func(o *Options) {
		o.TimeUnit = unit
	}
}

// Convert 执行类型转换，支持 Option
// src 为源对象，dst 为目标对象指针，返回 error
func Convert(src, dst any, opts ...Option) error {
	if isNil(src) {
		return nil
	}

	opt := Options{
		TagName:  "json",
		TimeUnit: "s",
	}
	for _, o := range opts {
		o(&opt)
	}

	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	fn := getConverter(srcType, dstType, opt)
	return fn(src, dst, opt)
}

type converter func(src, dst any, opt Options) error

// 优化的缓存键结构，使用指针地址避免 reflect.Type 比较开销
type tKey struct {
	srcPtr uintptr
	dstPtr uintptr
	tag    string
}

// typePair 用于自定义转换器的键结构
type typePair struct {
	srcPtr uintptr
	dstPtr uintptr
}

// typeKey 生成类型唯一 key，用于注册表索引
func typeKey(srcType, dstType reflect.Type, tagName string) tKey {
	return tKey{
		srcPtr: reflect.ValueOf(srcType).Pointer(),
		dstPtr: reflect.ValueOf(dstType).Pointer(),
		tag:    tagName,
	}
}

// makeTypePair 创建类型对键
func makeTypePair(srcType, dstType reflect.Type) typePair {
	return typePair{
		srcPtr: reflect.ValueOf(srcType).Pointer(),
		dstPtr: reflect.ValueOf(dstType).Pointer(),
	}
}

// RegisterConverter 注册自定义类型转换器
// srcType: 源类型
// dstType: 目标类型
// converter: 转换函数，负责将源值转换为目标值
//
// 示例:
//   type MyString string
//   type MyInt int
//   RegisterConverter(reflect.TypeOf(MyString("")), reflect.TypeOf(0), func(dst, src reflect.Value, opt options) {
//       str := src.Interface().(MyString)
//       dst.Set(reflect.ValueOf(len(str)))
//   })
func RegisterConverter(srcType, dstType reflect.Type, converter SetValueFunc) {
	// 支持指针类型
	srcType, _ = indirectType(srcType)
	dstType, _ = indirectType(dstType)

	pair := makeTypePair(srcType, dstType)
	customConverterRegistry.Store(pair, converter)
}

// getCustomConverter 获取自定义转换器
func getCustomConverter(srcType, dstType reflect.Type) (SetValueFunc, bool) {
	pair := makeTypePair(srcType, dstType)
	if fn, ok := customConverterRegistry.Load(pair); ok {
		return fn.(SetValueFunc), true
	}
	return nil, false
}

func getConverter(srcType, dstType reflect.Type, opt Options) converter {
	// 获取实际的类型（去除指针）
	srcType, _ = indirectType(srcType)
	dstType, _ = indirectType(dstType)

	if srcType.Kind() != reflect.Struct {
		return notSupportConvert
	}
	if dstType.Kind() != reflect.Struct {
		return notSupportConvert
	}

	key := typeKey(srcType, dstType, opt.TagName)

	// 使用 sync.Map 的 Load 方法，无需加锁
	if fn, ok := registry.Load(key); ok {
		return fn.(converter)
	}

	// 生成新的转换器
	fn := genConverter(srcType, dstType, opt)
	registry.Store(key, fn)
	return fn
}

// genConverter 支持递归结构体赋值
func genConverter(srcType, dstType reflect.Type, opt Options) converter {
	mapper := getCachedMapper(opt.TagName)
	dstFields := mapper.TypeMap(dstType)
	srcFields := mapper.TypeMap(srcType)

	// 预分配 map 容量，减少扩容开销
	fieldMapInfoList := make([]fieldMapInfo, 0, len(dstFields.Names))

	for name, dstField := range dstFields.Names {
		srcField, ok := srcFields.Names[name]
		if !ok {
			continue
		}

		srcFieldType := srcField.Field.Type
		dstFieldType := dstField.Field.Type

		// 检查是否有自定义转换器（优先于 ConvertibleTo 检查）
		if customConverter, exist := getCustomConverter(srcFieldType, dstFieldType); exist {
			fieldMapInfoList = append(fieldMapInfoList, fieldMapInfo{
				srcIndex:     srcField.Index,
				dstIndex:     dstField.Index,
				setValueFunc: customConverter,
			})
			continue
		}

		// 直接赋值（类型可赋值）
		if srcFieldType.AssignableTo(dstFieldType) {
			fieldMapInfoList = append(fieldMapInfoList, fieldMapInfo{
				srcIndex:     srcField.Index,
				dstIndex:     dstField.Index,
				setValueFunc: setAssignableTo,
			})
			continue
		}

		if srcFieldType.ConvertibleTo(dstFieldType) {
			fieldMapInfoList = append(fieldMapInfoList, fieldMapInfo{
				srcIndex:     srcField.Index,
				dstIndex:     dstField.Index,
				setValueFunc: setConvertibleTo,
			})
			continue
		}

		// 时间戳转换
		if srcFieldType == reflectx.Int64Type && dstFieldType == reflectx.TimeType {
			fieldMapInfoList = append(fieldMapInfoList, fieldMapInfo{
				srcIndex:     srcField.Index,
				dstIndex:     dstField.Index,
				setValueFunc: setInt64ToTime,
			})
			continue
		}

		if srcFieldType == reflectx.TimeType && dstFieldType == reflectx.Int64Type {
			fieldMapInfoList = append(fieldMapInfoList, fieldMapInfo{
				srcIndex:     srcField.Index,
				dstIndex:     dstField.Index,
				setValueFunc: setTimeToInt64,
			})
			continue
		}

		// 递归处理结构体
		if srcFieldType.Kind() == reflect.Struct && dstFieldType.Kind() == reflect.Struct {
			subConv := genConverter(srcFieldType, dstFieldType, opt)
			fieldMapInfoList = append(fieldMapInfoList, fieldMapInfo{
				srcIndex: srcField.Index,
				dstIndex: dstField.Index,
				setValueFunc: func(dst, src reflect.Value, opt Options) {
					_ = subConv(src.Addr().Interface(), dst.Addr().Interface(), opt)
				},
			})
			continue
		}
	}

	return func(from any, to any, opt Options) error {
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
		for _, fieldMap := range fieldMapInfoList {
			srcFieldVal := reflectx.MustFieldByIndexesReadOnly(fromVal, fieldMap.srcIndex)
			if !srcFieldVal.IsValid() {
				continue
			}

			dstFieldVal := reflectx.FieldByIndexes(toVal, fieldMap.dstIndex)
			if !dstFieldVal.CanSet() {
				continue
			}
			fieldMap.setValueFunc(dstFieldVal, srcFieldVal, opt)
		}
		return nil
	}
}

func notSupportConvert(src, dst any, opt Options) error {
	return ErrNotSupportType
}

type fieldMapInfo struct {
	srcIndex     []int        // 源字段索引
	dstIndex     []int        // 目标字段索引
	setValueFunc SetValueFunc // 设置值函数
}

// SetValueFunc 自定义转换器函数类型
type SetValueFunc func(dst, src reflect.Value, opt Options)

var setAssignableTo = func(dst, src reflect.Value, opt Options) {
	dst.Set(src)
}

var setConvertibleTo = func(dst, src reflect.Value, opt Options) {
	dst.Set(src.Convert(dst.Type()))
}

var setTimeToInt64 = func(dst, src reflect.Value, opt Options) {
	switch opt.TimeUnit {
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

var setInt64ToTime = func(dst, src reflect.Value, opt Options) {
	switch opt.TimeUnit {
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

// isNil 判断接口值是否为 nil
func isNil(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
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

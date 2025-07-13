package conv

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/fengjx/go-halo/reflectx"
)

// Converter 泛型类型定义
// 用于描述从 SRC 类型到 DIST 类型的转换函数签名
// 参数：src 源对象
// 返回值：DIST 目标对象，error 错误信息
type Converter[SRC, DIST any] func(src SRC) (DIST, error)

var (
	mu       sync.RWMutex
	registry = make(map[string]any)

	// mapperCache 用于缓存不同 tagName 的 reflectx.Mapper
	mapperCache sync.Map // map[string]*reflectx.Mapper
)

// Register 注册自定义类型转换器
// fn 为自定义的转换实现，注册后优先于默认转换器被调用
func Register[SRC, DIST any](fn Converter[SRC, DIST]) {
	key := typeKey[SRC, DIST]("")
	mu.Lock()
	defer mu.Unlock()
	registry[key] = fn
}

// Option 用于自定义转换行为的函数类型
// 例如可用于自定义结构体 tag 名称等
type Option func(*options)

type options struct {
	tagName string // 结构体字段映射时使用的 tag 名称
}

// WithTag 设置结构体 tag 名称（如 "json"、"db" 等）
// 用于控制字段名映射时采用的 tag
func WithTag(tag string) Option {
	return func(o *options) {
		o.tagName = tag
	}
}

// Result 封装类型转换结果和错误，支持链式调用
type Result[SRC, DIST any] struct {
	value DIST
	err   error
}

// Value 返回转换结果
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
// from 为源对象，返回 Result 对象，支持链式调用
// 可通过 Option 自定义 tagName 等行为
// 若已注册自定义转换器则优先使用，否则使用默认反射拷贝
// from 为 nil 时，返回 DIST 的零值
func Convert[SRC, DIST any](from SRC, opts ...Option) Result[SRC, DIST] {
	var zero DIST
	if isNil(any(from)) {
		return Result[SRC, DIST]{value: zero, err: nil}
	}
	opt := options{}
	for _, o := range opts {
		o(&opt)
	}
	key := typeKey[SRC, DIST](opt.tagName)
	mu.RLock()
	fn, ok := registry[key]
	mu.RUnlock()
	if ok {
		v, err := fn.(Converter[SRC, DIST])(from)
		return Result[SRC, DIST]{value: v, err: err}
	}
	v, err := defaultConvert[SRC, DIST](from, opt.tagName)
	return Result[SRC, DIST]{value: v, err: err}
}

// defaultConvert 默认转换器，支持 struct tag 和递归结构体拷贝
// tagName 用于指定结构体字段映射时采用的 tag
func defaultConvert[SRC, DIST any](from SRC, tagName string) (DIST, error) {
	var to DIST
	toVal := reflect.ValueOf(&to).Elem()
	if toVal.Kind() == reflect.Ptr {
		// DIST 是指针类型
		if toVal.IsNil() {
			toVal.Set(reflect.New(toVal.Type().Elem()))
		}
		copyStructRecursive(to, from, tagName)
		return to, nil
	} else {
		// DIST 是值类型
		copyStructRecursive(&to, from, tagName)
		return to, nil
	}
}

func copyStructRecursive(dst, src any, tagName string) {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
		panic("dst must be a non-nil pointer")
	}
	dstVal = dstVal.Elem()
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr && !srcVal.IsNil() {
		srcVal = srcVal.Elem()
	}

	dstType := dstVal.Type()
	srcType := srcVal.Type()

	dstMapper := getCachedMapper(tagName)
	srcMapper := getCachedMapper(tagName)
	dstFields := dstMapper.TypeMap(dstType)
	srcFields := srcMapper.TypeMap(srcType)

	for name, dstField := range dstFields.Names {
		srcField, ok := srcFields.Names[name]
		if !ok {
			continue
		}
		dstFieldVal := reflectx.FieldByIndexes(dstVal, dstField.Index)
		srcFieldVal := reflectx.FieldByIndexesReadOnly(srcVal, srcField.Index)

		if !srcFieldVal.IsValid() || !dstFieldVal.CanSet() {
			continue
		}

		// 递归处理结构体
		if dstFieldVal.Kind() == reflect.Struct && srcFieldVal.Kind() == reflect.Struct {
			copyStructRecursive(dstFieldVal.Addr().Interface(), srcFieldVal.Interface(), tagName)
			continue
		}
		// 指针类型递归
		if dstFieldVal.Kind() == reflect.Ptr && srcFieldVal.Kind() == reflect.Ptr && !srcFieldVal.IsNil() {
			if dstFieldVal.IsNil() {
				dstFieldVal.Set(reflect.New(dstFieldVal.Type().Elem()))
			}
			copyStructRecursive(dstFieldVal.Interface(), srcFieldVal.Interface(), tagName)
			continue
		}
		// 直接赋值（类型可赋值）
		if srcFieldVal.Type().AssignableTo(dstFieldVal.Type()) {
			dstFieldVal.Set(srcFieldVal)
		} else if srcFieldVal.Type().ConvertibleTo(dstFieldVal.Type()) {
			// 支持 alias type 与基础类型的自动转换
			dstFieldVal.Set(srcFieldVal.Convert(dstFieldVal.Type()))
		}
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

// typeKey 生成类型唯一 key，用于注册表索引
func typeKey[SRC, DIST any](tagName string) string {
	if tagName == "" {
		return fmt.Sprintf("%T->%T:%s", *new(SRC), *new(DIST), tagName)
	}
	return fmt.Sprintf("%T->%T", *new(SRC), *new(DIST))
}

// isNil 判断接口值是否为 nil（支持指针、interface、slice、map、chan、func）
// 用于泛型类型安全判断 nil
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

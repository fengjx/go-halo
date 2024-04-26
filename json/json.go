package json

import (
	"io"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

type (
	// Any alias to jsoniter.Any
	Any = jsoniter.Any

	// Encoder alias to jsoniter.Encoder
	Encoder = jsoniter.Encoder
	// Decoder alias to jsoniter.Decoder
	Decoder = jsoniter.Decoder
)

func init() {

}

// ToJson 对象转 json 字符串
func ToJson(data interface{}) (string, error) {
	b, err := jsoniter.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ToJsonDelay 对象转 json 字符串，在调用 String() 方法时才会执行转换方法
func ToJsonDelay(data interface{}) *DelayJsoniter {
	return &DelayJsoniter{
		data: data,
	}
}

// ToBytes 对象转字节数组
func ToBytes(data interface{}) ([]byte, error) {
	return jsoniter.Marshal(data)
}

// FromJson json 字符串转对象
func FromJson(jsonStr string, target interface{}) error {
	err := jsoniter.Unmarshal([]byte(jsonStr), &target)
	if err != nil {
		return err
	}
	return nil
}

// FromBytes 字节数组转对象
func FromBytes(bytes []byte, target interface{}) error {
	err := jsoniter.Unmarshal(bytes, &target)
	if err != nil {
		return err
	}
	return nil
}

// GetNodeFromString 通过路径，快速从json字符串解析数据
func GetNodeFromString(jsonStr string, path interface{}) Any {
	return jsoniter.Get([]byte(jsonStr), path)
}

// GetNodeFromBytes 通过路径，快速从字节数组解析数据
func GetNodeFromBytes(byt []byte, path interface{}) Any {
	return jsoniter.Get([]byte(byt), path)
}

// NewEncoder jsoniter.NewEncoder
func NewEncoder(w io.Writer) *Encoder {
	return jsoniter.NewEncoder(w)
}

// NewDecoder jsoniter.NewDecoder
func NewDecoder(r io.Reader) *Decoder {
	return jsoniter.NewDecoder(r)
}

// RegisterTimeAsInt64Codec 注册时间类型为 int64
func RegisterTimeAsInt64Codec(precision time.Duration) {
	extra.RegisterTimeAsInt64Codec(precision)
}

// RegisterFuzzyDecoders 注册解码器支持容错
// string 和 number 自动转换
func RegisterFuzzyDecoders() {
	extra.RegisterFuzzyDecoders()
}

// DelayJsoniter 通过重写 string() 方法来延迟执行json序列化
type DelayJsoniter struct {
	data any
}

// String 返回json字符串，如果失败则返回error信息
func (d *DelayJsoniter) String() string {
	jsonStr, err := ToJson(d.data)
	if err != nil {
		return err.Error()
	}
	return jsonStr
}

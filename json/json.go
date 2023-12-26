package json

import (
	jsoniter "github.com/json-iterator/go"
)

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
func GetNodeFromString(jsonStr string, path interface{}) jsoniter.Any {
	return jsoniter.Get([]byte(jsonStr), path)
}

// GetNodeFromBytes 通过路径，快速从字节数组解析数据
func GetNodeFromBytes(byt []byte, path interface{}) jsoniter.Any {
	return jsoniter.Get([]byte(byt), path)
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

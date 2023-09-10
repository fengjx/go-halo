package json

import (
	jsoniter "github.com/json-iterator/go"
)

func ToJson(data interface{}) (string, error) {
	b, err := jsoniter.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ToBytes(data interface{}) ([]byte, error) {
	return jsoniter.Marshal(data)
}

func FromJson(jsonStr string, target interface{}) error {
	err := jsoniter.Unmarshal([]byte(jsonStr), &target)
	if err != nil {
		return err
	}
	return nil
}

func FromBytes(bytes []byte, target interface{}) error {
	err := jsoniter.Unmarshal(bytes, &target)
	if err != nil {
		return err
	}
	return nil
}

func GetNodeFromString(jsonStr string, path interface{}) jsoniter.Any {
	return jsoniter.Get([]byte(jsonStr), path)
}

func GetNodeFromBytes(byt []byte, path interface{}) jsoniter.Any {
	return jsoniter.Get([]byte(byt), path)
}

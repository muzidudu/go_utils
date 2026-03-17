package cache

import (
	"encoding/json"
)

// valueToBytes 将 any 序列化为 []byte
func valueToBytes(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	if b, ok := v.([]byte); ok {
		return b, nil
	}
	return json.Marshal(v)
}

// bytesToAny 将 []byte 反序列化为 any
func bytesToAny(data []byte) (any, error) {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return data, nil // 非 JSON 时返回原始 []byte
	}
	return v, nil
}

// bytesToValue 将 []byte 反序列化到 dest（用于 GetInto）
func bytesToValue(data []byte, dest any) error {
	if dest == nil {
		return nil
	}
	if p, ok := dest.(*[]byte); ok {
		*p = append((*p)[0:0], data...)
		return nil
	}
	return json.Unmarshal(data, dest)
}

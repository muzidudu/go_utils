package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// BuildKey 构建缓存键：prefix 与 parts 用冒号连接，每个 part 使用默认格式 %v。
func BuildKey(prefix string, parts ...interface{}) string {
	key := prefix
	for _, part := range parts {
		key += fmt.Sprintf(":%v", part)
	}
	return key
}

// BuildQueryKey 基于查询参数构建缓存键：将 params 序列化为 JSON 后计算 MD5，格式为 prefix:query:十六进制摘要。
// 若 JSON 序列化失败，则退化为 prefix:query:params 的字符串形式（与 fmt 的 %v 一致）。
func BuildQueryKey(prefix string, params interface{}) string {
	jsonData, err := json.Marshal(params)
	if err != nil {
		return fmt.Sprintf("%s:query:%v", prefix, params)
	}
	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%s:query:%x", prefix, hash)
}

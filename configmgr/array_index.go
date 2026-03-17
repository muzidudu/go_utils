package configmgr

import (
	"encoding/json"
	"fmt"
	"sync"
)

// IdExtractor[T] 从数组元素中提取唯一 ID 的函数
type IdExtractor[T any] func(T) string

// ArrayIndex[T] 基于 ID 的数组配置索引，支持 O(1) 访问
type ArrayIndex[T any] struct {
	mu         sync.RWMutex
	items      []T
	index      map[string]int // id -> slice index
	extractID  IdExtractor[T]
	configKey  string
}

// LoadArrayIndex 从配置加载数组并构建 ID 索引
// key: 配置中的数组 key，如 "servers"
// extractID: 从元素提取唯一 ID 的函数
func LoadArrayIndex[T any](m *Manager, key string, extractID IdExtractor[T]) (*ArrayIndex[T], error) {
	var items []T
	if err := m.UnmarshalArrayKey(key, &items); err != nil {
		return nil, err
	}
	idx := buildArrayIndex(items, extractID, key)
	return idx, nil
}

// NewArrayIndex 从已有 slice 创建 ArrayIndex（不依赖 Manager）
func NewArrayIndex[T any](items []T, extractID IdExtractor[T], configKey string) *ArrayIndex[T] {
	return buildArrayIndex(items, extractID, configKey)
}

func buildArrayIndex[T any](items []T, extractID IdExtractor[T], configKey string) *ArrayIndex[T] {
	index := make(map[string]int, len(items))
	for i, item := range items {
		id := extractID(item)
		if id != "" {
			index[id] = i
		}
	}
	return &ArrayIndex[T]{
		items:     items,
		index:     index,
		extractID: extractID,
		configKey: configKey,
	}
}

// Get 按 ID 获取元素（O(1)），第二个返回值表示是否存在
func (ai *ArrayIndex[T]) Get(id string) (T, bool) {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	var zero T
	i, ok := ai.index[id]
	if !ok {
		return zero, false
	}
	if i < 0 || i >= len(ai.items) {
		return zero, false
	}
	return ai.items[i], true
}

// GetPtr 按 ID 获取元素指针（O(1)），便于原地修改
func (ai *ArrayIndex[T]) GetPtr(id string) (*T, bool) {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	i, ok := ai.index[id]
	if !ok {
		return nil, false
	}
	if i < 0 || i >= len(ai.items) {
		return nil, false
	}
	return &ai.items[i], true
}

// Has 检查 ID 是否存在
func (ai *ArrayIndex[T]) Has(id string) bool {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	_, ok := ai.index[id]
	return ok
}

// All 返回所有元素（副本）
func (ai *ArrayIndex[T]) All() []T {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	out := make([]T, len(ai.items))
	copy(out, ai.items)
	return out
}

// Len 返回元素数量
func (ai *ArrayIndex[T]) Len() int {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	return len(ai.items)
}

// Set 设置或更新元素，若 ID 已存在则覆盖
func (ai *ArrayIndex[T]) Set(id string, item T) {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	if i, ok := ai.index[id]; ok {
		ai.items[i] = item
		return
	}
	ai.index[id] = len(ai.items)
	ai.items = append(ai.items, item)
}

// Delete 按 ID 删除元素
func (ai *ArrayIndex[T]) Delete(id string) bool {
	ai.mu.Lock()
	defer ai.mu.Unlock()
	i, ok := ai.index[id]
	if !ok {
		return false
	}
	// 从 slice 移除：与最后一个交换后截断
	last := len(ai.items) - 1
	ai.items[i] = ai.items[last]
	ai.items = ai.items[:last]
	delete(ai.index, id)
	// 更新被移动元素的索引
	if i != last {
		movedID := ai.extractID(ai.items[i])
		if movedID != "" {
			ai.index[movedID] = i
		}
	}
	return true
}

// IDs 返回所有 ID 列表
func (ai *ArrayIndex[T]) IDs() []string {
	ai.mu.RLock()
	defer ai.mu.RUnlock()
	ids := make([]string, 0, len(ai.index))
	for id := range ai.index {
		ids = append(ids, id)
	}
	return ids
}

// Save 将当前数据保存回 Manager 的配置
func (ai *ArrayIndex[T]) Save(m *Manager) error {
	ai.mu.RLock()
	items := ai.items
	key := ai.configKey
	ai.mu.RUnlock()
	if key == "" {
		return fmt.Errorf("configmgr: ArrayIndex has no config key")
	}
	// 将 []T 转为 viper 可接受的格式
	raw, err := sliceToRaw(items)
	if err != nil {
		return fmt.Errorf("configmgr: marshal array: %w", err)
	}
	m.Set(key, raw)
	return m.Save()
}

// SaveAs 保存到指定 Manager 并写入指定路径
func (ai *ArrayIndex[T]) SaveAs(m *Manager, path string) error {
	ai.mu.RLock()
	items := ai.items
	key := ai.configKey
	ai.mu.RUnlock()
	if key == "" {
		return fmt.Errorf("configmgr: ArrayIndex has no config key")
	}
	raw, err := sliceToRaw(items)
	if err != nil {
		return fmt.Errorf("configmgr: marshal array: %w", err)
	}
	m.Set(key, raw)
	return m.SaveAs(path)
}

// sliceToRaw 将 []T 转为 []map[string]any 供 viper 使用
func sliceToRaw[T any](items []T) ([]any, error) {
	if len(items) == 0 {
		return []any{}, nil
	}
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	var raw []any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

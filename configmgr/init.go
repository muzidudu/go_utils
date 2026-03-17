package configmgr

import (
	"fmt"
	"reflect"
)

// InitObject 初始化单对象结构体：先应用默认值，再从配置解析
// defaultVal: 默认值结构体（按值传入），cfg 将合并配置
func (m *Manager) InitObject(key string, defaultVal, cfg any) error {
	if err := applyDefaults(cfg, defaultVal); err != nil {
		return fmt.Errorf("configmgr: apply defaults: %w", err)
	}
	return m.UnmarshalKey(key, cfg)
}

// InitArray 初始化多对象 slice：先应用默认值到每个元素，再从配置解析
// defaultVal: 单个元素的默认值（用于 slice 中每个新元素的初始值）
func (m *Manager) InitArray(key string, defaultVal any, cfg any) error {
	if err := m.UnmarshalArrayKey(key, cfg); err != nil {
		return err
	}
	// 对解析后的每个元素应用默认值（仅填充零值字段）
	return applyDefaultsToSlice(cfg, defaultVal)
}

// applyDefaults 将 defaultVal 的非零值复制到 dst 的零值字段
func applyDefaults(dst, defaultVal any) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Pointer || dstVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be pointer to struct")
	}
	defVal := reflect.ValueOf(defaultVal)
	if defVal.Kind() != reflect.Struct {
		return fmt.Errorf("defaultVal must be struct")
	}
	dstVal = dstVal.Elem()
	for i := 0; i < defVal.NumField(); i++ {
		df := defVal.Field(i)
		dd := dstVal.Field(i)
		if dd.CanSet() && dd.IsZero() && !df.IsZero() {
			dd.Set(df)
		}
	}
	return nil
}

// applyDefaultsToSlice 对 slice 中每个元素应用默认值
func applyDefaultsToSlice(slicePtr, defaultVal any) error {
	rv := reflect.ValueOf(slicePtr)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("slicePtr must be pointer to slice")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("slicePtr must be pointer to slice")
	}
	defVal := reflect.ValueOf(defaultVal)
	if defVal.Kind() != reflect.Struct {
		return nil
	}
	for i := 0; i < rv.Len(); i++ {
		el := rv.Index(i)
		if el.Kind() == reflect.Pointer {
			el = el.Elem()
		}
		if el.Kind() != reflect.Struct {
			continue
		}
		for j := 0; j < defVal.NumField(); j++ {
			df := defVal.Field(j)
			dd := el.Field(j)
			if dd.CanSet() && dd.IsZero() && !df.IsZero() {
				dd.Set(df)
			}
		}
	}
	return nil
}

// Package configmgr 基于 Viper 的配置管理工具库
// 支持对象类型（单对象）与数组类型（多对象）两种配置模式，支持结构体初始化
package configmgr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// Manager 配置管理器，封装 Viper 并提供对象/数组两种解析模式
type Manager struct {
	v          *viper.Viper
	configPath string // 配置文件路径，用于 LoadOrInit 创建文件
}

// Option 配置选项函数
type Option func(*Manager)

// New 创建新的配置管理器
// configPath: 配置文件所在目录，可为空
// configName: 配置文件名（不含扩展名）
func New(configPath, configName string, opts ...Option) *Manager {
	v := viper.New()
	if configName != "" {
		v.SetConfigName(configName)
	}
	if configPath != "" {
		v.AddConfigPath(configPath)
	}
	m := &Manager{v: v}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// extToType 扩展名到 Viper 类型映射
var extToType = map[string]string{
	".yaml": "yaml", ".yml": "yaml",
	".json": "json",
	".toml": "toml",
	".env":  "env",
}

// NewFromPath 根据完整路径创建配置管理器，自动推断目录、文件名、扩展名
// 示例: NewFromPath("examples/config.yaml") 自动设置路径、类型为 yaml
// 支持 .yaml/.yml/.json/.toml/.env
func NewFromPath(configPath string, opts ...Option) *Manager {
	path := filepath.Clean(configPath)
	dir := filepath.Dir(path)
	ext := strings.ToLower(filepath.Ext(path))
	configType := extToType[ext]
	if configType == "" {
		configType = "yaml"
	}
	m := New(dir, "", WithConfigFile(path), WithConfigType(configType))
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// WithConfigFile 指定完整配置文件路径（含扩展名）
func WithConfigFile(path string) Option {
	return func(m *Manager) {
		m.configPath = path
		m.v.SetConfigFile(path)
	}
}

// WithConfigType 指定配置文件类型（yaml, json, toml 等）
func WithConfigType(t string) Option {
	return func(m *Manager) {
		m.v.SetConfigType(t)
	}
}

// WithEnvPrefix 设置环境变量前缀
func WithEnvPrefix(prefix string) Option {
	return func(m *Manager) {
		m.v.SetEnvPrefix(prefix)
	}
}

// WithAutomaticEnv 自动绑定环境变量
func WithAutomaticEnv() Option {
	return func(m *Manager) {
		m.v.AutomaticEnv()
	}
}

// WithDefault 设置默认值
func WithDefault(key string, value any) Option {
	return func(m *Manager) {
		m.v.SetDefault(key, value)
	}
}

// WithDefaults 批量设置默认值
func WithDefaults(defaults map[string]any) Option {
	return func(m *Manager) {
		for k, v := range defaults {
			m.v.SetDefault(k, v)
		}
	}
}

// Viper 返回底层 Viper 实例，用于访问 Viper 原生功能
func (m *Manager) Viper() *viper.Viper {
	return m.v
}

// Load 加载配置文件
func (m *Manager) Load() error {
	return m.v.ReadInConfig()
}

// LoadOrInit 加载配置，若文件不存在则初始化创建空配置文件
// 需先通过 NewFromPath 或 WithConfigFile 指定路径
func (m *Manager) LoadOrInit() error {
	return m.loadOrInit(nil)
}

// LoadOrInitWithDefaults 加载配置，若文件不存在则创建并写入默认值
// defaults: 默认配置，key 支持点号路径如 "server.port"
func (m *Manager) LoadOrInitWithDefaults(defaults map[string]any) error {
	return m.loadOrInit(defaults)
}

func (m *Manager) loadOrInit(defaults map[string]any) error {
	if err := m.v.ReadInConfig(); err != nil {
		// SetConfigFile 时文件不存在返回 os 错误，非 ConfigFileNotFoundError
		_, isNotFound := err.(viper.ConfigFileNotFoundError)
		if isNotFound || errors.Is(err, os.ErrNotExist) {
			path := m.configPath
			if path == "" {
				path = m.v.ConfigFileUsed()
			}
			if path == "" {
				return fmt.Errorf("configmgr: config file path not set, use NewFromPath or WithConfigFile")
			}
			path = filepath.Clean(path)
			// 转为绝对路径，避免工作目录影响
			if absPath, err := filepath.Abs(path); err == nil {
				path = absPath
			}
			// 确保目录存在
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("configmgr: create config dir: %w", err)
			}
			m.v.SetConfigFile(path)
			if len(defaults) > 0 {
				for k, v := range defaults {
					m.v.Set(k, v)
				}
				if err := m.v.WriteConfig(); err != nil {
					return fmt.Errorf("configmgr: write default config: %w", err)
				}
			} else {
				ext := strings.ToLower(filepath.Ext(path))
				initial := initialContent(ext)
				if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
					return fmt.Errorf("configmgr: init config file: %w", err)
				}
			}
			return m.v.ReadInConfig()
		}
		return err
	}
	return nil
}

// initialContent 根据扩展名返回初始空配置内容
func initialContent(ext string) string {
	switch ext {
	case ".json":
		return "{}\n"
	case ".toml":
		return "# config\n\n"
	case ".env":
		return "# config\n"
	default:
		return "# config\n\n"
	}
}

// LoadOrCreate 加载配置，若文件不存在则根据 ConfigFile 路径创建空配置
// 已废弃：请使用 NewFromPath + LoadOrInit
func (m *Manager) LoadOrCreate(configPath string) error {
	if configPath != "" {
		m.v.SetConfigFile(configPath)
	}
	return m.LoadOrInit()
}

// WatchConfig 监听配置文件变化（需配合 OnConfigChange 使用）
func (m *Manager) WatchConfig() {
	m.v.WatchConfig()
}

// OnConfigChange 注册配置变更回调
func (m *Manager) OnConfigChange(fn func()) {
	m.v.OnConfigChange(func(_ fsnotify.Event) {
		fn()
	})
}

// WatchAndReload 监听配置文件变化，变更时自动重新加载并解析到 cfg
// cfg: 目标结构体指针，如 &myConfig
func (m *Manager) WatchAndReload(cfg any) {
	m.WatchConfig()
	m.OnConfigChange(func() {
		_ = m.Load()
		_ = m.UnmarshalObject(cfg)
	})
}

// WatchAndReloadFunc 监听配置文件变化，变更时执行回调
// 适用于需自定义重载逻辑的场景
func (m *Manager) WatchAndReloadFunc(fn func()) {
	m.WatchConfig()
	m.OnConfigChange(fn)
}

// UnmarshalObject 对象模式：将配置解析为单对象结构体
// 适用于配置为单个对象的场景，如 server: { host: "0.0.0.0", port: 8080 }
func (m *Manager) UnmarshalObject(v any) error {
	return m.UnmarshalKey("", v)
}

// UnmarshalObjectKey 对象模式：将指定 key 下的配置解析为单对象结构体
func (m *Manager) UnmarshalObjectKey(key string, v any) error {
	return m.UnmarshalKey(key, v)
}

// UnmarshalArray 数组模式：将配置解析为多对象 slice
// 适用于配置为数组的场景，如 servers: [{ name: "s1" }, { name: "s2" }]
// v 必须为指向 slice 的指针，如 &[]Server{}
// 注意：通常需配合 UnmarshalArrayKey 指定包含数组的 key
func (m *Manager) UnmarshalArray(v any) error {
	return m.UnmarshalArrayKey("", v)
}

// UnmarshalArrayKey 数组模式：将指定 key 下的配置解析为多对象 slice
// 示例配置 YAML:
//
//	servers:
//	  - name: "api"
//	    port: 8080
//	  - name: "web"
//	    port: 3000
//
// 调用: UnmarshalArrayKey("servers", &servers)
func (m *Manager) UnmarshalArrayKey(key string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("configmgr: v must be a pointer to slice, got %s", rv.Kind())
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("configmgr: v must be a pointer to slice, got pointer to %s", rv.Kind())
	}
	var raw any
	if key == "" {
		raw = m.v.AllSettings()
		// 根配置为数组时（如 YAML/JSON 根节点为 []）
		if slice, ok := toSlice(raw); ok {
			return decodeSlice(slice, v)
		}
		// 根配置为单 key 对象且值为数组时，尝试自动解析
		if cfgMap, ok := raw.(map[string]any); ok && len(cfgMap) == 1 {
			for _, val := range cfgMap {
				if slice, ok := toSlice(val); ok {
					return decodeSlice(slice, v)
				}
				break
			}
		}
		return fmt.Errorf("configmgr: root config is not an array, use UnmarshalArrayKey with key (e.g. \"servers\")")
	}
	raw = m.v.Get(key)
	if raw == nil {
		return nil
	}
	slice, ok := toSlice(raw)
	if !ok {
		return fmt.Errorf("configmgr: config at key %q is not an array", key)
	}
	return decodeSlice(slice, v)
}

// UnmarshalKey 将指定 key 下的配置解析到 v
// key 为空时解析整个配置
func (m *Manager) UnmarshalKey(key string, v any) error {
	dc := &mapstructure.DecoderConfig{
		Result:           v,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	decoder, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return fmt.Errorf("configmgr: create decoder: %w", err)
	}
	var raw any
	if key == "" {
		raw = m.v.AllSettings()
	} else {
		raw = m.v.Get(key)
	}
	if raw == nil {
		return nil
	}
	return decoder.Decode(raw)
}

// Get 获取配置值（Viper 原生）
func (m *Manager) Get(key string) any {
	return m.v.Get(key)
}

// GetString 获取字符串
func (m *Manager) GetString(key string) string {
	return m.v.GetString(key)
}

// GetInt 获取整数
func (m *Manager) GetInt(key string) int {
	return m.v.GetInt(key)
}

// GetBool 获取布尔值
func (m *Manager) GetBool(key string) bool {
	return m.v.GetBool(key)
}

// Set 设置配置值
func (m *Manager) Set(key string, value any) {
	m.v.Set(key, value)
}

// SetDefault 设置默认值
func (m *Manager) SetDefault(key string, value any) {
	m.v.SetDefault(key, value)
}

// SetDefaults 批量设置默认值
func (m *Manager) SetDefaults(defaults map[string]any) {
	for k, v := range defaults {
		m.v.SetDefault(k, v)
	}
}

// IsSet 检查 key 是否已设置
func (m *Manager) IsSet(key string) bool {
	return m.v.IsSet(key)
}

// AllSettings 获取所有配置
func (m *Manager) AllSettings() map[string]any {
	return m.v.AllSettings()
}

// Save 保存配置到当前配置文件
// 需先 Load 或 WithConfigFile 指定路径
func (m *Manager) Save() error {
	return m.v.WriteConfig()
}

// SaveAndReload 保存后自动重新加载并解析到 cfg，确保内存与文件一致
func (m *Manager) SaveAndReload(cfg any) error {
	if err := m.v.WriteConfig(); err != nil {
		return err
	}
	if err := m.v.ReadInConfig(); err != nil {
		return err
	}
	return m.UnmarshalObject(cfg)
}

// SaveAs 保存配置到指定路径
func (m *Manager) SaveAs(path string) error {
	return m.v.WriteConfigAs(path)
}

// SafeSave 若配置文件不存在则创建，否则覆盖保存
func (m *Manager) SafeSave(path string) error {
	return m.v.SafeWriteConfigAs(path)
}

// toSlice 将 any 转为 []any
func toSlice(raw any) ([]any, bool) {
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice {
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

// decodeSlice 将 []any 解码到 slice 指针
func decodeSlice(slice []any, v any) error {
	dc := &mapstructure.DecoderConfig{
		Result:           v,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}
	decoder, err := mapstructure.NewDecoder(dc)
	if err != nil {
		return fmt.Errorf("configmgr: create decoder: %w", err)
	}
	return decoder.Decode(slice)
}

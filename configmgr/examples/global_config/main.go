// 全局变量配置示例：初始化、默认值、读写、对象类型、数组类型
package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/muzidudu/go_utils/configmgr"
)

// ========== 全局配置变量 ==========

var (
	cfg     *Config
	cfgMgr  *configmgr.Manager
	cfgOnce sync.Once
)

// ServerConfig 对象类型：单对象配置
type ServerConfig struct {
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

// AppConfig 数组元素类型
type AppConfig struct {
	AppName string `mapstructure:"appName"`
	Version string `mapstructure:"version"`
	Port    int    `mapstructure:"port"`
}

// DBConfig 对象类型：数据库配置
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// Config 全局配置结构
type Config struct {
	Server   ServerConfig `mapstructure:"server"`   // 对象类型
	Database DBConfig     `mapstructure:"database"` // 对象类型
	Apps     []AppConfig  `mapstructure:"apps"`     // 数组类型
}

// ========== 默认值 ==========

var defaultConfig = Config{
	Server: ServerConfig{
		Host:  "0.0.0.0",
		Port:  8080,
		Debug: false,
	},
	Database: DBConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "myapp",
		User:     "admin",
		Password: "secret",
	},
	Apps: []AppConfig{
		{AppName: "web", Version: "1.0", Port: 3000},
		{AppName: "api", Version: "1.0", Port: 8080},
	},
}

// ========== 初始化 ==========

// InitConfig 初始化全局配置（单例，仅执行一次）
func InitConfig(configPath string) error {
	var err error
	cfgOnce.Do(func() {
		err = initConfig(configPath)
	})
	return err
}

func initConfig(configPath string) error {
	path, _ := filepath.Abs(configPath)
	cfgMgr = configmgr.NewFromPath(path)

	// 默认值：文件不存在时写入
	defaults := map[string]any{
		"server.host":       defaultConfig.Server.Host,
		"server.port":       defaultConfig.Server.Port,
		"server.debug":      defaultConfig.Server.Debug,
		"database.host":     defaultConfig.Database.Host,
		"database.port":     defaultConfig.Database.Port,
		"database.database": defaultConfig.Database.Database,
		"database.user":     defaultConfig.Database.User,
		"database.password": defaultConfig.Database.Password,
		"apps":              defaultConfig.Apps,
	}

	if err := cfgMgr.LoadOrInitWithDefaults(defaults); err != nil {
		return err
	}

	cfg = &Config{}
	if err := cfgMgr.UnmarshalObject(cfg); err != nil {
		return err
	}

	// 修改文件后自动重新加载
	cfgMgr.WatchAndReload(cfg)
	return nil
}

// ========== 读取 ==========

// GetConfig 获取全局配置（只读）
func GetConfig() *Config {
	return cfg
}

// GetServer 读取对象类型
func GetServer() ServerConfig {
	return cfg.Server
}

// GetDatabase 读取对象类型
func GetDatabase() DBConfig {
	return cfg.Database
}

// GetApps 读取数组类型
func GetApps() []AppConfig {
	return cfg.Apps
}

// ========== 写入 ==========

// SetServer 写入对象类型并保存，保存后自动重载
func SetServer(s ServerConfig) error {
	cfg.Server = s
	cfgMgr.Set("server", map[string]any{
		"host":  s.Host,
		"port":  s.Port,
		"debug": s.Debug,
	})
	return cfgMgr.SaveAndReload(cfg)
}

// SetServerPort 写入单个字段，保存后自动重载
func SetServerPort(port int) error {
	cfg.Server.Port = port
	cfgMgr.Set("server.port", port)
	return cfgMgr.SaveAndReload(cfg)
}

// SetApps 写入数组类型并保存，保存后自动重载
func SetApps(apps []AppConfig) error {
	cfg.Apps = apps
	raw := make([]map[string]any, len(apps))
	for i, a := range apps {
		raw[i] = map[string]any{"appName": a.AppName, "version": a.Version, "port": a.Port}
	}
	cfgMgr.Set("apps", raw)
	return cfgMgr.SaveAndReload(cfg)
}

// Reload 重新加载配置
func Reload() error {
	if err := cfgMgr.Load(); err != nil {
		return err
	}
	return cfgMgr.UnmarshalObject(cfg)
}

// ========== main 示例 ==========

func main() {
	// 1. 初始化（含默认值）
	if err := InitConfig("examples/config.yaml"); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	fmt.Println("=== 读取对象类型 ===")
	server := GetServer()
	fmt.Printf("Server: %s:%d debug=%v\n", server.Host, server.Port, server.Debug)

	db := GetDatabase()
	fmt.Printf("Database: %s@%s:%d/%s\n", db.User, db.Host, db.Port, db.Database)

	fmt.Println("\n=== 读取数组类型 ===")
	apps := GetApps()
	for i, app := range apps {
		fmt.Printf("  [%d] %s v%s :%d\n", i, app.AppName, app.Version, app.Port)
	}

	fmt.Println("\n=== 写入并保存（更新后自动重载） ===")
	if err := SetServerPort(9090); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("已修改并保存，port = %d\n", GetServer().Port)

	fmt.Println("\n=== 自动重载已启用：修改 config.yaml 后会自动生效 ===")
	fmt.Println("（10 秒内可编辑 examples/config.yaml 测试，Ctrl+C 提前退出）")
	time.Sleep(10 * time.Second)
}

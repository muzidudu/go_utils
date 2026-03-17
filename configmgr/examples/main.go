// 完整使用示例：对象模式、数组模式、ID 索引、保存配置
package main

import (
	"fmt"
	"log"

	"github.com/muzidudu/go_utils/configmgr"
)

// ServerConfig 单对象配置
type ServerConfig struct {
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	Debug bool   `mapstructure:"debug"`
}

// AppConfig 数组元素：appName、version、port
type AppConfig struct {
	AppName string `mapstructure:"appName"`
	Version string `mapstructure:"version"`
	Port    int    `mapstructure:"port"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func main() {
	// 从路径推断目录、文件名、扩展名，文件不存在则自动初始化
	// 从 configmgr 根目录运行: go run ./examples
	m := configmgr.NewFromPath("examples/config.yaml")
	if err := m.LoadOrInit(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	fmt.Println("=== 对象模式：单对象配置 ===")
	var server ServerConfig
	if err := m.UnmarshalObjectKey("server", &server); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Server: %s:%d debug=%v\n", server.Host, server.Port, server.Debug)

	fmt.Println("\n=== 数组模式：多对象配置 (appName, version, port) ===")
	var apps []AppConfig
	if err := m.UnmarshalArrayKey("apps", &apps); err != nil {
		log.Fatal(err)
	}
	for i, app := range apps {
		fmt.Printf("  [%d] %s v%s :%d\n", i, app.AppName, app.Version, app.Port)
	}

	fmt.Println("\n=== 基于 ID 的数组索引：O(1) 访问 ===")
	idx, err := configmgr.LoadArrayIndex(m, "apps", func(a AppConfig) string { return a.AppName })
	if err != nil {
		log.Fatal(err)
	}
	if app, ok := idx.Get("web"); ok {
		fmt.Printf("  Get(\"web\") -> %s v%s :%d\n", app.AppName, app.Version, app.Port)
	}
	if app, ok := idx.Get("myapp"); ok {
		fmt.Printf("  Get(\"myapp\") -> %s v%s :%d\n", app.AppName, app.Version, app.Port)
	}
	fmt.Printf("  所有 ID: %v\n", idx.IDs())

	fmt.Println("\n=== 单对象配置 ===")
	var db DBConfig
	if err := m.UnmarshalObjectKey("database", &db); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Database: %s@%s:%d/%s\n", db.User, db.Host, db.Port, db.Database)

	// 2. 修改并保存（可选，取消注释以启用）
	// m.Set("server.port", 9090)
	// idx.Set("web", AppConfig{AppName: "web", Version: "3.0", Port: 3001})
	// idx.Save(m)
	// fmt.Println("\n配置已保存")
}

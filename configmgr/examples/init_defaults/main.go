// 空配置写默认值示例：文件不存在时创建并写入默认配置
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/muzidudu/go_utils/configmgr"
)

func main() {
	// 使用临时目录，模拟首次运行（配置文件不存在）
	dir := filepath.Join(os.TempDir(), "configmgr-init-example")
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, 0755)
	cfgPath := filepath.Join(dir, "config.yaml")
	defer os.RemoveAll(dir)

	// 默认值：文件不存在时写入
	defaults := map[string]any{
		"server.host":       "0.0.0.0",
		"server.port":       8080,
		"server.debug":      true,
		"database.host":     "localhost",
		"database.port":     5432,
		"database.database": "myapp",
		"database.user":     "admin",
		"database.password": "secret",
		"apps": []map[string]any{
			{"appName": "web", "version": "2.0", "port": 3000},
			{"appName": "api", "version": "1.0", "port": 8080},
		},
	}

	m := configmgr.NewFromPath(cfgPath)
	if err := m.LoadOrInitWithDefaults(defaults); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	fmt.Println("=== 空配置已创建并写入默认值 ===")
	fmt.Printf("配置文件: %s\n\n", cfgPath)

	fmt.Println("server:")
	fmt.Printf("  host: %s\n", m.GetString("server.host"))
	fmt.Printf("  port: %d\n", m.GetInt("server.port"))
	fmt.Printf("  debug: %v\n", m.GetBool("server.debug"))

	fmt.Println("\ndatabase:")
	fmt.Printf("  %s@%s:%d/%s\n",
		m.GetString("database.user"),
		m.GetString("database.host"),
		m.GetInt("database.port"),
		m.GetString("database.database"))

	fmt.Println("\napps:")
	var apps []struct {
		AppName string `mapstructure:"appName"`
		Version string `mapstructure:"version"`
		Port    int    `mapstructure:"port"`
	}
	_ = m.UnmarshalArrayKey("apps", &apps)
	for _, app := range apps {
		fmt.Printf("  - %s v%s :%d\n", app.AppName, app.Version, app.Port)
	}

}

func setDefaults() {

}

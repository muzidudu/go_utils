package configmgr_test

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/muzidudu/go_utils/configmgr"
)

// ServerConfig 单对象配置结构体
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// 对象模式示例：单对象配置
func ExampleManager_UnmarshalObject() {
	dir := createTempConfig(`
server:
  host: "0.0.0.0"
  port: 8080
`)
	defer os.RemoveAll(dir)

	m := configmgr.New(dir, "config", configmgr.WithConfigFile(filepath.Join(dir, "config.yaml")))
	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	var cfg ServerConfig
	if err := m.UnmarshalObjectKey("server", &cfg); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("host=%s port=%d\n", cfg.Host, cfg.Port)
	// Output: host=0.0.0.0 port=8080
}

// 数组模式示例：多对象配置 (databases)
func ExampleManager_UnmarshalArray() {
	dir := createTempConfig(`
databases:
  - host: "localhost"
    port: 5432
    database: "app"
    user: "admin"
    password: "secret"
  - host: "replica.example.com"
    port: 5432
    database: "app"
    user: "readonly"
    password: "readonly"
`)
	defer os.RemoveAll(dir)

	m := configmgr.New(dir, "config", configmgr.WithConfigFile(filepath.Join(dir, "config.yaml")))
	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	var dbs []DBConfig
	if err := m.UnmarshalArrayKey("databases", &dbs); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("count=%d first=%s\n", len(dbs), dbs[0].Host)
	// Output: count=2 first=localhost
}

// 数组类型示例：appName、version、port 结构
func ExampleManager_UnmarshalArray_apps() {
	dir := createTempConfig(`
apps:
  - appName: "web"
    version: "2.0"
    port: 3000
  - appName: "myapp"
    version: "1.0"
    port: 8080
`)
	defer os.RemoveAll(dir)

	m := configmgr.New(dir, "config", configmgr.WithConfigFile(filepath.Join(dir, "config.yaml")))
	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	type AppConfig struct {
		AppName string `mapstructure:"appName"`
		Version string `mapstructure:"version"`
		Port    int    `mapstructure:"port"`
	}
	var apps []AppConfig
	if err := m.UnmarshalArrayKey("apps", &apps); err != nil {
		log.Fatal(err)
	}
	for _, app := range apps {
		fmt.Printf("%s v%s :%d\n", app.AppName, app.Version, app.Port)
	}
	// Output:
	// web v2.0 :3000
	// myapp v1.0 :8080
}

// 数组类型示例：JSON 格式 (appName, version, port)
func ExampleManager_UnmarshalArray_appsJSON() {
	dir, err := os.MkdirTemp("", "configmgr-json-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "config.json")
	cfg := `{"apps":[{"appName":"web","version":"2.0","port":3000},{"appName":"myapp","version":"1.0","port":8080}]}`
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		panic(err)
	}

	m := configmgr.New("", "", configmgr.WithConfigFile(path), configmgr.WithConfigType("json"))
	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	type AppConfig struct {
		AppName string `mapstructure:"appName"`
		Version string `mapstructure:"version"`
		Port    int    `mapstructure:"port"`
	}
	var apps []AppConfig
	if err := m.UnmarshalArrayKey("apps", &apps); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("web: %d, myapp: %d\n", apps[0].Port, apps[1].Port)
	// Output: web: 3000, myapp: 8080
}

// 基于 ID 的数组索引示例：O(1) 访问
func ExampleLoadArrayIndex() {
	dir := createTempConfig(`
items:
  - id: "user-1"
    name: "Alice"
  - id: "user-2"
    name: "Bob"
`)
	defer os.RemoveAll(dir)

	m := configmgr.New(dir, "config", configmgr.WithConfigFile(filepath.Join(dir, "config.yaml")))
	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	type User struct {
		ID   string `mapstructure:"id"`
		Name string `mapstructure:"name"`
	}
	idx, err := configmgr.LoadArrayIndex(m, "items", func(u User) string { return u.ID })
	if err != nil {
		log.Fatal(err)
	}
	user, _ := idx.Get("user-2")
	fmt.Printf("name=%s\n", user.Name)
	// Output: name=Bob
}

// 基于 appName 的数组索引：appName 作为 ID
func ExampleLoadArrayIndex_apps() {
	dir := createTempConfig(`
apps:
  - appName: "web"
    version: "2.0"
    port: 3000
  - appName: "myapp"
    version: "1.0"
    port: 8080
`)
	defer os.RemoveAll(dir)

	m := configmgr.New(dir, "config", configmgr.WithConfigFile(filepath.Join(dir, "config.yaml")))
	if err := m.Load(); err != nil {
		log.Fatal(err)
	}

	type AppConfig struct {
		AppName string `mapstructure:"appName"`
		Version string `mapstructure:"version"`
		Port    int    `mapstructure:"port"`
	}
	idx, err := configmgr.LoadArrayIndex(m, "apps", func(a AppConfig) string { return a.AppName })
	if err != nil {
		log.Fatal(err)
	}
	app, _ := idx.Get("web")
	fmt.Printf("%s v%s :%d\n", app.AppName, app.Version, app.Port)
	// Output: web v2.0 :3000
}

func createTempConfig(content string) string {
	dir, err := os.MkdirTemp("", "configmgr-example-*")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		os.RemoveAll(dir)
		panic(err)
	}
	return dir
}

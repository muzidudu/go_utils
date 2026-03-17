package configmgr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewFromPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	cfg := `server:
  port: 8080
`
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	m := NewFromPath(path)
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	if m.GetInt("server.port") != 8080 {
		t.Errorf("got port %d", m.GetInt("server.port"))
	}
}

func TestLoadOrInit(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	m := NewFromPath(path)
	// 文件不存在，应自动创建
	if err := m.LoadOrInit(); err != nil {
		t.Fatalf("LoadOrInit: %v (path=%q)", err, path)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file was not created")
	}
	// 再次加载应成功
	if err := m.LoadOrInit(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadOrInitWithDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.yaml")
	m := NewFromPath(path)
	defaults := map[string]any{
		"server.port": 9000,
		"app.name":    "test",
	}
	if err := m.LoadOrInitWithDefaults(defaults); err != nil {
		t.Fatal(err)
	}
	if m.GetInt("server.port") != 9000 || m.GetString("app.name") != "test" {
		t.Errorf("got port=%d name=%q", m.GetInt("server.port"), m.GetString("app.name"))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 10 {
		t.Error("config file should contain defaults")
	}
}

func TestManager_ObjectMode(t *testing.T) {
	dir := t.TempDir()
	cfg := `
server:
  host: "0.0.0.0"
  port: 8080
  debug: true
`
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir, "config", WithConfigFile(path))
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	var s struct {
		Host  string `mapstructure:"host"`
		Port  int    `mapstructure:"port"`
		Debug bool   `mapstructure:"debug"`
	}
	if err := m.UnmarshalObjectKey("server", &s); err != nil {
		t.Fatal(err)
	}
	if s.Host != "0.0.0.0" || s.Port != 8080 || !s.Debug {
		t.Errorf("got host=%q port=%d debug=%v", s.Host, s.Port, s.Debug)
	}
}

func TestManager_ArrayMode(t *testing.T) {
	dir := t.TempDir()
	cfg := `
servers:
  - name: "api"
    port: 8080
  - name: "web"
    port: 3000
`
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir, "config", WithConfigFile(path))
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	var servers []struct {
		Name string `mapstructure:"name"`
		Port int    `mapstructure:"port"`
	}
	if err := m.UnmarshalArrayKey("servers", &servers); err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 {
		t.Fatalf("got %d servers", len(servers))
	}
	if servers[0].Name != "api" || servers[0].Port != 8080 {
		t.Errorf("server[0]: got name=%q port=%d", servers[0].Name, servers[0].Port)
	}
	if servers[1].Name != "web" || servers[1].Port != 3000 {
		t.Errorf("server[1]: got name=%q port=%d", servers[1].Name, servers[1].Port)
	}
}

func TestManager_InitObject(t *testing.T) {
	dir := t.TempDir()
	cfg := `
app:
  port: 9000
`
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir, "config", WithConfigFile(path))
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	type AppConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}
	defaultCfg := AppConfig{Host: "127.0.0.1", Port: 8080}
	var result AppConfig
	if err := m.InitObject("app", defaultCfg, &result); err != nil {
		t.Fatal(err)
	}
	// port 来自配置 9000，host 来自默认 127.0.0.1
	if result.Host != "127.0.0.1" || result.Port != 9000 {
		t.Errorf("got host=%q port=%d", result.Host, result.Port)
	}
}

func TestManager_Save(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := `server:
  host: "0.0.0.0"
  port: 8080
`
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	m := New("", "", WithConfigFile(path))
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	m.Set("server.port", 9090)
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == cfg {
		t.Error("config file was not updated")
	}
	// 重新加载验证
	m2 := New("", "", WithConfigFile(path))
	if err := m2.Load(); err != nil {
		t.Fatal(err)
	}
	if m2.GetInt("server.port") != 9090 {
		t.Errorf("got port %d", m2.GetInt("server.port"))
	}
}

func TestArrayIndex(t *testing.T) {
	dir := t.TempDir()
	cfg := `
items:
  - id: "a"
    name: "item-a"
    value: 1
  - id: "b"
    name: "item-b"
    value: 2
  - id: "c"
    name: "item-c"
    value: 3
`
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(cfg), 0644); err != nil {
		t.Fatal(err)
	}
	m := New(dir, "config", WithConfigFile(path))
	if err := m.Load(); err != nil {
		t.Fatal(err)
	}
	type Item struct {
		ID    string `mapstructure:"id"`
		Name  string `mapstructure:"name"`
		Value int    `mapstructure:"value"`
	}
	idx, err := LoadArrayIndex(m, "items", func(i Item) string { return i.ID })
	if err != nil {
		t.Fatal(err)
	}
	// O(1) Get
	item, ok := idx.Get("b")
	if !ok {
		t.Fatal("Get(b) not found")
	}
	if item.Name != "item-b" || item.Value != 2 {
		t.Errorf("got %+v", item)
	}
	// Has
	if !idx.Has("a") || idx.Has("x") {
		t.Error("Has failed")
	}
	// Len
	if idx.Len() != 3 {
		t.Errorf("Len=%d", idx.Len())
	}
	// Set
	idx.Set("d", Item{ID: "d", Name: "item-d", Value: 4})
	if !idx.Has("d") || idx.Len() != 4 {
		t.Error("Set failed")
	}
	// Delete
	if !idx.Delete("c") || idx.Has("c") {
		t.Error("Delete failed")
	}
	if idx.Len() != 3 {
		t.Errorf("Len after delete=%d", idx.Len())
	}
	// GetPtr 修改
	if p, ok := idx.GetPtr("a"); ok {
		p.Value = 100
	}
	if item, _ := idx.Get("a"); item.Value != 100 {
		t.Errorf("GetPtr modify: got value=%d", item.Value)
	}
}

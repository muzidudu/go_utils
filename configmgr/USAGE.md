# configmgr 使用说明

基于 Viper 的配置管理工具库，支持对象类型（单对象）、数组类型（多对象）、结构体初始化、默认值、热重载等。

---

## 一、执行逻辑与流程

```
┌─────────────────────────────────────────────────────────────────┐
│                        初始化阶段                                │
├─────────────────────────────────────────────────────────────────┤
│  NewFromPath / New  →  Load / LoadOrInit / LoadOrInitWithDefaults │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                        读取阶段                                  │
├─────────────────────────────────────────────────────────────────┤
│  对象类型: UnmarshalObjectKey / UnmarshalObject / InitObject      │
│  数组类型: UnmarshalArrayKey / UnmarshalArray / InitArray         │
│  键值读取: Get / GetString / GetInt / GetBool                     │
│  数组索引: LoadArrayIndex → Get / GetPtr / Has / IDs              │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                        写入阶段                                  │
├─────────────────────────────────────────────────────────────────┤
│  Set → Save / SaveAndReload / SaveAs                             │
│  ArrayIndex: Set / Delete → Save(m)                              │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                        监听阶段（可选）                           │
├─────────────────────────────────────────────────────────────────┤
│  WatchAndReload / WatchAndReloadFunc                             │
│  → 文件变更时自动重载                                            │
└─────────────────────────────────────────────────────────────────┘
```

---

## 二、何时使用相应函数

### 2.1 创建 Manager

| 函数 | 使用场景 |
|------|----------|
| `NewFromPath(path)` | **推荐** 已知完整路径，自动推断目录、文件名、扩展名 |
| `New(dir, name)` | 需按目录+文件名搜索，或使用 Viper 默认搜索逻辑 |
| `WithConfigFile(path)` | 覆盖/指定配置文件路径 |
| `WithConfigType(t)` | 指定类型（yaml/json/toml/env） |
| `WithEnvPrefix(prefix)` | 环境变量前缀 |
| `WithAutomaticEnv()` | 自动绑定环境变量 |
| `WithDefault(s)/WithDefaults(map)` | 运行时默认值（文件未定义时生效） |

### 2.2 加载配置

| 函数 | 使用场景 |
|------|----------|
| `Load()` | 配置文件**必须存在**，直接读取 |
| `LoadOrInit()` | 文件**不存在时创建空文件**，再读取 |
| `LoadOrInitWithDefaults(defaults)` | 文件**不存在时创建并写入默认值**，再读取 |

### 2.3 解析配置

| 函数 | 使用场景 | 配置结构 |
|------|----------|----------|
| `UnmarshalObjectKey(key, &cfg)` | 解析**单对象**到结构体 | `server: { host, port }` |
| `UnmarshalObject(&cfg)` | 解析**整个配置**到结构体 | 根级对象 |
| `UnmarshalArrayKey(key, &slice)` | 解析**数组**到 slice | `apps: [{...}, {...}]` |
| `UnmarshalKey(key, v)` | 通用解析，key 为空时解析全部 | 任意 |
| `InitObject(key, defaultVal, &cfg)` | 单对象 + **默认值填充零值字段** | 同 UnmarshalObjectKey |
| `InitArray(key, defaultVal, &slice)` | 数组 + **默认值填充零值字段** | 同 UnmarshalArrayKey |

### 2.4 键值读写

| 函数 | 使用场景 |
|------|----------|
| `Get(key)` | 获取任意值 |
| `GetString/GetInt/GetBool(key)` | 获取类型化值 |
| `Set(key, value)` | 设置值（未持久化，需配合 Save） |
| `IsSet(key)` | 检查 key 是否存在 |
| `AllSettings()` | 获取全部配置 |

### 2.5 保存配置

| 函数 | 使用场景 |
|------|----------|
| `Save()` | 保存到当前配置文件 |
| `SaveAndReload(&cfg)` | 保存后**重新加载并解析到 cfg**，保证内存与文件一致 |
| `SaveAs(path)` | 保存到指定路径 |
| `SafeSave(path)` | 文件不存在则创建，存在则覆盖 |

### 2.6 热重载

| 函数 | 使用场景 |
|------|----------|
| `WatchAndReload(&cfg)` | 文件变更时**自动 Load + UnmarshalObject** |
| `WatchAndReloadFunc(fn)` | 文件变更时执行自定义回调 |
| `WatchConfig()` + `OnConfigChange(fn)` | 手动组合监听与回调 |

### 2.7 数组索引（ArrayIndex）

| 函数 | 使用场景 |
|------|----------|
| `LoadArrayIndex(m, key, extractID)` | 从配置加载数组并构建 **O(1) ID 索引** |
| `idx.Get(id)` | 按 ID 获取元素 |
| `idx.GetPtr(id)` | 按 ID 获取指针，便于原地修改 |
| `idx.Has(id)` | 检查 ID 是否存在 |
| `idx.Set(id, item)` | 新增或更新 |
| `idx.Delete(id)` | 按 ID 删除 |
| `idx.All()` / `idx.IDs()` | 获取全部元素 / 全部 ID |
| `idx.Save(m)` | 写回 Manager 并保存 |

---

## 三、使用方法

### 3.1 最简用法（推荐）

```go
m := configmgr.NewFromPath("config.yaml")
m.LoadOrInit()  // 文件不存在则创建

var server struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}
m.UnmarshalObjectKey("server", &server)
```

### 3.2 首次运行写默认值

```go
m := configmgr.NewFromPath("config.yaml")
defaults := map[string]any{
    "server.host": "0.0.0.0",
    "server.port": 8080,
    "apps": []map[string]any{
        {"appName": "web", "port": 3000},
    },
}
m.LoadOrInitWithDefaults(defaults)
```

### 3.3 对象类型配置

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
```

```go
type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}
var cfg ServerConfig
m.UnmarshalObjectKey("server", &cfg)
```

### 3.4 数组类型配置

```yaml
# config.yaml
apps:
  - appName: "web"
    version: "2.0"
    port: 3000
  - appName: "api"
    version: "1.0"
    port: 8080
```

```go
type AppConfig struct {
    AppName string `mapstructure:"appName"`
    Version string `mapstructure:"version"`
    Port    int    `mapstructure:"port"`
}
var apps []AppConfig
m.UnmarshalArrayKey("apps", &apps)
```

### 3.5 数组按 ID 访问（O(1)）

```go
idx, _ := configmgr.LoadArrayIndex(m, "apps", func(a AppConfig) string { return a.AppName })
app, ok := idx.Get("web")
idx.Set("newapp", AppConfig{AppName: "newapp", Port: 9000})
idx.Save(m)
```

### 3.6 带默认值的结构体初始化

```go
defaultServer := ServerConfig{Host: "127.0.0.1", Port: 8080}
var cfg ServerConfig
m.InitObject("server", defaultServer, &cfg)  // 配置覆盖默认值
```

### 3.7 修改并保存（更新后自动重载）

```go
m.Set("server.port", 9090)
m.SaveAndReload(&cfg)  // 保存 + 重新加载到 cfg
```

### 3.8 文件变更自动重载

```go
var cfg Config
m.UnmarshalObject(&cfg)
m.WatchAndReload(&cfg)  // 外部修改 config.yaml 时自动重载
```

### 3.9 全局配置完整流程

```go
// 1. 初始化
m := configmgr.NewFromPath("config.yaml")
m.LoadOrInitWithDefaults(defaults)
var cfg Config
m.UnmarshalObject(&cfg)

// 2. 监听文件变更
m.WatchAndReload(&cfg)

// 3. 修改时保存并重载
m.Set("server.port", 9090)
m.SaveAndReload(&cfg)
```

---

## 四、结构体标签

使用 `mapstructure` 标签映射配置字段：

```go
type AppConfig struct {
    AppName string `mapstructure:"appName"`  // 对应 YAML 中的 appName
    Version string `mapstructure:"version"`
    Port    int    `mapstructure:"port"`
}
```

---

## 五、示例目录

| 示例 | 说明 |
|------|------|
| `examples/main.go` | 完整流程：对象、数组、ID 索引 |
| `examples/init_defaults/` | 空配置写默认值 |
| `examples/global_config/` | 全局变量、初始化、读写、热重载、更新后自动重载 |

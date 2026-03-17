# configmgr 使用示例

## 运行

从 configmgr 根目录执行：

```bash
go run ./examples              # 完整示例
go run ./examples/init_defaults   # 空配置写默认值
go run ./examples/global_config   # 全局变量：初始化、默认值、读写、对象/数组
```

## 空配置写默认值 (init_defaults)

当配置文件不存在时，创建并写入默认值：

```go
defaults := map[string]any{
    "server.host":  "0.0.0.0",
    "server.port":  8080,
    "apps": []map[string]any{
        {"appName": "web", "version": "2.0", "port": 3000},
    },
}
m := configmgr.NewFromPath("config.yaml")
m.LoadOrInitWithDefaults(defaults)
```

## 全局变量 (global_config)

单例模式：全局配置变量 + 初始化 + 默认值 + 读写 + 对象/数组类型

```go
// 1. 初始化（含默认值）
InitConfig("config.yaml")

// 2. 读取 - 对象类型
server := GetServer()
db := GetDatabase()

// 3. 读取 - 数组类型
apps := GetApps()

// 4. 写入并保存（更新后自动重载）
SetServerPort(9090)
SetApps(newApps)

// 5. 文件变更自动重载 + 更新后自动重载保存（已内置）
```

**直接使用 Manager：**
```go
m := configmgr.NewFromPath("config.yaml")
m.LoadOrInit()
var cfg MyConfig
m.UnmarshalObject(&cfg)
m.WatchAndReload(&cfg)       // 文件变更时自动重载
m.Set("server.port", 9090)
m.SaveAndReload(&cfg)       // 更新后保存并重载
```

## 配置格式

支持 YAML 和 JSON，示例结构如下。

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080

apps:
  - appName: "web"
    version: "2.0"
    port: 3000
  - appName: "myapp"
    version: "1.0"
    port: 8080
```

### config.json

```json
{
  "apps": [
    { "appName": "web", "version": "2.0", "port": 3000 },
    { "appName": "myapp", "version": "1.0", "port": 8080 }
  ]
}
```

## 用法示例

| 模式 | 方法 | 说明 |
|------|------|------|
| 对象 | `UnmarshalObjectKey("server", &cfg)` | 单对象配置 |
| 数组 | `UnmarshalArrayKey("apps", &apps)` | 多对象 slice |
| ID 索引 | `LoadArrayIndex(m, "apps", func(a App) string { return a.AppName })` | O(1) 按 ID 访问 |
| 保存 | `m.Save()` / `idx.Save(m)` | 持久化配置 |

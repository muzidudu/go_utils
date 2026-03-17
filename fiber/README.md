# Fiber 快速启动框架

基于 Fiber v3 的 Web 应用脚手架，集成 configmgr、cache、GORM PostgreSQL，支持优雅关闭与 Bootstrap 启动。

## 特性

- **Fiber v3**：高性能 Web 框架
- **configmgr**：配置管理（YAML/JSON/TOML）
- **cache**：Redis / 内存缓存（自动降级）
- **GORM + PostgreSQL**：数据库
- **优雅关闭**：SIGINT/SIGTERM 时平滑退出
- **Bootstrap 启动**：自动加载配置、初始化缓存与数据库
- **路由分离**：`http_route.go`、`api_route.go`、`InstallRouter`

## 目录结构

```
fiber/
├── main.go              # 入口，优雅关闭
├── config/
│   └── config.yaml      # 配置文件
├── bootstrap/
│   └── bootstrap.go    # 启动引导
├── internal/
│   └── routes/
│       ├── install.go   # InstallRouter
│       ├── http_route.go # HTTP 页面路由
│       └── api_route.go  # API 路由
└── go.mod
```

## 快速开始

```bash
cd fiber
go run .
```

默认监听 `:3000`，访问：

- `GET /` - 首页
- `GET /health` - 健康检查
- `GET /api/ping` - API 示例
- `GET /api/cache/:key` - 读取缓存
- `POST /api/cache/:key` - 写入缓存

## 配置

编辑 `config/config.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 3000
  debug: false

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: app
  sslmode: disable

cache:
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    prefix: "app:"
  memory:
    max_count: 10000
    max_bytes: 67108864
```

- Redis 不可用时自动降级到内存缓存
- PostgreSQL 连接失败时应用仍可启动（无 DB 模式）

## 扩展路由

在 `internal/routes/http_route.go` 和 `api_route.go` 中新增路由，或创建新文件并在 `InstallRouter` 中调用。

## 依赖

- Go 1.25+
- 可选：PostgreSQL、Redis（不配置则使用内存缓存、无 DB）

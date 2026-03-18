# Fiber 快速启动框架

基于 Fiber v3 的 Web 应用脚手架，集成 configmgr、cache、GORM PostgreSQL、Django 模板、compress/logger 中间件，支持优雅关闭与 Bootstrap 启动。

## 特性

- **Fiber v3**：高性能 Web 框架
- **configmgr**：配置管理（YAML/JSON/TOML）
- **cache**：Redis / 内存缓存（自动降级）
- **GORM + PostgreSQL**：数据库
- **Django 模板**：`github.com/gofiber/template/django/v3`
- **中间件**：`compress`、`logger`
- **全局调用**：`app.Config()`、`app.Cache()`、`app.DB()`、`app.Fiber()`
- **分层架构**：Handler → Service → Repository，DTO 隔离 Model
- **数据库自动迁移**：`bootstrap.AutoMigrate`
- **优雅关闭**：SIGINT/SIGTERM 时平滑退出
- **路由分离**：`http_route.go`、`api_route.go`、`InstallRouter`

## 目录结构

```
fiber/
├── main.go
├── config/config.yaml
├── views/                    # Django 模板
│   ├── layouts/main.django
│   ├── index.django
│   └── users/index.django
├── bootstrap/
│   ├── app.go
│   ├── config.go
│   ├── cache.go
│   ├── database.go
│   └── fiber.go
└── internal/
    ├── app/                  # 全局应用
    ├── models/               # 数据模型
    ├── dto/                  # 请求/响应结构
    ├── repository/           # 数据访问层（纯 CRUD）
    ├── service/              # 业务逻辑
    ├── handlers/             # 控制器
    └── routes/               # 路由
```

## 快速开始

```bash
cd fiber
go run .
```

默认监听 `:3000`，访问：

- `GET /` - 首页（Django 模板）
- `GET /users` - 用户列表（模板）
- `GET /health` - 健康检查
- `GET /api/ping` - API 示例
- `GET /api/users` - 用户列表 API
- `POST /api/users` - 创建用户
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

## 全局调用示例

```go
import "github.com/muzidudu/go_utils/fiber/internal/app"

// 在 repository、handlers 等层使用
cfg := app.Config()
cache := app.Cache()
db := app.DB()
fiber := app.Fiber()
```

## 分层架构

| 层 | 职责 |
|----|------|
| **Handler** | 接收/解析请求，参数校验，调用 Service |
| **Service** | 业务逻辑处理 |
| **Repository** | 数据库读写（纯 CRUD，不写业务） |
| **Model** | 数据库结构体 |
| **DTO** | 请求/响应结构（隔离 Model） |
| **Router** | 注册路由，绑定 Handler |

## 扩展路由

在 `internal/routes/http_route.go` 和 `api_route.go` 中新增路由，或创建新文件并在 `InstallRouter` 中调用。

## 依赖

- Go 1.25+
- 可选：PostgreSQL、Redis（不配置则使用内存缓存、无 DB）

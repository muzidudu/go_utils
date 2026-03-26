# Fiber 快速启动框架

基于 Fiber v3 的 Web 应用脚手架，集成 configmgr、cache、GORM（PostgreSQL/MySQL/SQLite）、Django 模板、compress/logger 中间件，支持优雅关闭与 Bootstrap 启动。

## 特性

- **Fiber v3**：高性能 Web 框架
- **configmgr**：配置管理（YAML/JSON/TOML）
- **cache**：Redis / 内存缓存（自动降级）
- **GORM**：支持 PostgreSQL、MySQL、SQLite
- **Django 模板**：`github.com/gofiber/template/django/v3`
- **中间件**：`compress`、`logger`
- **全局调用**：`app.Config()`、`app.Cache()`、`app.DB()`、`app.Fiber()`
- **data 包**：`data.GetSiteByDomain()`、`data.GetCategoryTree()` 等全局数据访问
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
    ├── data/                 # 全局数据访问（sites、categories）
    ├── models/               # 数据模型
    ├── dto/                  # 请求/响应结构
    ├── repository/           # 数据访问层（纯 CRUD）
    ├── service/              # 业务逻辑
    ├── handlers/             # 控制器
    ├── middleware/           # 中间件
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
- `GET /api/sites` - 站点列表
- `GET /api/categories/tree` - 分类树
- `GET /api/categories/flat` - 分类扁平列表
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
  driver: postgres   # postgres | mysql | sqlite
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  database: app
  # path: data/app.db  # sqlite 时使用
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

- **数据库**：`driver` 可选 `postgres`、`mysql`、`sqlite`；SQLite 使用 `path` 或 `database` 作为文件路径
- **缓存**：Redis 不可用时自动降级到内存缓存
- **无 DB 模式**：数据库连接失败时应用仍可启动

## 全局调用示例

```go
import "github.com/muzidudu/go_utils/fiber/internal/app"
import "github.com/muzidudu/go_utils/fiber/internal/data"

// app：配置、缓存、数据库、Fiber 实例
cfg := app.Config()
cache := app.Cache()
db := app.DB()
fiber := app.Fiber()

// data：站点、分类（带缓存）
site := data.GetSiteByDomain(host)
site = data.GetDefaultSite()
sites := data.GetAllSites()

tree, _ := data.GetCategoryTree(0)    // 完整树
tree, _ = data.GetCategoryTree(5)    // 以 id=5 为根的子树
flat, _ := data.GetCategoryFlat(0)    // 根级扁平
cat, _ := data.GetCategoryByID(1)
```

## 分层架构


| 层              | 职责                      |
| -------------- | ----------------------- |
| **Handler**    | 接收/解析请求，参数校验，调用 Service |
| **Service**    | 业务逻辑处理                  |
| **Repository** | 数据库读写（纯 CRUD，不写业务）      |
| **Model**      | 数据库结构体                  |
| **DTO**        | 请求/响应结构（隔离 Model）       |
| **Router**     | 注册路由，绑定 Handler         |


## API 路由


| 方法     | 路径                                 | 说明     |
| ------ | ---------------------------------- | ------ |
| GET    | `/api/sites`                       | 站点列表   |
| GET    | `/api/sites/:id`                   | 站点详情   |
| POST   | `/api/sites`                       | 创建站点   |
| PUT    | `/api/sites/:id`                   | 更新站点   |
| DELETE | `/api/sites/:id`                   | 删除站点   |
| GET    | `/api/categories/tree?parent_id=0` | 分类树    |
| GET    | `/api/categories/flat?parent_id=0` | 分类扁平列表 |
| GET    | `/api/categories/:id`              | 分类详情   |
| POST   | `/api/categories`                  | 创建分类   |
| PUT    | `/api/categories/:id`              | 更新分类   |
| DELETE | `/api/categories/:id`              | 删除分类   |


## 扩展路由

在 `internal/routes/http_route.go` 和 `api_route.go` 中新增路由，或创建新文件并在 `InstallRouter` 中调用。

## 后台管理

`fiber/backend/` 为 SvelteKit + Tailwind 后台管理界面，支持站点管理和分类管理。左右布局，调用 Fiber API。

```bash
cd fiber/backend
npm install
npm run dev  # http://localhost:5173
```

配置 `VITE_API_URL=http://localhost:3000/api` 指向 Fiber 服务。Fiber 已启用 CORS 支持跨域。

## 依赖

- Go 1.25+
- 可选：PostgreSQL / MySQL / SQLite、Redis（不配置则使用内存缓存、无 DB）


# go_utils/template

基于 [pongo2](https://github.com/flosch/pongo2) 的多站点模板引擎，兼容 [Fiber](https://gofiber.io) 的 `Views` 接口，支持多主题（theme）按 binding 动态切换。

## 特性

- **多站点/多主题**：按 `Theme` 或 `siteBase.Template` 自动选择模板目录
- **Django 语法**：使用 pongo2 的 Django 风格模板语法
- **自定义 Tag**：支持 `RegisterTag` 注册 pongo2 自定义标签
- **Layout 布局**：支持 `layouts/main` 等布局模板
- **回退机制**：主题模板不存在时自动回退到 `default` 主题

## 安装

```bash
go get github.com/muzidudu/go_utils/template
```

## 目录结构

```
views/
└── template/
    ├── default/           # 默认主题
    │   ├── index.django
    │   └── layouts/
    │       └── main.django
    └── 321/               # 主题 321
        ├── index.django
        └── layouts/
            └── main.django
```

每个子目录名即为 theme 名称，模板 key 格式为 `{theme}/{name}`，如 `321/index`、`default/layouts/main`。

## 使用方法

### 1. 创建引擎并接入 Fiber

```go
package main

import (
    "path/filepath"

    "github.com/gofiber/fiber/v3"
    "github.com/muzidudu/go_utils/template"
)

func main() {
    templateDir := filepath.Join("views")
    engine := template.New(templateDir, ".django")

    // 可选：开发时热重载
    engine.Reload(true)
    engine.Debug(false)

    app := fiber.New(fiber.Config{
        Views:             engine,
        PassLocalsToViews: true,
    })

    app.Get("/", func(c fiber.Ctx) error {
        return c.Render("index", fiber.Map{
            "Title": "首页",
            "Theme": "321",  // 指定使用 321 主题
        }, "layouts/main")
    })
}
```

### 2. 通过 binding 指定主题

主题由 binding 中的 `Theme` 或 `siteBase.Template` 决定，优先级：`Theme` > `siteBase.Template` > `default`。

```go
// 方式一：直接传 Theme
c.Render("index", fiber.Map{
    "Theme": "321",
    "Title": "首页",
}, "layouts/main")

// 方式二：通过 siteBase（站点对象含 Template 字段）
c.Render("index", fiber.Map{
    "siteBase": site,  // site.Template == "321"
    "Title":    "首页",
}, "layouts/main")
```

### 3. 注册自定义 Tag

在 `Load` 之前调用 `RegisterTag`：

```go
engine := template.New("views", ".django")
engine.RegisterTag("uppercase", template.TagUppercaseParser)

// 模板中使用: {% uppercase %}hello{% enduppercase %} -> HELLO
```

### 4. 添加模板函数

```go
engine := template.New("views", ".django")
engine.AddFunc("upper", strings.ToUpper)
engine.AddFunc("now", func() time.Time { return time.Now() })
```

### 5. 中间件中 ViewBind 注入 Theme

在 Fiber 中间件中通过 `ViewBind` 注入站点信息，后续 `Render` 会自动使用对应主题：

```go
if site != nil {
    c.ViewBind(fiber.Map{
        "siteBase":  site,
        "Theme":     site.Template,
        "ThemePath": "template/" + site.Template,
    })
}
```

## API 概览

| 方法 | 说明 |
|------|------|
| `New(baseDir, extension string) *SitesEngine` | 创建引擎 |
| `RegisterTag(name string, parser pongo2.TagParser)` | 注册自定义 tag |
| `AddFunc(name string, fn interface{}) *SitesEngine` | 添加模板函数 |
| `SetAutoEscape(v bool) *SitesEngine` | 设置自动转义 |
| `Load() error` | 加载所有主题模板 |
| `Render(out, name, binding, layout...) error` | 渲染模板 |

## 依赖

- `github.com/flosch/pongo2/v6`
- `github.com/gofiber/template/v2`（实现 Engine 接口）

## License

与 go_utils 项目一致。

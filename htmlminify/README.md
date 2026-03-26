# htmlminify

基于 [github.com/tdewolff/minify/v2](https://github.com/tdewolff/minify) 封装的 [Fiber v3](https://github.com/gofiber/fiber) 中间件，在响应写出后对 `text/html` 正文做压缩，减小 HTML 体积。

## 依赖

- `github.com/gofiber/fiber/v3`
- `github.com/tdewolff/minify/v2`（HTML，可选内联 CSS / JS、SVG）

## 安装

Module 路径：`github.com/muzidudu/go_utils/htmlminify`。

```bash
go get github.com/muzidudu/go_utils/htmlminify@latest
```

在业务项目中使用本仓库的本地副本时，可在 `go.mod` 中增加：

```go
replace github.com/muzidudu/go_utils/htmlminify => ../path/to/go_utils/htmlminify
```

（将右侧路径改为本包实际所在目录。）

## 快速开始

```go
package main

import (
	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/htmlminify"
)

func main() {
	app := fiber.New()

	app.Use(htmlminify.HTMLMinify())

	app.Get("/", func(c fiber.Ctx) error {
		return c.Type("html").SendString(`<!DOCTYPE html><html><body>Hello</body></html>`)
	})

	app.Listen(":3000")
}
```

中间件在 `c.Next()` 之后执行，仅当响应头 `Content-Type` 包含 `text/html` 时才压缩响应体；压缩失败时保留原始正文，不中断请求。

## 配置

`HTMLMinify` 可传入可选的 `HTMLMinifyConfig`：

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `Skip` | `func(c fiber.Ctx) bool` | `nil` | 返回 `true` 时跳过压缩（例如调试、特定路由） |
| `MinifyInlineCSS` | `bool` | `false` | 是否压缩内联 CSS |
| `MinifyInlineJS` | `bool` | `false` | 是否压缩内联 JavaScript |
| `RemoveComments` | `bool` | `true` | 为 `true` 时移除普通 HTML 注释；为 `false` 时保留条件注释等特殊注释（见下方） |
| `MinifySvg` | `bool` | `false` | 是否压缩内联 SVG |

示例：跳过 `/debug`、开启内联 CSS/JS、保留注释：

```go
import (
	"strings"

	"github.com/muzidudu/go_utils/htmlminify"
)

app.Use(htmlminify.HTMLMinify(htmlminify.HTMLMinifyConfig{
	Skip: func(c fiber.Ctx) bool {
		return strings.HasPrefix(c.Path(), "/debug")
	},
	MinifyInlineCSS: true,
	MinifyInlineJS:  true,
	RemoveComments:  false,
	MinifySvg:       true,
}))
```

## 行为说明

- **执行顺序**：先执行后续处理器，再根据最终响应判断是否压缩，适合与模板渲染、错误页等配合。
- **Content-Type**：仅处理包含 `text/html` 的 `Content-Type`。
- **空响应**：正文为空时不处理。
- **注释**：`RemoveComments: true` 时移除常规注释；底层 HTML minifier 仍可通过 `KeepSpecialComments` 保留 IE 条件注释等（与 `RemoveComments` 联动）。

## 许可证

与本仓库一致；`tdewolff/minify` 遵循其自身开源协议。

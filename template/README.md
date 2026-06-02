# go_utils/template

基于 [pongo2](https://github.com/flosch/pongo2) 的多站点模板引擎，兼容 [Fiber](https://gofiber.io) 的 `Views` 接口，支持多主题（theme）按 binding 动态切换。

## 特性

- **多站点/多主题**：按 `Theme` 或 `siteBase.Template` 自动选择模板目录
- **Django 语法**：使用 pongo2 的 Django 风格模板语法
- **自定义 Tag**：支持 `RegisterTag` 注册 pongo2 自定义标签
- **内置 Filter**：16 个扩展 filter（`New` 后自动注册，见下文「内置 Filter」）
- **扩展 Filter**：支持 `RegisterFilter` / `ReplaceFilter` / `FilterExists`
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

### 4. Filter

模板中可使用两类 filter：

1. **pongo2 自带**（无需注册）：`upper`、`lower`、`escape`、`default`、`truncatewords`、`date` 等，见 [pongo2 文档](https://github.com/flosch/pongo2)。
2. **本包内置**（见下一节「内置 Filter」）：`template.New` 时加入注册队列，首次 `Load()` 时写入 pongo2。

**业务自定义 filter**：在 `Load()` 之前 `RegisterFilter`；与内置或 pongo2 同名时用 `ReplaceFilter` 覆盖：

```go
engine := template.New("views", ".django")
engine.RegisterFilter("slug", mySlugFilter)
engine.ReplaceFilter("suffix", mySuffix) // 覆盖本包内置实现
_ = engine.Load()
```

Filter 签名：`func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error)`；模板中带参数写法见下方代码块（`变量|filter名:参数`）。

## 内置 Filter

`SitesEngine` 在 `New()` 时通过 `initBuiltinFilters()` 注册下列 filter（定义在 `filters_builtin.go` 的 `packageBuiltinFilters`）。**无需在业务代码里手动注册**，创建引擎并 `Load()` 后即可在模板中使用。

> **与 pongo2 同名 filter**：pongo2 的 filter 表是进程级全局的。若名称已被 pongo2 占用（例如 `split`、`wordwrap`），本包对应实现会在 `Load` 时跳过注册；若要使用本包实现，请在 `Load` 前调用 `ReplaceFilter("split", ...)` 等覆盖。

> **README 说明**：Markdown 表格用 `|` 分列，因此**不要把 Django 模板示例写在表格里**（会被截断）。下表仅列名称与说明，示例统一放在代码块中。

### 字符串

| 名称 | 说明 |
|------|------|
| `contains` | 是否包含子串 |
| `trim` | 去两端空白；有参数时按参数字符集裁剪 |
| `trim_left` | 去左侧（`strings.TrimLeft`，参数为 cutset） |
| `trim_right` | 去右侧 |
| `replace` | 全局替换；参数为 `原串,新串`（逗号分隔） |
| `suffix` | 末尾追加；无参默认 `!` |
| `default_empty` | 仅当值为空字符串（含空白）时用默认值 |
| `truncate_chars` | 按 **rune** 截断，超出加 `...`；无参默认最长 50 |
| `repeat` | 重复字符串 N 次 |

```django
{{ "hello"|contains:"ell" }}
{{ "  hi  "|trim }}
{{ " hi "|trim_left:" " }}
{{ " hi "|trim_right:" " }}
{{ "a-b"|replace:"-","_" }}
{{ title|suffix }}
{{ title|suffix:"~" }}
{{ name|default_empty:"匿名" }}
{{ body|truncate_chars:80 }}
{{ "ab"|repeat:3 }}
```

### 拆分与列表

| 名称 | 说明 |
|------|------|
| `split` | 按分隔符拆成字符串数组；`\n` 表示换行 |
| `list` | 解析类 Python 列表字面量字符串为数组 |
| `fields` | 按空白拆分为单词列表 |

```django
{{ "a,b"|split:"," }}
{{ "a,b"|list }}
{{ "a b c"|fields }}
```

### 序列 / 调试

| 名称 | 说明 |
|------|------|
| `json` | `json.Marshal` 输入值 |
| `dump` | `%+v` 格式化（调试用） |

```django
{{ obj|json }}
{{ obj|dump }}
```

### 集合与查找

对**字符串**：`count` / `index` 使用 `strings.Count`、`strings.Index`。对**可迭代 slice**：在元素中查找匹配项。

| 名称 | 说明 |
|------|------|
| `count` | 子串出现次数，或 slice 中匹配元素个数 |
| `index` | 子串首次下标（未找到为 `-1`），或 slice 中首次匹配下标 |

```django
{{ "hello"|count:"l" }}
{{ "hello"|index:"l" }}
```

### 排版

| 名称 | 说明 |
|------|------|
| `wordwrap` | 按**单词**折行，参数为每行最多单词数 |

```django
{{ text|wordwrap:5 }}
```

### 新增内置 filter（维护者）

在 `filters_builtin.go` 中：

1. 实现 `filterXxx(in, param *pongo2.Value) (*pongo2.Value, *pongo2.Error)`
2. 在 `packageBuiltinFilters` 切片追加 `{"name", filterXxx}`
3. 同步更新本 README 表格与对应分类下的 `django` 示例代码块

`New()` → `initBuiltinFilters()` → `Load()` → `registerFilters()` 会自动完成注册。

### 5. 添加模板函数

```go
engine := template.New("views", ".django")
engine.AddFunc("upper", strings.ToUpper)
engine.AddFunc("now", func() time.Time { return time.Now() })
```

### 6. 中间件中 ViewBind 注入 Theme

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
| `RegisterFilter(name string, fn pongo2.FilterFunction) error` | 注册自定义 filter |
| `ReplaceFilter(name string, fn pongo2.FilterFunction) error` | 替换已有 filter |
| `FilterExists(name string) bool` | 判断 filter 是否已存在 |
| `AddFunc(name string, fn interface{}) *SitesEngine` | 添加模板函数 |
| `SetAutoEscape(v bool) *SitesEngine` | 设置自动转义 |
| `Load() error` | 加载所有主题模板 |
| `Render(out, name, binding, layout...) error` | 渲染模板 |

## 依赖

- `github.com/flosch/pongo2/v6`
- `github.com/gofiber/template/v2`（实现 Engine 接口）

## License

与 go_utils 项目一致。

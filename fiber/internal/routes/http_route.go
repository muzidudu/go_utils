// Package routes HTTP 页面路由（HTML、静态资源等）
package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
	"github.com/muzidudu/go_utils/fiber/internal/handlers"
	middleware "github.com/muzidudu/go_utils/fiber/internal/middleware"
	"github.com/muzidudu/go_utils/fiber/internal/sites"
)

type HTTPRoute struct{}

func NewHTTPRoute() *HTTPRoute {
	return &HTTPRoute{}
}

// InstallRouter 注册 HTTP 页面路由
func (h *HTTPRoute) InstallRouter(app *bootstrap.App) {

	f := app.Fiber
	f.Use(middleware.SiteMiddleware())
	// HTML压缩中间件（仅压缩HTML响应）
	f.Use(middleware.HTMLMinify(middleware.HTMLMinifyConfig{
		Skip: func(c fiber.Ctx) bool {
			// 跳过管理后台和API
			// return strings.HasPrefix(c.Path(), config.AppConfig.Admin.Path) || strings.HasPrefix(c.Path(), "/api")
			return false
		},

		RemoveComments:  true,
		MinifyInlineCSS: true,
		MinifyInlineJS:  true,
	}))

	// 健康检查
	f.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// 首页（Django 模板）
	f.Get("/", func(c fiber.Ctx) error {

		site := sites.GetSiteByDomain(c.Host())
		if site == nil {
			site = sites.GetDefaultSite()
		}

		return c.Render("index", fiber.Map{
			"Title":  "Fiber App",
			"MySite": site,
		}, "layouts/main")
	})

	// 用户列表（handlers 控制器）
	f.Get("/users", handlers.User.ListPage)

	// 静态文件（可选，使用 middleware/static）
	// f.Use("/static", static.New("./public"))
}

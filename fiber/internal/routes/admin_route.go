// Package routes API 路由
package routes

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
	"github.com/muzidudu/go_utils/fiber/internal/handlers"
)

type AdminRoute struct{}

func NewAdminRoute() *AdminRoute {
	return &AdminRoute{}
}

// InstallRouter 注册 API 路由
func (h *AdminRoute) InstallRouter(app *bootstrap.App) {
	f := app.Fiber
	adminPath := app.Config.Admin.AdminPath
	if adminPath == "" {
		adminPath = "/admin"
	}
	// BasicAuth 中间件配置
	basicAuthMiddleware := basicauth.New(basicauth.Config{
		Users: map[string]string{
			app.Config.Admin.Username: app.Config.Admin.Password,
		},
		Realm: app.Config.Admin.Realm,
		Authorizer: func(username, password string, _ fiber.Ctx) bool {
			return username == app.Config.Admin.Username && password == app.Config.Admin.Password
		},
	})

	// 包装 BasicAuth 中间件，跳过静态资源和 API 的认证
	adminAuthMiddleware := func(c fiber.Ctx) error {
		path := c.Path()
		// 跳过静态资源
		if strings.HasPrefix(path, adminPath+"/_app") {
			return c.Next()
		}
		// 跳过 API 请求（API 路由组会单独处理认证）
		if strings.HasPrefix(path, adminPath+"/api") {
			return c.Next()
		}
		// 其他路径需要认证
		return basicAuthMiddleware(c)
	}

	admin := f.Group(adminPath)

	admin.Use(adminAuthMiddleware)

	// 用户 API（handlers 控制器）
	admin.Get("/users", handlers.User.List)
	admin.Get("/users/:id", handlers.User.GetByID)
	admin.Post("/users", handlers.User.Create)

}

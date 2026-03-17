// Package routes HTTP 页面路由（HTML、静态资源等）
package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
)

// InstallHTTPRoutes 注册 HTTP 页面路由
func InstallHTTPRoutes(app *bootstrap.App) {
	f := app.Fiber

	// 健康检查
	f.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// 首页
	f.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, Fiber!")
	})

	// 静态文件（可选，需存在 ./public 目录）
	// f.Static("/static", "./public")
}

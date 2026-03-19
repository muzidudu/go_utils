// Package routes API 路由
package routes

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/cache"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
	"github.com/muzidudu/go_utils/fiber/internal/handlers"
)

type APIRoute struct{}

func NewAPIRoute() *APIRoute {
	return &APIRoute{}
}

// InstallRouter 注册 API 路由
func (h *APIRoute) InstallRouter(app *bootstrap.App) {
	f := app.Fiber
	api := f.Group("/api")

	// 示例：使用缓存的 API（全局 app.Cache）
	api.Get("/ping", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "pong", "time": time.Now().Format(time.RFC3339)})
	})

	api.Get("/cache/:key", func(c fiber.Ctx) error {
		key := c.Params("key")
		if key == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key required"})
		}
		val, err := app.Cache.Get(key)
		if err != nil {
			if errors.Is(err, cache.ErrNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"key": key, "value": val})
	})

	api.Post("/cache/:key", func(c fiber.Ctx) error {
		key := c.Params("key")
		if key == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "key required"})
		}
		body := c.Body()
		if err := app.Cache.Set(key, string(body), 5*time.Minute); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"ok": true, "key": key})
	})

	// 用户 API（handlers 控制器）
	api.Get("/users", handlers.User.List)
	api.Get("/users/:id", handlers.User.GetByID)
	api.Post("/users", handlers.User.Create)

	// 站点 API（sites.json 增删改查）
	api.Get("/sites", handlers.Site.List)
	api.Get("/sites/:id", handlers.Site.GetByID)
	api.Post("/sites", handlers.Site.Create)
	api.Put("/sites/:id", handlers.Site.Update)
	api.Delete("/sites/:id", handlers.Site.Delete)

	// 分类 API（多级，支持树形/扁平）
	api.Get("/categories/tree", handlers.Category.ListTree)
	api.Get("/categories/flat", handlers.Category.ListFlat)
	api.Get("/categories/:id", handlers.Category.GetByID)
	api.Post("/categories", handlers.Category.Create)
	api.Put("/categories/:id", handlers.Category.Update)
	api.Delete("/categories/:id", handlers.Category.Delete)

	api.Get("/templates", handlers.Template.ListTemplates)
}

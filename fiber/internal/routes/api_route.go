// Package routes API 路由
package routes

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/cache"
	"github.com/muzidudu/go_utils/fiber/bootstrap"
)

// InstallAPIRoutes 注册 API 路由
func InstallAPIRoutes(app *bootstrap.App) {
	f := app.Fiber
	api := f.Group("/api")

	// 示例：使用缓存的 API
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
}

package bootstrap

import (
	"github.com/gofiber/fiber/v3"
)

// initFiber 创建 Fiber 应用
func initFiber() *fiber.App {
	return fiber.New(fiber.Config{
		AppName: "Fiber App",
	})
}

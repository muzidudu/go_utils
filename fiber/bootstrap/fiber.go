package bootstrap

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/muzidudu/go_utils/template"
)

// initFiber 创建 Fiber 应用（多站点 Pongo2 模板、compress、logger 中间件）
func initFiber(cfg *Config) *fiber.App {
	// 多站点 Pongo2 模板引擎，支持 RegisterTag
	engine := template.New("./views", ".django")
	engine.Reload(true)
	engine.Debug(false)
	// 注册自定义 tag 示例：engine.RegisterTag("uppercase", template.TagUppercaseParser)
	engine.RegisterTag("uppercase", template.TagUppercaseParser)

	app := fiber.New(fiber.Config{
		AppName:           "Fiber App",
		Views:             engine,
		PassLocalsToViews: true,
	})

	// 中间件
	app.Use(logger.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	// favicon中间件
	app.Use(favicon.New())
	// 恢复中间件
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	return app
}

package sitesmiddleware

import "github.com/gofiber/fiber/v3"

// GetSite 从 Locals 读取当前站点；若无则调用 defaultSite（可为 nil）。
func GetSite[S any](c fiber.Ctx, defaultSite func(fiber.Ctx) *S) *S {
	if raw := c.Locals("site"); raw != nil {
		if s, ok := raw.(*S); ok && s != nil {
			return s
		}
	}
	if defaultSite != nil {
		return defaultSite(c)
	}
	return nil
}

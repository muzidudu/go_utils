package sitesmiddleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// DefaultHostFromRequest 从请求中获取 Host（不含端口），可与 Resolve 配合使用。
func DefaultHostFromRequest(c fiber.Ctx) string {
	host := c.Host()
	if host == "" {
		host = c.Get("Host")
	}
	host = strings.TrimSpace(strings.ToLower(host))
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

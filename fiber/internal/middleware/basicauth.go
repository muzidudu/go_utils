package middleware

import (
	"encoding/base64"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/internal/app"
)

// BasicAuthConfig BasicAuth 配置
type BasicAuthConfig struct {
	Username string
	Password string
	Realm    string
}

// BasicAuth BasicAuth 中间件
func BasicAuth(cfg BasicAuthConfig) fiber.Handler {
	if cfg.Realm == "" {
		cfg.Realm = "Restricted"
	}

	return func(c fiber.Ctx) error {
		// 获取 Authorization header
		auth := c.Get("Authorization")
		if auth == "" {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
			return c.JSON(fiber.Map{
				"error": "未授权访问",
			})
		}

		// 检查 Basic Auth 格式
		if !strings.HasPrefix(auth, "Basic ") {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
			return c.JSON(fiber.Map{
				"error": "无效的认证格式",
			})
		}

		// 解码 Base64
		encoded := strings.TrimPrefix(auth, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
			return c.JSON(fiber.Map{
				"error": "认证信息解码失败",
			})
		}

		// 解析用户名和密码
		credentials := strings.SplitN(string(decoded), ":", 2)
		if len(credentials) != 2 {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
			return c.JSON(fiber.Map{
				"error": "无效的认证信息",
			})
		}

		username := credentials[0]
		password := credentials[1]

		// 验证用户名和密码
		if username != cfg.Username || password != cfg.Password {
			c.Status(fiber.StatusUnauthorized)
			c.Set("WWW-Authenticate", `Basic realm="`+cfg.Realm+`"`)
			return c.JSON(fiber.Map{
				"error": "用户名或密码错误",
			})
		}

		// 将用户名存储到 locals
		c.Locals("admin_username", username)

		return c.Next()
	}
}

// BasicAuthFromConfig 从配置创建 BasicAuth 中间件
func BasicAuthFromConfig() fiber.Handler {
	cfg := BasicAuthConfig{
		Username: app.Config().Admin.Username,
		Password: app.Config().Admin.Password,
		Realm:    "Admin",
	}
	return BasicAuth(cfg)
}

package middleware

import (
	"bytes"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

// HTMLMinifyConfig HTML压缩配置
type HTMLMinifyConfig struct {
	// Skip 用于跳过某些路径的压缩
	Skip func(c fiber.Ctx) bool
	// MinifyInlineCSS 是否压缩内联 CSS
	MinifyInlineCSS bool
	// MinifyInlineJS 是否压缩内联 JavaScript
	MinifyInlineJS bool
	// RemoveComments 是否移除 HTML 注释
	RemoveComments bool
}

// HTMLMinify HTML压缩中间件
// 使用 github.com/tdewolff/minify/v2 进行高效压缩
func HTMLMinify(config ...HTMLMinifyConfig) fiber.Handler {
	cfg := HTMLMinifyConfig{
		MinifyInlineCSS: false,
		MinifyInlineJS:  false,
		RemoveComments:  true,
	}

	if len(config) > 0 {
		cfg = config[0]
	}

	// 创建 minify 实例
	m := minify.New()

	// 配置 HTML minifier（支持保留特殊注释）
	htmlMinifier := &html.Minifier{
		KeepSpecialComments: !cfg.RemoveComments, // 保留条件注释等特殊注释
		KeepDefaultAttrVals: true,
	}
	m.Add("text/html", htmlMinifier)

	// 配置 CSS minifier（如果启用）
	if cfg.MinifyInlineCSS {
		m.AddFunc("text/css", css.Minify)
	}

	// 配置 JS minifier（如果启用）
	if cfg.MinifyInlineJS {
		m.AddFunc("text/javascript", js.Minify)
		m.AddFunc("application/javascript", js.Minify)
	}

	return func(c fiber.Ctx) error {
		// 执行下一个处理器
		if err := c.Next(); err != nil {
			return err
		}

		// 检查是否跳过压缩
		if cfg.Skip != nil && cfg.Skip(c) {
			return nil
		}

		// 检查 Content-Type 是否为 HTML
		contentType := c.GetRespHeader("Content-Type")
		if !strings.Contains(contentType, "text/html") {
			return nil
		}

		// 获取响应体
		body := c.Response().Body()
		if len(body) == 0 {
			return nil
		}

		// 压缩 HTML
		var minified bytes.Buffer
		reader := bytes.NewReader(body)

		// 使用 minify 库压缩
		if err := m.Minify("text/html", &minified, reader); err != nil {
			// 如果压缩失败，返回原始内容
			return nil
		}

		// 更新响应体
		c.Response().SetBody(minified.Bytes())

		return nil
	}
}

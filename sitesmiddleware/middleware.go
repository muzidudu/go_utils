package sitesmiddleware

import (
	"github.com/gofiber/fiber/v3"
)

// New 返回站点中间件：按域名匹配站点、解析路径参数、写入 Locals / ViewBind。
func New[S any](opts Options[S]) fiber.Handler {
	if opts.Resolve == nil {
		// 如果Resolve为空，则panic
		panic("sitesmiddleware: Options.Resolve is required")
	}
	b := opts.SiteBinding
	// 返回站点中间件
	return func(c fiber.Ctx) error {
		// 获取User-Agent
		userAgent := c.Get("User-Agent")
		// 判断是否为搜索引擎爬虫
		isBot := IsSearchEngineBot(c, userAgent, opts.BotUserAgents)

		// 垃圾/采集类爬虫拦截
		if !opts.SkipRubbishBotBlock {
			// 获取垃圾UA列表
			rubbishList := rubbishListForRequest(c, opts.RubbishBotUserAgents)
			// 如果UA命中垃圾列表，则返回403
			if IsRubbishBotWithList(userAgent, rubbishList) {
				return c.Status(fiber.StatusForbidden).SendString("Access denied")
			}
		}

		c.Locals("isBot", isBot)
		c.ViewBind(fiber.Map{
			"isBot": isBot,
		})
		// 非爬虫且未跳过语言检查时，检查Accept-Language
		if !isBot && !opts.SkipLanguageCheck {
			// 检查Accept-Language是否在允许列表中
			allowed := opts.PreferredLanguageTags
			if len(allowed) == 0 {
				// 如果允许列表为空，则使用默认允许列表
				allowed = DefaultPreferredLanguageTags
			}
			// 检查Accept-Language是否在允许列表中
			if !IsPreferredLanguage(c.AcceptLanguage(), allowed) {
				// 如果不在允许列表中，则返回401
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
			}
		}

		// 解析站点
		host, site := opts.Resolve(c)

		if site != nil {
			// 如果站点不为空，则写入Locals
			if b.SetHost == nil || b.SiteID == nil || b.Template == nil {
				// 如果SiteBinding的SetHost, SiteID和Template为空，则panic
				panic("sitesmiddleware: Options.SiteBinding.SetHost, SiteID and Template are required when site is set")
			}
			b.SetHost(site, host)
			// 获取站点ID
			siteID := b.SiteID(site)
			// 获取站点模板
			tmpl := b.Template(site)
			// 写入Locals
			c.Locals("ctx", c)
			// 写入site
			c.Locals("siteBase", site)
			// 写入siteID
			c.Locals("siteID", siteID)
			// 写入host
			c.Locals("host", host)
			// 写入ThemePath
			c.Locals("ThemePath", "template/"+tmpl)
			// 写入Theme
			c.Locals("Theme", tmpl)

			// 写入ViewBind
			binding := fiber.Map{
				"ctx":       c,
				"siteBase":  site,
				"siteID":    siteID,
				"ThemePath": "template/" + tmpl,
				"Theme":     tmpl,
			}
			c.ViewBind(binding)
		}

		return c.Next()
	}
}

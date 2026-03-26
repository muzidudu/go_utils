package sitesmiddleware

import "github.com/gofiber/fiber/v3"

// SiteBinding 描述如何从站点值上读写中间件需要的字段（避免依赖具体模型类型）
type SiteBinding[S any] struct {
	SetHost  func(s *S, host string)
	SiteID   func(s *S) uint
	Template func(s *S) string
}

// Options 站点中间件依赖；由应用层注入解析与工具函数。
type Options[S any] struct {
	// SiteBinding 在 Resolve 可能返回非 nil 站点时必须填写完整（SetHost / SiteID / Template）
	SiteBinding SiteBinding[S]

	// BotUserAgents 可选：返回自定义爬虫 UA 列表；nil 或空切片则使用 DefaultBotUserAgents
	BotUserAgents func(c fiber.Ctx) []string

	// SkipRubbishBotBlock 为 true 时不拦截垃圾 UA（不返回 403）
	SkipRubbishBotBlock bool

	// RubbishBotUserAgents 可选：返回用于垃圾 UA 匹配的子串列表；nil 或空切片则使用 DefaultRubbishBotSubstrings
	RubbishBotUserAgents func(c fiber.Ctx) []string

	// SkipLanguageCheck 为 true 时，非爬虫不校验 Accept-Language
	SkipLanguageCheck bool

	// PreferredLanguageTags 在非爬虫且未跳过语言检查时使用：首选语言主标签须在此列表中。
	// nil 或空切片则使用 DefaultPreferredLanguageTags（["zh"]）
	PreferredLanguageTags []string

	// Resolve 必须：返回当前请求的 host 与按域名解析出的站点（无匹配时可返回默认站点）
	Resolve func(c fiber.Ctx) (host string, site *S)
}

package middleware

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/internal/app"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/sites"
	"github.com/muzidudu/go_utils/fiber/pkg/utils"
)

// defaultBotUserAgents 默认搜索引擎爬虫列表
var defaultBotUserAgents = []string{
	"googlebot", "bingbot", "baiduspider", "yandexbot",
	"yisouspider", "360spider", "spider", "ucbrowser",
}

// isSearchEngineBot 检查请求是否来自搜索引擎爬虫
func isSearchEngineBot(userAgent string) bool {
	if userAgent == "" {
		return false
	}
	userAgent = strings.ToLower(userAgent)
	botUserAgents := defaultBotUserAgents
	if cfg := app.Config(); cfg != nil && len(cfg.Site.BotUserAgents) > 0 {
		botUserAgents = cfg.Site.BotUserAgents
	}
	for _, bot := range botUserAgents {
		if strings.Contains(userAgent, strings.ToLower(bot)) {
			return true
		}
	}

	return false
}

func isRubbishBot(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	botUserAgents := []string{
		"PetalBot",
		"DataForSeoBot",
		"SemrushBot",
		"DotBot",
		"MJ12bot",
		"AhrefsBot",
		"MauiBot",
		"BLEXBot",
		"ZoominfoBot",
		"ExtLinksBot",
		"hubspot",
		"leiki",
		"webmeup",
		"psbot",
		"GPTBot",
		"anthropic-ai",
		"Google-Extended",
		"ChatGPT-User",
		"PerplexityBot",
		"cohere-ai",
		"CCBot",
		"omgili",
		"omgilibot",
		"FacebookBot",
		"Twitterbot",
	}
	if len(botUserAgents) == 0 {
		return false
	}

	for _, bot := range botUserAgents {
		if strings.Contains(userAgent, strings.ToLower(bot)) {
			return true
		}
	}

	userAgent = strings.ToLower(userAgent)

	return false
}

// isChinesePreferredLanguage 检查 Accept-Language 首选项是否为中文
func isChinesePreferredLanguage(acceptLanguage string) bool {
	if acceptLanguage == "" {
		// 如果没有 Accept-Language 头，默认不允许通过
		return false
	}

	// Accept-Language 格式示例: "zh-CN,zh;q=0.9,en;q=0.8"
	// 按逗号分割语言标签
	languages := strings.Split(acceptLanguage, ",")
	if len(languages) == 0 {
		return false
	}

	// 获取第一个语言标签（优先级最高的）
	firstLang := strings.TrimSpace(languages[0])
	// 移除质量值（q=0.9 等）
	if idx := strings.Index(firstLang, ";"); idx != -1 {
		firstLang = firstLang[:idx]
	}
	// 移除区域代码（如 -CN）
	if idx := strings.Index(firstLang, "-"); idx != -1 {
		firstLang = firstLang[:idx]
	}

	// 检查是否为中文（zh）
	firstLang = strings.ToLower(firstLang)
	return firstLang == "zh"
}

// SiteMiddleware 站点中间件，根据域名匹配站点
// 支持多个域名和泛域名匹配
func SiteMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 检查是否是搜索引擎爬虫
		userAgent := c.Get("User-Agent")
		isBot := isSearchEngineBot(userAgent)
		// 检查是否是垃圾爬虫
		isRubbish := isRubbishBot(userAgent)
		if isRubbish {
			return c.Status(fiber.StatusForbidden).SendString("Access denied")
		}

		c.Locals("isBot", isBot)
		c.ViewBind(fiber.Map{
			"isBot": isBot,
		})

		// 如果不是爬虫，检查语言首选项
		if !isBot {
			acceptLanguage := c.AcceptLanguage()
			if !isChinesePreferredLanguage(acceptLanguage) {
				// 首选语言不是中文，拒绝请求
				return c.Status(fiber.StatusUnauthorized).SendString("Unauthorized")
			}
		}

		// 获取请求的 Host
		host := sites.GetHostFromRequest(c)
		site := sites.GetSiteByDomain(host)
		if site == nil {
			site = sites.GetDefaultSite()
		}

		// 注意：由于中间件在路由匹配之前执行，无法使用 c.Params() 获取路由参数
		// 需要手动解析路径来获取参数
		path := c.Path()

		// 手动解析路径参数
		var hash, typeParam, keyword, pageParam string
		var page int = 1

		// 解析 /type/:type/:page? 路由
		// 例如：/type/abc 或 /type/abc/2
		if strings.HasPrefix(path, "/type/") {
			parts := strings.Split(strings.TrimPrefix(path, "/type/"), "/")
			if len(parts) > 0 {
				typeParam = parts[0]
			}
			if len(parts) > 1 && parts[1] != "" {
				pageParam = parts[1]
				if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
					page = p
				}
			}
		}

		// 解析 /detail/:hash 路由
		// 例如：/detail/abc123
		if strings.HasPrefix(path, "/detail/") {
			hash = strings.TrimPrefix(path, "/detail/")
		}

		// 解析 /search/:keyword/:page? 路由
		// 例如：/search/关键词 或 /search/关键词/2
		if strings.HasPrefix(path, "/search/") {
			parts := strings.Split(strings.TrimPrefix(path, "/search/"), "/")
			if len(parts) > 0 {
				keyword = parts[0]
			}
			if len(parts) > 1 && parts[1] != "" {
				pageParam = parts[1]
				if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
					page = p
				}
			}
		}

		// 如果路径参数中没有 page，尝试从查询参数获取
		if pageParam == "" {
			if queryPage := fiber.Query[int](c, "page", 0); queryPage > 0 {
				page = queryPage
			}
		}

		// 获取分类ID（如果有 type 参数，且 site 存在）
		var categoryID int64
		if site != nil && typeParam != "" {
			categoryID = utils.Hash2ID(typeParam, int64(site.ID), 0)
		}

		// 尝试使用 c.Params() 获取参数（如果路由已匹配）
		if site != nil && c.Params("type") != "" {
			typeParam = c.Params("type")
			categoryID = utils.Hash2ID(typeParam, int64(site.ID), 0)
		}
		if c.Params("hash") != "" {
			hash = c.Params("hash")
		}
		if c.Params("keyword") != "" {
			keyword = c.Params("keyword")
		}
		if c.Params("page") != "" {
			if p, err := strconv.Atoi(c.Params("page")); err == nil && p > 0 {
				page = p
			}
		}

		// 获取查询字符串
		queryString := c.OriginalURL()
		if idx := strings.Index(queryString, path); idx != -1 {
			queryString = queryString[idx+len(path):]
		}

		// 构建 page 对象并设置到模板上下文
		pageInfo := fiber.Map{
			"page":    page,
			"hash":    hash,
			"type":    typeParam,
			"keyword": keyword,
			"path":    path,
			"query":   queryString,
		}

		// 将站点信息存储到上下文中
		if site != nil {
			// 设置当前请求的Host（用于区分实际访问的域名）
			site.Host = host
			c.Locals("ctx", c)
			c.Locals("site", site)
			c.Locals("siteID", site.ID)
			c.Locals("host", host)
			c.Locals("ThemePath", "template/"+site.Template)
			c.Locals("Theme", site.Template)
			// fmt.Println("[middleware] site.ID", site.ID)

			// seed := utils.PathToSeed(c.OriginalURL())
			// keywords := utils.GetRandomWords(config.AppConfig.Site.SiteKeywords, 40, false, int64(site.ID), seed)

			binding := fiber.Map{
				"ctx":        c,
				"siteBase":   site,
				"siteID":     site.ID,
				"ThemePath":  "template/" + site.Template,
				"Theme":      site.Template,
				"params":     pageInfo,
				"categoryID": categoryID,
			}
			c.ViewBind(binding)
		}

		return c.Next()
	}
}

func getDomainNames(c fiber.Ctx) []fiber.Map {
	site := GetSite(c)
	if site == nil {
		return nil
	}
	seed := utils.PathToSeed(c.OriginalURL())
	allSites := sites.GetAllSites()

	type domainInfo struct {
		domain string
		siteID uint
	}
	domainInfos := make([]domainInfo, 0)
	for _, s := range allSites {
		if s.Domain != "" && !strings.HasPrefix(s.Domain, "*") {
			domainInfos = append(domainInfos, domainInfo{domain: s.Domain, siteID: s.ID})
		}
		for _, d := range s.Subdomains {
			d = strings.TrimSpace(d)
			if d != "" && !strings.HasPrefix(d, "*") {
				domainInfos = append(domainInfos, domainInfo{domain: d, siteID: s.ID})
			}
		}
	}

	randomNames := defaultBotUserAgents // 占位，无 SiteWords 时用默认
	if cfg := app.Config(); cfg != nil && len(cfg.Site.SiteWords) > 0 {
		randomNames = utils.GetRandomWords(cfg.Site.SiteWords, len(domainInfos), false, int64(site.ID), seed)
		if len(randomNames) == 0 {
			randomNames = cfg.Site.SiteWords
		}
	}

	// 组合域名、siteid 和随机名称
	domainNames := make([]fiber.Map, 0)
	for i, info := range domainInfos {
		name := ""
		if i < len(randomNames) {
			name = randomNames[i]
		}
		domainNames = append(domainNames, fiber.Map{
			"domain": info.domain,
			"url":    "https://" + info.domain,
			"id":     info.siteID,
			"name":   name,
		})
	}
	return domainNames
}

// GetSite 从上下文获取当前站点
func GetSite(c fiber.Ctx) *models.Site {
	if site, ok := c.Locals("site").(*models.Site); ok {
		return site
	}
	return sites.GetDefaultSite()
}

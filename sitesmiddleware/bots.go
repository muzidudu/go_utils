package sitesmiddleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// DefaultBotUserAgents 默认搜索引擎爬虫列表
var DefaultBotUserAgents = []string{
	"googlebot", "bingbot", "baiduspider", "yandexbot",
	"yisouspider", "360spider", "spider", "ucbrowser",
}

// DefaultRubbishBotSubstrings 默认垃圾/采集类 UA 子串列表（用于包含匹配）
var DefaultRubbishBotSubstrings = []string{
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

func botListForRequest(c fiber.Ctx, botUserAgents func(fiber.Ctx) []string) []string {
	if botUserAgents != nil {
		if list := botUserAgents(c); len(list) > 0 {
			return list
		}
	}
	return DefaultBotUserAgents
}

func rubbishListForRequest(c fiber.Ctx, rubbishBotUserAgents func(fiber.Ctx) []string) []string {
	if rubbishBotUserAgents != nil {
		if list := rubbishBotUserAgents(c); len(list) > 0 {
			return list
		}
	}
	return DefaultRubbishBotSubstrings
}

// IsSearchEngineBot 检查请求是否来自搜索引擎爬虫
func IsSearchEngineBot(c fiber.Ctx, userAgent string, botUserAgents func(fiber.Ctx) []string) bool {
	if userAgent == "" {
		return false
	}
	userAgent = strings.ToLower(userAgent)
	bots := botListForRequest(c, botUserAgents)
	for _, bot := range bots {
		if strings.Contains(userAgent, strings.ToLower(bot)) {
			return true
		}
	}
	return false
}

// IsRubbishBotWithList 检查 UA 是否命中任一子串（大小写不敏感）
func IsRubbishBotWithList(userAgent string, substrings []string) bool {
	if userAgent == "" || len(substrings) == 0 {
		return false
	}
	uaLower := strings.ToLower(userAgent)
	for _, s := range substrings {
		if s == "" {
			continue
		}
		if strings.Contains(uaLower, strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// IsRubbishBot 检查是否为垃圾/采集类爬虫（使用 DefaultRubbishBotSubstrings）
func IsRubbishBot(userAgent string) bool {
	return IsRubbishBotWithList(userAgent, DefaultRubbishBotSubstrings)
}

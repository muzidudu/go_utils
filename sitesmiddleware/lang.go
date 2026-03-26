package sitesmiddleware

import "strings"

// DefaultPreferredLanguageTags 未指定 PreferredLanguageTags 时使用的默认首选语言主标签（ISO 639-1）
var DefaultPreferredLanguageTags = []string{"zh"}

// PrimaryLanguageTag 从 Accept-Language 取出第一个语言标签的主语言部分（小写），空串表示无法解析
func PrimaryLanguageTag(acceptLanguage string) string {
	if acceptLanguage == "" {
		return ""
	}
	languages := strings.Split(acceptLanguage, ",")
	if len(languages) == 0 {
		return ""
	}
	firstLang := strings.TrimSpace(languages[0])
	if idx := strings.Index(firstLang, ";"); idx != -1 {
		firstLang = firstLang[:idx]
	}
	if idx := strings.Index(firstLang, "-"); idx != -1 {
		firstLang = firstLang[:idx]
	}
	return strings.ToLower(strings.TrimSpace(firstLang))
}

// IsPreferredLanguage 检查 Accept-Language 首选语言的主标签是否在 allowed 中（忽略大小写）
func IsPreferredLanguage(acceptLanguage string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	primary := PrimaryLanguageTag(acceptLanguage)
	if primary == "" {
		return false
	}
	for _, a := range allowed {
		if strings.ToLower(strings.TrimSpace(a)) == primary {
			return true
		}
	}
	return false
}

// IsChinesePreferredLanguage 检查 Accept-Language 首选项是否为中文
func IsChinesePreferredLanguage(acceptLanguage string) bool {
	return IsPreferredLanguage(acceptLanguage, []string{"zh"})
}

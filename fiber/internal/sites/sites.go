// Package sites 站点相关方法（基于数据库）
package sites

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
)

// GetHostFromRequest 从请求中获取 Host（不含端口）
func GetHostFromRequest(c fiber.Ctx) string {
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

// GetSiteByDomain 根据域名匹配站点
func GetSiteByDomain(host string) *models.Site {
	site, err := repository.Site.GetByDomain(host)
	if err != nil || site == nil {
		return nil
	}
	return site
}

// GetDefaultSite 获取默认站点
func GetDefaultSite() *models.Site {
	site, err := repository.Site.GetDefault()
	if err != nil || site == nil {
		return nil
	}
	return site
}

// GetAllSites 获取所有启用站点
func GetAllSites() []*models.Site {
	list, err := repository.Site.List()
	if err != nil || len(list) == 0 {
		return nil
	}
	result := make([]*models.Site, 0, len(list))
	for i := range list {
		if list[i].Status != 0 {
			result = append(result, &list[i])
		}
	}
	return result
}

// Package models 数据模型
package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Site 站点模型
type Site struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	Domain     string         `gorm:"size:255;uniqueIndex;not null" json:"domain"`
	Bind       uint           `gorm:"default:0;index" json:"bind"` // 绑定分类菜单ID
	Subdomains []string       `gorm:"serializer:json" json:"subdomains"`
	Template   string         `gorm:"size:50;default:default" json:"template"`
	IsDefault  bool           `gorm:"default:false" json:"is_default"`
	Status     int            `gorm:"default:1" json:"status"`
	Host       string         `gorm:"-" json:"-"` // 运行时：实际请求的 Host
}

// TableName 表名
func (Site) TableName() string {
	return "sites"
}

// MatchDomain 检查 host 是否匹配该站点的 domain 或 subdomains
func (s *Site) MatchDomain(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return false
	}
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if s.Domain != "" {
		domain := strings.TrimSpace(strings.ToLower(s.Domain))
		if domain != "" && !strings.HasPrefix(domain, "*") {
			if host == domain || strings.HasSuffix(host, "."+domain) {
				return true
			}
		}
	}
	for _, sub := range s.Subdomains {
		sub = strings.TrimSpace(strings.ToLower(sub))
		if sub != "" && !strings.HasPrefix(sub, "*") {
			if host == sub || strings.HasSuffix(host, "."+sub) {
				return true
			}
		}
	}
	return false
}

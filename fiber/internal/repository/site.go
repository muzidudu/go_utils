// Package repository 数据访问层
package repository

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/muzidudu/go_utils/cache"
	"github.com/muzidudu/go_utils/fiber/internal/app"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"gorm.io/gorm"
)

const (
	siteListKey  = "site:list"
	siteCacheTTL = 12 * time.Hour
)

func siteIDKey(id uint) string { return fmt.Sprintf("site:id:%d", id) }

var (
	siteMemCache  cache.Cache
	siteCacheOnce sync.Once
)

// siteCache 站点专用缓存，默认使用内存缓存
func siteCache() cache.Cache {
	siteCacheOnce.Do(func() {
		siteMemCache = cache.NewMemoryCache(cache.MemoryConfig{
			MaxCount: 500,
			MaxBytes: 2 * 1024 * 1024, // 2MB
		})
	})
	return siteMemCache
}

// SiteRepository 站点仓储
type SiteRepository struct{}

// Site 站点仓储实例
var Site = &SiteRepository{}

// List 获取站点列表（含已禁用），带内存缓存
func (r *SiteRepository) List() ([]models.Site, error) {
	c := siteCache()
	var list []models.Site
	if err := c.GetInto(siteListKey, &list); err == nil {
		return list, nil
	}
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	if err := db.Order("id").Find(&list).Error; err != nil {
		return nil, err
	}
	_ = c.Set(siteListKey, list, siteCacheTTL)
	return list, nil
}

// GetByID 根据 ID 获取站点，带内存缓存
func (r *SiteRepository) GetByID(id uint) (*models.Site, error) {
	c := siteCache()
	var s models.Site
	if err := c.GetInto(siteIDKey(id), &s); err == nil {
		return &s, nil
	}
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	if err := db.First(&s, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	_ = c.Set(siteIDKey(id), &s, siteCacheTTL)
	return &s, nil
}

// invalidateSiteCache 增删改后清除站点内存缓存
func invalidateSiteCache(id uint) {
	c := siteCache()
	_ = c.Delete(siteListKey)
	if id > 0 {
		_ = c.Delete(siteIDKey(id))
	}
}

// GetByDomain 根据域名获取站点（精确匹配 domain 或 subdomains）
func (r *SiteRepository) GetByDomain(host string) (*models.Site, error) {
	list, err := r.List()
	if err != nil {
		return nil, err
	}
	host = trimHost(host)
	for i := range list {
		if list[i].Status != 0 && list[i].MatchDomain(host) {
			return &list[i], nil
		}
	}
	return nil, nil
}

// GetDefault 获取默认站点
func (r *SiteRepository) GetDefault() (*models.Site, error) {
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	var s models.Site
	if err := db.Where("is_default = ? AND status != 0", true).First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 无默认时返回第一个
			if err := db.Where("status != 0").Order("id").First(&s).Error; err != nil {
				return nil, nil
			}
		} else {
			return nil, err
		}
	}
	return &s, nil
}

// Create 创建站点
func (r *SiteRepository) Create(s *models.Site) error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Create(s).Error; err != nil {
		return err
	}
	invalidateSiteCache(0) // 列表变更
	return nil
}

// Update 更新站点
func (r *SiteRepository) Update(s *models.Site) error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Save(s).Error; err != nil {
		return err
	}
	invalidateSiteCache(s.ID)
	return nil
}

// Delete 删除站点（软删除）
func (r *SiteRepository) Delete(id uint) error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Delete(&models.Site{}, id).Error; err != nil {
		return err
	}
	invalidateSiteCache(id)
	return nil
}

// ClearDefault 清除所有默认标记
func (r *SiteRepository) ClearDefault() error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Model(&models.Site{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return err
	}
	invalidateSiteCache(0)
	return nil
}

func trimHost(host string) string {
	host = strings.TrimSpace(strings.ToLower(host))
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

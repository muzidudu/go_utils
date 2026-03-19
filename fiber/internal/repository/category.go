// Package repository 数据访问层
package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/muzidudu/go_utils/cache"
	"github.com/muzidudu/go_utils/fiber/internal/app"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"gorm.io/gorm"
)

const (
	categoryAllKey   = "category:all"
	categoryCacheTTL = 12 * time.Hour
)

func categoryIDKey(id uint) string { return fmt.Sprintf("category:id:%d", id) }

var (
	categoryMemCache  cache.Cache
	categoryCacheOnce sync.Once
)

func categoryCache() cache.Cache {
	categoryCacheOnce.Do(func() {
		categoryMemCache = cache.NewMemoryCache(cache.MemoryConfig{
			MaxCount: 500,
			MaxBytes: 2 * 1024 * 1024, // 2MB
		})
	})
	return categoryMemCache
}

func invalidateCategoryCache(id uint) {
	c := categoryCache()
	_ = c.Delete(categoryAllKey)
	if id > 0 {
		_ = c.Delete(categoryIDKey(id))
	}
}

// CategoryRepository 分类仓储
type CategoryRepository struct{}

// Category 分类仓储实例
var Category = &CategoryRepository{}

// ListFlat 获取扁平列表，parentID=0 返回根级，parentID>0 返回该分类的直接子级（带内存缓存）
func (r *CategoryRepository) ListFlat(parentID uint) ([]models.Category, error) {
	list, err := r.ListAll()
	if err != nil {
		return nil, err
	}
	var result []models.Category
	for i := range list {
		if list[i].ParentID == parentID {
			result = append(result, list[i])
		}
	}
	return result, nil
}

// ListAll 获取全部扁平列表（带内存缓存）
func (r *CategoryRepository) ListAll() ([]models.Category, error) {
	c := categoryCache()
	var list []models.Category
	if err := c.GetInto(categoryAllKey, &list); err == nil {
		return list, nil
	}
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	if err := db.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	_ = c.Set(categoryAllKey, list, categoryCacheTTL)
	return list, nil
}

// ListTree 获取树形结构，parentID=0 返回完整树，parentID>0 返回以该 id 为根的子树（带内存缓存）
func (r *CategoryRepository) ListTree(parentID uint) ([]*models.Category, error) {
	list, err := r.ListAll()
	if err != nil {
		return nil, err
	}
	return buildCategoryTree(list, parentID), nil
}

// buildCategoryTree 将扁平列表转为树形
func buildCategoryTree(list []models.Category, parentID uint) []*models.Category {
	var tree []*models.Category
	for i := range list {
		if list[i].ParentID == parentID {
			item := &list[i]
			item.Children = buildCategoryTree(list, item.ID)
			tree = append(tree, item)
		}
	}
	return tree
}

// GetByID 根据 ID 获取（带内存缓存）
func (r *CategoryRepository) GetByID(id uint) (*models.Category, error) {
	c := categoryCache()
	var m models.Category
	if err := c.GetInto(categoryIDKey(id), &m); err == nil {
		return &m, nil
	}
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	if err := db.First(&m, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	_ = c.Set(categoryIDKey(id), &m, categoryCacheTTL)
	return &m, nil
}

// Create 创建
func (r *CategoryRepository) Create(m *models.Category) error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Create(m).Error; err != nil {
		return err
	}
	invalidateCategoryCache(0)
	return nil
}

// Update 更新
func (r *CategoryRepository) Update(m *models.Category) error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Save(m).Error; err != nil {
		return err
	}
	invalidateCategoryCache(m.ID)
	return nil
}

// Delete 删除（软删除）
func (r *CategoryRepository) Delete(id uint) error {
	db := app.DB()
	if db == nil {
		return ErrNoDB
	}
	if err := db.Delete(&models.Category{}, id).Error; err != nil {
		return err
	}
	invalidateCategoryCache(id)
	return nil
}

// Package models 数据模型
package models

import (
	"time"
)

// Category 分类模型，支持多级
type Category struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ParentID uint   `gorm:"default:0;index" json:"parent_id"` // 0 表示根级
	Name     string `gorm:"size:100;not null" json:"name"`
	Slug     string `gorm:"size:255;index" json:"slug"`
	Type     string `gorm:"size:20;default:category" json:"type"` // category | link
	Link     string `gorm:"size:500" json:"link"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Status   int    `gorm:"default:1" json:"status"` // 0 禁用 1 启用

	Children []*Category `gorm:"-" json:"children,omitempty"` // 子分类，不持久化
}

// TableName 表名
func (Category) TableName() string {
	return "categories"
}

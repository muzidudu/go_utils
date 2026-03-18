// Package models 数据模型
package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

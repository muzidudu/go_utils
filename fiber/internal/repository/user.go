// Package repository 数据访问层
package repository

import (
	"github.com/muzidudu/go_utils/fiber/internal/app"
	"github.com/muzidudu/go_utils/fiber/internal/models"
)

// UserRepository 用户仓储
type UserRepository struct{}

// User 用户仓储实例
var User = &UserRepository{}

// GetByID 根据 ID 获取用户
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	var u models.User
	if err := db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// List 获取用户列表
func (r *UserRepository) List(limit int) ([]models.User, error) {
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	if limit <= 0 {
		limit = 10
	}
	var users []models.User
	if err := db.Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Create 创建用户
func (r *UserRepository) Create(name, email string) (*models.User, error) {
	db := app.DB()
	if db == nil {
		return nil, ErrNoDB
	}
	u := &models.User{Name: name, Email: email}
	if err := db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

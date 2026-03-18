// Package service 业务逻辑处理
package service

import (
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
)

// UserService 用户业务
type UserService struct{}

// User 用户服务实例
var User = &UserService{}

// GetByID 根据 ID 获取用户
func (s *UserService) GetByID(id uint) (*models.User, error) {
	return repository.User.GetByID(id)
}

// List 获取用户列表（业务层可加过滤、排序等逻辑）
func (s *UserService) List(limit int) ([]models.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	return repository.User.List(limit)
}

// Create 创建用户（业务校验：邮箱格式、重名等）
func (s *UserService) Create(name, email string) (*models.User, error) {
	// 业务校验示例
	if name == "" || email == "" {
		return nil, ErrInvalidInput
	}
	return repository.User.Create(name, email)
}

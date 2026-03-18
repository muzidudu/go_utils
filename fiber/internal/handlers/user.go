// Package handlers 接收/解析请求，参数校验，调用 Service
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/internal/dto"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
	"github.com/muzidudu/go_utils/fiber/internal/service"
)

// UserHandler 用户控制器
type UserHandler struct{}

// User 用户控制器实例
var User = &UserHandler{}

// GetByID 根据 ID 获取用户
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	u, err := service.User.GetByID(uint(id))
	if err != nil {
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(modelToResp(u))
}

// List 获取用户列表（JSON）
func (h *UserHandler) List(c fiber.Ctx) error {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	users, err := service.User.List(limit)
	if err != nil {
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	resp := make([]dto.UserResp, len(users))
	for i := range users {
		resp[i] = *modelToResp(&users[i])
	}
	return c.JSON(resp)
}

// ListPage 用户列表页面（模板渲染）
func (h *UserHandler) ListPage(c fiber.Ctx) error {
	users, err := service.User.List(20)
	if err != nil && err != repository.ErrNoDB {
		return c.Status(500).SendString(err.Error())
	}
	return c.Render("users/index", fiber.Map{
		"users": users,
	}, "layouts/main")
}

// Create 创建用户
func (h *UserHandler) Create(c fiber.Ctx) error {
	var req dto.CreateUserReq
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	// 参数校验
	if req.Name == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name and email required"})
	}
	u, err := service.User.Create(req.Name, req.Email)
	if err != nil {
		if err == service.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(modelToResp(u))
}

func modelToResp(u *models.User) *dto.UserResp {
	if u == nil {
		return nil
	}
	return &dto.UserResp{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

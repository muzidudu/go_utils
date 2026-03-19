// Package handlers 接收/解析请求，参数校验，调用 Service
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/internal/dto"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
	"github.com/muzidudu/go_utils/fiber/internal/service"
)

// CategoryHandler 分类 API 控制器
type CategoryHandler struct{}

// Category 分类控制器实例
var Category = &CategoryHandler{}

// ListTree 获取树形分类，?parent_id=0 完整树，?parent_id=5 以 id=5 为根的子树
func (h *CategoryHandler) ListTree(c fiber.Ctx) error {
	parentID := parseParentID(c)
	tree, err := service.Categories.ListTree(parentID)
	if err != nil {
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(tree)
}

// ListFlat 获取扁平列表，?parent_id=0 根级，?parent_id=5 该分类的直接子级
func (h *CategoryHandler) ListFlat(c fiber.Ctx) error {
	parentID := parseParentID(c)
	list, err := service.Categories.ListFlat(parentID)
	if err != nil {
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(list)
}

// GetByID 根据 ID 获取
func (h *CategoryHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	cat, err := service.Categories.GetByID(uint(id))
	if err != nil {
		if err == service.ErrCategoryNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cat)
}

// Create 创建
func (h *CategoryHandler) Create(c fiber.Ctx) error {
	var req dto.CreateCategoryReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	cat, err := service.Categories.Create(req)
	if err != nil {
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(cat)
}

// Update 更新
func (h *CategoryHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req dto.UpdateCategoryReq
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	cat, err := service.Categories.Update(uint(id), req)
	if err != nil {
		if err == service.ErrCategoryNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(cat)
}

// Delete 删除
func (h *CategoryHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := service.Categories.Delete(uint(id)); err != nil {
		if err == service.ErrCategoryNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "category not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseParentID(c fiber.Ctx) uint {
	s := c.Query("id")
	if s == "" {
		s = c.Query("parent_id")
	}
	if s == "" {
		return 0
	}
	id, _ := strconv.ParseUint(s, 10, 32)
	return uint(id)
}

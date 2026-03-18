// Package handlers 接收/解析请求，参数校验，调用 Service
package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/muzidudu/go_utils/fiber/internal/dto"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
	"github.com/muzidudu/go_utils/fiber/internal/service"
)

// SiteHandler 站点 API 控制器
type SiteHandler struct{}

// Site 站点控制器实例
var Site = &SiteHandler{}

// List 获取站点列表
func (h *SiteHandler) List(c fiber.Ctx) error {
	sites, err := service.Sites.List()
	if err != nil {
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(sites)
}

// GetByID 根据 ID 获取站点
func (h *SiteHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	site, err := service.Sites.GetByID(uint(id))
	if err != nil {
		if err == service.ErrSiteNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "site not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(site)
}

// Create 创建站点
func (h *SiteHandler) Create(c fiber.Ctx) error {
	var req dto.CreateSiteReq
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Domain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "domain required"})
	}
	site, err := service.Sites.Create(req)
	if err != nil {
		if err == service.ErrSiteDuplicateDomain {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(site)
}

// Update 更新站点
func (h *SiteHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	var req dto.UpdateSiteReq
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	site, err := service.Sites.Update(uint(id), req)
	if err != nil {
		if err == service.ErrSiteNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "site not found"})
		}
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(site)
}

// Delete 删除站点
func (h *SiteHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	if err := service.Sites.Delete(uint(id)); err != nil {
		if err == service.ErrSiteNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "site not found"})
		}
		if err == repository.ErrNoDB {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

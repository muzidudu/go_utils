package handlers

import (
	"os"

	"github.com/gofiber/fiber/v3"
)

type TemplateHandler struct{}

var Template = &TemplateHandler{}

// ListTemplates 列出所有模板
func (h *TemplateHandler) ListTemplates(c fiber.Ctx) error {
	templateDir := "./views/template"
	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	templates := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "common" {
			templates = append(templates, entry.Name())
		}
	}
	return c.JSON(templates)
}

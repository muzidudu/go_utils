// Package service 业务逻辑处理
package service

import (
	"errors"

	"github.com/muzidudu/go_utils/fiber/internal/dto"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
)

// ErrCategoryNotFound 分类不存在
var ErrCategoryNotFound = errors.New("category not found")

// CategoryService 分类业务
type CategoryService struct{}

// Categories 分类服务实例
var Categories = &CategoryService{}

// ListTree 获取树形分类，parentID=0 完整树，parentID>0 以该 id 为根的子树
func (s *CategoryService) ListTree(parentID uint) ([]*dto.CategoryResp, error) {
	tree, err := repository.Category.ListTree(parentID)
	if err != nil {
		return nil, err
	}
	return categoryTreeToResp(tree), nil
}

// ListFlat 获取扁平列表，parentID=0 根级，parentID>0 该分类的直接子级
func (s *CategoryService) ListFlat(parentID uint) ([]dto.CategoryResp, error) {
	list, err := repository.Category.ListFlat(parentID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryResp, len(list))
	for i := range list {
		result[i] = categoryToResp(&list[i])
	}
	return result, nil
}

// GetByID 根据 ID 获取
func (s *CategoryService) GetByID(id uint) (*dto.CategoryResp, error) {
	m, err := repository.Category.GetByID(id)
	if err != nil || m == nil {
		return nil, ErrCategoryNotFound
	}
	resp := categoryToResp(m)
	return &resp, nil
}

// Create 创建
func (s *CategoryService) Create(req dto.CreateCategoryReq) (*dto.CategoryResp, error) {
	if req.Name == "" {
		return nil, errors.New("name required")
	}
	typ := req.Type
	if typ == "" {
		typ = "category"
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	m := &models.Category{
		ParentID: req.ParentID,
		Name:     req.Name,
		Slug:     req.Slug,
		Type:     typ,
		Link:     req.Link,
		Sort:     req.Sort,
		Status:   status,
	}
	if err := repository.Category.Create(m); err != nil {
		return nil, err
	}
	resp := categoryToResp(m)
	return &resp, nil
}

// Update 更新
func (s *CategoryService) Update(id uint, req dto.UpdateCategoryReq) (*dto.CategoryResp, error) {
	m, err := repository.Category.GetByID(id)
	if err != nil || m == nil {
		return nil, ErrCategoryNotFound
	}
	if req.ParentID != nil {
		m.ParentID = *req.ParentID
	}
	if req.Name != nil {
		m.Name = *req.Name
	}
	if req.Slug != nil {
		m.Slug = *req.Slug
	}
	if req.Type != nil {
		m.Type = *req.Type
	}
	if req.Link != nil {
		m.Link = *req.Link
	}
	if req.Sort != nil {
		m.Sort = *req.Sort
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := repository.Category.Update(m); err != nil {
		return nil, err
	}
	resp := categoryToResp(m)
	return &resp, nil
}

// Delete 删除
func (s *CategoryService) Delete(id uint) error {
	_, err := repository.Category.GetByID(id)
	if err != nil {
		return ErrCategoryNotFound
	}
	return repository.Category.Delete(id)
}

func categoryToResp(m *models.Category) dto.CategoryResp {
	return dto.CategoryResp{
		ID:       m.ID,
		ParentID: m.ParentID,
		Name:     m.Name,
		Slug:     m.Slug,
		Type:     m.Type,
		Link:     m.Link,
		Sort:     m.Sort,
		Status:   m.Status,
	}
}

func categoryTreeToResp(tree []*models.Category) []*dto.CategoryResp {
	if tree == nil {
		return nil
	}
	result := make([]*dto.CategoryResp, len(tree))
	for i, m := range tree {
		r := categoryToResp(m)
		r.Children = categoryTreeToResp(m.Children)
		result[i] = &r
	}
	return result
}

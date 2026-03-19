// Package dto 请求/响应结构
package dto

// CategoryResp 分类响应
type CategoryResp struct {
	ID       uint           `json:"id"`
	ParentID uint           `json:"parent_id"`
	Name     string         `json:"name"`
	Slug     string         `json:"slug"`
	Type     string         `json:"type"`
	Link     string         `json:"link"`
	Sort     int            `json:"sort"`
	Status   int            `json:"status"`
	Children []*CategoryResp `json:"children,omitempty"`
}

// CreateCategoryReq 创建分类请求
type CreateCategoryReq struct {
	ParentID uint   `json:"parent_id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Type     string `json:"type"`
	Link     string `json:"link"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

// UpdateCategoryReq 更新分类请求
type UpdateCategoryReq struct {
	ParentID *uint   `json:"parent_id,omitempty"`
	Name     *string `json:"name,omitempty"`
	Slug     *string `json:"slug,omitempty"`
	Type     *string `json:"type,omitempty"`
	Link     *string `json:"link,omitempty"`
	Sort     *int    `json:"sort,omitempty"`
	Status   *int    `json:"status,omitempty"`
}

// Package dto 请求/响应结构
package dto

// SiteResp 站点响应
type SiteResp struct {
	ID         uint     `json:"id"`
	Name       string   `json:"name"`
	Domain     string   `json:"domain"`
	Bind       uint     `json:"bind"`
	Subdomains []string `json:"subdomains"`
	Template   string   `json:"template"`
	IsDefault  bool     `json:"is_default"`
	Status     int      `json:"status"`
}

// CreateSiteReq 创建站点请求
type CreateSiteReq struct {
	Name       string   `json:"name"`
	Domain     string   `json:"domain"`
	Bind       uint     `json:"bind"`
	Subdomains []string `json:"subdomains"`
	Template   string   `json:"template"`
	IsDefault  bool     `json:"is_default"`
	Status     int      `json:"status"`
}

// UpdateSiteReq 更新站点请求（字段可选）
type UpdateSiteReq struct {
	Name       *string   `json:"name,omitempty"`
	Domain     *string   `json:"domain,omitempty"`
	Bind       *uint     `json:"bind,omitempty"`
	Subdomains *[]string `json:"subdomains,omitempty"`
	Template   *string   `json:"template,omitempty"`
	IsDefault  *bool     `json:"is_default,omitempty"`
	Status     *int      `json:"status,omitempty"`
}

// Package service 业务逻辑处理
package service

import (
	"errors"

	"github.com/muzidudu/go_utils/fiber/internal/dto"
	"github.com/muzidudu/go_utils/fiber/internal/models"
	"github.com/muzidudu/go_utils/fiber/internal/repository"
)

// ErrSiteNotFound 站点不存在
var ErrSiteNotFound = errors.New("site not found")

// ErrSiteDuplicateDomain 域名已存在
var ErrSiteDuplicateDomain = errors.New("domain already exists")

// SitesService 站点业务
type SitesService struct{}

// Sites 站点服务实例
var Sites = &SitesService{}

// List 获取站点列表（含已禁用）
func (s *SitesService) List() ([]dto.SiteResp, error) {
	list, err := repository.Site.List()
	if err != nil {
		return nil, err
	}
	result := make([]dto.SiteResp, len(list))
	for i := range list {
		result[i] = siteToResp(&list[i])
	}
	return result, nil
}

// GetByID 根据 ID 获取站点
func (s *SitesService) GetByID(id uint) (*dto.SiteResp, error) {
	site, err := repository.Site.GetByID(id)
	if err != nil || site == nil {
		return nil, ErrSiteNotFound
	}
	resp := siteToResp(site)
	return &resp, nil
}

// Create 创建站点
func (s *SitesService) Create(req dto.CreateSiteReq) (*dto.SiteResp, error) {
	if req.Domain == "" {
		return nil, errors.New("domain required")
	}
	// 校验域名唯一（通过 DB unique 约束，或先查询）
	exist, _ := repository.Site.GetByDomain(req.Domain)
	if exist != nil {
		return nil, ErrSiteDuplicateDomain
	}
	// 若设为默认，先取消其他默认
	if req.IsDefault {
		_ = repository.Site.ClearDefault()
	}
	template := req.Template
	if template == "" {
		template = "default"
	}
	status := req.Status
	if status == 0 {
		status = 1
	}
	site := &models.Site{
		Name:       req.Name,
		Domain:     req.Domain,
		Subdomains: req.Subdomains,
		Template:   template,
		IsDefault:  req.IsDefault,
		Status:     status,
	}
	if err := repository.Site.Create(site); err != nil {
		return nil, err
	}
	resp := siteToResp(site)
	return &resp, nil
}

// Update 更新站点
func (s *SitesService) Update(id uint, req dto.UpdateSiteReq) (*dto.SiteResp, error) {
	site, err := repository.Site.GetByID(id)
	if err != nil || site == nil {
		return nil, ErrSiteNotFound
	}
	if req.Name != nil {
		site.Name = *req.Name
	}
	if req.Domain != nil {
		site.Domain = *req.Domain
	}
	if req.Subdomains != nil {
		site.Subdomains = *req.Subdomains
	}
	if req.Template != nil {
		site.Template = *req.Template
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			_ = repository.Site.ClearDefault()
			site.IsDefault = true
		} else {
			site.IsDefault = false
		}
	}
	if req.Status != nil {
		site.Status = *req.Status
	}
	if err := repository.Site.Update(site); err != nil {
		return nil, err
	}
	resp := siteToResp(site)
	return &resp, nil
}

// Delete 删除站点
func (s *SitesService) Delete(id uint) error {
	site, err := repository.Site.GetByID(id)
	if err != nil || site == nil {
		return ErrSiteNotFound
	}
	return repository.Site.Delete(id)
}

func siteToResp(s *models.Site) dto.SiteResp {
	return dto.SiteResp{
		ID:         s.ID,
		Name:       s.Name,
		Domain:     s.Domain,
		Subdomains: s.Subdomains,
		Template:   s.Template,
		IsDefault:  s.IsDefault,
		Status:     s.Status,
	}
}

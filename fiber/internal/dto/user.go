// Package dto 请求/响应结构，隔离 Model
package dto

// CreateUserReq 创建用户请求
type CreateUserReq struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserResp 用户响应
type UserResp struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

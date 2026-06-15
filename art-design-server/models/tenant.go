package models

import "time"

// Tenant 租户模型
type Tenant struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	Status       int       `json:"status"` // 1:启用 0:禁用
	ContactName  string    `json:"contactName"`
	ContactPhone string    `json:"contactPhone"`
	ContactEmail string    `json:"contactEmail"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"createTime"`
	UpdatedAt    time.Time `json:"updateTime"`
}

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=128"`
	Code         string `json:"code" binding:"required,min=2,max=64"`
	ContactName  string `json:"contactName" binding:"max=64"`
	ContactPhone string `json:"contactPhone" binding:"max=20"`
	ContactEmail string `json:"contactEmail" binding:"max=128"`
	Description  string `json:"description" binding:"max=256"`

	// 租户管理员账号（创建时手动输入）
	AdminUserName string `json:"adminUserName" binding:"required,min=2,max=64"`
	AdminPassword string `json:"adminPassword" binding:"required,min=6,max=32"`
	AdminNickName string `json:"adminNickName" binding:"max=64"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	ID           uint   `json:"id"`
	Name         string `json:"name" binding:"required,min=2,max=128"`
	ContactName  string `json:"contactName" binding:"max=64"`
	ContactPhone string `json:"contactPhone" binding:"max=20"`
	ContactEmail string `json:"contactEmail" binding:"max=128"`
	Description  string `json:"description" binding:"max=256"`
	Status       int    `json:"status"`
}

// TenantSwitchRequest 租户切换请求
type TenantSwitchRequest struct {
	TenantID uint `json:"tenantId"` // 0 = 回到全局视角
}

// Permission 权限点模型
type Permission struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createTime"`
}

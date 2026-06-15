package models

import "time"

// Role 角色模型（匹配数据库 roles 表）
type Role struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	RoleName    string    `json:"roleName" gorm:"column:role_name;size:64;not null"`
	RoleCode    string    `json:"roleCode" gorm:"column:role_code;uniqueIndex;size:32;not null"`
	Description string    `json:"description" gorm:"column:description;size:256"`
	MenuIDs     []int64   `json:"menuIds" gorm:"column:menu_ids;type:integer[]"`
	Status      int       `json:"status" gorm:"default:1"`
	CreatedAt   time.Time `json:"createTime" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updateTime" gorm:"column:updated_at"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	RoleName    string  `json:"roleName" binding:"required,min=2,max=64"`
	RoleCode    string  `json:"roleCode" binding:"required,min=2,max=32"`
	Description string  `json:"description" binding:"max=256"`
	Status      int     `json:"status"`
	MenuIDs     []int64 `json:"menuIds"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID          uint    `json:"id"`
	RoleName    string  `json:"roleName" binding:"required,min=2,max=64"`
	RoleCode    string  `json:"roleCode" binding:"required,min=2,max=32"`
	Description string  `json:"description" binding:"max=256"`
	Status      int     `json:"status"`
	MenuIDs     []int64 `json:"menuIds"`
}

package models

import "time"

// Role 角色模型
type Role struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" binding:"required" gorm:"uniqueIndex;size:64;not null"`
	Code      string    `json:"code" gorm:"uniqueIndex;size:64;not null"`
	Desc      string    `json:"desc" gorm:"size:256"`
	Status    int       `json:"status" gorm:"default:1"` // 1:启用 0:禁用
	MenuIDs   []uint    `json:"menuIds" gorm:"-"`
	MenuJSON  string    `json:"-" gorm:"type:text"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=64"`
	Code    string `json:"code" binding:"required,min=2,max=64"`
	Desc    string `json:"desc" binding:"max=256"`
	Status  int    `json:"status"`
	MenuIDs []uint `json:"menuIds"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	ID      uint   `json:"id" binding:"required"`
	Name    string `json:"name" binding:"required,min=2,max=64"`
	Code    string `json:"code" binding:"required,min=2,max=64"`
	Desc    string `json:"desc" binding:"max=256"`
	Status  int    `json:"status"`
	MenuIDs []uint `json:"menuIds"`
}

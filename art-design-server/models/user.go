package models

import "time"

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserName  string    `json:"userName" binding:"required" gorm:"uniqueIndex;size:64;not null"`
	Password  string    `json:"password,omitempty" gorm:"size:128;not null"`
	NickName  string    `json:"nickName" gorm:"size:64"`
	Email     string    `json:"email" gorm:"size:128"`
	Phone     string    `json:"phone" gorm:"size:20"`
	Avatar    string    `json:"avatar" gorm:"size:256"`
	Status    int       `json:"status" gorm:"default:1"` // 1:启用 0:禁用
	RoleID    uint      `json:"roleId" gorm:"default:0"`
	RoleName  string    `json:"roleName" gorm:"size:64"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	UserName string `json:"userName" binding:"required,min=2,max=64"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	NickName string `json:"nickName" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"max=20"`
	Avatar   string `json:"avatar" binding:"max=256"`
	Status   int    `json:"status"`
	RoleID   uint   `json:"roleId"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	NickName string `json:"nickName" binding:"max=64"`
	Email    string `json:"email" binding:"omitempty,email"`
	Phone    string `json:"phone" binding:"max=20"`
	Avatar   string `json:"avatar" binding:"max=256"`
	Status   int    `json:"status"`
	RoleID   uint   `json:"roleId"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	UserInfo     User   `json:"userInfo"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=32"`
}

// UserInfoResponse 用户信息响应（匹配前端 Api.Auth.UserInfo 类型）
// 前端定义: src/types/api/api.d.ts -> interface UserInfo { buttons, roles, userId, userName, email, avatar }
type UserInfoResponse struct {
	Buttons  []string `json:"buttons"`
	Roles    []string `json:"roles"`
	UserID   uint     `json:"userId"`
	UserName string   `json:"userName"`
	Email    string   `json:"email"`
	Avatar   string   `json:"avatar,omitempty"`
}

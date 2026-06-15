package models

import "time"

// Menu 菜单模型
type Menu struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ParentID  uint      `json:"parentId" gorm:"default:0"`
	Name      string    `json:"name" binding:"required" gorm:"size:64;not null"`
	Path      string    `json:"path" gorm:"size:256"`
	Component string    `json:"component" gorm:"size:256"`
	Icon      string    `json:"icon" gorm:"size:64"`
	Sort      int       `json:"sort" gorm:"default:0"`
	Type      int       `json:"type" gorm:"default:1"` // 1:目录 2:菜单 3:按钮
	Status    int       `json:"status" gorm:"default:1"`
	AuthMark  string    `json:"authMark" gorm:"size:64"`
	Locale    string    `json:"locale" gorm:"size:128"`
	Children  []Menu    `json:"children,omitempty" gorm:"-"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name" binding:"required,min=2,max=64"`
	Path      string `json:"path" binding:"max=256"`
	Component string `json:"component" binding:"max=256"`
	Icon      string `json:"icon" binding:"max=64"`
	Sort      int    `json:"sort"`
	Type      int    `json:"type"`
	Status    int    `json:"status"`
	AuthMark  string `json:"authMark" binding:"max=64"`
	Locale    string `json:"locale" binding:"max=128"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	ID        uint   `json:"id" binding:"required"`
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name" binding:"required,min=2,max=64"`
	Path      string `json:"path" binding:"max=256"`
	Component string `json:"component" binding:"max=256"`
	Icon      string `json:"icon" binding:"max=64"`
	Sort      int    `json:"sort"`
	Type      int    `json:"type"`
	Status    int    `json:"status"`
	AuthMark  string `json:"authMark" binding:"max=64"`
	Locale    string `json:"locale" binding:"max=128"`
}

// MenuTree 菜单树节点（用于前端动态路由）
type MenuTree struct {
	ID        uint       `json:"id"`
	ParentID  uint       `json:"parentId"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Component string     `json:"component"`
	Icon      string     `json:"icon"`
	Sort      int        `json:"sort"`
	Type      int        `json:"type"`
	AuthMark  string     `json:"authMark"`
	Locale    string     `json:"locale"`
	Children  []MenuTree `json:"children,omitempty"`
}

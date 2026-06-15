package models

import "time"

// Menu 菜单模型（匹配数据库 menus 表）
type Menu struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ParentID  uint      `json:"parentId" gorm:"column:parent_id;default:0"`
	Title     string    `json:"title" gorm:"column:title;size:64;not null"`
	Name      string    `json:"name" gorm:"column:name;size:64;not null"`
	Path      string    `json:"path" gorm:"column:path;size:128"`
	Component string    `json:"component" gorm:"column:component;size:128"`
	Icon      string    `json:"icon" gorm:"column:icon;size:64"`
	SortOrder int       `json:"sort" gorm:"column:sort_order;default:0"`
	MenuType  string    `json:"menuType" gorm:"column:menu_type;size:16;default:menu"` // directory|menu|button
	Hidden    bool      `json:"hidden" gorm:"column:hidden;default:false"`
	Status    int       `json:"status" gorm:"default:1"`
	Children  []Menu    `json:"children,omitempty" gorm:"-"`
	CreatedAt time.Time `json:"createTime" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updateTime" gorm:"column:updated_at"`
}

// CreateMenuRequest 创建菜单请求
type CreateMenuRequest struct {
	ParentID  uint   `json:"parentId"`
	Title     string `json:"title" binding:"required,min=2,max=64"`
	Name      string `json:"name" binding:"required,min=2,max=64"`
	Path      string `json:"path" binding:"max=128"`
	Component string `json:"component" binding:"max=128"`
	Icon      string `json:"icon" binding:"max=64"`
	SortOrder int    `json:"sort"`
	MenuType  string `json:"menuType"`
	Hidden    *bool  `json:"hidden"`
	Status    int    `json:"status"`
}

// UpdateMenuRequest 更新菜单请求
type UpdateMenuRequest struct {
	ID        uint   `json:"id" binding:"required"`
	ParentID  uint   `json:"parentId"`
	Title     string `json:"title" binding:"required,min=2,max=64"`
	Name      string `json:"name" binding:"required,min=2,max=64"`
	Path      string `json:"path" binding:"max=128"`
	Component string `json:"component" binding:"max=128"`
	Icon      string `json:"icon" binding:"max=64"`
	SortOrder int    `json:"sort"`
	MenuType  string `json:"menuType"`
	Hidden    *bool  `json:"hidden"`
	Status    int    `json:"status"`
}

// MenuTree 菜单树节点（用于前端动态路由）
type MenuTree struct {
	ID        uint       `json:"id"`
	ParentID  uint       `json:"parentId"`
	Title     string     `json:"title"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	Component string     `json:"component"`
	Icon      string     `json:"icon"`
	SortOrder int        `json:"sort"`
	MenuType  string     `json:"menuType"`
	Hidden    bool       `json:"hidden"`
	Children  []MenuTree `json:"children,omitempty"`
}

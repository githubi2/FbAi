package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// RoleHandler 角色处理器
type RoleHandler struct{}

var DefaultRoleHandler = &RoleHandler{}

// List GET /api/v1/roles
func (h *RoleHandler) List(c *gin.Context) {
	var tenantID *uint
	if tid, exists := c.Get("tenantID"); exists {
		if t, ok := tid.(*uint); ok && t != nil {
			tenantID = t
		}
	}
	list := services.DefaultRoleService.List(tenantID)
	c.JSON(http.StatusOK, models.Success(list))
}

// GetByID GET /api/v1/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	role, err := services.DefaultRoleService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(role))
}

// Create POST /api/v1/roles
func (h *RoleHandler) Create(c *gin.Context) {
	var req models.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	if req.Status == 0 {
		req.Status = 1
	}

	// 从上下文获取 tenantID（租户管理员创建角色时自动绑定）
	if req.TenantID == nil {
		if tid, exists := c.Get("tenantID"); exists {
			if t, ok := tid.(*uint); ok && t != nil {
				req.TenantID = t
			}
		}
	}

	role, err := services.DefaultRoleService.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.Success(role))
}

// Update PUT /api/v1/roles/:id
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}
	req.ID = uint(id)

	if req.Status == 0 {
		req.Status = 1
	}

	role, err := services.DefaultRoleService.Update(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(role))
}

// Delete DELETE /api/v1/roles/:id
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	if err := services.DefaultRoleService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("删除成功", nil))
}

// GetMenus GET /api/v1/roles/:id/menus — 获取角色的菜单权限
func (h *RoleHandler) GetMenus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	role, err := services.DefaultRoleService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	// 获取当前用户的角色菜单权限（用于租户上下文中过滤 allMenus）
	var userMenuIDs []int64
	if userID, exists := c.Get("userID"); exists {
		if uid, ok := userID.(uint); ok {
			userMenuIDs = services.DefaultRoleService.GetUserMenuIDs(uid)
		}
	}

	allMenus := services.DefaultMenuService.List()

	// 如果当前用户有角色菜单限制（租户上下文），过滤 allMenus
	if len(userMenuIDs) > 0 {
		userMenuSet := make(map[int64]bool)
		for _, mid := range userMenuIDs {
			userMenuSet[mid] = true
		}
		var filtered []models.Menu
		for _, m := range allMenus {
			if userMenuSet[int64(m.ID)] {
				filtered = append(filtered, m)
			}
		}
		allMenus = filtered
	}

	c.JSON(http.StatusOK, models.Success(gin.H{
		"allMenus":  allMenus,
		"roleMenus": role.MenuIDs,
	}))
}

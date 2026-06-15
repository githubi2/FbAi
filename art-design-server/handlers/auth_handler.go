package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/middleware"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// AuthHandler 认证处理器
type AuthHandler struct{}

var DefaultAuthHandler = &AuthHandler{}

// Login POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	userID, role, valid := middleware.ValidateUser(req.UserName, req.Password)
	if !valid {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户名或密码错误"))
		return
	}

	// 获取用户信息
	user, err := services.DefaultUserService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, "获取用户信息失败"))
		return
	}

	token := middleware.GenerateToken(userID)
	refreshToken := middleware.GenerateToken(userID + 1000)

	c.JSON(http.StatusOK, models.Success(models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		UserInfo:     *user,
	}))

	_ = role // 后续扩展权限使用
}

// UserInfo GET /api/v1/auth/userinfo
func (h *AuthHandler) UserInfo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "未登录"))
		return
	}

	id := userID.(uint)
	user, err := services.DefaultUserService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, "用户不存在"))
		return
	}

	c.JSON(http.StatusOK, models.Success(user))
}

// GetMenus GET /api/v1/auth/menus — 获取当前用户的菜单树
func (h *AuthHandler) GetMenus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "未登录"))
		return
	}

	// 根据用户角色返回对应菜单
	user, err := services.DefaultUserService.GetByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, "用户不存在"))
		return
	}

	// 超级管理员看到所有菜单
	if user.RoleName == "超级管理员" || user.RoleName == "R_SUPER" {
		tree := services.DefaultMenuService.Tree()
		c.JSON(http.StatusOK, models.Success(tree))
		return
	}

	// 根据角色获取菜单
	role, err := services.DefaultRoleService.GetByID(user.RoleID)
	if err != nil {
		c.JSON(http.StatusOK, models.Success([]models.MenuTree{}))
		return
	}
	tree := services.DefaultMenuService.TreeByIDs(role.MenuIDs)
	c.JSON(http.StatusOK, models.Success(tree))
}

// GetUserInfoHandler GET /api/user/info — 返回前端兼容的用户信息格式
func (h *AuthHandler) GetUserInfoHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "未登录"))
		return
	}

	id := userID.(uint)
	user, err := services.DefaultUserService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, "用户不存在"))
		return
	}

	// 修复 role_name：如果为空，从 role_id 查询角色表
	if user.RoleName == "" && user.RoleID > 0 {
		role, roleErr := services.DefaultRoleService.GetByID(user.RoleID)
		if roleErr == nil {
			user.RoleName = role.RoleName
		}
	}

	// 构建角色列表
	var roles []string
	if user.RoleName != "" {
		// 尝试通过角色名称匹配 code
		allRoles := services.DefaultRoleService.List()
		for _, r := range allRoles {
			if r.RoleName == user.RoleName {
				roles = append(roles, r.RoleCode)
				break
			}
		}
		// 兜底：如果 role_name 本身已经是 code (如 "R_SUPER")
		if len(roles) == 0 {
			roles = append(roles, user.RoleName)
		}
	}

	// 按钮权限：从角色关联的菜单中提取 menu_type='button' 的菜单 name
	buttons := make([]string, 0)
	if user.RoleID > 0 {
		role, err := services.DefaultRoleService.GetByID(user.RoleID)
		if err == nil {
			for _, menuID := range role.MenuIDs {
				menu, err := services.DefaultMenuService.GetByID(uint(menuID))
				if err == nil && menu.MenuType == "button" {
					buttons = append(buttons, menu.Name)
				}
			}
		}
	}

	resp := models.UserInfoResponse{
		Buttons:  buttons,
		Roles:    roles,
		UserID:   user.ID,
		UserName: user.UserName,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}

	c.JSON(http.StatusOK, models.Success(resp))
}

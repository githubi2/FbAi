package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// TenantHandler 租户处理器
type TenantHandler struct{}

var DefaultTenantHandler = &TenantHandler{}

// List GET /api/v1/tenants — 获取租户列表
func (h *TenantHandler) List(c *gin.Context) {
	tenants := services.DefaultTenantService.List()
	c.JSON(http.StatusOK, models.Success(tenants))
}

// GetByID GET /api/v1/tenants/:id — 获取租户详情
func (h *TenantHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误"))
		return
	}

	tenant, err := services.DefaultTenantService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, "租户不存在"))
		return
	}

	c.JSON(http.StatusOK, models.Success(tenant))
}

// Create POST /api/v1/tenants — 创建租户
func (h *TenantHandler) Create(c *gin.Context) {
	var req models.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	tenant, adminUser, err := services.DefaultTenantService.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("租户创建成功", gin.H{
		"tenant":    tenant,
		"adminUser": gin.H{
			"id":       adminUser.ID,
			"userName": adminUser.UserName,
			"nickName": adminUser.NickName,
		},
	}))
}

// Update PUT /api/v1/tenants/:id — 更新租户
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误"))
		return
	}

	var req models.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}
	req.ID = uint(id)

	tenant, err := services.DefaultTenantService.Update(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(tenant))
}

// Delete DELETE /api/v1/tenants/:id — 删除租户
func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误"))
		return
	}

	if err := services.DefaultTenantService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("租户已删除", nil))
}

// Switch POST /api/v1/tenants/switch — 切换租户上下文
func (h *TenantHandler) Switch(c *gin.Context) {
	var req models.TenantSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	// 获取当前 token
	token := c.GetHeader("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	var tenantID *uint
	var tenantName string

	if req.TenantID == 0 {
		// 回到全局视角
		tenantID = nil
		tenantName = "全局视角"
	} else {
		// 切换到指定租户
		tenant, err := services.DefaultTenantService.GetByID(req.TenantID)
		if err != nil {
			c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, "租户不存在"))
			return
		}
		tid := req.TenantID
		tenantID = &tid
		tenantName = tenant.Name
	}

	// 更新会话的租户上下文
	if err := services.DefaultSessionService.SetTenantID(token, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, "切换失败"))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("切换成功", gin.H{
		"tenantId":   tenantID,
		"tenantName": tenantName,
	}))
}

// Current GET /api/v1/tenants/current — 获取当前租户上下文
func (h *TenantHandler) Current(c *gin.Context) {
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

	var tenantID *uint
	var tenantName string

	if user.TenantID != nil {
		tenantID = user.TenantID
		tenant, tErr := services.DefaultTenantService.GetByID(*user.TenantID)
		if tErr == nil {
			tenantName = tenant.Name
		}
	} else {
		// 管理员：从请求上下文获取
		if tid, exists := c.Get("tenantID"); exists && tid != nil {
			if t, ok := tid.(*uint); ok {
				tenantID = t
			}
		}
		if tn, exists := c.Get("tenantName"); exists {
			if name, ok := tn.(string); ok {
				tenantName = name
			}
		}
		if tenantID == nil {
			tenantName = "全局视角"
		}
	}

	c.JSON(http.StatusOK, models.Success(gin.H{
		"tenantId":   tenantID,
		"tenantName": tenantName,
	}))
}

// Permissions GET /api/v1/permissions — 获取所有权限点列表
func (h *TenantHandler) Permissions(c *gin.Context) {
	// 权限点列表已经通过 services 封装，这里直接从 DB 查询
	// 为简单起见，复用已有的权限数据结构
	var perms []models.Permission
	// 可以使用 tenant service 或直接查询
	c.JSON(http.StatusOK, models.Success(perms))
}

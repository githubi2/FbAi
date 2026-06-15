package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// RequirePermission 权限校验中间件工厂
// permCode: 权限编码，如 "system:user:create"
func RequirePermission(permCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未登录",
				"data": nil,
			})
			return
		}

		id := userID.(uint)

		// 超级管理员直接放行
		if isSuperAdmin(id) {
			c.Next()
			return
		}

		// 检查用户是否有该权限
		if !hasPermission(id, permCode) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "权限不足: " + permCode,
				"data": nil,
			})
			return
		}

		c.Next()
	}
}

// TenantContext 租户上下文中间件 — 从 session 读取 tenant_id，设置到上下文
func TenantContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.Next()
			return
		}

		id := userID.(uint)

		// 从 session 读取租户上下文（支持管理员切换租户）
		token := c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		var tenantID *uint
		var tenantName string

		// 优先从 session 读取（管理员切换租户后的上下文）
		sessionTenantID, _ := services.DefaultSessionService.GetTenantID(token)
		if sessionTenantID != nil {
			tenantID = sessionTenantID
			tenant, tErr := services.DefaultTenantService.GetByID(*sessionTenantID)
			if tErr == nil {
				tenantName = tenant.Name
			}
		} else {
			// 租户用户：从用户记录读取
			user, err := services.DefaultUserService.GetByID(id)
			if err == nil && user.TenantID != nil {
				tenantID = user.TenantID
				tenant, tErr := services.DefaultTenantService.GetByID(*user.TenantID)
				if tErr == nil {
					tenantName = tenant.Name
				}
			}
		}

		// 设置租户上下文到 gin.Context
		c.Set("tenantID", tenantID)
		c.Set("tenantName", tenantName)
		
		// 静默处理未使用的变量
		_ = id

		// 设置 PostgreSQL RLS 参数
		if tenantID != nil && db.Pool != nil {
			_, _ = db.Pool.Exec(c.Request.Context(),
				"SELECT set_config('app.current_tenant_id', $1::text, true)",
				tenantID,
			)
		}

		c.Next()
	}
}

// isSuperAdmin 检查用户是否为超级管理员
func isSuperAdmin(userID uint) bool {
	user, err := services.DefaultUserService.GetByID(userID)
	if err != nil {
		return false
	}
	// 全局管理员（tenant_id 为空）视为超级管理员
	return user.TenantID == nil
}

// hasPermission 检查用户是否有指定权限
func hasPermission(userID uint, permCode string) bool {
	perms := services.DefaultTenantService.GetUserPermissions(userID)
	for _, p := range perms {
		if p == permCode {
			return true
		}
	}
	return false
}

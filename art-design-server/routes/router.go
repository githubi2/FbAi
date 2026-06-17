package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/handlers"
	"github.com/githubi2/FbAi/art-design-server/middleware"
)

// SetupRouter 配置所有路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"msg":     "pong",
			"data":    nil,
			"version": "1.0.0",
		})
	})

	// ==================== 向后兼容路由（前端旧 API 路径） ====================

	r.GET("/api/user/info", middleware.AuthRequired(), middleware.TenantContext(), handlers.DefaultAuthHandler.GetUserInfoHandler)
	r.GET("/api/user/list", middleware.AuthRequired(), middleware.TenantContext(), handlers.DefaultUserHandler.List)
	r.GET("/api/role/list", middleware.AuthRequired(), middleware.TenantContext(), handlers.DefaultRoleHandler.List)
	r.GET("/api/v3/system/menus/simple", middleware.AuthRequired(), middleware.TenantContext(), handlers.DefaultMenuHandler.Tree)

	// ==================== API v1 路由组 ====================
	v1 := r.Group("/api/v1")
	{
		// 认证接口（无需登录）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.DefaultAuthHandler.Login)
		}

		// 隐私政策页面（无需登录，Facebook App Review 需要）
		v1.GET("/privacy-policy", handlers.DefaultFbHandler.PrivacyPolicy)

		// Facebook OAuth 回调（无需登录，由 Facebook 重定向调用）
		v1.GET("/fb/callback", handlers.DefaultFbHandler.Callback)

		// Facebook 用户数据删除回调（无需登录，由 Facebook 调用）
		v1.POST("/fb/data-deletion", handlers.DefaultFbHandler.DataDeletion)

		// Facebook OAuth 短链接重定向（无需登录，用户直接访问）
		v1.GET("/fb/go/:token", handlers.DefaultFbHandler.ShortRedirect)

		// 需要登录的接口
		authorized := v1.Group("")
		authorized.Use(middleware.AuthRequired())
		authorized.Use(middleware.TenantContext())
		{
			// 用户信息（无需额外权限）
			authorized.GET("/auth/userinfo", handlers.DefaultAuthHandler.UserInfo)
			authorized.GET("/auth/menus", handlers.DefaultAuthHandler.GetMenus)

			// 租户切换 & 当前租户
			authorized.POST("/tenants/switch", handlers.DefaultTenantHandler.Switch)
			authorized.GET("/tenants/current", handlers.DefaultTenantHandler.Current)

			// 权限点列表
			authorized.GET("/permissions", handlers.DefaultTenantHandler.Permissions)

			// ==================== 租户管理（超级管理员） ====================
			tenants := authorized.Group("/tenants")
			{
				tenants.GET("", middleware.RequirePermission("system:tenant:list"), handlers.DefaultTenantHandler.List)
				tenants.GET("/:id", middleware.RequirePermission("system:tenant:list"), handlers.DefaultTenantHandler.GetByID)
				tenants.POST("", middleware.RequirePermission("system:tenant:create"), handlers.DefaultTenantHandler.Create)
				tenants.PUT("/:id", middleware.RequirePermission("system:tenant:edit"), handlers.DefaultTenantHandler.Update)
				tenants.DELETE("/:id", middleware.RequirePermission("system:tenant:delete"), handlers.DefaultTenantHandler.Delete)
			}

			// ==================== 用户管理 ====================
			users := authorized.Group("/users")
			{
				users.GET("", middleware.RequirePermission("system:user:list"), handlers.DefaultUserHandler.List)
				users.GET("/:id", middleware.RequirePermission("system:user:list"), handlers.DefaultUserHandler.GetByID)
				users.POST("", middleware.RequirePermission("system:user:create"), handlers.DefaultUserHandler.Create)
				users.PUT("/:id", middleware.RequirePermission("system:user:edit"), handlers.DefaultUserHandler.Update)
				users.DELETE("/:id", middleware.RequirePermission("system:user:delete"), handlers.DefaultUserHandler.Delete)
				users.POST("/batch-delete", middleware.RequirePermission("system:user:delete"), handlers.DefaultUserHandler.BatchDelete)
			}

			// ==================== 角色管理 ====================
			roles := authorized.Group("/roles")
			{
				roles.GET("", middleware.RequirePermission("system:role:list"), handlers.DefaultRoleHandler.List)
				roles.GET("/:id", middleware.RequirePermission("system:role:list"), handlers.DefaultRoleHandler.GetByID)
				roles.GET("/:id/menus", middleware.RequirePermission("system:role:list"), handlers.DefaultRoleHandler.GetMenus)
				roles.POST("", middleware.RequirePermission("system:role:create"), handlers.DefaultRoleHandler.Create)
				roles.PUT("/:id", middleware.RequirePermission("system:role:edit"), handlers.DefaultRoleHandler.Update)
				roles.DELETE("/:id", middleware.RequirePermission("system:role:delete"), handlers.DefaultRoleHandler.Delete)
			}

			// ==================== 菜单管理 ====================
			menus := authorized.Group("/menus")
			{
				menus.GET("", middleware.RequirePermission("system:menu:list"), handlers.DefaultMenuHandler.List)
				menus.GET("/tree", middleware.RequirePermission("system:menu:list"), handlers.DefaultMenuHandler.Tree)
				menus.GET("/:id", middleware.RequirePermission("system:menu:list"), handlers.DefaultMenuHandler.GetByID)
				menus.POST("", middleware.RequirePermission("system:menu:create"), handlers.DefaultMenuHandler.Create)
				menus.PUT("/:id", middleware.RequirePermission("system:menu:edit"), handlers.DefaultMenuHandler.Update)
				menus.DELETE("/:id", middleware.RequirePermission("system:menu:delete"), handlers.DefaultMenuHandler.Delete)
			}

			// ==================== Facebook 广告管理 ====================
			fb := authorized.Group("/fb")
			{
				fb.GET("/auth-url", handlers.DefaultFbHandler.AuthURL)
				fb.GET("/status", handlers.DefaultFbHandler.ConnectionStatus)
				fb.GET("/ad-accounts", handlers.DefaultFbHandler.AdAccounts)
				fb.GET("/ad-accounts/detail", handlers.DefaultFbHandler.AdAccountsDetail)
				fb.GET("/ad-accounts/:id/payments", handlers.DefaultFbHandler.PaymentHistory)
				fb.DELETE("/disconnect", handlers.DefaultFbHandler.Disconnect)

				// 多账号改造 — 新增路由
				fbAccounts := fb.Group("/accounts")
				{
					fbAccounts.GET("", handlers.DefaultFbHandler.ListAccounts)
					fbAccounts.DELETE("/:id", handlers.DefaultFbHandler.Disconnect)
					fbAccounts.PUT("/:id/label", handlers.DefaultFbHandler.UpdateLabel)
					fbAccounts.POST("/:id/refresh", handlers.DefaultFbHandler.RefreshStats)
				}
			}
		}
	}

	return r
}

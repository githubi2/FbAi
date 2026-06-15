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

	// /api/user/info — 前端 fetchGetUserInfo() 直接调用（不在 /api/v1 组下）
	r.GET("/api/user/info", middleware.AuthRequired(), handlers.DefaultAuthHandler.GetUserInfoHandler)

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 认证接口（无需登录）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.DefaultAuthHandler.Login)
		}

		// 需要登录的接口
		authorized := v1.Group("")
		authorized.Use(middleware.AuthRequired())
		{
			// 用户信息
			authorized.GET("/auth/userinfo", handlers.DefaultAuthHandler.UserInfo)
			authorized.GET("/auth/menus", handlers.DefaultAuthHandler.GetMenus)

			// 用户管理
			users := authorized.Group("/users")
			{
				users.GET("", handlers.DefaultUserHandler.List)
				users.GET("/:id", handlers.DefaultUserHandler.GetByID)
				users.POST("", handlers.DefaultUserHandler.Create)
				users.PUT("/:id", handlers.DefaultUserHandler.Update)
				users.DELETE("/:id", handlers.DefaultUserHandler.Delete)
				users.POST("/batch-delete", handlers.DefaultUserHandler.BatchDelete)
			}

			// 角色管理
			roles := authorized.Group("/roles")
			{
				roles.GET("", handlers.DefaultRoleHandler.List)
				roles.GET("/:id", handlers.DefaultRoleHandler.GetByID)
				roles.GET("/:id/menus", handlers.DefaultRoleHandler.GetMenus)
				roles.POST("", handlers.DefaultRoleHandler.Create)
				roles.PUT("/:id", handlers.DefaultRoleHandler.Update)
				roles.DELETE("/:id", handlers.DefaultRoleHandler.Delete)
			}

			// 菜单管理
			menus := authorized.Group("/menus")
			{
				menus.GET("", handlers.DefaultMenuHandler.List)
				menus.GET("/tree", handlers.DefaultMenuHandler.Tree)
				menus.GET("/:id", handlers.DefaultMenuHandler.GetByID)
				menus.POST("", handlers.DefaultMenuHandler.Create)
				menus.PUT("/:id", handlers.DefaultMenuHandler.Update)
				menus.DELETE("/:id", handlers.DefaultMenuHandler.Delete)
			}
		}
	}

	return r
}

package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/crypto"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// ValidateUser 验证用户名密码（使用 PostgreSQL）
func ValidateUser(userName, password string) (uint, string, *uint, bool) {
	userID, hashedPassword, role, tenantID, err := services.DefaultUserService.GetPasswordHash(userName)
	if err != nil {
		return 0, "", nil, false
	}

	if !crypto.CheckPassword(hashedPassword, password) {
		return 0, "", nil, false
	}

	return userID, role, tenantID, true
}

// GenerateToken 生成 token 并存入数据库 sessions 表
// rememberMe: true=3天过期, false=24小时过期
// tenantID: 租户上下文。nil=全局管理员视角
func GenerateToken(userID uint, rememberMe bool, tenantID *uint) string {
	token := fmt.Sprintf("token_%d_art-design-%d", userID, time.Now().UnixNano())
	refreshToken := fmt.Sprintf("refresh_%d_art-design-%d", userID, time.Now().UnixNano()+1)

	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(72 * time.Hour)
	} else {
		expiresAt = time.Now().Add(24 * time.Hour)
	}

	// SSO 单点登录：仅管理员 tenantID=nil 时踢旧会话
	_ = services.DefaultSessionService.Create(userID, token, refreshToken, expiresAt, tenantID)

	return token
}

// ValidateToken 验证 token（从数据库查询）
func ValidateToken(token string) (uint, bool) {
	return services.DefaultSessionService.Validate(token)
}

// InvalidateUserSessions 失效指定用户的所有会话（管理员改密码时调用）
func InvalidateUserSessions(userID uint) {
	_ = services.DefaultSessionService.InvalidateUser(userID)
}

// AuthRequired 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "未登录或登录已过期",
				"data": nil,
			})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		userID, valid := ValidateToken(token)
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "token无效或已过期",
				"data": nil,
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

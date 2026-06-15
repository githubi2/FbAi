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
func ValidateUser(userName, password string) (uint, string, bool) {
	userID, hashedPassword, role, err := services.DefaultUserService.GetPasswordHash(userName)
	if err != nil {
		return 0, "", false
	}

	if !crypto.CheckPassword(hashedPassword, password) {
		return 0, "", false
	}

	return userID, role, true
}

// GenerateToken 生成 token 并存入数据库 sessions 表
// rememberMe: true=3天过期, false=24小时过期
func GenerateToken(userID uint, rememberMe bool) string {
	token := fmt.Sprintf("token_%d_art-design-%d", userID, time.Now().UnixNano())
	refreshToken := fmt.Sprintf("refresh_%d_art-design-%d", userID, time.Now().UnixNano()+1)

	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(72 * time.Hour) // 记住密码：3天
	} else {
		expiresAt = time.Now().Add(24 * time.Hour) // 不记住：24小时
	}

	// SSO 单点登录：Create 内部会先删除该用户所有旧会话
	_ = services.DefaultSessionService.Create(userID, token, refreshToken, expiresAt)

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

package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/services"
	"golang.org/x/crypto/bcrypt"
)

// CheckPassword 校验密码
func CheckPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// HashPassword 生成密码哈希
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ValidateUser 验证用户名密码（使用 PostgreSQL）
func ValidateUser(userName, password string) (uint, string, bool) {
	// 从数据库获取用户认证信息
	userID, hashedPassword, role, err := services.DefaultUserService.GetPasswordHash(userName)
	if err != nil {
		return 0, "", false
	}

	// 验证密码
	if !CheckPassword(hashedPassword, password) {
		return 0, "", false
	}

	return userID, role, true
}

// 简单的 Token 存储（生产环境用 JWT）
var tokenStore = map[string]uint{}

// GenerateToken 生成简单 token
func GenerateToken(userID uint) string {
	token := fmt.Sprintf("token_%d_art-design-%d", userID, time.Now().UnixNano())
	tokenStore[token] = userID
	return token
}

// ValidateToken 验证 token
func ValidateToken(token string) (uint, bool) {
	userID, exists := tokenStore[token]
	return userID, exists
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

		// Bearer token 解析
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

		// 将用户信息存入上下文
		c.Set("userID", userID)
		c.Next()
	}
}

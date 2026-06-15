package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// 模拟用户存储（开发阶段用内存数据，后续替换为数据库）
var mockUsers = map[string]struct {
	Password string
	UserID   uint
	Role     string
}{
	"admin": {Password: hashPassword("admin123"), UserID: 1, Role: "R_SUPER"},
	"user":  {Password: hashPassword("user123"), UserID: 2, Role: "R_USER"},
}

func hashPassword(pwd string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(hash)
}

// CheckPassword 校验密码
func CheckPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// ValidateUser 验证用户名密码
func ValidateUser(userName, password string) (uint, string, bool) {
	user, exists := mockUsers[userName]
	if !exists {
		return 0, "", false
	}
	if !CheckPassword(user.Password, password) {
		return 0, "", false
	}
	return user.UserID, user.Role, true
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

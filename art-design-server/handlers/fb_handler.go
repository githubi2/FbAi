package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// FbHandler Facebook 处理器
type FbHandler struct{}

var DefaultFbHandler = &FbHandler{}

// AuthURL GET /api/v1/fb/auth-url — 获取 Facebook OAuth 授权链接
func (h *FbHandler) AuthURL(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	authURL, err := services.DefaultFbService.GetAuthURL(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(models.FbAuthURLResponse{AuthURL: authURL}))
}

// Callback GET /api/v1/fb/callback — Facebook OAuth 回调
func (h *FbHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少授权码 code"))
		return
	}

	// Exchange code for token（state 中已编码 userID）
	token, userID, err := services.DefaultFbService.ExchangeCodeForToken(code, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, "授权失败: "+err.Error()))
		return
	}

	// 保存 token
	if err := services.DefaultFbService.SaveToken(uint(userID), token); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, "保存 token 失败: "+err.Error()))
		return
	}

	// 回调成功，重定向到前端页面
	frontendURL := "http://localhost:3006/#/ad-account/list?fb_connected=success"
	c.Redirect(http.StatusFound, frontendURL)
}

// ConnectionStatus GET /api/v1/fb/status — 获取 Facebook 连接状态
func (h *FbHandler) ConnectionStatus(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	status := services.DefaultFbService.GetConnectionStatus(userID)
	c.JSON(http.StatusOK, models.Success(status))
}

// AdAccounts GET /api/v1/fb/ad-accounts — 获取广告账户列表
func (h *FbHandler) AdAccounts(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	result, err := services.DefaultFbService.GetAdAccounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// Disconnect DELETE /api/v1/fb/disconnect — 断开 Facebook 连接
func (h *FbHandler) Disconnect(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	if err := services.DefaultFbService.Disconnect(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("已断开 Facebook 连接", nil))
}

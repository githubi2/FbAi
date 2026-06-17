package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// FbHandler Facebook 处理器
type FbHandler struct{}

var DefaultFbHandler = &FbHandler{}

// privacyPolicyHTML 隐私政策页面
const privacyPolicyHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Privacy Policy — AIFB</title>
<style>
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;max-width:800px;margin:40px auto;padding:0 20px;color:#333;line-height:1.7}
h1{color:#1a1a2e;font-size:24px;border-bottom:2px solid #eee;padding-bottom:8px}
h2{color:#444;font-size:18px;margin-top:24px}
.update{color:#999;font-size:14px}
</style>
</head>
<body>
<p class="update">Last Updated: June 17, 2026</p>
<h1>PRIVACY POLICY</h1>
<h2>1. INTRODUCTION</h2>
<p>Welcome to AIFB. We are committed to protecting your personal information and your right to privacy.</p>
<h2>2. INFORMATION WE COLLECT</h2>
<p>We collect personal information that you voluntarily provide, including email address.</p>
<h2>3. HOW WE USE YOUR INFORMATION</h2>
<p>We use information to provide and maintain services, improve our website, communicate with you, and comply with legal obligations.</p>
<h2>4. THIRD-PARTY SERVICES</h2>
<p>We may share information with Facebook Pixel. These services have their own privacy policies.</p>
<h2>5. COOKIES</h2>
<p>We use essential, analytics, functional, and marketing cookies.</p>
<h2>6. DATA RETENTION</h2>
<p>We retain personal information only as long as necessary.</p>
<h2>7. SECURITY</h2>
<p>We use administrative, technical, and physical security measures to protect your information.</p>
<h2>8. YOUR RIGHTS (GDPR)</h2>
<p>EEA residents have the right to access, rectify, erase, restrict processing, data portability, and object. Contact: zengyxiansheng@gmail.com.</p>
<h2>9. YOUR RIGHTS (CCPA)</h2>
<p>California residents have the right to know, delete, opt-out, and non-discrimination. Contact: zengyxiansheng@gmail.com.</p>
<h2>10. CONTACT US</h2>
<p>Email: zengyxiansheng@gmail.com</p>
</body>
</html>`

// getTenantID 从 gin context 提取租户 ID（nil = 超级管理员）
func getTenantID(c *gin.Context) *uint {
	if tid, exists := c.Get("tenantID"); exists {
		if t, ok := tid.(*uint); ok && t != nil {
			return t
		}
	}
	return nil
}

// PrivacyPolicy GET /privacy-policy — 隐私政策页面（无需登录）
func (h *FbHandler) PrivacyPolicy(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, privacyPolicyHTML)
}

// AuthURL GET /api/v1/fb/auth-url — 获取 Facebook OAuth 授权链接
func (h *FbHandler) AuthURL(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	authURL, shortURL, err := services.DefaultFbService.GetShortAuthURL(userID, tenantID, c.Request.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(models.FbAuthURLResponse{
		AuthURL:  authURL,
		ShortURL: shortURL,
	}))
}

// Callback GET /api/v1/fb/callback — Facebook OAuth 回调
func (h *FbHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少授权码 code"))
		return
	}

	// Exchange code for token（state 中已编码 userID，从 pending 记录中获取 tenantID）
	token, userID, tenantID, err := services.DefaultFbService.ExchangeCodeForToken(code, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, "授权失败: "+err.Error()))
		return
	}

	// 保存 token（含 tenant_id）
	if err := services.DefaultFbService.SaveToken(uint(userID), tenantID, token); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, "保存 token 失败: "+err.Error()))
		return
	}

	// 回调成功，返回 HTML 成功页面（不重定向到前端，因为用户可能在不同浏览器授权）
	// 样式与后台 ArtResultPage 结果页保持一致
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Facebook 授权成功</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background: #f5f7fa;
    min-height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .page-content {
    text-align: center;
    max-width: 500px;
    width: 90%;
    padding: 64px 20px;
  }
  .icon-circle {
    width: 88px;
    height: 88px;
    margin: 0 auto;
    background: #19BE6B;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .icon-circle svg {
    width: 56px;
    height: 56px;
    stroke: #fff;
    stroke-width: 3;
    fill: none;
  }
  .title {
    margin-top: 32px;
    font-size: 30px;
    font-weight: 500;
    color: #323251;
    line-height: 1.4;
  }
  .msg {
    margin-top: 20px;
    font-size: 16px;
    color: #7987a1;
    line-height: 1.6;
  }
  .info-box {
    margin-top: 30px;
    border-radius: 6px;
    background: rgba(242, 244, 245, 0.8);
    padding: 22px 30px;
    text-align: left;
  }
  .info-box p {
    display: flex;
    align-items: flex-start;
    padding: 8px 0;
    font-size: 14px;
    color: #808695;
    line-height: 1.7;
  }
  .info-box .dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    min-width: 6px;
    background: #19BE6B;
    border-radius: 50%;
    margin-right: 10px;
    margin-top: 7px;
  }
</style>
</head>
<body>
<div class="page-content">
  <div class="icon-circle">
    <svg viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="20 6 9 17 4 12"></polyline>
    </svg>
  </div>
  <h1 class="title">授权成功！</h1>
  <p class="msg">Facebook 广告账户授权已完成。</p>
  <div class="info-box">
    <p><span class="dot"></span>您可以关闭此页面，回到后台管理系统查看广告账户。</p>
    <p><span class="dot"></span>此页面可安全关闭</p>
  </div>
</div>
</body>
</html>`)
}

// ConnectionStatus GET /api/v1/fb/status — 获取 Facebook 连接状态
func (h *FbHandler) ConnectionStatus(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	status := services.DefaultFbService.GetConnectionStatus(userID, tenantID)
	c.JSON(http.StatusOK, models.Success(status))
}

// AdAccounts GET /api/v1/fb/ad-accounts — 获取广告账户列表
func (h *FbHandler) AdAccounts(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.GetAdAccounts(userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// ShortRedirect GET /api/v1/fb/go/:token — 短链接重定向到 Facebook OAuth 授权页
// 无需登录，由用户在浏览器中直接访问
func (h *FbHandler) ShortRedirect(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少 token"))
		return
	}

	authURL, err := services.DefaultFbService.ResolveShortToken(token)
	if err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		// 样式与后台 ArtResultPage 失败页保持一致
		c.String(http.StatusGone, `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>链接已过期</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background: #f5f7fa;
    min-height: 100vh;
    display: flex;
    justify-content: center;
    align-items: center;
  }
  .page-content {
    text-align: center;
    max-width: 500px;
    width: 90%;
    padding: 64px 20px;
  }
  .icon-circle {
    width: 88px;
    height: 88px;
    margin: 0 auto;
    background: #ED4014;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .icon-circle svg {
    width: 56px;
    height: 56px;
    stroke: #fff;
    stroke-width: 3;
    fill: none;
  }
  .title {
    margin-top: 32px;
    font-size: 30px;
    font-weight: 500;
    color: #323251;
    line-height: 1.4;
  }
  .msg {
    margin-top: 20px;
    font-size: 16px;
    color: #7987a1;
    line-height: 1.6;
  }
  .info-box {
    margin-top: 30px;
    border-radius: 6px;
    background: rgba(242, 244, 245, 0.8);
    padding: 22px 30px;
    text-align: left;
  }
  .info-box p {
    display: flex;
    align-items: flex-start;
    padding: 8px 0;
    font-size: 14px;
    color: #808695;
    line-height: 1.7;
  }
  .info-box .dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    min-width: 6px;
    background: #ED4014;
    border-radius: 50%;
    margin-right: 10px;
    margin-top: 7px;
  }
</style>
</head>
<body>
<div class="page-content">
  <div class="icon-circle">
    <svg viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round">
      <line x1="18" y1="6" x2="6" y2="18"></line>
      <line x1="6" y1="6" x2="18" y2="18"></line>
    </svg>
  </div>
  <h1 class="title">链接已过期</h1>
  <p class="msg">授权链接有效期 5 分钟，请回到后台重新生成。</p>
  <div class="info-box">
    <p><span class="dot"></span>请返回后台管理系统，重新点击"连接 Facebook"获取新的授权链接。</p>
  </div>
</div>
</body>
</html>`)
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

// Disconnect DELETE /api/v1/fb/disconnect — 断开 Facebook 连接
// 多账号改造：按主键 ID 断开指定连接
func (h *FbHandler) Disconnect(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	// 从路径参数获取 ID（多账号改造）
	idStr := c.Param("id")
	if idStr == "" {
		// 兼容旧版：如果只有 user_id，断开所有连接
		tenantID := getTenantID(c)
		// 旧版行为：断开所有
		if err := services.DefaultFbService.DisconnectAll(userID, tenantID); err != nil {
			c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
			return
		}
		c.JSON(http.StatusOK, models.SuccessWithMsg("已断开所有 Facebook 连接", nil))
		return
	}

	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的 ID"))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbService.Disconnect(uint(id), userID, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("已断开 Facebook 连接", nil))
}

// DataDeletion POST /api/v1/fb/data-deletion — Facebook 用户数据删除回调
func (h *FbHandler) DataDeletion(c *gin.Context) {
	signedRequest := c.PostForm("signed_request")
	if signedRequest != "" {
		// Facebook 发送 signed_request，解码获取 user_id
		log.Printf("[FB] 收到数据删除请求 (signed_request present)")
	}

	// 返回确认信息
	c.JSON(http.StatusOK, gin.H{
		"url":               "http://localhost:3006/privacy-policy.html",
		"confirmation_code": "data_deletion_confirmed",
	})
}

// ==================== 多账号改造 — 新增 handler ====================

// ListAccounts GET /api/v1/fb/accounts — 获取用户所有已授权 FB 账号列表
func (h *FbHandler) ListAccounts(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.ListAccounts(userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// UpdateLabel PUT /api/v1/fb/accounts/:id/label — 更新 FB 账号备注
func (h *FbHandler) UpdateLabel(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的 ID"))
		return
	}

	var req models.FbUpdateLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbService.UpdateLabel(uint(id), userID, tenantID, req.Label); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("备注已更新", nil))
}

// RefreshStats POST /api/v1/fb/accounts/:id/refresh — 刷新 FB 账号的 BM 和广告账户统计
func (h *FbHandler) RefreshStats(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	idStr := c.Param("id")
	id, err := parseUint(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的 ID"))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbService.RefreshAccountStats(uint(id), userID, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("统计已刷新", nil))
}

// parseUint 解析字符串为 uint64
func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// ==================== 广告账户管理 ====================

// AdAccountsDetail GET /api/v1/fb/ad-accounts/detail — 获取所有已授权FB账号下的广告账户详细信息
func (h *FbHandler) AdAccountsDetail(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.GetAdAccountsDetail(userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

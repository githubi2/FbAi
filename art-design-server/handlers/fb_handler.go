package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/db"
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

	// 后台异步刷新该用户的 FB 账号统计（BM/广告账户数量等），不阻塞授权结果页
	// 前端列表页轮询 refresh-status，完成后自动重载即可看到最新数据
	go h.refreshAccountsCache(uint(userID), tenantID)

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
// 重构后：纯读接口，无任何副作用。后台刷新由 POST /accounts/refresh-all 显式触发
func (h *FbHandler) ListAccounts(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	// 1. 缓存直出（毫秒级）
	cached, _ := cacheService.GetCachedAccounts(userID, tenantID)
	if cached != nil && len(cached.Accounts) > 0 {
		c.JSON(http.StatusOK, models.Success(cached))
		return
	}

	// 2. 无缓存：读 fb_tokens 表（纯 DB 查询，不调 FB API），异步写入缓存
	result, err := services.DefaultFbService.ListAccounts(userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	go h.saveAccountsToCache(userID, tenantID, result)

	c.JSON(http.StatusOK, models.Success(result))
}

// RefreshAllAccounts POST /api/v1/fb/accounts/refresh-all — 显式触发后台刷新所有 FB 账号统计
// 5 分钟冷却期内为 no-op；幂等，重复调用安全
func (h *FbHandler) RefreshAllAccounts(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	started := false
	switch {
	case cacheService.IsRefreshing(userID, tenantID, "accounts"):
		// 已在刷新中
	case cacheService.ShouldRefresh(userID, tenantID, "accounts", 5*time.Minute):
		go h.refreshAccountsCache(userID, tenantID)
		started = true
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"started": started}))
}

// refreshAccountsCache 异步刷新FB账号缓存
// 流程：逐个 token 调 FB Graph API 更新统计（走限速队列）→ 重读 fb_tokens → 写入缓存表
func (h *FbHandler) refreshAccountsCache(userID uint, tenantID *uint) {
	cacheService := services.DefaultFbCacheService

	// 检查是否正在刷新
	if cacheService.IsRefreshing(userID, tenantID, "accounts") {
		return
	}

	// 创建刷新任务
	refreshID, err := cacheService.StartRefresh(userID, tenantID, "accounts")
	if err != nil {
		log.Printf("[FB-HANDLER] 创建刷新任务失败: %v", err)
		return
	}

	// 1. 查询所有有效 token，逐个从 FB API 刷新统计（结果回写 fb_tokens 的 bm_list/ad_accounts）
	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT id FROM fb_tokens WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	)
	if err != nil {
		cacheService.CompleteRefresh(refreshID, err.Error())
		return
	}
	var tokenIDs []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err == nil {
			tokenIDs = append(tokenIDs, id)
		}
	}
	rows.Close()

	for _, tokenID := range tokenIDs {
		// 单个账号失败不影响其他账号（错误已写入 last_error，状态显示"异常"）
		_ = services.DefaultFbService.RefreshAccountStats(tokenID, userID, tenantID)
	}

	// 2. 重新从 fb_tokens 读取最新统计
	result, err := services.DefaultFbService.ListAccounts(userID, tenantID)
	if err != nil {
		cacheService.CompleteRefresh(refreshID, err.Error())
		return
	}

	// 3. 写入缓存表
	if err := cacheService.SaveAccountsCache(userID, tenantID, result.Accounts); err != nil {
		log.Printf("[FB-HANDLER] 保存账号缓存失败: %v", err)
	}

	cacheService.CompleteRefresh(refreshID, "")
}

// saveAccountsToCache 保存账号数据到缓存
func (h *FbHandler) saveAccountsToCache(userID uint, tenantID *uint, result *models.FbAccountListResponse) {
	if err := services.DefaultFbCacheService.SaveAccountsCache(userID, tenantID, result.Accounts); err != nil {
		log.Printf("[FB-HANDLER] 保存账号缓存失败: %v", err)
	}
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

	// 同步更新缓存表，列表页立即展示最新统计（无需等后台刷新任务）
	if result, err := services.DefaultFbService.ListAccounts(userID, tenantID); err == nil {
		_ = services.DefaultFbCacheService.SaveAccountsCache(userID, tenantID, result.Accounts)
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("统计已刷新", nil))
}

// parseUint 解析字符串为 uint64
func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// ==================== 广告账户管理 ====================

// AdAccountsDetail GET /api/v1/fb/ad-accounts/detail — 获取所有已授权FB账号下的广告账户详细信息
// 重构后：纯读接口，无任何副作用。后台刷新由 POST /ad-accounts/refresh-all 显式触发
func (h *FbHandler) AdAccountsDetail(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	// 1. 缓存直出（毫秒级）
	cached, _ := cacheService.GetCachedAdAccounts(userID, tenantID)
	if cached != nil && len(cached.Accounts) > 0 {
		c.JSON(http.StatusOK, models.Success(cached))
		return
	}

	// 2. 无缓存：返回空列表（前端触发刷新后轮询自动更新），绝不同步等待 FB API
	c.JSON(http.StatusOK, models.Success(&models.FbAdAccountDetailListResponse{
		Accounts: []models.FbAdAccountDetail{},
		Total:    0,
	}))
}

// RefreshAllAdAccounts POST /api/v1/fb/ad-accounts/refresh-all — 显式触发后台刷新广告账户详情
// 5 分钟冷却期内为 no-op；幂等，重复调用安全
func (h *FbHandler) RefreshAllAdAccounts(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	started := false
	switch {
	case cacheService.IsRefreshing(userID, tenantID, "ad_accounts"):
		// 已在刷新中
	case cacheService.ShouldRefresh(userID, tenantID, "ad_accounts", 5*time.Minute):
		go h.refreshAdAccountsCache(userID, tenantID)
		started = true
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"started": started}))
}

// refreshAdAccountsCache 异步刷新广告账户缓存
func (h *FbHandler) refreshAdAccountsCache(userID uint, tenantID *uint) {
	cacheService := services.DefaultFbCacheService

	// 检查是否正在刷新
	if cacheService.IsRefreshing(userID, tenantID, "ad_accounts") {
		return
	}

	// 创建刷新任务
	refreshID, err := cacheService.StartRefresh(userID, tenantID, "ad_accounts")
	if err != nil {
		log.Printf("[FB-HANDLER] 创建刷新任务失败: %v", err)
		return
	}

	// 获取最新数据
	result, err := services.DefaultFbService.GetAdAccountsDetail(userID, tenantID)
	if err != nil {
		cacheService.CompleteRefresh(refreshID, err.Error())
		return
	}

	// 保存缓存
	h.saveAdAccountsToCache(userID, tenantID, result)

	cacheService.CompleteRefresh(refreshID, "")
}

// UpdateAdAccountRemark PUT /api/v1/fb/ad-accounts/:id/remark — 更新广告账户本地备注
func (h *FbHandler) UpdateAdAccountRemark(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	adAccountID := c.Param("id") // act_xxx
	if adAccountID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少广告账户ID"))
		return
	}

	var req struct {
		Remark string `json:"remark" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "备注最长255字符"))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbCacheService.UpdateAdAccountRemark(userID, tenantID, adAccountID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"remark": req.Remark}))
}

// saveAdAccountsToCache 保存广告账户数据到缓存
func (h *FbHandler) saveAdAccountsToCache(userID uint, tenantID *uint, result *models.FbAdAccountDetailListResponse) {
	if err := services.DefaultFbCacheService.SaveAdAccountsCache(userID, tenantID, result.Accounts); err != nil {
		log.Printf("[FB-HANDLER] 保存广告账户缓存失败: %v", err)
	}
}

// RefreshStatus GET /api/v1/fb/refresh-status — 获取刷新状态
func (h *FbHandler) RefreshStatus(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	refreshType := c.Query("type") // accounts / ad_accounts
	if refreshType == "" {
		refreshType = "all"
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	// 检查两种刷新状态
	if refreshType == "all" {
		accountsStatus, _ := cacheService.GetRefreshStatus(userID, tenantID, "accounts")
		adAccountsStatus, _ := cacheService.GetRefreshStatus(userID, tenantID, "ad_accounts")

		isRunning := (accountsStatus != nil && accountsStatus.IsRunning) ||
			(adAccountsStatus != nil && adAccountsStatus.IsRunning)

		c.JSON(http.StatusOK, models.Success(models.FbRefreshStatusResponse{
			Status:    "running",
			IsRunning: isRunning,
		}))
		return
	}

	status, err := cacheService.GetRefreshStatus(userID, tenantID, refreshType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(status))
}

// PaymentHistory GET /api/v1/fb/ad-accounts/:id/payments — 获取广告账户支付记录
func (h *FbHandler) PaymentHistory(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	adAccountID := c.Param("id")
	if adAccountID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少广告账户 ID"))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.GetPaymentHistory(userID, tenantID, adAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// AssignUser POST /api/v1/fb/ad-accounts/assign-user — 将用户分配到广告账户
func (h *FbHandler) AssignUser(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	var req models.FbAssignUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.AssignAdAccountUser(userID, tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// LookupUsers POST /api/v1/fb/users/lookup — 查找 Facebook 用户信息
func (h *FbHandler) LookupUsers(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	var req models.FbLookupUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.LookupFacebookUsers(userID, tenantID, req.UIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// RemoveUser POST /api/v1/fb/ad-accounts/remove-user — 删除广告账号权限
func (h *FbHandler) RemoveUser(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	var req models.FbRemoveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.RemoveAdAccountUser(userID, tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(result))
}

// ==================== FB 广告投放（只读监控）====================

// CampaignList GET /api/v1/fb/campaigns?accountId=act_xxx — 广告系列列表（含近7天统计）
func (h *FbHandler) CampaignList(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}
	accountID := c.Query("accountId")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少 accountId"))
		return
	}
	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.GetCampaigns(userID, tenantID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.Success(result))
}

// AdSetList GET /api/v1/fb/campaigns/:id/adsets?accountId=act_xxx — 广告组列表
func (h *FbHandler) AdSetList(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}
	campaignID := c.Param("id")
	accountID := c.Query("accountId")
	if campaignID == "" || accountID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少参数"))
		return
	}
	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.GetAdSets(userID, tenantID, campaignID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.Success(result))
}

// AdList GET /api/v1/fb/adsets/:id/ads?accountId=act_xxx — 广告列表
func (h *FbHandler) AdList(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}
	adsetID := c.Param("id")
	accountID := c.Query("accountId")
	if adsetID == "" || accountID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少参数"))
		return
	}
	tenantID := getTenantID(c)
	result, err := services.DefaultFbService.GetAds(userID, tenantID, adsetID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, models.Success(result))
}

// ==================== BM 列表 ====================

// BmList GET /api/v1/fb/bm-list — 获取 BM 列表
// 纯读接口：缓存直出（毫秒级），后台刷新由 POST /bm-list/refresh 显式触发
func (h *FbHandler) BmList(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	cached, _ := cacheService.GetCachedBms(userID, tenantID)
	if cached != nil && len(cached.List) > 0 {
		c.JSON(http.StatusOK, models.Success(cached))
		return
	}

	// 无缓存：返回空列表（前端触发刷新后轮询自动更新）
	c.JSON(http.StatusOK, models.Success(&models.FbBmListResponse{
		List:  []models.FbBmListItem{},
		Total: 0,
	}))
}

// RefreshAllBms POST /api/v1/fb/bm-list/refresh — 显式触发后台刷新 BM 列表
// 5 分钟冷却期内为 no-op；幂等，重复调用安全
func (h *FbHandler) RefreshAllBms(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	started := false
	switch {
	case cacheService.IsRefreshing(userID, tenantID, "bm"):
		// 已在刷新中
	case cacheService.ShouldRefresh(userID, tenantID, "bm", 5*time.Minute):
		go h.refreshBmsCache(userID, tenantID)
		started = true
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"started": started}))
}

// refreshBmsCache 异步刷新 BM 缓存
func (h *FbHandler) refreshBmsCache(userID uint, tenantID *uint) {
	cacheService := services.DefaultFbCacheService

	if cacheService.IsRefreshing(userID, tenantID, "bm") {
		return
	}

	refreshID, err := cacheService.StartRefresh(userID, tenantID, "bm")
	if err != nil {
		log.Printf("[FB-HANDLER] 创建BM刷新任务失败: %v", err)
		return
	}

	result, err := services.DefaultFbService.GetBmList(userID, tenantID)
	if err != nil {
		cacheService.CompleteRefresh(refreshID, err.Error())
		return
	}

	if err := cacheService.SaveBmsCache(userID, tenantID, result.List); err != nil {
		log.Printf("[FB-HANDLER] 保存BM缓存失败: %v", err)
	}

	cacheService.CompleteRefresh(refreshID, "")
}

// UpdateBmRemark PUT /api/v1/fb/bm-list/:id/remark — 更新 BM 本地备注
func (h *FbHandler) UpdateBmRemark(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	bmID := c.Param("id")
	if bmID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少BM ID"))
		return
	}

	var req struct {
		Remark string `json:"remark" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "备注最长255字符"))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbCacheService.UpdateBmRemark(userID, tenantID, bmID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"remark": req.Remark}))
}

// ==================== FB 公共主页 ====================

// PageList GET /api/v1/fb/pages — 获取公共主页列表（缓存直出）
func (h *FbHandler) PageList(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	cached, _ := cacheService.GetCachedPages(userID, tenantID)
	if cached != nil && len(cached.List) > 0 {
		c.JSON(http.StatusOK, models.Success(cached))
		return
	}

	// 无缓存：返回空列表（前端触发刷新后轮询自动更新）
	c.JSON(http.StatusOK, models.Success(&models.FbPageListResponse{
		List:  []models.FbPageItem{},
		Total: 0,
	}))
}

// RefreshAllPages POST /api/v1/fb/pages/refresh-all — 显式触发后台刷新主页列表
// 5 分钟冷却期内为 no-op；幂等，重复调用安全
func (h *FbHandler) RefreshAllPages(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	started := false
	switch {
	case cacheService.IsRefreshing(userID, tenantID, "pages"):
		// 已在刷新中
	case cacheService.ShouldRefresh(userID, tenantID, "pages", 5*time.Minute):
		go h.refreshPagesCache(userID, tenantID)
		started = true
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"started": started}))
}

// refreshPagesCache 异步刷新主页缓存
func (h *FbHandler) refreshPagesCache(userID uint, tenantID *uint) {
	cacheService := services.DefaultFbCacheService

	if cacheService.IsRefreshing(userID, tenantID, "pages") {
		return
	}

	refreshID, err := cacheService.StartRefresh(userID, tenantID, "pages")
	if err != nil {
		log.Printf("[FB-HANDLER] 创建主页刷新任务失败: %v", err)
		return
	}

	result, err := services.DefaultFbService.GetPageList(userID, tenantID)
	if err != nil {
		cacheService.CompleteRefresh(refreshID, err.Error())
		return
	}

	if err := cacheService.SavePagesCache(userID, tenantID, result.List); err != nil {
		log.Printf("[FB-HANDLER] 保存主页缓存失败: %v", err)
	}

	cacheService.CompleteRefresh(refreshID, "")
}

// UpdatePageRemark PUT /api/v1/fb/pages/:id/remark — 更新主页本地备注
func (h *FbHandler) UpdatePageRemark(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	pageID := c.Param("id")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少主页ID"))
		return
	}

	var req struct {
		Remark string `json:"remark" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "备注最长255字符"))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbCacheService.UpdatePageRemark(userID, tenantID, pageID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"remark": req.Remark}))
}

// ==================== FB 像素 ====================

// PixelList GET /api/v1/fb/pixels — 获取像素列表（缓存直出）
func (h *FbHandler) PixelList(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	cached, _ := cacheService.GetCachedPixels(userID, tenantID)
	if cached != nil && len(cached.List) > 0 {
		c.JSON(http.StatusOK, models.Success(cached))
		return
	}

	// 无缓存：返回空列表（前端触发刷新后轮询自动更新）
	c.JSON(http.StatusOK, models.Success(&models.FbPixelListResponse{
		List:  []models.FbPixelItem{},
		Total: 0,
	}))
}

// RefreshAllPixels POST /api/v1/fb/pixels/refresh-all — 显式触发后台刷新像素列表
// 5 分钟冷却期内为 no-op；幂等，重复调用安全
func (h *FbHandler) RefreshAllPixels(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	tenantID := getTenantID(c)
	cacheService := services.DefaultFbCacheService

	started := false
	switch {
	case cacheService.IsRefreshing(userID, tenantID, "pixels"):
		// 已在刷新中
	case cacheService.ShouldRefresh(userID, tenantID, "pixels", 5*time.Minute):
		go h.refreshPixelsCache(userID, tenantID)
		started = true
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"started": started}))
}

// refreshPixelsCache 异步刷新像素缓存
func (h *FbHandler) refreshPixelsCache(userID uint, tenantID *uint) {
	cacheService := services.DefaultFbCacheService

	if cacheService.IsRefreshing(userID, tenantID, "pixels") {
		return
	}

	refreshID, err := cacheService.StartRefresh(userID, tenantID, "pixels")
	if err != nil {
		log.Printf("[FB-HANDLER] 创建像素刷新任务失败: %v", err)
		return
	}

	result, err := services.DefaultFbService.GetPixelList(userID, tenantID)
	if err != nil {
		cacheService.CompleteRefresh(refreshID, err.Error())
		return
	}

	if err := cacheService.SavePixelsCache(userID, tenantID, result.List); err != nil {
		log.Printf("[FB-HANDLER] 保存像素缓存失败: %v", err)
	}

	cacheService.CompleteRefresh(refreshID, "")
}

// UpdatePixelRemark PUT /api/v1/fb/pixels/:id/remark — 更新像素本地备注
func (h *FbHandler) UpdatePixelRemark(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	pixelID := c.Param("id")
	if pixelID == "" {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "缺少像素ID"))
		return
	}

	var req struct {
		Remark string `json:"remark" binding:"max=255"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "备注最长255字符"))
		return
	}

	tenantID := getTenantID(c)
	if err := services.DefaultFbCacheService.UpdatePixelRemark(userID, tenantID, pixelID, req.Remark); err != nil {
		c.JSON(http.StatusInternalServerError, models.Error(models.CodeServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"remark": req.Remark}))
}

// CreatePixel POST /api/v1/fb/pixels — 在指定广告账户下创建像素
func (h *FbHandler) CreatePixel(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.Error(models.CodeUnauthorized, "用户未登录"))
		return
	}

	var req struct {
		AdAccountID string `json:"adAccountId" binding:"required"`
		Name        string `json:"name" binding:"required,max=256"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "广告账户和像素名称不能为空"))
		return
	}

	tenantID := getTenantID(c)
	pixelID, err := services.DefaultFbService.CreatePixel(userID, tenantID, req.AdAccountID, req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(gin.H{"pixelId": pixelID}))
}

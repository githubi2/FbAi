package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
)

// FbService Facebook 服务
type FbService struct {
	appID       string
	appSecret   string
	configID    string // 企业版 Facebook 登录的配置 ID
	redirectURI string
	graphAPI    string
	graphVer    string
	httpClient  *http.Client
}

var DefaultFbService = &FbService{}

// 短链接存储（内存中，5 分钟后自动过期）
var (
	shortTokensMu sync.RWMutex
	shortTokens   = make(map[string]shortTokenEntry)
)

type shortTokenEntry struct {
	authURL   string
	createdAt time.Time
}

// init 从环境变量加载 Facebook 配置，含代理支持
func (s *FbService) init() {
	if s.appID == "" {
		s.appID = os.Getenv("FB_APP_ID")
		s.appSecret = os.Getenv("FB_APP_SECRET")
		s.configID = os.Getenv("FB_CONFIG_ID")
		s.redirectURI = os.Getenv("FB_REDIRECT_URI")
		s.graphAPI = "https://graph.facebook.com"
		s.graphVer = os.Getenv("FB_GRAPH_VERSION")
		if s.graphVer == "" {
			s.graphVer = "v22.0"
		}

		// 初始化 HTTP 客户端（支持 SOCKS5 代理）
		s.httpClient = &http.Client{Timeout: 30 * time.Second}
		fbProxy := os.Getenv("FB_PROXY")
		if fbProxy == "" {
			fbProxy = os.Getenv("HTTPS_PROXY")
		}
		if fbProxy != "" {
			proxyURL, err := url.Parse(fbProxy)
			if err == nil {
				if proxyURL.Scheme == "socks5" {
					dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, nil, proxy.Direct)
					if err == nil {
						s.httpClient.Transport = &http.Transport{Dial: dialer.Dial}
						log.Printf("[FB] 使用 SOCKS5 代理: %s", proxyURL.Host)
					}
				} else {
					s.httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
					log.Printf("[FB] 使用 HTTP 代理: %s", fbProxy)
				}
			}
		}
	}
}

// GetAuthURL 生成 Facebook OAuth 授权链接
// state 参数用于 CSRF 防护并携带 userID，回调时验证
// tenantID: 租户 ID，nil 表示超级管理员
func (s *FbService) GetAuthURL(userID uint, tenantID *uint) (string, error) {
	s.init()

	if s.appID == "" || s.appSecret == "" {
		return "", fmt.Errorf("Facebook 应用未配置，请在 .env 中设置 FB_APP_ID 和 FB_APP_SECRET")
	}

	// 生成 CSRF state token，包含 userID 编码
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("生成 nonce 失败: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	// state 格式: hex(userID):nonce
	state := fmt.Sprintf("%x:%s", userID, nonce)

	// 存储 state 用于回调验证（5分钟有效），同时存储 tenant_id
	// 多账号改造：pending 记录 status=0，不受 unique 约束限制，直接 INSERT
	if db.Pool != nil {
		ctx := context.Background()
		_, err := db.Pool.Exec(ctx,
			`INSERT INTO fb_tokens (user_id, tenant_id, access_token, status, created_at, updated_at)
			 VALUES ($1, $2, $3, 0, NOW(), NOW())`,
			userID, tenantID, "pending:"+state,
		)
		if err != nil {
			log.Printf("[FB] 存储 state 失败: %v", err)
		}
	}

	// 验证配置 ID（企业版 Facebook 登录必需）
	if s.configID == "" {
		return "", fmt.Errorf("Facebook 配置 ID 未设置，请在 .env 中设置 FB_CONFIG_ID")
	}

	// 企业版 Facebook 登录：使用 config_id 替代 scope
	// 权限在 Facebook 应用面板的"企业版 Facebook 登录 → 配置"中管理
	authURL := fmt.Sprintf(
		"%s/%s/oauth/authorize?client_id=%s&redirect_uri=%s&config_id=%s&response_type=code&override_default_response_type=true&state=%s",
		s.graphAPI, s.graphVer,
		url.QueryEscape(s.appID),
		url.QueryEscape(s.redirectURI),
		url.QueryEscape(s.configID),
		state,
	)

	return authURL, nil
}

// GetShortAuthURL 生成短链接版本的授权 URL
// 返回完整授权链接和对应的短链接
func (s *FbService) GetShortAuthURL(userID uint, tenantID *uint, serverHost string) (authURL, shortURL string, err error) {
	authURL, err = s.GetAuthURL(userID, tenantID)
	if err != nil {
		return "", "", err
	}

	// 生成 8 位随机 token
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("生成短链接失败: %w", err)
	}
	token := hex.EncodeToString(b) // 8 字符

	shortTokensMu.Lock()
	shortTokens[token] = shortTokenEntry{
		authURL:   authURL,
		createdAt: time.Now(),
	}
	shortTokensMu.Unlock()

	// 清理过期 token
	go s.cleanExpiredShortTokens()

	shortURL = fmt.Sprintf("http://%s/api/v1/fb/go/%s", serverHost, token)
	return authURL, shortURL, nil
}

// ResolveShortToken 根据短 token 获取完整的 Facebook 授权链接
func (s *FbService) ResolveShortToken(token string) (string, error) {
	shortTokensMu.RLock()
	entry, ok := shortTokens[token]
	shortTokensMu.RUnlock()

	if !ok {
		return "", fmt.Errorf("链接已过期或无效，请重新生成")
	}

	// 检查是否过期（5 分钟）
	if time.Since(entry.createdAt) > 5*time.Minute {
		shortTokensMu.Lock()
		delete(shortTokens, token)
		shortTokensMu.Unlock()
		return "", fmt.Errorf("链接已过期，请重新生成")
	}

	return entry.authURL, nil
}

// cleanExpiredShortTokens 清理过期的短链接 token
func (s *FbService) cleanExpiredShortTokens() {
	shortTokensMu.Lock()
	defer shortTokensMu.Unlock()
	for token, entry := range shortTokens {
		if time.Since(entry.createdAt) > 5*time.Minute {
			delete(shortTokens, token)
		}
	}
}

// ExchangeCodeForToken 用授权码换取 access token
// state 包含编码的 userID（格式: hex(userID):nonce）
// 返回 FbToken、userID 和 tenantID
func (s *FbService) ExchangeCodeForToken(code, state string) (*models.FbToken, uint, *uint, error) {
	s.init()

	if s.appID == "" || s.appSecret == "" {
		return nil, 0, nil, fmt.Errorf("Facebook 应用未配置")
	}

	// 从 state 解析 userID
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 {
		return nil, 0, nil, fmt.Errorf("无效的 state 参数")
	}

	userIDHex, nonce := parts[0], parts[1]

	// 解码 userID（hex → uint）
	var userID uint64
	if _, err := fmt.Sscanf(userIDHex, "%x", &userID); err != nil {
		return nil, 0, nil, fmt.Errorf("无效的 userID 编码: %w", err)
	}

	// 验证 state（CSRF 防护 + 确认是本人发起的请求），同时获取 tenant_id
	var tenantID *uint
	if db.Pool != nil {
		var storedToken string
		var tid *uint
		ctx := context.Background()
		err := db.Pool.QueryRow(ctx,
			`SELECT access_token, tenant_id FROM fb_tokens
			 WHERE user_id = $1 AND access_token LIKE 'pending:%'
			   AND status = 0 AND updated_at > NOW() - INTERVAL '5 minutes'`,
			userID,
		).Scan(&storedToken, &tid)
		if err != nil || storedToken != "pending:"+state {
			return nil, 0, nil, fmt.Errorf("无效的 state 参数，可能为 CSRF 攻击或授权已过期")
		}
		tenantID = tid
		_ = nonce
	}

	// 构建 token 交换请求
	tokenURL := fmt.Sprintf("%s/%s/oauth/access_token", s.graphAPI, s.graphVer)
	resp, err := s.httpClient.Get(fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s",
		tokenURL,
		url.QueryEscape(s.appID),
		url.QueryEscape(s.redirectURI),
		url.QueryEscape(s.appSecret),
		url.QueryEscape(code),
	))
	if err != nil {
		return nil, 0, nil, fmt.Errorf("请求 Facebook token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("读取 Facebook 响应失败: %w", err)
	}

	// 解析响应
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		Error       *struct {
			Message   string `json:"message"`
			Type      string `json:"type"`
			Code      int    `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, 0, nil, fmt.Errorf("解析 Facebook token 响应失败: %w", err)
	}

	if tokenResp.Error != nil {
		return nil, 0, nil, fmt.Errorf("Facebook 返回错误: %s (type=%s, code=%d)",
			tokenResp.Error.Message, tokenResp.Error.Type, tokenResp.Error.Code)
	}

	// 用短期 token 换取长期 token
	longToken, err := s.exchangeLongLivedToken(tokenResp.AccessToken)
	if err != nil {
		log.Printf("[FB] 换取长期 token 失败: %v，使用短期 token", err)
		longToken = tokenResp.AccessToken
		// 短期 token 约 1-2 小时有效
		tokenResp.ExpiresIn = 7200
	}

	// 获取 Facebook 用户信息
	fbUserID, fbUserName := s.getFbUserInfo(longToken)

	// 计算过期时间
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return &models.FbToken{
		FbUserID:    fbUserID,
		FbUserName:  fbUserName,
		AccessToken: longToken,
		TokenType:   tokenResp.TokenType,
		ExpiresAt:   expiresAt,
		Scopes:      []string{"ads_read", "ads_management", "business_management"}, // 企业版登录配置中的权限
		Status:      1,
	}, uint(userID), tenantID, nil
}

// exchangeLongLivedToken 用短期 token 换取长期 token（60天有效）
func (s *FbService) exchangeLongLivedToken(shortToken string) (string, error) {
	s.init()

	tokenURL := fmt.Sprintf("%s/%s/oauth/access_token", s.graphAPI, s.graphVer)
	resp, err := s.httpClient.Get(fmt.Sprintf(
		"%s?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
		tokenURL,
		url.QueryEscape(s.appID),
		url.QueryEscape(s.appSecret),
		url.QueryEscape(shortToken),
	))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"` // 通常 5184000 秒（60天）
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("换取长期 token 失败: %s", string(body))
	}

	return result.AccessToken, nil
}

// getFbUserInfo 获取 Facebook 用户信息
func (s *FbService) getFbUserInfo(accessToken string) (userID, userName string) {
	s.init()

	resp, err := s.httpClient.Get(fmt.Sprintf(
		"%s/%s/me?fields=id,name&access_token=%s",
		s.graphAPI, s.graphVer, url.QueryEscape(accessToken),
	))
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ""
	}

	var user struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return "", ""
	}

	return user.ID, user.Name
}

// SaveToken 保存或更新用户的 Facebook token（多账号支持）
// 同一用户授权同一个 FB 账号 → 刷新 token（UPDATE）
// 同一用户授权不同 FB 账号 → 新增记录（INSERT）
func (s *FbService) SaveToken(userID uint, tenantID *uint, token *models.FbToken) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	// 转换为 PostgreSQL TEXT[] 格式：["a","b"] → {a,b}
	scopesArr := "{"
	for i, s := range token.Scopes {
		if i > 0 {
			scopesArr += ","
		}
		scopesArr += fmt.Sprintf("\"%s\"", s)
	}
	scopesArr += "}"

	// 多账号改造：ON CONFLICT (user_id, fb_user_id) WHERE status=1
	// 部分唯一索引：同一用户+同一FB账号只保留一条有效记录
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO fb_tokens (user_id, tenant_id, fb_user_id, fb_user_name, access_token, token_type, expires_at, scopes, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, NOW(), NOW())
		 ON CONFLICT (user_id, fb_user_id) WHERE status = 1 DO UPDATE
		 SET tenant_id = $2, fb_user_name = $4, access_token = $5, token_type = $6,
		     expires_at = $7, scopes = $8, status = 1, updated_at = NOW()`,
		userID, tenantID, token.FbUserID, token.FbUserName, token.AccessToken,
		token.TokenType, token.ExpiresAt, scopesArr,
	)
	if err != nil {
		return fmt.Errorf("保存 token 失败: %w", err)
	}

	return nil
}

// GetToken 获取用户的有效 Facebook token（租户隔离）
func (s *FbService) GetToken(userID uint, tenantID *uint) (*models.FbToken, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	var token models.FbToken
	var scopesStr string
	var bmListStr, adAccsStr string
	var expiresAt time.Time
	var tid *uint

	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, tenant_id, fb_user_id, fb_user_name, COALESCE(label, ''), access_token, token_type, expires_at,
		        COALESCE(scopes::text, '[]'), COALESCE(bm_list::text, '[]'), COALESCE(ad_accounts::text, '[]'),
		        selected_ad_account_id, status, created_at, updated_at
		 FROM fb_tokens WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	).Scan(&token.ID, &token.UserID, &tid, &token.FbUserID, &token.FbUserName,
		&token.Label, &token.AccessToken, &token.TokenType, &expiresAt,
		&scopesStr, &bmListStr, &adAccsStr,
		&token.SelectedAdAccountID, &token.Status, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("未找到有效的 Facebook 授权: %w", err)
	}

	token.TenantID = tid
	token.ExpiresAt = expiresAt
	token.BmList = bmListStr
	token.AdAccounts = adAccsStr

	// 解析 scopes（TEXT[] 格式: {a,b} 或 {"a","b"}）
	token.Scopes = parsePgArray(scopesStr)

	return &token, nil
}

// GetConnectionStatus 获取用户的 Facebook 连接状态（租户隔离）
func (s *FbService) GetConnectionStatus(userID uint, tenantID *uint) *models.FbConnectionStatusResponse {
	token, err := s.GetToken(userID, tenantID)
	if err != nil {
		return &models.FbConnectionStatusResponse{Connected: false}
	}

	return &models.FbConnectionStatusResponse{
		Connected:           true,
		FbUserID:            token.FbUserID,
		FbUserName:          token.FbUserName,
		ExpiresAt:           token.ExpiresAt.Format(time.RFC3339),
		SelectedAdAccountID: token.SelectedAdAccountID,
		Scopes:              token.Scopes,
	}
}

// GetAdAccounts 获取用户可访问的广告账户列表（租户隔离）
func (s *FbService) GetAdAccounts(userID uint, tenantID *uint) (*models.FbAdAccountListResponse, error) {
	token, err := s.GetToken(userID, tenantID)
	if err != nil {
		return nil, err
	}

	s.init()

	// 获取广告账户
	adAccResp, err := s.fbGet(
		fmt.Sprintf("/%s/me/adaccounts", s.graphVer),
		map[string]string{
			"fields":       "id,name,account_status,currency,business{name}",
			"access_token": token.AccessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取广告账户失败: %w", err)
	}

	var adAccounts []models.FbAdAccount
	if data, ok := adAccResp["data"].([]interface{}); ok {
		for _, item := range data {
			if acc, ok := item.(map[string]interface{}); ok {
				fbAcc := models.FbAdAccount{
					ID:            getString(acc, "id"),
					AccountID:     getString(acc, "account_id"),
					Name:          getString(acc, "name"),
					AccountStatus: getInt(acc, "account_status"),
					Currency:      getString(acc, "currency"),
				}
				// 获取关联的 BM 名称
				if business, ok := acc["business"].(map[string]interface{}); ok {
					fbAcc.BusinessName = getString(business, "name")
				}
				adAccounts = append(adAccounts, fbAcc)
			}
		}
	}

	// 获取商务管理平台列表
	bmResp, err := s.fbGet(
		fmt.Sprintf("/%s/me/businesses", s.graphVer),
		map[string]string{
			"fields":       "id,name",
			"access_token": token.AccessToken,
			"limit":        "100",
		},
	)

	var businesses []models.FbBusinessManager
	if err == nil {
		if data, ok := bmResp["data"].([]interface{}); ok {
			for _, item := range data {
				if bm, ok := item.(map[string]interface{}); ok {
					businesses = append(businesses, models.FbBusinessManager{
						ID:   getString(bm, "id"),
						Name: getString(bm, "name"),
					})
				}
			}
		}
	}

	// 缓存广告账户列表
	if adAccounts != nil {
		accJSON, _ := json.Marshal(adAccounts)
		bmJSON, _ := json.Marshal(businesses)
		ctx := context.Background()
		db.Pool.Exec(ctx,
			`UPDATE fb_tokens SET ad_accounts = $1, bm_list = $2, updated_at = NOW()
			 WHERE user_id = $3 AND tenant_id IS NOT DISTINCT FROM $4`,
			string(accJSON), string(bmJSON), userID, tenantID,
		)
	}

	return &models.FbAdAccountListResponse{
		AdAccounts: adAccounts,
		Businesses: businesses,
	}, nil
}

// Disconnect 断开指定 Facebook 连接（按主键 ID，租户隔离）
func (s *FbService) Disconnect(id uint, userID uint, tenantID *uint) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET status = 0, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND tenant_id IS NOT DISTINCT FROM $3`,
		id, userID, tenantID,
	)
	return err
}

// DisconnectAll 断开用户所有已连接的 FB 账号（租户隔离）
func (s *FbService) DisconnectAll(userID uint, tenantID *uint) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET status = 0, updated_at = NOW()
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	)
	return err
}

// GetTokenByID 按主键获取指定 token（租户隔离）
func (s *FbService) GetTokenByID(id uint, userID uint, tenantID *uint) (*models.FbToken, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	var token models.FbToken
	var scopesStr string
	var bmListStr, adAccsStr string
	var expiresAt time.Time
	var tid *uint

	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, tenant_id, fb_user_id, fb_user_name, COALESCE(label, ''), access_token, token_type, expires_at,
		        COALESCE(scopes::text, '[]'), COALESCE(bm_list::text, '[]'), COALESCE(ad_accounts::text, '[]'),
		        selected_ad_account_id, status, created_at, updated_at
		 FROM fb_tokens WHERE id = $1 AND user_id = $2 AND tenant_id IS NOT DISTINCT FROM $3 AND status = 1`,
		id, userID, tenantID,
	).Scan(&token.ID, &token.UserID, &tid, &token.FbUserID, &token.FbUserName,
		&token.Label, &token.AccessToken, &token.TokenType, &expiresAt,
		&scopesStr, &bmListStr, &adAccsStr,
		&token.SelectedAdAccountID, &token.Status, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("未找到指定的 Facebook 授权: %w", err)
	}

	token.TenantID = tid
	token.ExpiresAt = expiresAt
	token.BmList = bmListStr
	token.AdAccounts = adAccsStr
	token.Scopes = parsePgArray(scopesStr)

	return &token, nil
}

// ListAccounts 获取用户所有已授权的 FB 账号列表（租户隔离）
func (s *FbService) ListAccounts(userID uint, tenantID *uint) (*models.FbAccountListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT id, fb_user_id, fb_user_name, COALESCE(label, ''), COALESCE(scopes::text, '{}'),
		        expires_at, created_at, COALESCE(bm_list::text, '[]'), COALESCE(ad_accounts::text, '[]')
		 FROM fb_tokens
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1
		 ORDER BY created_at DESC`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询 FB 账号列表失败: %w", err)
	}
	defer rows.Close()

	var accounts []models.FbAccountListItem
	now := time.Now()

	for rows.Next() {
		var (
			id            uint
			fbUserID      string
			fbUserName    string
			label         string
			scopesStr     string
			expiresAt     time.Time
			createdAt     time.Time
			bmListStr     string
			adAccountsStr string
		)
		if err := rows.Scan(&id, &fbUserID, &fbUserName, &label, &scopesStr,
			&expiresAt, &createdAt, &bmListStr, &adAccountsStr); err != nil {
			log.Printf("[FB] 扫描账号行失败: %v", err)
			continue
		}

		scopes := parsePgArray(scopesStr)

		// 检查是否有广告权限
		hasAdPerm := false
		for _, sc := range scopes {
			if sc == "ads_read" || sc == "ads_management" {
				hasAdPerm = true
				break
			}
		}

		// 计算 BM 数量和个人/BM 广告账户数量
		bmCount := 0
		personalAdCount := 0
		bmAdCount := 0

		if bmListStr != "" {
			var bmList []map[string]interface{}
			if err := json.Unmarshal([]byte(bmListStr), &bmList); err == nil {
				bmCount = len(bmList)
			}
		}

		if adAccountsStr != "" {
			var adAccs []map[string]interface{}
			if err := json.Unmarshal([]byte(adAccountsStr), &adAccs); err == nil {
				for _, acc := range adAccs {
					if _, hasBusiness := acc["business"]; hasBusiness {
						bmAdCount++
					} else {
						personalAdCount++
					}
				}
			}
		}

		// 计算剩余天数
		daysUntilExpiry := int(expiresAt.Sub(now).Hours() / 24)

		// 判断账号状态
		accountStatus := "正常"
		if daysUntilExpiry < 0 {
			accountStatus = "已过期"
		}

		accounts = append(accounts, models.FbAccountListItem{
			ID:              id,
			FbUserID:        fbUserID,
			FbUserName:      fbUserName,
			Label:           label,
			Scopes:          scopes,
			ExpiresAt:       expiresAt.Format(time.RFC3339),
			CreatedAt:       createdAt.Format(time.RFC3339),
			DaysUntilExpiry: daysUntilExpiry,
			HasAdPerm:       hasAdPerm,
			AccountStatus:   accountStatus,
			BmCount:         bmCount,
			PersonalAdCount: personalAdCount,
			BmAdCount:       bmAdCount,
		})
	}

	if accounts == nil {
		accounts = []models.FbAccountListItem{}
	}

	return &models.FbAccountListResponse{
		Accounts: accounts,
		Total:    len(accounts),
	}, nil
}

// UpdateLabel 更新 FB 账号备注（租户隔离）
func (s *FbService) UpdateLabel(id uint, userID uint, tenantID *uint, label string) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	result, err := db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET label = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3 AND tenant_id IS NOT DISTINCT FROM $4 AND status = 1`,
		label, id, userID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("更新备注失败: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("未找到指定的 FB 账号")
	}

	return nil
}

// RefreshAccountStats 刷新指定 FB 账号的 BM 和广告账户缓存（租户隔离）
func (s *FbService) RefreshAccountStats(id uint, userID uint, tenantID *uint) error {
	token, err := s.GetTokenByID(id, userID, tenantID)
	if err != nil {
		return err
	}

	s.init()

	// 获取广告账户
	adAccResp, err := s.fbGet(
		fmt.Sprintf("/%s/me/adaccounts", s.graphVer),
		map[string]string{
			"fields":       "id,name,account_status,currency,business{name}",
			"access_token": token.AccessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return fmt.Errorf("获取广告账户失败: %w", err)
	}

	var adAccounts []map[string]interface{}
	if data, ok := adAccResp["data"].([]interface{}); ok {
		for _, item := range data {
			if acc, ok := item.(map[string]interface{}); ok {
				adAccounts = append(adAccounts, acc)
			}
		}
	}

	// 获取 BM 列表
	bmResp, err := s.fbGet(
		fmt.Sprintf("/%s/me/businesses", s.graphVer),
		map[string]string{
			"fields":       "id,name",
			"access_token": token.AccessToken,
			"limit":        "100",
		},
	)

	var businesses []map[string]interface{}
	if err == nil {
		if data, ok := bmResp["data"].([]interface{}); ok {
			for _, item := range data {
				if bm, ok := item.(map[string]interface{}); ok {
					businesses = append(businesses, bm)
				}
			}
		}
	}

	// 缓存到数据库
	accJSON, _ := json.Marshal(adAccounts)
	bmJSON, _ := json.Marshal(businesses)
	ctx := context.Background()
	_, err = db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET ad_accounts = $1, bm_list = $2, updated_at = NOW()
		 WHERE id = $3 AND user_id = $4 AND tenant_id IS NOT DISTINCT FROM $5`,
		string(accJSON), string(bmJSON), id, userID, tenantID,
	)
	if err != nil {
		return fmt.Errorf("缓存账户数据失败: %w", err)
	}

	return nil
}

// fbGet 调用 Facebook Graph API (GET)
func (s *FbService) fbGet(endpoint string, params map[string]string) (map[string]interface{}, error) {
	s.init()

	// 构建 URL
	u, _ := url.Parse(s.graphAPI + endpoint)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	resp, err := s.httpClient.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析 Facebook API 响应失败: %w", err)
	}

	// 检查 Facebook 错误
	if errMsg, ok := result["error"].(map[string]interface{}); ok {
		return nil, fmt.Errorf("Facebook API 错误: %v", errMsg["message"])
	}

	return result, nil
}

// 辅助函数
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

// parsePgArray 解析 PostgreSQL TEXT[] 格式: {a,b} 或 {"a","b"}
func parsePgArray(s string) []string {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return []string{}
	}
	// 去掉首尾花括号
	inner := s[1 : len(s)-1]
	if inner == "" {
		return []string{}
	}
	var result []string
	var current strings.Builder
	inQuote := false
	for _, ch := range inner {
		switch {
		case ch == '"':
			inQuote = !inQuote
		case ch == ',' && !inQuote:
			result = append(result, strings.Trim(current.String(), `"`))
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}
	result = append(result, strings.Trim(current.String(), `"`))
	return result
}

// GetAdAccountsDetail 获取当前用户所有已授权FB账号下的广告账户详细信息
func (s *FbService) GetAdAccountsDetail(userID uint, tenantID *uint) (*models.FbAdAccountDetailListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	s.init()

	// 获取所有已授权的 token
	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT id, fb_user_id, fb_user_name, access_token
		 FROM fb_tokens
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询 FB token 失败: %w", err)
	}
	defer rows.Close()

	var allAccounts []models.FbAdAccountDetail

	for rows.Next() {
		var (
			tokenID    uint
			fbUserID   string
			fbUserName string
			accessToken string
		)
		if err := rows.Scan(&tokenID, &fbUserID, &fbUserName, &accessToken); err != nil {
			log.Printf("[FB] 扫描 token 行失败: %v", err)
			continue
		}

		// 调用 Facebook API 获取该账号下的广告账户详细信息
		adAccResp, err := s.fbGet(
			fmt.Sprintf("/%s/me/adaccounts", s.graphVer),
			map[string]string{
				"fields":       "id,account_id,name,account_status,currency,amount_spent,spend_cap,balance,business{name},owner,users{name},timezone_name,created_time",
				"access_token": accessToken,
				"limit":        "100",
			},
		)
		if err != nil {
			log.Printf("[FB] 获取广告账户详情失败 (fbUserId=%s): %v", fbUserID, err)
			continue
		}

		if data, ok := adAccResp["data"].([]interface{}); ok {
			for _, item := range data {
				if acc, ok := item.(map[string]interface{}); ok {
					detail := s.parseAdAccountDetail(acc, fbUserID, fbUserName)
					allAccounts = append(allAccounts, detail)
				}
			}
		}
	}

	if allAccounts == nil {
		allAccounts = []models.FbAdAccountDetail{}
	}

	return &models.FbAdAccountDetailListResponse{
		Accounts: allAccounts,
		Total:    len(allAccounts),
	}, nil
}

// parseAdAccountDetail 解析单个广告账户的详细信息
func (s *FbService) parseAdAccountDetail(acc map[string]interface{}, fbUserID, fbUserName string) models.FbAdAccountDetail {
	status := getInt(acc, "account_status")
	statusLabel := s.getAccountStatusLabel(status)

	// 解析 BM 名称
	businessName := ""
	if business, ok := acc["business"].(map[string]interface{}); ok {
		businessName = getString(business, "name")
	}

	// 解析管理员信息
	adminName := ""
	hiddenAdmins := 0
	if users, ok := acc["users"].(map[string]interface{}); ok {
		if userData, ok := users["data"].([]interface{}); ok {
			for i, u := range userData {
				if userMap, ok := u.(map[string]interface{}); ok {
					uname := getString(userMap, "name")
					// 第一个用户作为主管理员
					if i == 0 {
						adminName = uname
					}
				}
			}
			// 计算隐藏管理员数：总用户数 - 1（显示的主管理员）
			if len(userData) > 1 {
				hiddenAdmins = len(userData) - 1
			}
		}
	}

	// 格式化创建时间
	createdTime := ""
	if ct, ok := acc["created_time"].(string); ok {
		createdTime = ct
	}

	// 时区（用于显示国家/地区）
	timezoneName := getString(acc, "timezone_name")

	// 获取金额相关字段
	amountSpent := 0.0
	if v, ok := acc["amount_spent"]; ok {
		amountSpent = toFloat64(v)
	}

	spendCap := 0.0
	if v, ok := acc["spend_cap"]; ok {
		spendCap = toFloat64(v)
	}

	balance := 0.0
	if v, ok := acc["balance"]; ok {
		balance = toFloat64(v)
	}

	return models.FbAdAccountDetail{
		ID:            getString(acc, "id"),
		AccountID:     getString(acc, "account_id"),
		Name:          getString(acc, "name"),
		FbOwnerName:   fbUserName,
		FbOwnerID:     fbUserID,
		BusinessName:  businessName,
		AccountStatus: status,
		StatusLabel:   statusLabel,
		Platform:      "Facebook",
		AmountSpent:   amountSpent,
		Currency:      getString(acc, "currency"),
		SpendCap:      spendCap,
		Balance:       balance,
		AdminName:     adminName,
		HiddenAdmins:  hiddenAdmins,
		TimezoneName:  timezoneName,
		CreatedTime:   createdTime,
	}
}

// getAccountStatusLabel 获取广告账户状态的中文标签
func (s *FbService) getAccountStatusLabel(status int) string {
	switch status {
	case 1:
		return "活跃"
	case 2:
		return "已禁用"
	case 3:
		return "未结算"
	case 7:
		return "待审核"
	case 9:
		return "非活跃"
	case 100:
		return "待关闭"
	case 101:
		return "已关闭"
	default:
		return fmt.Sprintf("未知(%d)", status)
	}
}

// toFloat64 将 interface{} 转换为 float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case json.Number:
		f, _ := val.Float64()
		return f
	default:
		return 0
	}
}

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
		s.httpClient = &http.Client{Timeout: 120 * time.Second}
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

	result, err := DefaultFbRateLimiter.Do(context.Background(), "/oauth/access_token", func() (interface{}, error) {
		tokenURL := fmt.Sprintf("%s/%s/oauth/access_token", s.graphAPI, s.graphVer)
		resp, reqErr := s.httpClient.Get(fmt.Sprintf(
			"%s?grant_type=fb_exchange_token&client_id=%s&client_secret=%s&fb_exchange_token=%s",
			tokenURL,
			url.QueryEscape(s.appID),
			url.QueryEscape(s.appSecret),
			url.QueryEscape(shortToken),
		))
		if reqErr != nil {
			return nil, reqErr
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}

		var tokenResult struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}
		if jsonErr := json.Unmarshal(body, &tokenResult); jsonErr != nil {
			return nil, jsonErr
		}

		if tokenResult.AccessToken == "" {
			return nil, fmt.Errorf("换取长期 token 失败: %s", string(body))
		}

		return tokenResult.AccessToken, nil
	})

	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// getFbUserInfo 获取 Facebook 用户信息 — 自动走限速队列
func (s *FbService) getFbUserInfo(accessToken string) (userID, userName string) {
	s.init()

	result, err := DefaultFbRateLimiter.Do(context.Background(), "/me", func() (interface{}, error) {
		resp, reqErr := s.httpClient.Get(fmt.Sprintf(
			"%s/%s/me?fields=id,name&access_token=%s",
			s.graphAPI, s.graphVer, url.QueryEscape(accessToken),
		))
		if reqErr != nil {
			return nil, reqErr
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}

		var user struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if jsonErr := json.Unmarshal(body, &user); jsonErr != nil {
			return nil, jsonErr
		}

		return &user, nil
	})

	if err != nil {
		return "", ""
	}
	u := result.(*struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	})
	return u.ID, u.Name
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
	if err != nil {
		return err
	}

	// 清理缓存
	db.Pool.Exec(ctx,
		`DELETE FROM fb_accounts_cache WHERE fb_token_id = $1 AND user_id = $2`,
		id, userID,
	)
	db.Pool.Exec(ctx,
		`DELETE FROM fb_ad_accounts_cache WHERE fb_token_id = $1 AND user_id = $2`,
		id, userID,
	)

	return nil
}

// DisconnectAll 断开用户所有已连接的 FB 账号（租户隔离）
func (s *FbService) DisconnectAll(userID uint, tenantID *uint) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()

	// 获取要断开的 token IDs
	rows, err := db.Pool.Query(ctx,
		`SELECT id FROM fb_tokens WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []uint
	for rows.Next() {
		var id uint
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}

	// 更新状态
	_, err = db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET status = 0, updated_at = NOW()
	 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	)
	if err != nil {
		return err
	}

	// 清理缓存
	for _, id := range ids {
		db.Pool.Exec(ctx,
			`DELETE FROM fb_accounts_cache WHERE fb_token_id = $1 AND user_id = $2`,
			id, userID,
		)
		db.Pool.Exec(ctx,
			`DELETE FROM fb_ad_accounts_cache WHERE fb_token_id = $1 AND user_id = $2`,
			id, userID,
		)
	}

	return nil
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
		        expires_at, created_at, COALESCE(bm_list::text, '[]'), COALESCE(ad_accounts::text, '[]'),
		        COALESCE(last_error, '')
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
			lastError     string
		)
		if err := rows.Scan(&id, &fbUserID, &fbUserName, &label, &scopesStr,
			&expiresAt, &createdAt, &bmListStr, &adAccountsStr, &lastError); err != nil {
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

		// 判断账号状态：异常 > 已过期 > 正常
		accountStatus := "正常"
		if lastError != "" {
			accountStatus = "异常"
		} else if daysUntilExpiry < 0 {
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
			DataError:       lastError,
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
		errMsg := fmt.Sprintf("获取广告账户失败: %v", err)
		s.setLastError(id, errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
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

	// 刷新成功，清除之前的错误
	s.clearLastError(id)

	return nil
}

// ==================== FB 广告投放（只读监控，v26.0）====================
// 数据源：公开 Marketing API（限量 v26.0，已实测可用）：
//
//	/{act_id}/campaigns?fields=id,name,status,effective_status,objective,daily_budget,...
//	/{campaign_id}/adsets?fields=id,name,status,effective_status,optimization_goal,...
//	/{adset_id}/ads?fields=id,name,status,effective_status,creative{id,name},...
//	/{act_id}/insights?date_preset=last_7d&level=campaign&fields=campaign_id,spend,...

// resolveTokenByAccountID 按广告账户 ID 解析当前用户/租户下的访问 token（账户必须属于用户）
func (s *FbService) resolveTokenByAccountID(userID uint, tenantID *uint, accountID string) (string, error) {
	if accountID == "" {
		return "", fmt.Errorf("缺少 accountId")
	}
	if db.Pool == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	ctx := context.Background()

	var tokenID uint
	err := db.Pool.QueryRow(ctx,
		`SELECT fb_token_id FROM fb_ad_accounts_cache
		 WHERE ad_account_id = $1 AND user_id = $2 AND tenant_id IS NOT DISTINCT FROM $3
		 LIMIT 1`,
		accountID, userID, tenantID,
	).Scan(&tokenID)
	if err != nil {
		return "", fmt.Errorf("广告账户 %s 不属于当前用户: %w", accountID, err)
	}

	var accessToken string
	if err := db.Pool.QueryRow(ctx,
		`SELECT access_token FROM fb_tokens WHERE id = $1 AND user_id = $2 AND status = 1`,
		tokenID, userID,
	).Scan(&accessToken); err != nil {
		return "", fmt.Errorf("获取授权 token 失败: %w", err)
	}
	return accessToken, nil
}

// GetCampaigns 获取广告系列列表 + 近 7 天统计（insights 一次性合并，避免 N+1）
func (s *FbService) GetCampaigns(userID uint, tenantID *uint, accountID string) (*models.FbCampaignListResponse, error) {
	s.init()
	accessToken, err := s.resolveTokenByAccountID(userID, tenantID, accountID)
	if err != nil {
		return nil, err
	}

	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/campaigns", s.graphVer, accountID),
		map[string]string{
			"fields":       "id,name,status,effective_status,objective,daily_budget,lifetime_budget,bid_strategy,start_time,stop_time,created_time,updated_time",
			"access_token": accessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取广告系列失败: %w", err)
	}

	list := []models.FbCampaign{}
	if data, ok := resp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				list = append(list, models.FbCampaign{
					ID:              getString(m, "id"),
					Name:            getString(m, "name"),
					Status:          getString(m, "status"),
					EffectiveStatus: getString(m, "effective_status"),
					Objective:       getString(m, "objective"),
					DailyBudget:     getString(m, "daily_budget"),
					LifetimeBudget:  getString(m, "lifetime_budget"),
					BidStrategy:     getString(m, "bid_strategy"),
					StartTime:       getString(m, "start_time"),
					StopTime:        getString(m, "stop_time"),
					CreatedTime:     getString(m, "created_time"),
					UpdatedTime:     getString(m, "updated_time"),
				})
			}
		}
	}

	// 近 7 天统计：每账户 1 次 insights 调用（level=campaign 汇总全部系列）
	s.mergeCampaignInsights(accountID, accessToken, list)

	if list == nil {
		list = []models.FbCampaign{}
	}
	return &models.FbCampaignListResponse{List: list, Total: len(list), AccountID: accountID}, nil
}

// mergeCampaignInsights 合并近 7 天统计到系列列表（失败仅记日志，不影响列表）
func (s *FbService) mergeCampaignInsights(accountID, accessToken string, list []models.FbCampaign) {
	if len(list) == 0 {
		return
	}
	insResp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/insights", s.graphVer, accountID),
		map[string]string{
			"date_preset":  "last_7d",
			"level":        "campaign",
			"fields":       "campaign_id,spend,impressions,clicks,ctr,cpc",
			"access_token": accessToken,
			"limit":        "500",
		},
	)
	if err != nil {
		log.Printf("[FB-CAMPAIGN] 获取 insights 失败 (act=%s): %v", accountID, err)
		return
	}
	insMap := make(map[string]*models.FbInsight)
	if data, ok := insResp["data"].([]interface{}); ok {
		for _, item := range data {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			cid := getString(m, "campaign_id")
			if cid == "" {
				continue
			}
			insMap[cid] = &models.FbInsight{
				Spend:       getString(m, "spend"),
				Impressions: getString(m, "impressions"),
				Clicks:      getString(m, "clicks"),
				CTR:         getString(m, "ctr"),
				CPC:         getString(m, "cpc"),
			}
		}
	}
	for i := range list {
		if ins, ok := insMap[list[i].ID]; ok {
			list[i].Insight = ins
		}
	}
}

// GetAdSetsByAccount 获取广告账户下全部广告组（一次调用，含所属系列）
func (s *FbService) GetAdSetsByAccount(userID uint, tenantID *uint, accountID string) (*models.FbAdSetListResponse, error) {
	s.init()
	accessToken, err := s.resolveTokenByAccountID(userID, tenantID, accountID)
	if err != nil {
		return nil, err
	}
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/adsets", s.graphVer, accountID),
		map[string]string{
			"fields":       "id,name,status,effective_status,optimization_goal,billing_event,daily_budget,lifetime_budget,start_time,stop_time,created_time,campaign{id,name}",
			"access_token": accessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取广告组失败: %w", err)
	}
	list := []models.FbAdSet{}
	if data, ok := resp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				campaignName := ""
				if camp, ok := m["campaign"].(map[string]interface{}); ok {
					campaignName = getString(camp, "name")
				}
				list = append(list, models.FbAdSet{
					ID:               getString(m, "id"),
					Name:             getString(m, "name"),
					Status:           getString(m, "status"),
					EffectiveStatus:  getString(m, "effective_status"),
					OptimizationGoal: getString(m, "optimization_goal"),
					BillingEvent:     getString(m, "billing_event"),
					DailyBudget:      getString(m, "daily_budget"),
					LifetimeBudget:   getString(m, "lifetime_budget"),
					StartTime:        getString(m, "start_time"),
					StopTime:         getString(m, "stop_time"),
					CreatedTime:      getString(m, "created_time"),
					CampaignName:     campaignName, // 所属系列
				})
			}
		}
	}
	if list == nil {
		list = []models.FbAdSet{}
	}
	return &models.FbAdSetListResponse{List: list, Total: len(list)}, nil
}

// GetAdsByAccount 获取广告账户下全部广告（一次调用，含所属系列/广告组）
func (s *FbService) GetAdsByAccount(userID uint, tenantID *uint, accountID string) (*models.FbAdListResponse, error) {
	s.init()
	accessToken, err := s.resolveTokenByAccountID(userID, tenantID, accountID)
	if err != nil {
		return nil, err
	}
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/ads", s.graphVer, accountID),
		map[string]string{
			"fields":       "id,name,status,effective_status,campaign{id,name},adset{id,name},creative{id,name},created_time,updated_time",
			"access_token": accessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取广告失败: %w", err)
	}
	list := []models.FbAd{}
	if data, ok := resp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				campaignName, adsetName := "", ""
				if camp, ok := m["campaign"].(map[string]interface{}); ok {
					campaignName = getString(camp, "name")
				}
				if as, ok := m["adset"].(map[string]interface{}); ok {
					adsetName = getString(as, "name")
				}
				creativeID, creativeName := "", ""
				if cr, ok := m["creative"].(map[string]interface{}); ok {
					creativeID = getString(cr, "id")
					creativeName = getString(cr, "name")
				}
				list = append(list, models.FbAd{
					ID:              getString(m, "id"),
					Name:            getString(m, "name"),
					Status:          getString(m, "status"),
					EffectiveStatus: getString(m, "effective_status"),
					CreativeID:      creativeID,
					CreativeName:    creativeName,
					CreatedTime:     getString(m, "created_time"),
					UpdatedTime:     getString(m, "updated_time"),
					CampaignName:    campaignName, // 所属系列
					AdsetName:       adsetName,    // 所属广告组
				})
			}
		}
	}
	if list == nil {
		list = []models.FbAd{}
	}
	return &models.FbAdListResponse{List: list, Total: len(list)}, nil
}

// GetAdSets 获取广告组列表（按 campaign 单个查询，保留兼容）
func (s *FbService) GetAdSets(userID uint, tenantID *uint, campaignID, accountID string) (*models.FbAdSetListResponse, error) {
	s.init()
	accessToken, err := s.resolveTokenByAccountID(userID, tenantID, accountID)
	if err != nil {
		return nil, err
	}
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/adsets", s.graphVer, campaignID),
		map[string]string{
			"fields":       "id,name,status,effective_status,optimization_goal,billing_event,daily_budget,lifetime_budget,start_time,stop_time,created_time",
			"access_token": accessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取广告组失败: %w", err)
	}
	list := []models.FbAdSet{}
	if data, ok := resp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				list = append(list, models.FbAdSet{
					ID:               getString(m, "id"),
					Name:             getString(m, "name"),
					Status:           getString(m, "status"),
					EffectiveStatus:  getString(m, "effective_status"),
					OptimizationGoal: getString(m, "optimization_goal"),
					BillingEvent:     getString(m, "billing_event"),
					DailyBudget:      getString(m, "daily_budget"),
					LifetimeBudget:   getString(m, "lifetime_budget"),
					StartTime:        getString(m, "start_time"),
					StopTime:         getString(m, "stop_time"),
					CreatedTime:      getString(m, "created_time"),
				})
			}
		}
	}
	if list == nil {
		list = []models.FbAdSet{}
	}
	return &models.FbAdSetListResponse{List: list, Total: len(list)}, nil
}

// GetAds 获取广告列表
func (s *FbService) GetAds(userID uint, tenantID *uint, adsetID, accountID string) (*models.FbAdListResponse, error) {
	s.init()
	accessToken, err := s.resolveTokenByAccountID(userID, tenantID, accountID)
	if err != nil {
		return nil, err
	}
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/ads", s.graphVer, adsetID),
		map[string]string{
			"fields":       "id,name,status,effective_status,creative{id,name},created_time,updated_time",
			"access_token": accessToken,
			"limit":        "100",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("获取广告失败: %w", err)
	}
	list := []models.FbAd{}
	if data, ok := resp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				creativeID, creativeName := "", ""
				if cr, ok := m["creative"].(map[string]interface{}); ok {
					creativeID = getString(cr, "id")
					creativeName = getString(cr, "name")
				}
				list = append(list, models.FbAd{
					ID:              getString(m, "id"),
					Name:            getString(m, "name"),
					Status:          getString(m, "status"),
					EffectiveStatus: getString(m, "effective_status"),
					CreativeID:      creativeID,
					CreativeName:    creativeName,
					CreatedTime:     getString(m, "created_time"),
					UpdatedTime:     getString(m, "updated_time"),
				})
			}
		}
	}
	if list == nil {
		list = []models.FbAd{}
	}
	return &models.FbAdListResponse{List: list, Total: len(list)}, nil
}

// fbGet 调用 Facebook Graph API (GET) — 自动走限速队列
func (s *FbService) fbGet(endpoint string, params map[string]string) (map[string]interface{}, error) {
	s.init()

	result, err := DefaultFbRateLimiter.Do(context.Background(), endpoint, func() (interface{}, error) {
		// 构建 URL
		u, _ := url.Parse(s.graphAPI + endpoint)
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()

		resp, reqErr := s.httpClient.Get(u.String())
		if reqErr != nil {
			return nil, reqErr
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}

		var result map[string]interface{}
		if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
			return nil, fmt.Errorf("解析 Facebook API 响应失败: %w", jsonErr)
		}

		// 检查 Facebook 错误
		if errMsg, ok := result["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("Facebook API 错误: %v", errMsg["message"])
		}

		return result, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(map[string]interface{}), nil
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

// isNumericUID 判断是否为纯数字 Facebook UID
func isNumericUID(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// setLastError 更新 fb_tokens 的 last_error 字段
func (s *FbService) setLastError(tokenID uint, errMsg string) {
	if db.Pool == nil || tokenID == 0 {
		return
	}
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET last_error = $1, last_error_at = NOW(), updated_at = NOW()
		 WHERE id = $2`,
		errMsg, tokenID,
	)
	if err != nil {
		log.Printf("[FB] 更新 last_error 失败 (tokenID=%d): %v", tokenID, err)
	}
}

// clearLastError 清除 fb_tokens 的 last_error 字段（刷新成功时调用）
func (s *FbService) clearLastError(tokenID uint) {
	if db.Pool == nil || tokenID == 0 {
		return
	}
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET last_error = '', last_error_at = NULL, updated_at = NOW()
		 WHERE id = $1`,
		tokenID,
	)
	if err != nil {
		log.Printf("[FB] 清除 last_error 失败 (tokenID=%d): %v", tokenID, err)
	}
}
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
			tokenID     uint
			fbUserID    string
			fbUserName  string
			accessToken string
		)
		if err := rows.Scan(&tokenID, &fbUserID, &fbUserName, &accessToken); err != nil {
			log.Printf("[FB] 扫描 token 行失败: %v", err)
			continue
		}

		// 调用 Facebook API 获取该账号下的广告账户基本信息
		adAccResp, err := s.fbGet(
			fmt.Sprintf("/%s/me/adaccounts", s.graphVer),
			map[string]string{
				"fields":       "id,account_id,name,account_status,currency,amount_spent,spend_cap,balance,is_prepay_account,user_tasks,business{name},owner,users{name},timezone_name,timezone_offset_hours_utc,created_time",
				"access_token": accessToken,
				"limit":        "100",
			},
		)
		if err != nil {
			errMsg := fmt.Sprintf("获取广告账户详情失败: %v", err)
			log.Printf("[FB] %s (fbUserId=%s)", errMsg, fbUserID)
			s.setLastError(tokenID, errMsg)
			continue
		}

		// API 调用成功，清除之前的错误
		s.clearLastError(tokenID)

		if data, ok := adAccResp["data"].([]interface{}); ok {
			for _, item := range data {
				if acc, ok := item.(map[string]interface{}); ok {
					accID := getString(acc, "id") // act_xxx 格式
					detail := s.parseAdAccountDetail(acc, fbUserID, fbUserName)
					detail.TokenID = tokenID // 记录所属 token，供缓存表关联

					// 通过单个账户端点获取高级字段
					if accID != "" {
						enriched := s.enrichAdAccountDetail(accID, accessToken, detail)
						allAccounts = append(allAccounts, enriched)
					} else {
						allAccounts = append(allAccounts, detail)
					}
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

// enrichAdAccountDetail 通过单个广告账户端点获取详细字段并合并
func (s *FbService) enrichAdAccountDetail(adAccountID, accessToken string, base models.FbAdAccountDetail) models.FbAdAccountDetail {
	log.Printf("[FB-ENRICH] 开始获取 %s 的高级字段...", adAccountID)
	// 分批请求：先测试哪些字段可用
	// 第一批：已验证可用的字段
	detailFields := "funding_source_details{display_string,type},disable_reason,business_country_code,is_personal"
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s", s.graphVer, adAccountID),
		map[string]string{
			"fields":       detailFields,
			"access_token": accessToken,
		},
	)
	if err != nil {
		log.Printf("[FB-ENRICH] %s 第一批字段失败: %v", adAccountID, err)
	} else {
		log.Printf("[FB-ENRICH] %s 第一批字段响应: funding_source_details=%v, disable_reason=%v, business_country_code=%v, is_personal=%v",
			adAccountID, resp["funding_source_details"], resp["disable_reason"], resp["business_country_code"], resp["is_personal"])
		// 支付方法
		if fs, ok := resp["funding_source_details"].(map[string]interface{}); ok {
			base.FundingSource = getString(fs, "display_string")
			if base.FundingSource == "" {
				base.FundingSource = getString(fs, "type")
			}
		}
		// 锁定原因
		if v, ok := resp["disable_reason"]; ok {
			base.DisableReason = int(toFloat64(v))
			base.DisableReasonLabel = s.getDisableReasonLabel(base.DisableReason)
		}
		// 国家编码
		if v, ok := resp["business_country_code"].(string); ok && v != "" {
			base.CountryCode = v
		}
		// 是否个人账户
		if v, ok := resp["is_personal"]; ok {
			base.IsPersonal = int(toFloat64(v))
		}
	}

	// 第二批：日限额（可能已弃用，需要特殊权限）和账单日期
	detailFields2 := "next_bill_date"
	resp2, err2 := s.fbGet(
		fmt.Sprintf("/%s/%s", s.graphVer, adAccountID),
		map[string]string{
			"fields":       detailFields2,
			"access_token": accessToken,
		},
	)
	if err2 != nil {
		log.Printf("[FB-ENRICH] %s 第二批字段失败: %v", adAccountID, err2)
	} else {
		log.Printf("[FB-ENRICH] %s 第二批字段响应: %v", adAccountID, resp2)
		if v, ok := resp2["next_bill_date"].(string); ok {
			base.NextBillDate = v
		}
	}

	// 注意：daily_spend_limit 字段已在 Graph API v22 移除（实测返回 #100 nonexisting field），
	// 不再单独发请求探测，DailySpendLimit 保持 0，前端显示「无限制」
	// 所有者角色由列表请求的 user_tasks 字段解析（见 parseAdAccountDetail），禁用账户也可靠

	// 第三批：管理员列表兜底
	// 禁用账户的内联 users{name} 会被 FB 剥离，导致管理员数为 0；
	// BM 账户走 assigned_users 边（必须带 business 参数），个人账户走 /users 边
	usersURL := fmt.Sprintf("/%s/%s/users", s.graphVer, adAccountID)
	usersParams := map[string]string{
		"fields":       "id,name,tasks",
		"access_token": accessToken,
	}
	if base.BusinessName != "" && base.OwnerBusinessID != "" {
		usersURL = fmt.Sprintf("/%s/%s/assigned_users", s.graphVer, adAccountID)
		usersParams["business"] = base.OwnerBusinessID
	}
	respU, errU := s.fbGet(usersURL, usersParams)
	if errU != nil {
		log.Printf("[FB-ENRICH] %s 用户列表获取失败: %v", adAccountID, errU)
	} else if data, ok := respU["data"].([]interface{}); ok && len(data) > 0 {
		names := []string{}
		for _, item := range data {
			if u, ok := item.(map[string]interface{}); ok {
				names = append(names, getString(u, "name"))
			}
		}
		if len(names) > 0 {
			base.AdminName = names[0]
			base.OtherAdminNames = names[1:]
			base.HiddenAdmins = len(names) - 1
		}
	}

	log.Printf("[FB-ENRICH] %s 完成: fundingSource=%s, disableReason=%d, nextBillDate=%s, countryCode=%s, isPersonal=%d, ownerRole=%s, admin=%s(+%d)",
		adAccountID, base.FundingSource, base.DisableReason, base.NextBillDate, base.CountryCode, base.IsPersonal, base.OwnerRole, base.AdminName, base.HiddenAdmins)
	return base
}

// deriveRoleFromTasks 从 FB tasks 列表推导账户角色
func deriveRoleFromTasks(tasksRaw interface{}) string {
	tasks, ok := tasksRaw.([]interface{})
	if !ok {
		return ""
	}
	has := func(name string) bool {
		for _, t := range tasks {
			if s, ok := t.(string); ok && s == name {
				return true
			}
		}
		return false
	}
	switch {
	case has("MANAGE"):
		return "Admin"
	case has("ADVERTISE"):
		return "Advertiser"
	case has("ANALYZE"):
		return "Analyst"
	}
	return ""
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

	// 解析所有者 BM ID
	ownerBusinessID := ""
	if owner, ok := acc["owner"].(map[string]interface{}); ok {
		ownerBusinessID = getString(owner, "id")
	} else if ownerStr, ok := acc["owner"].(string); ok {
		ownerBusinessID = ownerStr
	}

	// 解析管理员信息
	adminName := ""
	hiddenAdmins := 0
	otherAdminNames := []string{}
	if users, ok := acc["users"].(map[string]interface{}); ok {
		if userData, ok := users["data"].([]interface{}); ok {
			for i, u := range userData {
				if userMap, ok := u.(map[string]interface{}); ok {
					uname := getString(userMap, "name")
					// 第一个用户作为主管理员
					if i == 0 {
						adminName = uname
					} else {
						otherAdminNames = append(otherAdminNames, uname)
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
	timezoneOffset := 0.0
	if v, ok := acc["timezone_offset_hours_utc"]; ok {
		timezoneOffset = toFloat64(v)
	}

	// 获取金额相关字段
	// 注意：Facebook Marketing API 的金额字段统一以「分」（货币单位的 1/100）返回字符串，
	// 包括 USD/TWD/JPY 等所有货币，这里统一转换为元
	amountSpent := 0.0
	if v, ok := acc["amount_spent"]; ok {
		amountSpent = toFloat64(v) / 100
	}

	spendCap := 0.0
	if v, ok := acc["spend_cap"]; ok {
		spendCap = toFloat64(v) / 100
	}

	balance := 0.0
	if v, ok := acc["balance"]; ok {
		balance = toFloat64(v) / 100
	}

	// 日限额（v22 已移除该字段，保留解析兼容旧版本）
	dailySpendLimit := 0.0
	if v, ok := acc["daily_spend_limit"]; ok {
		dailySpendLimit = toFloat64(v) / 100
	}

	// 支付方法
	fundingSource := ""
	if fs, ok := acc["funding_source_details"].(map[string]interface{}); ok {
		fundingSource = getString(fs, "display_string")
		if fundingSource == "" {
			fundingSource = getString(fs, "type")
		}
	}

	// 锁定原因
	disableReason := getInt(acc, "disable_reason")
	disableReasonLabel := s.getDisableReasonLabel(disableReason)

	// 下个账单日期
	nextBillDate := getString(acc, "next_bill_date")

	// 国家编码
	countryCode := getString(acc, "business_country_code")
	if countryCode == "" {
		// 从时区名称推断国家编码（取 Continent/City 的 Continent 部分）
		if timezoneName != "" {
			// 简单映射常见时区到国家编码
			countryCode = deriveCountryCode(timezoneName)
		}
	}

	// 是否为个人广告账户
	isPersonal := getInt(acc, "is_personal")

	// 是否预付费账户（false=后付费）
	isPrepay := 0
	if v, ok := acc["is_prepay_account"].(bool); ok && v {
		isPrepay = 1
	}

	// 授权用户在该账户的角色：user_tasks 直接挂在账户节点上（禁用账户也返回，比 /users 边可靠）
	ownerRole := deriveRoleFromTasks(acc["user_tasks"])

	return models.FbAdAccountDetail{
		ID:                 getString(acc, "id"),
		AccountID:          getString(acc, "account_id"),
		Name:               getString(acc, "name"),
		FbOwnerName:        fbUserName,
		FbOwnerID:          fbUserID,
		BusinessName:       businessName,
		OwnerBusinessID:    ownerBusinessID,
		AccountStatus:      status,
		StatusLabel:        statusLabel,
		Platform:           "Facebook",
		AmountSpent:        amountSpent,
		Currency:           getString(acc, "currency"),
		SpendCap:           spendCap,
		Balance:            balance,
		DailySpendLimit:    dailySpendLimit,
		AdminName:          adminName,
		HiddenAdmins:       hiddenAdmins,
		OtherAdminNames:    otherAdminNames,
		TimezoneName:       timezoneName,
		TimezoneOffset:     timezoneOffset,
		CountryCode:        countryCode,
		IsPersonal:         isPersonal,
		FundingSource:      fundingSource,
		DisableReason:      disableReason,
		DisableReasonLabel: disableReasonLabel,
		NextBillDate:       nextBillDate,
		CreatedTime:        createdTime,
		IsPrepay:           isPrepay,
		OwnerRole:          ownerRole,
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

// getDisableReasonLabel 获取广告账户禁用原因的中文标签
func (s *FbService) getDisableReasonLabel(reason int) string {
	switch reason {
	case 0:
		return "无" // 未锁定
	case 1:
		return "广告拒登"
	case 2:
		return "广告被限制"
	case 3:
		return "账单问题"
	case 4:
		return "政策违规"
	case 5:
		return "广告账户可疑"
	case 6:
		return "用户请求关闭"
	case 7:
		return "风险支付"
	case 8:
		return "需要确认身份"
	case 9:
		return "广告主已列入黑名单"
	default:
		if reason > 0 {
			return fmt.Sprintf("锁定(%d)", reason)
		}
		return "—"
	}
}

// deriveCountryCode 从时区名称推断国家编码
func deriveCountryCode(tz string) string {
	// 常见时区→国家编码映射
	mapping := map[string]string{
		"America/New_York":    "US",
		"America/Chicago":     "US",
		"America/Denver":      "US",
		"America/Los_Angeles": "US",
		"America/Toronto":     "CA",
		"America/Vancouver":   "CA",
		"America/Mexico_City": "MX",
		"America/Sao_Paulo":   "BR",
		"America/Buenos_Aires": "AR",
		"Europe/London":       "GB",
		"Europe/Paris":        "FR",
		"Europe/Berlin":       "DE",
		"Europe/Madrid":       "ES",
		"Europe/Rome":         "IT",
		"Europe/Amsterdam":    "NL",
		"Europe/Stockholm":    "SE",
		"Europe/Moscow":       "RU",
		"Asia/Shanghai":       "CN",
		"Asia/Tokyo":          "JP",
		"Asia/Seoul":          "KR",
		"Asia/Taipei":         "TW",
		"Asia/Hong_Kong":      "HK",
		"Asia/Singapore":      "SG",
		"Asia/Bangkok":        "TH",
		"Asia/Ho_Chi_Minh":    "VN",
		"Asia/Kolkata":        "IN",
		"Asia/Dubai":          "AE",
		"Asia/Jerusalem":      "IL",
		"Asia/Manila":         "PH",
		"Asia/Jakarta":        "ID",
		"Australia/Sydney":    "AU",
		"Pacific/Auckland":    "NZ",
	}
	if code, ok := mapping[tz]; ok {
		return code
	}
	// 尝试从时区名称提取洲部分作为标识
	parts := strings.SplitN(tz, "/", 2)
	if len(parts) == 2 {
		return parts[0] + "/" + parts[1]
	}
	return tz
}

// GetPaymentHistory 获取广告账户的支付/交易记录
func (s *FbService) GetPaymentHistory(userID uint, tenantID *uint, adAccountID string) (*models.FbPaymentListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	s.init()

	// 获取有效的 FB token
	var accessToken string
	ctx := context.Background()
	err := db.Pool.QueryRow(ctx,
		`SELECT access_token FROM fb_tokens
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1
		 ORDER BY updated_at DESC LIMIT 1`,
		userID, tenantID,
	).Scan(&accessToken)
	if err != nil {
		return nil, fmt.Errorf("未找到有效的 Facebook 授权: %w", err)
	}

	// 支付记录数据源：/{ad-account-id}/activities 边的账单类事件
	// 依据 FB 官方 AdActivity 文档（v22）：旧的 /transactions、/adspayments 边均已移除，
	// 账单事件（扣费/退款/退单/添加支付方式）通过 activities 边的 event_type 暴露
	billingEvents := map[string]struct{ Label, Status string }{
		"ad_account_billing_charge":              {"扣费", "已支付"},
		"ad_account_billing_charge_failed":       {"扣费", "失败"},
		"ad_account_billing_decline":             {"扣费", "失败"},
		"ad_account_billing_refund":              {"退款", "已退款"},
		"ad_account_billing_chargeback":          {"退单", "退单"},
		"ad_account_billing_chargeback_reversal": {"退单撤销", "已撤销"},
		"funding_event_initiated":                {"发起添加支付方式", "处理中"},
		"funding_event_successful":               {"添加支付方式", "成功"},
	}

	// activities 边官方文档明确「没有任何参数」，不支持事件类型过滤，只能分页拉取后本地过滤
	// 实测 since 参数（未文档化）有效：不带时默认只返回最近几天，带 since=一年前 可拉满 FB 保留的全部历史
	// 游标分页直至最后一页，上限 10 页（约 5000 条事件），防止异常账户历史过长拖垮限速器
	since := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	records := []models.FbPaymentRecord{}
	after := ""
	for page := 0; page < 10; page++ {
		params := map[string]string{
			"fields":       "event_type,event_time,extra_data",
			"access_token": accessToken,
			"limit":        "500",
			"since":        since,
		}
		if after != "" {
			params["after"] = after
		}

		resp, err := s.fbGet(fmt.Sprintf("/%s/%s/activities", s.graphVer, adAccountID), params)
		if err != nil {
			if page == 0 {
				return nil, fmt.Errorf("获取支付记录失败: %w", err)
			}
			log.Printf("[FB-PAY] %s 第%d页失败，返回已获取的%d条: %v", adAccountID, page+1, len(records), err)
			break
		}

		if data, ok := resp["data"].([]interface{}); ok {
			for _, item := range data {
				ev, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				meta, isBilling := billingEvents[getString(ev, "event_type")]
				if !isBilling {
					continue
				}

				// extra_data 是 JSON 字符串：{"currency":"USD","new_value":3090(单位为分),"transaction_id":"..."}
				amount := 0.0
				currency := ""
				txID := ""
				if raw := getString(ev, "extra_data"); raw != "" {
					var extra map[string]interface{}
					if json.Unmarshal([]byte(raw), &extra) == nil {
						amount = toFloat64(extra["new_value"]) / 100 // FB 金额以分返回
						currency = getString(extra, "currency")
						txID = getString(extra, "transaction_id")
					}
				}

				records = append(records, models.FbPaymentRecord{
					ID:          txID,
					AccountID:   adAccountID,
					Time:        getString(ev, "event_time"),
					Description: meta.Label,
					Amount:      amount,
					Currency:    currency,
					Status:      meta.Status,
				})
			}
		}

		// 翻页：有 next 且有 after 游标才继续
		paging, ok := resp["paging"].(map[string]interface{})
		if !ok {
			break
		}
		if _, hasNext := paging["next"]; !hasNext {
			break
		}
		nextAfter := ""
		if cursors, ok := paging["cursors"].(map[string]interface{}); ok {
			nextAfter = getString(cursors, "after")
		}
		if nextAfter == "" {
			break
		}
		after = nextAfter
	}

	return &models.FbPaymentListResponse{
		Records: records,
		Total:   len(records),
	}, nil
}

// ==================== 广告账户授权 ====================

// fbPost 向 Facebook Graph API 发送 POST 请求
func (s *FbService) fbPost(endpoint string, params map[string]string) (map[string]interface{}, error) {
	s.init()

	result, err := DefaultFbRateLimiter.Do(context.Background(), endpoint, func() (interface{}, error) {
		// 构建表单数据
		formData := url.Values{}
		for k, v := range params {
			formData.Set(k, v)
		}

		resp, reqErr := s.httpClient.PostForm(s.graphAPI+endpoint, formData)
		if reqErr != nil {
			return nil, reqErr
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}

		var result map[string]interface{}
		if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
			return nil, fmt.Errorf("解析 Facebook API 响应失败: %w", jsonErr)
		}

		// 检查 Facebook 错误
		if errMsg, ok := result["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("Facebook API 错误: %v", errMsg["message"])
		}

		return result, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(map[string]interface{}), nil
}

// AssignAdAccountUser 将用户分配到广告账户（调用 FB Graph API）
func (s *FbService) AssignAdAccountUser(userID uint, tenantID *uint, req *models.FbAssignUserRequest) (*models.FbAssignUserResponse, error) {
	token, err := s.GetToken(userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取 FB token 失败: %w", err)
	}

	s.init()

	response := &models.FbAssignUserResponse{
		Results: make([]models.FbAssignUserResult, 0, len(req.AdAccountIDs)),
		Total:   len(req.AdAccountIDs),
	}

	// Facebook 广告账户 /users 边要求 role 为数字：
	// 1001=ADMIN（管理员）1002=ADVERTISER（广告管理员）1003=ANALYST（分析员）
	fbRoleNum := map[string]string{
		"ADMIN":      "1001",
		"ADVERTISER": "1002",
		"ANALYST":    "1003",
	}[req.Role]
	if fbRoleNum == "" {
		fbRoleNum = "1002"
	}

	// BM 账户 assigned_users 边使用 tasks 数组而不是数字 role
	fbTasks := map[string]string{
		"ADMIN":      `["MANAGE","ADVERTISE","ANALYZE"]`,
		"ADVERTISER": `["ADVERTISE","ANALYZE"]`,
		"ANALYST":    `["ANALYZE"]`,
	}[req.Role]
	if fbTasks == "" {
		fbTasks = `["ADVERTISE","ANALYZE"]`
	}

	// /users 边要求数字 UID：用户名（如 adamjumaa.adamjumaa）先解析为数字 ID
	fbUserID := req.UserID
	if !isNumericUID(fbUserID) {
		userResp, resolveErr := s.fbGet(
			fmt.Sprintf("/%s/%s", s.graphVer, fbUserID),
			map[string]string{
				"fields":       "id",
				"access_token": token.AccessToken,
			},
		)
		if resolveErr == nil {
			if id := getString(userResp, "id"); id != "" {
				fbUserID = id
			}
		}
	}

	for _, adAccountID := range req.AdAccountIDs {
		// 判断账户归属：BM 名下账户必须走 assigned_users 边（带 business 参数 + tasks），
		// 个人账户走 /users 边（数字 role）
		businessID := ""
		accResp, accErr := s.fbGet(
			fmt.Sprintf("/%s/%s", s.graphVer, adAccountID),
			map[string]string{
				"fields":       "business",
				"access_token": token.AccessToken,
			},
		)
		if accErr == nil {
			if biz, ok := accResp["business"].(map[string]interface{}); ok {
				businessID = getString(biz, "id")
			}
		}

		var err error
		if businessID != "" {
			_, err = s.fbPost(
				fmt.Sprintf("/%s/%s/assigned_users", s.graphVer, adAccountID),
				map[string]string{
					"user":         fbUserID,
					"tasks":        fbTasks,
					"business":     businessID,
					"access_token": token.AccessToken,
				},
			)
		} else {
			_, err = s.fbPost(
				fmt.Sprintf("/%s/%s/users", s.graphVer, adAccountID),
				map[string]string{
					"user":         fbUserID,
					"role":         fbRoleNum,
					"access_token": token.AccessToken,
				},
			)
		}

		result := models.FbAssignUserResult{
			AdAccountID: adAccountID,
		}

		if err != nil {
			result.Success = false
			result.Message = err.Error()
			response.Failed++
			log.Printf("[FB-AUTH] 授权失败 %s -> %s (role=%s): %v", req.UserID, adAccountID, req.Role, err)
		} else {
			result.Success = true
			result.Message = "授权成功"
			response.Success++
			log.Printf("[FB-AUTH] 授权成功 %s -> %s (role=%s)", req.UserID, adAccountID, req.Role)
		}

		response.Results = append(response.Results, result)
	}

	return response, nil
}

// LookupFacebookUsers 查找 Facebook 用户信息（通过 UID）
func (s *FbService) LookupFacebookUsers(userID uint, tenantID *uint, uids []string) (*models.FbLookupUserResponse, error) {
	token, err := s.GetToken(userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取 FB token 失败: %w", err)
	}

	s.init()

	// 获取当前用户的好友列表
	friendsResp, friendErr := s.fbGet(
		fmt.Sprintf("/%s/me/friends", s.graphVer),
		map[string]string{
			"access_token": token.AccessToken,
			"limit":        "5000",
		},
	)

	friendSet := make(map[string]bool)
	if friendErr == nil {
		if data, ok := friendsResp["data"].([]interface{}); ok {
			for _, f := range data {
				if fm, ok := f.(map[string]interface{}); ok {
					friendSet[getString(fm, "id")] = true
				}
			}
		}
	}

	// 逐个查找用户信息
	users := make([]models.FbLookupUserResult, 0, len(uids))
	for _, uid := range uids {
		result := models.FbLookupUserResult{
			UID:      uid,
			IsFriend: friendSet[uid],
		}

		// 尝试获取用户名称
		userResp, err := s.fbGet(
			fmt.Sprintf("/%s/%s", s.graphVer, uid),
			map[string]string{
				"fields":       "name,picture",
				"access_token": token.AccessToken,
			},
		)
		if err == nil {
			result.Name = getString(userResp, "name")
			if pic, ok := userResp["picture"].(map[string]interface{}); ok {
				if data, ok := pic["data"].(map[string]interface{}); ok {
					result.Avatar = getString(data, "url")
				}
			}
		}

		users = append(users, result)
	}

	return &models.FbLookupUserResponse{Users: users}, nil
}

// fbDelete 向 Facebook Graph API 发送 DELETE 请求
func (s *FbService) fbDelete(endpoint string, params map[string]string) (map[string]interface{}, error) {
	s.init()

	result, err := DefaultFbRateLimiter.Do(context.Background(), endpoint, func() (interface{}, error) {
		// 构建 URL（DELETE 参数放 query string）
		u, _ := url.Parse(s.graphAPI + endpoint)
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()

		req, reqErr := http.NewRequest("DELETE", u.String(), nil)
		if reqErr != nil {
			return nil, reqErr
		}

		resp, doErr := s.httpClient.Do(req)
		if doErr != nil {
			return nil, doErr
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, readErr
		}

		var result map[string]interface{}
		if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
			return nil, fmt.Errorf("解析 Facebook API 响应失败: %w", jsonErr)
		}

		// 检查 Facebook 错误
		if errMsg, ok := result["error"].(map[string]interface{}); ok {
			return nil, fmt.Errorf("Facebook API 错误: %v", errMsg["message"])
		}

		return result, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(map[string]interface{}), nil
}

// RemoveAdAccountUser 从广告账户删除用户权限（调用 FB Graph API）
func (s *FbService) RemoveAdAccountUser(userID uint, tenantID *uint, req *models.FbRemoveUserRequest) (*models.FbAssignUserResponse, error) {
	token, err := s.GetToken(userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("获取 FB token 失败: %w", err)
	}

	s.init()

	response := &models.FbAssignUserResponse{
		Results: make([]models.FbAssignUserResult, 0),
	}

	switch req.Mode {
	case "deleteTheirs":
		// 删除指定用户在选中广告账户上的权限
		for _, adAccountID := range req.AdAccountIDs {
			for _, uid := range req.UIDs {
				result := s.removeUserFromAdAccount(token.AccessToken, adAccountID, uid)
				response.Results = append(response.Results, result)
				response.Total++
				if result.Success {
					response.Success++
				} else {
					response.Failed++
				}
			}
		}

	case "deleteExceptTheirs":
		// 删除除指定用户外的所有人权限
		for _, adAccountID := range req.AdAccountIDs {
			// 获取当前广告账户的所有用户
			allUsers, listErr := s.listAdAccountUsers(token.AccessToken, adAccountID)
			if listErr != nil {
				log.Printf("[FB-REMOVE] 获取广告账户用户列表失败 %s: %v", adAccountID, listErr)
				continue
			}
			exceptSet := make(map[string]bool)
			for _, uid := range req.UIDs {
				exceptSet[uid] = true
			}
			for _, u := range allUsers {
				if exceptSet[u] {
					continue
				}
				result := s.removeUserFromAdAccount(token.AccessToken, adAccountID, u)
				response.Results = append(response.Results, result)
				response.Total++
				if result.Success {
					response.Success++
				} else {
					response.Failed++
				}
			}
		}

	case "deleteExceptSelf":
		// 删除除自己外的所有人权限
		for _, adAccountID := range req.AdAccountIDs {
			allUsers, listErr := s.listAdAccountUsers(token.AccessToken, adAccountID)
			if listErr != nil {
				log.Printf("[FB-REMOVE] 获取广告账户用户列表失败 %s: %v", adAccountID, listErr)
				continue
			}
			// 获取当前 FB 用户 ID
			selfID := s.getCurrentFbUserID(token.AccessToken)
			for _, u := range allUsers {
				if u == selfID {
					continue
				}
				result := s.removeUserFromAdAccount(token.AccessToken, adAccountID, u)
				response.Results = append(response.Results, result)
				response.Total++
				if result.Success {
					response.Success++
				} else {
					response.Failed++
				}
			}
		}

	case "deleteSelf":
		// 删除自己的权限
		selfID := s.getCurrentFbUserID(token.AccessToken)
		if selfID == "" {
			return nil, fmt.Errorf("无法获取当前 Facebook 用户 ID")
		}
		for _, adAccountID := range req.AdAccountIDs {
			result := s.removeUserFromAdAccount(token.AccessToken, adAccountID, selfID)
			response.Results = append(response.Results, result)
			response.Total++
			if result.Success {
				response.Success++
			} else {
				response.Failed++
			}
		}

	case "deleteBM":
		// 删除 BM（从广告账户移除 Business Manager 关联）
		for _, adAccountID := range req.AdAccountIDs {
			_, delErr := s.fbDelete(
				fmt.Sprintf("/%s/%s", s.graphVer, adAccountID),
				map[string]string{
					"access_token": token.AccessToken,
				},
			)
			result := models.FbAssignUserResult{
				AdAccountID: adAccountID,
			}
			if delErr != nil {
				result.Success = false
				result.Message = delErr.Error()
				response.Failed++
				log.Printf("[FB-REMOVE] 删除BM失败 %s: %v", adAccountID, delErr)
			} else {
				result.Success = true
				result.Message = "删除成功"
				response.Success++
				log.Printf("[FB-REMOVE] 删除BM成功 %s", adAccountID)
			}
			response.Total++
			response.Results = append(response.Results, result)
		}

	default:
		return nil, fmt.Errorf("不支持的删除模式: %s", req.Mode)
	}

	return response, nil
}

// removeUserFromAdAccount 从单个广告账户移除单个用户
func (s *FbService) removeUserFromAdAccount(accessToken, adAccountID, uid string) models.FbAssignUserResult {
	result := models.FbAssignUserResult{
		AdAccountID: adAccountID,
	}

	_, err := s.fbDelete(
		fmt.Sprintf("/%s/%s/users", s.graphVer, adAccountID),
		map[string]string{
			"user":         uid,
			"access_token": accessToken,
		},
	)

	if err != nil {
		result.Success = false
		result.Message = err.Error()
		log.Printf("[FB-REMOVE] 移除用户失败 %s from %s: %v", uid, adAccountID, err)
	} else {
		result.Success = true
		result.Message = "移除成功"
		log.Printf("[FB-REMOVE] 移除用户成功 %s from %s", uid, adAccountID)
	}

	return result
}

// listAdAccountUsers 获取广告账户的所有用户 ID
func (s *FbService) listAdAccountUsers(accessToken, adAccountID string) ([]string, error) {
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/users", s.graphVer, adAccountID),
		map[string]string{
			"fields":       "id",
			"access_token": accessToken,
		},
	)
	if err != nil {
		return nil, err
	}

	var users []string
	if data, ok := resp["data"].([]interface{}); ok {
		for _, item := range data {
			if m, ok := item.(map[string]interface{}); ok {
				if id := getString(m, "id"); id != "" {
					users = append(users, id)
				}
			}
		}
	}
	return users, nil
}

// getCurrentFbUserID 获取当前 Facebook 用户的 UID
func (s *FbService) getCurrentFbUserID(accessToken string) string {
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/me", s.graphVer),
		map[string]string{
			"fields":       "id",
			"access_token": accessToken,
		},
	)
	if err != nil {
		log.Printf("[FB-REMOVE] 获取当前用户ID失败: %v", err)
		return ""
	}
	return getString(resp, "id")
}

// ==================== BM（Business Manager）列表 ====================

// countBmEdge 统计某个 BM 边的条目数（单次 limit=500，够用；失败返回 0）
func (s *FbService) countBmEdge(bmID, edge, accessToken string) int {
	resp, err := s.fbGet(
		fmt.Sprintf("/%s/%s/%s", s.graphVer, bmID, edge),
		map[string]string{
			"fields":       "id",
			"access_token": accessToken,
			"limit":        "500",
		},
	)
	if err != nil {
		log.Printf("[FB-BM] 获取边 %s/%s 失败: %v", bmID, edge, err)
		return 0
	}
	if data, ok := resp["data"].([]interface{}); ok {
		return len(data)
	}
	return 0
}

// GetBmList 从 FB API 获取所有已授权账号下的 BM 列表（供后台刷新任务调用，勿在请求链路同步调用）
// 字段来源均为官方公开 API（已实测）：
//
//	me/businesses?fields=id,name,created_time,verification_status
//	/{bm-id}/business_users?fields=id,name,role   — 管理员数 / 授权用户角色
//	owned_ad_accounts + client_ad_accounts        — 广告账户数
//	owned_businesses + agencies                   — 合作伙伴数
func (s *FbService) GetBmList(userID uint, tenantID *uint) (*models.FbBmListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	s.init()

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

	var allBms []models.FbBmListItem

	for rows.Next() {
		var (
			tokenID     uint
			fbUserID    string
			fbUserName  string
			accessToken string
		)
		if err := rows.Scan(&tokenID, &fbUserID, &fbUserName, &accessToken); err != nil {
			log.Printf("[FB-BM] 扫描 token 行失败: %v", err)
			continue
		}

		bmResp, err := s.fbGet(
			fmt.Sprintf("/%s/me/businesses", s.graphVer),
			map[string]string{
				// permitted_roles：当前授权用户在该 BM 中的角色（如 ["ADMIN"]），官方字段，实测可用
				"fields":       "id,name,created_time,verification_status,permitted_roles",
				"access_token": accessToken,
				"limit":        "100",
			},
		)
		if err != nil {
			log.Printf("[FB-BM] 获取BM列表失败 (fbUserId=%s): %v", fbUserID, err)
			continue
		}

		data, _ := bmResp["data"].([]interface{})
		for _, item := range data {
			bm, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			bmID := getString(bm, "id")
			if bmID == "" {
				continue
			}

			bmItem := models.FbBmListItem{
				BmID:               bmID,
				Name:               getString(bm, "name"),
				FbOwnerName:        fbUserName,
				FbOwnerID:          fbUserID,
				StatusLabel:        "正常", // API 可达即正常；官方无 BM 状态字段
				VerificationStatus: getString(bm, "verification_status"),
				CreatedTime:        getString(bm, "created_time"),
				TokenID:            tokenID,
			}

			// 授权用户角色：permitted_roles（["DEVELOPER","ADMIN"]），取 ADMIN > EMPLOYEE > 其他
			if roles, ok := bm["permitted_roles"].([]interface{}); ok {
				var roleStrs []string
				for _, r := range roles {
					if rs, ok := r.(string); ok {
						roleStrs = append(roleStrs, rs)
					}
				}
				for _, want := range []string{"ADMIN", "EMPLOYEE"} {
					for _, rs := range roleStrs {
						if rs == want {
							bmItem.OwnerRole = rs
							break
						}
					}
					if bmItem.OwnerRole != "" {
						break
					}
				}
				if bmItem.OwnerRole == "" && len(roleStrs) > 0 {
					bmItem.OwnerRole = roleStrs[0]
				}
			}

			// business_users → 管理员总数（summary=total_count）+ 可见名单
			// 注意：灰号/企业账号用户在 data 里不可枚举，但 summary.total_count 是真实总数
			usersResp, err := s.fbGet(
				fmt.Sprintf("/%s/%s/business_users", s.graphVer, bmID),
				map[string]string{
					"fields":       "id,name,role",
					"access_token": accessToken,
					"limit":        "100",
				},
			)
			if err == nil {
				if users, ok := usersResp["data"].([]interface{}); ok {
					for _, u := range users {
						user, ok := u.(map[string]interface{})
						if !ok {
							continue
						}
						uName := getString(user, "name")
						uRole := getString(user, "role")
						if uRole == "ADMIN" {
							bmItem.AdminNames = append(bmItem.AdminNames, uName)
						}
					}
				}
			} else {
				log.Printf("[FB-BM] 获取 business_users 失败 (bm=%s): %v", bmID, err)
			}

			// summary=total_count 取管理员真实总数（含不可枚举的用户）
			summaryResp, err := s.fbGet(
				fmt.Sprintf("/%s/%s/business_users", s.graphVer, bmID),
				map[string]string{
					"access_token": accessToken,
					"summary":      "total_count",
					"limit":        "0",
				},
			)
			if err == nil {
				if summary, ok := summaryResp["summary"].(map[string]interface{}); ok {
					bmItem.AdminCount = getInt(summary, "total_count")
				}
			}
			if bmItem.AdminCount == 0 {
				bmItem.AdminCount = len(bmItem.AdminNames)
			}

			// pending_users → 邀请中的管理员（已发出邀请未接受；无 name 字段）
			pendingResp, err := s.fbGet(
				fmt.Sprintf("/%s/%s/pending_users", s.graphVer, bmID),
				map[string]string{
					"fields":       "id,role",
					"access_token": accessToken,
					"limit":        "100",
				},
			)
			if err == nil {
				if pending, ok := pendingResp["data"].([]interface{}); ok {
					for _, u := range pending {
						user, ok := u.(map[string]interface{})
						if !ok {
							continue
						}
						if getString(user, "role") == "ADMIN" {
							bmItem.PendingAdminCount++
						}
					}
				}
			} else {
				log.Printf("[FB-BM] 获取 pending_users 失败 (bm=%s): %v", bmID, err)
			}

			// 广告账户数 = 自有 + 客户
			bmItem.AdAccountCount = s.countBmEdge(bmID, "owned_ad_accounts", accessToken) +
				s.countBmEdge(bmID, "client_ad_accounts", accessToken)

			// 合作伙伴数 = 子BM + 代理商
			bmItem.PartnerCount = s.countBmEdge(bmID, "owned_businesses", accessToken) +
				s.countBmEdge(bmID, "agencies", accessToken)

			allBms = append(allBms, bmItem)
		}
	}

	if allBms == nil {
		allBms = []models.FbBmListItem{}
	}

	return &models.FbBmListResponse{
		List:  allBms,
		Total: len(allBms),
	}, nil
}

// ==================== FB 公共主页 ====================

// GetPageList 获取所有已授权 FB 账号下的公共主页列表
func (s *FbService) GetPageList(userID uint, tenantID *uint) (*models.FbPageListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	s.init()

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

	var allPages []models.FbPageItem

	for rows.Next() {
		var (
			tokenID     uint
			fbUserID    string
			fbUserName  string
			accessToken string
		)
		if err := rows.Scan(&tokenID, &fbUserID, &fbUserName, &accessToken); err != nil {
			log.Printf("[FB-PAGE] 扫描 token 行失败: %v", err)
			continue
		}

		pageResp, err := s.fbGet(
			fmt.Sprintf("/%s/me/accounts", s.graphVer),
			map[string]string{
				"fields": "id,name,link,category,fan_count,followers_count,is_published," +
					"verification_status,website,phone,emails,location,access_token,tasks,business",
				"access_token": accessToken,
				"limit":        "100",
			},
		)
		if err != nil {
			log.Printf("[FB-PAGE] 获取主页列表失败 (fbUserId=%s): %v", fbUserID, err)
			continue
		}

		data, _ := pageResp["data"].([]interface{})
		for _, item := range data {
			page, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			pageID := getString(page, "id")
			if pageID == "" {
				continue
			}

			pageItem := models.FbPageItem{
				PageID:             pageID,
				Name:               getString(page, "name"),
				Link:               getString(page, "link"),
				FbOwnerName:        fbUserName,
				FbOwnerID:          fbUserID,
				Category:           getString(page, "category"),
				FanCount:           getInt(page, "fan_count"),
				FollowersCount:     getInt(page, "followers_count"),
				VerificationStatus: getString(page, "verification_status"),
				Website:            getString(page, "website"),
				Phone:              getString(page, "phone"),
				AdPerm:             -1,
				BlockedCount:       0,
				TokenID:            tokenID,
			}

			// 新版公共主页体验的 /roles、/blocked、/settings 边必须使用主页访问口令
			pageToken := getString(page, "access_token")
			if pageToken == "" {
				pageToken = accessToken
			}

			// 广告权限：tasks 中包含 ADVERTISE 即广告功能正常
			if tasks, ok := page["tasks"].([]interface{}); ok {
				pageItem.AdPerm = 0
				for _, task := range tasks {
					if ts, ok := task.(string); ok && ts == "ADVERTISE" {
						pageItem.AdPerm = 1
						break
					}
				}
			}

			// 所属 BM：Page 的 business 字段（未绑定 BM 时为空）
			if biz, ok := page["business"].(map[string]interface{}); ok {
				pageItem.BusinessName = getString(biz, "name")
			}

			if v, ok := page["is_published"].(bool); ok && !v {
				pageItem.IsPublished = 0
			}

			// emails 为数组，取第一个
			if emails, ok := page["emails"].([]interface{}); ok && len(emails) > 0 {
				if e, ok := emails[0].(string); ok {
					pageItem.Email = e
				}
			}

			// location 拼成单行地址
			if loc, ok := page["location"].(map[string]interface{}); ok {
				parts := []string{}
				for _, k := range []string{"street", "city", "state", "country"} {
					if v := getString(loc, k); v != "" {
						parts = append(parts, v)
					}
				}
				pageItem.Address = strings.Join(parts, ", ")
			}

			// 管理员名单：/roles 边（必须用主页访问口令），tasks 含 MANAGE 视为管理员，失败容忍
			rolesResp, rolesErr := s.fbGet(
				fmt.Sprintf("/%s/%s/roles", s.graphVer, pageID),
				map[string]string{
					"fields":       "name,tasks",
					"access_token": pageToken,
					"limit":        "100",
				},
			)
			if rolesErr == nil {
				if rolesData, ok := rolesResp["data"].([]interface{}); ok {
					for _, r := range rolesData {
						ru, ok := r.(map[string]interface{})
						if !ok {
							continue
						}
						isAdmin := false
						if rtasks, ok := ru["tasks"].([]interface{}); ok {
							for _, task := range rtasks {
								if ts, ok := task.(string); ok && ts == "MANAGE" {
									isAdmin = true
									break
								}
							}
						}
						if isAdmin {
							if n := getString(ru, "name"); n != "" {
								pageItem.AdminNames = append(pageItem.AdminNames, n)
							}
						}
					}
				}
			} else {
				log.Printf("[FB-PAGE] 获取主页管理员失败 (page=%s): %v", pageID, rolesErr)
			}

			// 黑名单数量：/blocked 边（主页访问口令），失败容忍
			blockedResp, blockedErr := s.fbGet(
				fmt.Sprintf("/%s/%s/blocked", s.graphVer, pageID),
				map[string]string{
					"access_token": pageToken,
					"limit":        "100",
				},
			)
			if blockedErr == nil {
				if blockedData, ok := blockedResp["data"].([]interface{}); ok {
					pageItem.BlockedCount = len(blockedData)
				}
			} else {
				log.Printf("[FB-PAGE] 获取主页黑名单失败 (page=%s): %v", pageID, blockedErr)
			}

			// 隐藏不文明用语：/settings 边的 PROFANITY_FILTER（none/medium/strong），失败容忍
			settingsResp, settingsErr := s.fbGet(
				fmt.Sprintf("/%s/%s/settings", s.graphVer, pageID),
				map[string]string{
					"access_token": pageToken,
					"limit":        "100",
				},
			)
			if settingsErr == nil {
				if settingsData, ok := settingsResp["data"].([]interface{}); ok {
					for _, item := range settingsData {
						if so, ok := item.(map[string]interface{}); ok {
							if getString(so, "setting") == "PROFANITY_FILTER" {
								if v, ok := so["value"].(string); ok {
									pageItem.ProfanityFilter = v
								}
							}
						}
					}
				}
			} else {
				log.Printf("[FB-PAGE] 获取主页设置失败 (page=%s): %v", pageID, settingsErr)
			}

			allPages = append(allPages, pageItem)
		}
	}

	if allPages == nil {
		allPages = []models.FbPageItem{}
	}

	return &models.FbPageListResponse{
		List:  allPages,
		Total: len(allPages),
	}, nil
}

// ==================== FB 像素 ====================

// parseFbTime 解析 FB ISO8601 时间（如 2024-01-02T15:04:05+0000）
func parseFbTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05-0700", s); err == nil {
		return &t
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

// GetPixelList 获取所有已授权 FB 账号下各广告账户的像素列表
// 数据源：/{act}/adspixels（需 ads_read/ads_management，无权限的广告账户跳过并记录）
func (s *FbService) GetPixelList(userID uint, tenantID *uint) (*models.FbPixelListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	s.init()

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

	type tokenRow struct {
		id    uint
		fbUid string
		fbNam string
		token string
	}
	var tokens []tokenRow
	for rows.Next() {
		var tr tokenRow
		if err := rows.Scan(&tr.id, &tr.fbUid, &tr.fbNam, &tr.token); err != nil {
			continue
		}
		tokens = append(tokens, tr)
	}

	var allPixels []models.FbPixelItem

	for _, tr := range tokens {
		// 该 FB 账号下的广告账户
		actResp, err := s.fbGet(
			fmt.Sprintf("/%s/me/adaccounts", s.graphVer),
			map[string]string{
				"fields":       "id,name",
				"access_token": tr.token,
				"limit":        "100",
			},
		)
		if err != nil {
			log.Printf("[FB-PIXEL] 获取广告账户失败 (fbUserId=%s): %v", tr.fbUid, err)
			continue
		}

		acts, _ := actResp["data"].([]interface{})
		for _, a := range acts {
			act, ok := a.(map[string]interface{})
			if !ok {
				continue
			}
			actID := getString(act, "id")
			actName := getString(act, "name")
			if actID == "" {
				continue
			}

			// 该广告账户下的像素（无 ads 权限的账户会报错，跳过即可）
			pxResp, err := s.fbGet(
				fmt.Sprintf("/%s/%s/adspixels", s.graphVer, actID),
				map[string]string{
					"fields": "id,name,creation_time,last_fired_time,is_unavailable," +
						"owner_business,creator",
					"access_token": tr.token,
					"limit":        "100",
				},
			)
			if err != nil {
				log.Printf("[FB-PIXEL] 获取像素失败 (act=%s): %v", actID, err)
				continue
			}

			pxData, _ := pxResp["data"].([]interface{})
			for _, p := range pxData {
				px, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				pixelID := getString(px, "id")
				if pixelID == "" {
					continue
				}

				item := models.FbPixelItem{
					PixelID:       pixelID,
					Name:          getString(px, "name"),
					AdAccountID:   actID,
					AdAccountName: actName,
					FbOwnerName:   tr.fbNam,
					CreationTime:  parseFbTime(getString(px, "creation_time")),
					LastFiredTime: parseFbTime(getString(px, "last_fired_time")),
					TokenID:       tr.id,
				}
				if un, ok := px["is_unavailable"].(bool); ok && un {
					item.IsUnavailable = 1
				}
				if biz, ok := px["owner_business"].(map[string]interface{}); ok {
					item.OwnerBmID = getString(biz, "id")
					item.OwnerBmName = getString(biz, "name")
				}
				if creator, ok := px["creator"].(map[string]interface{}); ok {
					item.CreatorName = getString(creator, "name")
				}

				// 以下边需要 BM 上下文，仅在像素归属 BM 时尝试；失败容忍
				if item.OwnerBmID != "" {
					// 像素上分配的用户（管理员/角色）
					auResp, auErr := s.fbGet(
						fmt.Sprintf("/%s/%s/assigned_users", s.graphVer, pixelID),
						map[string]string{
							"fields":       "name,role",
							"business":     item.OwnerBmID,
							"access_token": tr.token,
							"limit":        "100",
						},
					)
					if auErr == nil {
						if auData, ok := auResp["data"].([]interface{}); ok {
							for _, u := range auData {
								if au, ok := u.(map[string]interface{}); ok {
									if n := getString(au, "name"); n != "" {
										item.AdminNames = append(item.AdminNames, n)
									}
									if r := getString(au, "role"); r != "" {
										item.RoleNames = append(item.RoleNames, r)
									}
								}
							}
						}
					} else {
						log.Printf("[FB-PIXEL] 获取像素分配用户失败 (pixel=%s): %v", pixelID, auErr)
					}

					// 共享给的合作伙伴（agency）
					saResp, saErr := s.fbGet(
						fmt.Sprintf("/%s/%s/shared_agencies", s.graphVer, pixelID),
						map[string]string{
							"fields":       "name",
							"business":     item.OwnerBmID,
							"access_token": tr.token,
							"limit":        "100",
						},
					)
					if saErr == nil {
						if saData, ok := saResp["data"].([]interface{}); ok {
							for _, ag := range saData {
								if am, ok := ag.(map[string]interface{}); ok {
									if n := getString(am, "name"); n != "" {
										item.SharedAgencies = append(item.SharedAgencies, n)
									}
								}
							}
						}
					} else {
						log.Printf("[FB-PIXEL] 获取像素共享伙伴失败 (pixel=%s): %v", pixelID, saErr)
					}
				}

				allPixels = append(allPixels, item)
			}
		}
	}

	if allPixels == nil {
		allPixels = []models.FbPixelItem{}
	}

	return &models.FbPixelListResponse{
		List:  allPixels,
		Total: len(allPixels),
	}, nil
}

// CreatePixel 在指定广告账户下创建像素（POST /{act}/adspixels）
// 逐个尝试该用户已启用的 FB token，第一个成功的为准
func (s *FbService) CreatePixel(userID uint, tenantID *uint, adAccountID, name string) (string, error) {
	if db.Pool == nil {
		return "", fmt.Errorf("数据库未连接")
	}
	if adAccountID == "" || name == "" {
		return "", fmt.Errorf("广告账户和像素名称不能为空")
	}

	s.init()

	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT access_token FROM fb_tokens
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND status = 1`,
		userID, tenantID,
	)
	if err != nil {
		return "", fmt.Errorf("查询 FB token 失败: %w", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var tk string
		if err := rows.Scan(&tk); err != nil {
			continue
		}
		tokens = append(tokens, tk)
	}
	if len(tokens) == 0 {
		return "", fmt.Errorf("没有已启用的 FB 账号，请先连接 Facebook")
	}

	var lastErr error
	for _, tk := range tokens {
		resp, err := s.fbPost(
			fmt.Sprintf("/%s/%s/adspixels", s.graphVer, adAccountID),
			map[string]string{
				"name":         name,
				"access_token": tk,
			},
		)
		if err != nil {
			lastErr = err
			continue
		}
		if id := getString(resp, "id"); id != "" {
			return id, nil
		}
		lastErr = fmt.Errorf("Facebook 未返回像素 ID")
	}

	return "", fmt.Errorf("创建像素失败: %v", lastErr)
}

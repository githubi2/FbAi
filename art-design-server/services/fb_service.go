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
	"time"

	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
)

// FbService Facebook 服务
type FbService struct {
	appID       string
	appSecret   string
	redirectURI string
	graphAPI    string
	graphVer    string
}

var DefaultFbService = &FbService{}

// init 从环境变量加载 Facebook 配置
func (s *FbService) init() {
	if s.appID == "" {
		s.appID = os.Getenv("FB_APP_ID")
		s.appSecret = os.Getenv("FB_APP_SECRET")
		s.redirectURI = os.Getenv("FB_REDIRECT_URI")
		s.graphAPI = "https://graph.facebook.com"
		s.graphVer = os.Getenv("FB_GRAPH_VERSION")
		if s.graphVer == "" {
			s.graphVer = "v22.0"
		}
	}
}

// GetAuthURL 生成 Facebook OAuth 授权链接
// state 参数用于 CSRF 防护并携带 userID，回调时验证
func (s *FbService) GetAuthURL(userID uint) (string, error) {
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

	// 存储 state 用于回调验证（5分钟有效）
	if db.Pool != nil {
		ctx := context.Background()
		_, err := db.Pool.Exec(ctx,
			`INSERT INTO fb_tokens (user_id, access_token, status, created_at, updated_at)
			 VALUES ($1, $2, 0, NOW(), NOW())
			 ON CONFLICT (user_id) DO UPDATE
			 SET access_token = $2, status = 0, updated_at = NOW()`,
			userID, "pending:"+state,
		)
		if err != nil {
			log.Printf("[FB] 存储 state 失败: %v", err)
		}
	}

	// 需要的权限
	scopes := []string{
		"ads_management",     // 广告管理
		"ads_read",           // 广告读取
		"business_management", // 商务管理平台
	}

	authURL := fmt.Sprintf(
		"%s/%s/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s&response_type=code",
		s.graphAPI, s.graphVer,
		url.QueryEscape(s.appID),
		url.QueryEscape(s.redirectURI),
		url.QueryEscape(strings.Join(scopes, ",")),
		state,
	)

	return authURL, nil
}

// ExchangeCodeForToken 用授权码换取 access token
// state 包含编码的 userID（格式: hex(userID):nonce）
// 返回 FbToken 和对应的 userID
func (s *FbService) ExchangeCodeForToken(code, state string) (*models.FbToken, uint, error) {
	s.init()

	if s.appID == "" || s.appSecret == "" {
		return nil, 0, fmt.Errorf("Facebook 应用未配置")
	}

	// 从 state 解析 userID
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("无效的 state 参数")
	}

	userIDHex, nonce := parts[0], parts[1]

	// 解码 userID（hex → uint）
	var userID uint64
	if _, err := fmt.Sscanf(userIDHex, "%x", &userID); err != nil {
		return nil, 0, fmt.Errorf("无效的 userID 编码: %w", err)
	}

	// 验证 state（CSRF 防护 + 确认是本人发起的请求）
	if db.Pool != nil {
		var storedToken string
		ctx := context.Background()
		err := db.Pool.QueryRow(ctx,
			`SELECT access_token FROM fb_tokens
			 WHERE user_id = $1 AND access_token LIKE 'pending:%'
			   AND status = 0 AND updated_at > NOW() - INTERVAL '5 minutes'`,
			userID,
		).Scan(&storedToken)
		if err != nil || storedToken != "pending:"+state {
			return nil, 0, fmt.Errorf("无效的 state 参数，可能为 CSRF 攻击或授权已过期")
		}
		// 用 nonce 做二次验证
		_ = nonce
	}

	// 构建 token 交换请求
	tokenURL := fmt.Sprintf("%s/%s/oauth/access_token", s.graphAPI, s.graphVer)
	resp, err := http.Get(fmt.Sprintf(
		"%s?client_id=%s&redirect_uri=%s&client_secret=%s&code=%s",
		tokenURL,
		url.QueryEscape(s.appID),
		url.QueryEscape(s.redirectURI),
		url.QueryEscape(s.appSecret),
		url.QueryEscape(code),
	))
	if err != nil {
		return nil, 0, fmt.Errorf("请求 Facebook token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("读取 Facebook 响应失败: %w", err)
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
		return nil, 0, fmt.Errorf("解析 Facebook token 响应失败: %w", err)
	}

	if tokenResp.Error != nil {
		return nil, 0, fmt.Errorf("Facebook 返回错误: %s (type=%s, code=%d)",
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
		Scopes:      []string{"ads_management", "ads_read", "business_management"},
		Status:      1,
	}, uint(userID), nil
}

// exchangeLongLivedToken 用短期 token 换取长期 token（60天有效）
func (s *FbService) exchangeLongLivedToken(shortToken string) (string, error) {
	s.init()

	tokenURL := fmt.Sprintf("%s/%s/oauth/access_token", s.graphAPI, s.graphVer)
	resp, err := http.Get(fmt.Sprintf(
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

	resp, err := http.Get(fmt.Sprintf(
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

// SaveToken 保存或更新用户的 Facebook token
func (s *FbService) SaveToken(userID uint, token *models.FbToken) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	scopesJSON, _ := json.Marshal(token.Scopes)

	_, err := db.Pool.Exec(ctx,
		`INSERT INTO fb_tokens (user_id, fb_user_id, fb_user_name, access_token, token_type, expires_at, scopes, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 1, NOW(), NOW())
		 ON CONFLICT (user_id) DO UPDATE
		 SET fb_user_id = $2, fb_user_name = $3, access_token = $4, token_type = $5,
		     expires_at = $6, scopes = $7, status = 1, updated_at = NOW()`,
		userID, token.FbUserID, token.FbUserName, token.AccessToken,
		token.TokenType, token.ExpiresAt, string(scopesJSON),
	)
	if err != nil {
		return fmt.Errorf("保存 token 失败: %w", err)
	}

	return nil
}

// GetToken 获取用户的有效 Facebook token
func (s *FbService) GetToken(userID uint) (*models.FbToken, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	var token models.FbToken
	var scopesStr string
	var bmListStr, adAccsStr string
	var expiresAt time.Time

	err := db.Pool.QueryRow(ctx,
		`SELECT id, user_id, fb_user_id, fb_user_name, access_token, token_type, expires_at,
		        COALESCE(scopes::text, '[]'), COALESCE(bm_list::text, '[]'), COALESCE(ad_accounts::text, '[]'),
		        selected_ad_account_id, status, created_at, updated_at
		 FROM fb_tokens WHERE user_id = $1 AND status = 1`,
		userID,
	).Scan(&token.ID, &token.UserID, &token.FbUserID, &token.FbUserName,
		&token.AccessToken, &token.TokenType, &expiresAt,
		&scopesStr, &bmListStr, &adAccsStr,
		&token.SelectedAdAccountID, &token.Status, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("未找到有效的 Facebook 授权: %w", err)
	}

	token.ExpiresAt = expiresAt
	token.BmList = bmListStr
	token.AdAccounts = adAccsStr

	// 解析 scopes
	json.Unmarshal([]byte(scopesStr), &token.Scopes)

	return &token, nil
}

// GetConnectionStatus 获取用户的 Facebook 连接状态
func (s *FbService) GetConnectionStatus(userID uint) *models.FbConnectionStatusResponse {
	token, err := s.GetToken(userID)
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

// GetAdAccounts 获取用户可访问的广告账户列表
func (s *FbService) GetAdAccounts(userID uint) (*models.FbAdAccountListResponse, error) {
	token, err := s.GetToken(userID)
	if err != nil {
		return nil, err
	}

	s.init()

	// 获取广告账户
	adAccResp, err := s.fbGet(
		fmt.Sprintf("/%s/adaccounts", s.graphVer),
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
			`UPDATE fb_tokens SET ad_accounts = $1, bm_list = $2, updated_at = NOW() WHERE user_id = $3`,
			string(accJSON), string(bmJSON), userID,
		)
	}

	return &models.FbAdAccountListResponse{
		AdAccounts: adAccounts,
		Businesses: businesses,
	}, nil
}

// Disconnect 断开 Facebook 连接
func (s *FbService) Disconnect(userID uint) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx,
		`UPDATE fb_tokens SET status = 0, updated_at = NOW() WHERE user_id = $1`,
		userID,
	)
	return err
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

	resp, err := http.Get(u.String())
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

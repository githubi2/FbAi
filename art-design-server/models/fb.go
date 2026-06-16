package models

import "time"

// FbToken Facebook 授权令牌模型
type FbToken struct {
	ID                   uint      `json:"id"`
	UserID               int       `json:"userId"`
	FbUserID             string    `json:"fbUserId"`
	FbUserName           string    `json:"fbUserName"`
	AccessToken          string    `json:"-"` // 不序列化到 JSON（安全）
	TokenType            string    `json:"tokenType"`
	ExpiresAt            time.Time `json:"expiresAt"`
	Scopes               []string  `json:"scopes"`
	BmList              string    `json:"bmList"`     // JSONB 字符串
	AdAccounts          string    `json:"adAccounts"` // JSONB 字符串
	SelectedAdAccountID  string    `json:"selectedAdAccountId"`
	Status               int       `json:"status"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// FbAuthURLResponse 授权链接响应
type FbAuthURLResponse struct {
	AuthURL string `json:"authUrl"`
}

// FbConnectionStatusResponse 连接状态响应
type FbConnectionStatusResponse struct {
	Connected           bool   `json:"connected"`
	FbUserID            string `json:"fbUserId"`
	FbUserName          string `json:"fbUserName"`
	ExpiresAt           string `json:"expiresAt"`
	SelectedAdAccountID string `json:"selectedAdAccountId"`
	Scopes              []string `json:"scopes"`
}

// FbAdAccount 广告账户
type FbAdAccount struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	Name          string `json:"name"`
	AccountStatus int    `json:"accountStatus"` // 1=active, 2=disabled, 3=unsettled, 7=pending, 9=inactive
	Currency      string `json:"currency"`
	BusinessName  string `json:"businessName"`
}

// FbBusinessManager Facebook Business Manager
type FbBusinessManager struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FbAdAccountListResponse 广告账户列表响应
type FbAdAccountListResponse struct {
	AdAccounts   []FbAdAccount       `json:"adAccounts"`
	Businesses   []FbBusinessManager `json:"businesses"`
}

package models

import "time"

// FbToken Facebook 授权令牌模型
type FbToken struct {
	ID                   uint      `json:"id"`
	UserID               int       `json:"userId"`
	TenantID             *uint     `json:"tenantId"` // NULL=超级管理员
	FbUserID             string    `json:"fbUserId"`
	FbUserName           string    `json:"fbUserName"`
	Label                string    `json:"label"` // 用户自定义备注（如"主账号"）
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
	AuthURL  string `json:"authUrl"`
	ShortURL string `json:"shortUrl"`
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

// ==================== FB 账号列表（多账号改造） ====================

// FbAccountListItem FB 账号列表项（前端表格行数据）
type FbAccountListItem struct {
	ID              uint      `json:"id"`
	FbUserID        string    `json:"fbUserId"`
	FbUserName      string    `json:"fbUserName"`
	Label           string    `json:"label"`
	Scopes          []string  `json:"scopes"`
	ExpiresAt       string    `json:"expiresAt"`       // ISO 时间字符串
	CreatedAt       string    `json:"createdAt"`       // ISO 时间字符串
	DaysUntilExpiry int       `json:"daysUntilExpiry"` // 剩余天数（负数=已过期）
	HasAdPerm       bool      `json:"hasAdPerm"`       // 是否有广告权限
	AccountStatus   string    `json:"accountStatus"`   // "正常" / "已过期"
	BmCount         int       `json:"bmCount"`         // BM 总个数
	PersonalAdCount int       `json:"personalAdCount"` // 个人广告账户数量
	BmAdCount       int       `json:"bmAdCount"`       // BM 下广告账户数量
}

// FbAccountListResponse FB 账号列表响应
type FbAccountListResponse struct {
	Accounts []FbAccountListItem `json:"accounts"`
	Total    int                 `json:"total"`
}

// FbUpdateLabelRequest 更新备注请求
type FbUpdateLabelRequest struct {
	Label string `json:"label" binding:"max=64"`
}

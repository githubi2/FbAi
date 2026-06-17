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
	LastError            string    `json:"lastError"`     // 最近一次 FB API 调用失败的错误信息
	LastErrorAt          *time.Time `json:"lastErrorAt"`  // 最近一次错误发生时间
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

// FbAdAccount 广告账户（基础信息）
type FbAdAccount struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	Name          string `json:"name"`
	AccountStatus int    `json:"accountStatus"` // 1=active, 2=disabled, 3=unsettled, 7=pending, 9=inactive
	Currency      string `json:"currency"`
	BusinessName  string `json:"businessName"`
}

// FbAdAccountDetail 广告账户详细信息（含消耗/限额/管理员等）
type FbAdAccountDetail struct {
	ID             string  `json:"id"`             // act_xxx 格式
	AccountID      string  `json:"accountId"`      // 数字ID
	Name           string  `json:"name"`           // 账户名称
	FbOwnerName    string  `json:"fbOwnerName"`    // 所属FB账号名称
	FbOwnerID      string  `json:"fbOwnerId"`      // 所属FB账号ID
	BusinessName   string  `json:"businessName"`   // 所属BM名称
	AccountStatus  int     `json:"accountStatus"`  // 状态码
	StatusLabel    string  `json:"statusLabel"`    // 状态显示文本
	Platform       string  `json:"platform"`       // 平台（facebook）
	AmountSpent    float64 `json:"amountSpent"`    // 总消耗金额
	Currency       string  `json:"currency"`       // 货币
	SpendCap       float64 `json:"spendCap"`       // 限额
	Balance        float64 `json:"balance"`        // 余额/下笔扣款额度
	AdminName      string  `json:"adminName"`      // 主管理员名称
	HiddenAdmins   int     `json:"hiddenAdmins"`   // 隐藏管理员个数
	TimezoneName   string  `json:"timezoneName"`   // 时区名称（如 Asia/Taipei）
	TimezoneOffset float64 `json:"timezoneOffset"`  // UTC偏移小时（如 8）
	CreatedTime    string  `json:"createdTime"`     // 创建时间
}

// FbAdAccountDetailListResponse 广告账户详细列表响应
type FbAdAccountDetailListResponse struct {
	Accounts []FbAdAccountDetail `json:"accounts"`
	Total    int                 `json:"total"`
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
	AccountStatus   string    `json:"accountStatus"`   // "正常" / "已过期" / "异常"
	BmCount         int       `json:"bmCount"`         // BM 总个数
	PersonalAdCount int       `json:"personalAdCount"` // 个人广告账户数量
	BmAdCount       int       `json:"bmAdCount"`       // BM 下广告账户数量
	DataError       string    `json:"dataError"`       // 数据拉取失败时的错误信息
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

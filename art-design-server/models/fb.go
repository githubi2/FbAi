package models

import "time"

// FbToken Facebook 授权令牌模型
type FbToken struct {
	ID                  uint       `json:"id"`
	UserID              int        `json:"userId"`
	TenantID            *uint      `json:"tenantId"` // NULL=超级管理员
	FbUserID            string     `json:"fbUserId"`
	FbUserName          string     `json:"fbUserName"`
	Label               string     `json:"label"` // 用户自定义备注（如"主账号"）
	AccessToken         string     `json:"-"`     // 不序列化到 JSON（安全）
	TokenType           string     `json:"tokenType"`
	ExpiresAt           time.Time  `json:"expiresAt"`
	Scopes              []string   `json:"scopes"`
	BmList              string     `json:"bmList"`     // JSONB 字符串
	AdAccounts          string     `json:"adAccounts"` // JSONB 字符串
	SelectedAdAccountID string     `json:"selectedAdAccountId"`
	Status              int        `json:"status"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	LastError           string     `json:"lastError"`    // 最近一次 FB API 调用失败的错误信息
	LastErrorAt         *time.Time `json:"lastErrorAt"`  // 最近一次错误发生时间
}

// FbAuthURLResponse 授权链接响应
type FbAuthURLResponse struct {
	AuthURL  string `json:"authUrl"`
	ShortURL string `json:"shortUrl"`
}

// FbConnectionStatusResponse 连接状态响应
type FbConnectionStatusResponse struct {
	Connected           bool     `json:"connected"`
	FbUserID            string   `json:"fbUserId"`
	FbUserName          string   `json:"fbUserName"`
	ExpiresAt           string   `json:"expiresAt"`
	SelectedAdAccountID string   `json:"selectedAdAccountId"`
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
	ID                 string   `json:"id"`                 // act_xxx 格式
	AccountID          string   `json:"accountId"`          // 数字ID
	Name               string   `json:"name"`               // 账户名称
	FbOwnerName        string   `json:"fbOwnerName"`        // 所属FB账号名称
	FbOwnerID          string   `json:"fbOwnerId"`          // 所属FB账号ID
	BusinessName       string   `json:"businessName"`       // 所属BM名称
	OwnerBusinessID    string   `json:"ownerBusinessId"`    // 所有者BM ID
	AccountStatus      int      `json:"accountStatus"`      // 状态码
	StatusLabel        string   `json:"statusLabel"`        // 状态显示文本
	Platform           string   `json:"platform"`           // 平台（facebook）
	AmountSpent        float64  `json:"amountSpent"`        // 总消耗/已花费金额
	Currency           string   `json:"currency"`           // 货币
	SpendCap           float64  `json:"spendCap"`           // 花费限额
	Balance            float64  `json:"balance"`            // 余额
	DailySpendLimit    float64  `json:"dailySpendLimit"`    // 日限额
	AdminName          string   `json:"adminName"`          // 主管理员名称
	HiddenAdmins       int      `json:"hiddenAdmins"`       // 隐藏管理员个数
	OtherAdminNames    []string `json:"otherAdminNames"`    // 其他管理员名称列表（排除主管理员）
	TimezoneName       string   `json:"timezoneName"`       // 时区名称（如 Asia/Taipei）
	TimezoneOffset     float64  `json:"timezoneOffset"`     // UTC偏移小时（如 8）
	CountryCode        string   `json:"countryCode"`        // 国家编码
	IsPersonal         int      `json:"isPersonal"`         // 是否个人广告账户（1=个人, 0=BM）
	FundingSource      string   `json:"fundingSource"`      // 支付方法
	DisableReason      int      `json:"disableReason"`      // 锁定原因状态码
	DisableReasonLabel string   `json:"disableReasonLabel"` // 锁定原因显示文本
	NextBillDate       string   `json:"nextBillDate"`       // 下个账单日期
	CreatedTime        string   `json:"createdTime"`        // 创建时间
	IsPrepay           int      `json:"isPrepay"`           // 是否预付费账户（1=预付费, 0=后付费）
	OwnerRole          string   `json:"ownerRole"`          // 授权用户在该账户的角色（Admin/Advertiser/Analyst）
	Remark             string   `json:"remark"`             // 本地备注（仅存缓存表，刷新不覆盖）
	TokenID            uint     `json:"-"`                  // 所属 fb_tokens.id（内部缓存用，不返回前端）
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

// FbBmListItem BM 列表项（缓存表 fb_bm_cache 映射）
// 字段来源说明（官方 Marketing API 实测可用）：
//   - name / createdTime / verificationStatus → GET /me/businesses?fields=id,name,created_time,verification_status
//   - ownerRole / adminCount / adminNames     → GET /{bm-id}/business_users?fields=id,name,role
//   - adAccountCount                          → owned_ad_accounts + client_ad_accounts 边计数
//   - partnerCount                            → owned_businesses + agencies 边计数
//   - statusLabel                             —— 官方无 BM 状态字段，API 可达即为「正常」
//   - pushStatus / remark                     —— 本地字段，FB 刷新不覆盖
// 官方公开 API 不提供：BM类型、日限额、隐藏管理员、BM质量（前端显示 —）
type FbBmListItem struct {
	BmID               string     `json:"bmId"`
	Name               string     `json:"name"`
	FbOwnerName        string     `json:"fbOwnerName"`
	FbOwnerID          string     `json:"fbOwnerId"`
	StatusLabel        string     `json:"statusLabel"`
	PushStatus         string     `json:"pushStatus"`
	Remark             string     `json:"remark"`
	OwnerRole          string     `json:"ownerRole"`
	VerificationStatus string     `json:"verificationStatus"`
	AdminCount         int        `json:"adminCount"`        // 管理员总数 = 在职 + 邀请中
	PendingAdminCount  int        `json:"pendingAdminCount"` // 邀请中管理员数
	AdminNames         []string   `json:"adminNames"`        // 在职管理员名单
	PartnerCount       int        `json:"partnerCount"`
	AdAccountCount     int        `json:"adAccountCount"`
	CreatedTime        string     `json:"createdTime"`
	LastRefreshAt      *time.Time `json:"lastRefreshAt"`
	TokenID            uint       `json:"-"` // 所属 fb_tokens.id（内部缓存用，不返回前端）
}

// FbBmListResponse BM 列表响应
type FbBmListResponse struct {
	List  []FbBmListItem `json:"list"`
	Total int            `json:"total"`
}

// ==================== FB 公共主页 ====================

// FbPageItem 公共主页列表项
type FbPageItem struct {
	PageID             string     `json:"pageId"`
	Name               string     `json:"name"`
	Link               string     `json:"link"`
	FbOwnerName        string     `json:"fbOwnerName"`
	FbOwnerID          string     `json:"fbOwnerId"`
	Category           string     `json:"category"`
	FanCount           int        `json:"fanCount"`
	FollowersCount     int        `json:"followersCount"`
	IsPublished        int        `json:"isPublished"` // 1=已发布 0=未发布
	VerificationStatus string     `json:"verificationStatus"`
	Website            string     `json:"website"`
	Phone              string     `json:"phone"`
	Email              string     `json:"email"`
	Address            string     `json:"address"`
	AdminNames         []string   `json:"adminNames"`
	BusinessName       string     `json:"bmName"`         // 所属 BM 名称（Page business 字段）
	AdPerm             int        `json:"adPerm"`         // 广告权限 1=正常 0=无权限 -1=未知（tasks 含 ADVERTISE）
	ProfanityFilter    string     `json:"profanityFilter"` // 隐藏不文明用语 none/medium/strong
	BlockedCount       int        `json:"blockedCount"`   // 黑名单数量（/blocked 边）
	PushStatus         string     `json:"pushStatus"` // 本地字段：推送状态
	Remark             string     `json:"remark"`     // 本地字段：备注
	LastRefreshAt      *time.Time `json:"lastRefreshAt"`
	TokenID            uint       `json:"-"` // 所属 fb_tokens.id（内部缓存用，不返回前端）
}

// FbPageListResponse 公共主页列表响应
type FbPageListResponse struct {
	List  []FbPageItem `json:"list"`
	Total int          `json:"total"`
}

// ==================== FB 广告投放（只读监控，v26.0）====================

// FbInsight 广告数据统计（近 7 天）
type FbInsight struct {
	Spend       string `json:"spend"`
	Impressions string `json:"impressions"`
	Clicks      string `json:"clicks"`
	CTR         string `json:"ctr"`
	CPC         string `json:"cpc"`
}

// FbCampaign 广告系列
type FbCampaign struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`          // ACTIVE/PAUSED/ARCHIVED/DELETED
	EffectiveStatus string     `json:"effectiveStatus"` // 实际生效状态
	Objective       string     `json:"objective"`       // 目标（OUTCOME_*）
	DailyBudget     string     `json:"dailyBudget"`     // FB 返回字符串（"0"=未设置）
	LifetimeBudget  string     `json:"lifetimeBudget"`
	BidStrategy     string     `json:"bidStrategy"`
	StartTime       string     `json:"startTime"`
	StopTime        string     `json:"stopTime"`
	CreatedTime     string     `json:"createdTime"`
	UpdatedTime     string     `json:"updatedTime"`
	Insight         *FbInsight `json:"insight,omitempty"` // 近 7 天统计（insights 合并）
	AccountID       string     `json:"accountId"`         // 所属广告账户（聚合时填充）
	AccountName     string     `json:"accountName"`       // 所属广告账户名（聚合时填充）
	AccountBM       string     `json:"accountBm"`         // 所属 BM 名称（可为空）
}

// FbCampaignListResponse 广告系列响应
type FbCampaignListResponse struct {
	List      []FbCampaign `json:"list"`
	Total     int          `json:"total"`
	AccountID string       `json:"accountId"`
}

// FbAdSet 广告组
type FbAdSet struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	EffectiveStatus  string `json:"effectiveStatus"`
	OptimizationGoal string `json:"optimizationGoal"`
	BillingEvent     string `json:"billingEvent"`
	DailyBudget      string `json:"dailyBudget"`
	LifetimeBudget   string `json:"lifetimeBudget"`
	StartTime        string `json:"startTime"`
	StopTime         string `json:"stopTime"`
	CreatedTime      string `json:"createdTime"`
	CampaignName     string `json:"campaignName"` // 所属系列名（账户级查询时填充）
	AccountID        string `json:"accountId"`    // 所属广告账户（聚合时填充）
	AccountName      string `json:"accountName"`  // 所属广告账户名（聚合时填充）
	AccountBM        string `json:"accountBm"`    // 所属 BM 名称（可为空）
}

// FbAdSetListResponse 广告组响应
type FbAdSetListResponse struct {
	List  []FbAdSet `json:"list"`
	Total int       `json:"total"`
}

// FbAd 广告
type FbAd struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	EffectiveStatus string `json:"effectiveStatus"`
	CreativeID      string `json:"creativeId"`
	CreativeName    string `json:"creativeName"`
	CreatedTime     string `json:"createdTime"`
	UpdatedTime     string `json:"updatedTime"`
	CampaignName    string `json:"campaignName"` // 所属系列名（账户级查询时填充）
	AdsetName       string `json:"adsetName"`    // 所属广告组名（账户级查询时填充）
	AccountID       string `json:"accountId"`    // 所属广告账户（聚合时填充）
	AccountName     string `json:"accountName"`  // 所属广告账户名（聚合时填充）
	AccountBM       string `json:"accountBm"`    // 所属 BM 名称（可为空）
}

// FbAdListResponse 广告响应
type FbAdListResponse struct {
	List  []FbAd `json:"list"`
	Total int    `json:"total"`
}

// ==================== FB 像素 ====================

// FbPixelItem 像素列表项
type FbPixelItem struct {
	PixelID        string     `json:"pixelId"`
	Name           string     `json:"name"`
	AdAccountID    string     `json:"adAccountId"`   // 所属广告账号 act_xxx
	AdAccountName  string     `json:"adAccountName"` // 所属广告账号名称
	OwnerBmID      string     `json:"ownerBmId"`
	OwnerBmName    string     `json:"ownerBmName"`
	CreatorName    string     `json:"creatorName"`
	IsUnavailable  int        `json:"isUnavailable"` // 1=不可用 0=正常
	CreationTime   *time.Time `json:"creationTime"`
	LastFiredTime  *time.Time `json:"lastFiredTime"` // 最近一次上报事件时间
	RoleNames      []string   `json:"roleNames"`     // 当前用户在像素上的角色
	AdminNames     []string   `json:"adminNames"`
	SharedAgencies []string   `json:"sharedAgencies"` // 共享合作伙伴名单
	Remark         string     `json:"remark"`         // 本地字段：备注
	LastRefreshAt  *time.Time `json:"lastRefreshAt"`
	FbOwnerName    string     `json:"fbOwnerName"` // 所属 FB 账号名
	TokenID        uint       `json:"-"`           // 所属 fb_tokens.id（内部缓存用，不返回前端）
}

// FbPixelListResponse 像素列表响应
type FbPixelListResponse struct {
	List  []FbPixelItem `json:"list"`
	Total int           `json:"total"`
}

// FbAdAccountListResponse 广告账户列表响应
type FbAdAccountListResponse struct {
	AdAccounts []FbAdAccount       `json:"adAccounts"`
	Businesses []FbBusinessManager `json:"businesses"`
}

// ==================== FB 账号列表（多账号改造） ====================

// FbAccountListItem FB 账号列表项（前端表格行数据）
type FbAccountListItem struct {
	ID              uint     `json:"id"`
	FbUserID        string   `json:"fbUserId"`
	FbUserName      string   `json:"fbUserName"`
	Label           string   `json:"label"`
	Scopes          []string `json:"scopes"`
	ExpiresAt       string   `json:"expiresAt"`       // ISO 时间字符串
	CreatedAt       string   `json:"createdAt"`       // ISO 时间字符串
	DaysUntilExpiry int      `json:"daysUntilExpiry"` // 剩余天数（负数=已过期）
	HasAdPerm       bool     `json:"hasAdPerm"`       // 是否有广告权限
	AccountStatus   string   `json:"accountStatus"`   // "正常" / "已过期" / "异常"
	BmCount         int      `json:"bmCount"`         // BM 总个数
	PersonalAdCount int      `json:"personalAdCount"` // 个人广告账户数量
	BmAdCount       int      `json:"bmAdCount"`       // BM 下广告账户数量
	DataError       string   `json:"dataError"`       // 数据拉取失败时的错误信息
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

// FbPaymentRecord 支付/交易记录
type FbPaymentRecord struct {
	ID            string  `json:"id"`
	AccountID     string  `json:"accountId"`
	Time          string  `json:"time"`
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	BillingStart  string  `json:"billingStart"`
	BillingEnd    string  `json:"billingEnd"`
	Status        string  `json:"status"`
	PaymentMethod string  `json:"paymentMethod"`
}

// FbPaymentListResponse 支付记录列表响应
type FbPaymentListResponse struct {
	Records []FbPaymentRecord `json:"records"`
	Total   int               `json:"total"`
}

// ==================== 广告账户授权 ====================

// FbAssignUserRequest 广告账户授权请求
type FbAssignUserRequest struct {
	AdAccountIDs []string `json:"adAccountIds" binding:"required,min=1"`                      // act_xxx 格式的广告账户 ID 列表
	UserID       string   `json:"userId" binding:"required"`                                  // Facebook 用户 UID
	Role         string   `json:"role" binding:"required,oneof=ADMIN ADVERTISER ANALYST"`     // 角色
}

// FbAssignUserResult 单个广告账户的授权结果
type FbAssignUserResult struct {
	AdAccountID string `json:"adAccountId"`
	Success     bool   `json:"success"`
	Message     string `json:"message"` // 成功或错误信息
}

// FbAssignUserResponse 广告账户授权响应
type FbAssignUserResponse struct {
	Results []FbAssignUserResult `json:"results"`
	Total   int                  `json:"total"`
	Success int                  `json:"success"`
	Failed  int                  `json:"failed"`
}

// FbLookupUserRequest 查找 Facebook 用户请求
type FbLookupUserRequest struct {
	UIDs []string `json:"uids" binding:"required,min=1"` // Facebook UID 列表
}

// FbLookupUserResult 单个用户查找结果
type FbLookupUserResult struct {
	UID      string `json:"uid"`
	Name     string `json:"name"`
	IsFriend bool   `json:"isFriend"`
	Avatar   string `json:"avatar"`
}

// FbLookupUserResponse 查找用户响应
type FbLookupUserResponse struct {
	Users []FbLookupUserResult `json:"users"`
}

// ==================== 删除广告账号权限 ====================

// FbRemoveUserRequest 删除广告账号权限请求
type FbRemoveUserRequest struct {
	AdAccountIDs  []string `json:"adAccountIds" binding:"required,min=1"` // 广告账户 ID 列表
	UIDs          []string `json:"uids"`                                  // 要删除的 Facebook UID 列表（部分模式可为空）
	Mode          string   `json:"mode" binding:"required"`               // 删除模式
	DeleteCurrent bool     `json:"deleteCurrent"`                         // 是否删除当前 FB 账号权限
}

// ==================== 缓存表模型 ====================

// FbAccountCache FB账号缓存
type FbAccountCache struct {
	ID              int        `json:"id"`
	UserID          int        `json:"userId"`
	TenantID        *int       `json:"tenantId"`
	FbTokenID       int        `json:"fbTokenId"`
	FbUserID        string     `json:"fbUserId"`
	FbUserName      string     `json:"fbUserName"`
	Label           string     `json:"label"`
	Scopes          string     `json:"scopes"` // JSON array string
	ExpiresAt       *time.Time `json:"expiresAt"`
	DaysUntilExpiry int        `json:"daysUntilExpiry"`
	HasAdPerm       bool       `json:"hasAdPerm"`
	AccountStatus   string     `json:"accountStatus"`
	BmCount         int        `json:"bmCount"`
	PersonalAdCount int        `json:"personalAdCount"`
	BmAdCount       int        `json:"bmAdCount"`
	DataError       string     `json:"dataError"`
	LastRefreshAt   *time.Time `json:"lastRefreshAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// FbAdAccountCache 广告账户缓存
type FbAdAccountCache struct {
	ID                 int        `json:"id"`
	UserID             int        `json:"userId"`
	TenantID           *int       `json:"tenantId"`
	FbTokenID          int        `json:"fbTokenId"`
	AdAccountID        string     `json:"adAccountID"`      // act_xxx
	AccountID          string     `json:"accountId"`        // 数字ID
	Name               string     `json:"name"`
	FbOwnerName        string     `json:"fbOwnerName"`
	FbOwnerID          string     `json:"fbOwnerId"`
	BusinessName       string     `json:"businessName"`
	OwnerBusinessID    string     `json:"ownerBusinessId"`
	AccountStatus      int        `json:"accountStatus"`
	StatusLabel        string     `json:"statusLabel"`
	Platform           string     `json:"platform"`
	AmountSpent        float64    `json:"amountSpent"`
	Currency           string     `json:"currency"`
	SpendCap           float64    `json:"spendCap"`
	Balance            float64    `json:"balance"`
	DailySpendLimit    float64    `json:"dailySpendLimit"`
	AdminName          string     `json:"adminName"`
	HiddenAdmins       int        `json:"hiddenAdmins"`
	OtherAdminNames    string     `json:"otherAdminNames"` // JSON array string
	TimezoneName       string     `json:"timezoneName"`
	TimezoneOffset     float64    `json:"timezoneOffset"`
	CountryCode        string     `json:"countryCode"`
	IsPersonal         int        `json:"isPersonal"`
	FundingSource      string     `json:"fundingSource"`
	DisableReason      int        `json:"disableReason"`
	DisableReasonLabel string     `json:"disableReasonLabel"`
	NextBillDate       string     `json:"nextBillDate"`
	CreatedTime        string     `json:"createdTime"`
	IsPrepay           int        `json:"isPrepay"`  // 是否预付费账户（1=预付费, 0=后付费）
	OwnerRole          string     `json:"ownerRole"` // 授权用户角色
	Remark             string     `json:"remark"`    // 本地备注（刷新不覆盖）
	LastRefreshAt      *time.Time `json:"lastRefreshAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// FbRefreshStatus 刷新状态
type FbRefreshStatus struct {
	ID           int        `json:"id"`
	UserID       int        `json:"userId"`
	TenantID     *int       `json:"tenantId"`
	RefreshType  string     `json:"refreshType"`
	Status       string     `json:"status"` // pending/running/completed/failed
	StartedAt    time.Time  `json:"startedAt"`
	CompletedAt  *time.Time `json:"completedAt"`
	ErrorMessage string     `json:"errorMessage"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// FbRefreshStatusResponse 刷新状态响应
type FbRefreshStatusResponse struct {
	Status    string `json:"status"`    // pending/running/completed/failed
	IsRunning bool   `json:"isRunning"` // 是否正在刷新
	Error     string `json:"error"`     // 错误信息（如果有）
}

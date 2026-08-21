package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
)

// FbCacheService FB缓存服务
type FbCacheService struct {
	mu sync.Mutex // 保护并发刷新
}

var DefaultFbCacheService = &FbCacheService{}

// ==================== FB 账号缓存 ====================

// GetCachedAccounts 从缓存表获取FB账号列表
func (s *FbCacheService) GetCachedAccounts(userID uint, tenantID *uint) (*models.FbAccountListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT c.id, c.fb_token_id, c.fb_user_id, c.fb_user_name, c.label, c.scopes,
		        c.expires_at, c.days_until_expiry, c.has_ad_perm, c.account_status,
		        c.bm_count, c.personal_ad_count, c.bm_ad_count, c.data_error, c.created_at
		 FROM fb_accounts_cache c
		 JOIN fb_tokens t ON t.id = c.fb_token_id AND t.status = 1
		 WHERE c.user_id = $1 AND c.tenant_id IS NOT DISTINCT FROM $2
		 ORDER BY c.created_at DESC`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询FB账号缓存失败: %w", err)
	}
	defer rows.Close()

	var accounts []models.FbAccountListItem

	for rows.Next() {
		var (
			cacheID         int
			fbTokenID       int
			fbUserID        string
			fbUserName      string
			label           string
			scopesStr       string
			expiresAt       *time.Time
			daysUntilExpiry int
			hasAdPerm       bool
			accountStatus   string
			bmCount         int
			personalAdCount int
			bmAdCount       int
			dataError       string
			createdAt       *time.Time
		)
		if err := rows.Scan(&cacheID, &fbTokenID, &fbUserID, &fbUserName, &label, &scopesStr,
			&expiresAt, &daysUntilExpiry, &hasAdPerm, &accountStatus,
			&bmCount, &personalAdCount, &bmAdCount, &dataError, &createdAt); err != nil {
			log.Printf("[FB-CACHE] 扫描账号缓存行失败: %v", err)
			continue
		}

		var scopes []string
		if scopesStr != "" {
			json.Unmarshal([]byte(scopesStr), &scopes)
		}
		if scopes == nil {
			scopes = []string{}
		}

		expiresAtStr := ""
		if expiresAt != nil {
			expiresAtStr = expiresAt.Format(time.RFC3339)
		}

		createdAtStr := ""
		if createdAt != nil {
			createdAtStr = createdAt.Format(time.RFC3339)
		}

		accounts = append(accounts, models.FbAccountListItem{
			ID:              uint(fbTokenID),  // 用 fb_token_id，而不是缓存表的 id
			FbUserID:        fbUserID,
			FbUserName:      fbUserName,
			Label:           label,
			Scopes:          scopes,
			ExpiresAt:       expiresAtStr,
			CreatedAt:       createdAtStr,
			DaysUntilExpiry: daysUntilExpiry,
			HasAdPerm:       hasAdPerm,
			AccountStatus:   accountStatus,
			BmCount:         bmCount,
			PersonalAdCount: personalAdCount,
			BmAdCount:       bmAdCount,
			DataError:       dataError,
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

// SaveAccountsCache 保存FB账号列表到缓存表
// 注意：FbAccountListItem.ID 即 fb_tokens.id，逐条按自身 ID 关联，避免多账号时错配
func (s *FbCacheService) SaveAccountsCache(userID uint, tenantID *uint, accounts []models.FbAccountListItem) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	for _, acc := range accounts {
		tokenID := int(acc.ID)
		if tokenID == 0 {
			continue
		}
		scopesJSON, _ := json.Marshal(acc.Scopes)

		var expiresAt *time.Time
		if acc.ExpiresAt != "" {
			if t, err := time.Parse(time.RFC3339, acc.ExpiresAt); err == nil {
				expiresAt = &t
			}
		}

		// UPSERT: 存在则更新，不存在则插入
		_, err := db.Pool.Exec(ctx,
			`INSERT INTO fb_accounts_cache 
				(user_id, tenant_id, fb_token_id, fb_user_id, fb_user_name, label, scopes,
				 expires_at, days_until_expiry, has_ad_perm, account_status,
				 bm_count, personal_ad_count, bm_ad_count, data_error, last_refresh_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
			ON CONFLICT (fb_token_id) DO UPDATE SET
				fb_user_name = EXCLUDED.fb_user_name,
				label = EXCLUDED.label,
				scopes = EXCLUDED.scopes,
				expires_at = EXCLUDED.expires_at,
				days_until_expiry = EXCLUDED.days_until_expiry,
				has_ad_perm = EXCLUDED.has_ad_perm,
				account_status = EXCLUDED.account_status,
				bm_count = EXCLUDED.bm_count,
				personal_ad_count = EXCLUDED.personal_ad_count,
				bm_ad_count = EXCLUDED.bm_ad_count,
				data_error = EXCLUDED.data_error,
				last_refresh_at = EXCLUDED.last_refresh_at,
				updated_at = EXCLUDED.updated_at`,
			userID, tenantID, tokenID, acc.FbUserID, acc.FbUserName, acc.Label, string(scopesJSON),
			expiresAt, acc.DaysUntilExpiry, acc.HasAdPerm, acc.AccountStatus,
			acc.BmCount, acc.PersonalAdCount, acc.BmAdCount, acc.DataError, now, now,
		)
		if err != nil {
			log.Printf("[FB-CACHE] 保存账号缓存失败 (token_id=%d): %v", tokenID, err)
			return err
		}
	}

	return nil
}

// ==================== 广告账户缓存 ====================

// GetCachedAdAccounts 从缓存表获取广告账户列表
func (s *FbCacheService) GetCachedAdAccounts(userID uint, tenantID *uint) (*models.FbAdAccountDetailListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT c.ad_account_id, c.account_id, c.name, c.fb_owner_name, c.fb_owner_id,
		        c.business_name, c.owner_business_id, c.account_status, c.status_label, c.platform,
		        c.amount_spent, c.currency, c.spend_cap, c.balance, c.daily_spend_limit,
		        c.admin_name, c.hidden_admins, c.other_admin_names,
		        c.timezone_name, c.timezone_offset, c.country_code, c.is_personal,
		        c.funding_source, c.disable_reason, c.disable_reason_label,
		        c.next_bill_date, c.created_time, c.is_prepay, c.owner_role, c.remark
		 FROM fb_ad_accounts_cache c
		 JOIN fb_tokens t ON t.id = c.fb_token_id AND t.status = 1
		 WHERE c.user_id = $1 AND c.tenant_id IS NOT DISTINCT FROM $2
		 ORDER BY c.name`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询广告账户缓存失败: %w", err)
	}
	defer rows.Close()

	var accounts []models.FbAdAccountDetail

	for rows.Next() {
		var (
			adAccountID        string
			accountID          string
			name               string
			fbOwnerName        string
			fbOwnerID          string
			businessName       string
			ownerBusinessID    string
			accountStatus      int
			statusLabel        string
			platform           string
			amountSpent        float64
			currency           string
			spendCap           float64
			balance            float64
			dailySpendLimit    float64
			adminName          string
			hiddenAdmins       int
			otherAdminNamesStr string
			timezoneName       string
			timezoneOffset     float64
			countryCode        string
			isPersonal         int
			fundingSource      string
			disableReason      int
			disableReasonLabel string
			nextBillDate       string
			createdTime        string
			isPrepay           int
			ownerRole          string
			remark             string
		)
		if err := rows.Scan(
			&adAccountID, &accountID, &name, &fbOwnerName, &fbOwnerID,
			&businessName, &ownerBusinessID, &accountStatus, &statusLabel, &platform,
			&amountSpent, &currency, &spendCap, &balance, &dailySpendLimit,
			&adminName, &hiddenAdmins, &otherAdminNamesStr,
			&timezoneName, &timezoneOffset, &countryCode, &isPersonal,
			&fundingSource, &disableReason, &disableReasonLabel,
			&nextBillDate, &createdTime, &isPrepay, &ownerRole, &remark,
		); err != nil {
			log.Printf("[FB-CACHE] 扫描广告账户缓存行失败: %v", err)
			continue
		}

		var otherAdminNames []string
		if otherAdminNamesStr != "" {
			json.Unmarshal([]byte(otherAdminNamesStr), &otherAdminNames)
		}
		if otherAdminNames == nil {
			otherAdminNames = []string{}
		}

		accounts = append(accounts, models.FbAdAccountDetail{
			ID:                 adAccountID,
			AccountID:          accountID,
			Name:               name,
			FbOwnerName:        fbOwnerName,
			FbOwnerID:          fbOwnerID,
			BusinessName:       businessName,
			OwnerBusinessID:    ownerBusinessID,
			AccountStatus:      accountStatus,
			StatusLabel:        statusLabel,
			Platform:           platform,
			AmountSpent:        amountSpent,
			Currency:           currency,
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
			Remark:             remark,
		})
	}

	if accounts == nil {
		accounts = []models.FbAdAccountDetail{}
	}

	return &models.FbAdAccountDetailListResponse{
		Accounts: accounts,
		Total:    len(accounts),
	}, nil
}

// SaveAdAccountsCache 保存广告账户列表到缓存表
// 注意：按每条记录的 TokenID 关联 fb_tokens，避免多账号时错配
func (s *FbCacheService) SaveAdAccountsCache(userID uint, tenantID *uint, accounts []models.FbAdAccountDetail) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	for _, acc := range accounts {
		tokenID := int(acc.TokenID)
		if tokenID == 0 {
			continue
		}
		otherAdminNamesJSON, _ := json.Marshal(acc.OtherAdminNames)

		// UPSERT
		_, err := db.Pool.Exec(ctx,
			`INSERT INTO fb_ad_accounts_cache 
				(user_id, tenant_id, fb_token_id, ad_account_id, account_id, name,
				 fb_owner_name, fb_owner_id, business_name, owner_business_id,
				 account_status, status_label, platform,
				 amount_spent, currency, spend_cap, balance, daily_spend_limit,
				 admin_name, hidden_admins, other_admin_names,
				 timezone_name, timezone_offset, country_code, is_personal,
				 funding_source, disable_reason, disable_reason_label,
				 next_bill_date, created_time, is_prepay, owner_role, last_refresh_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34)
			ON CONFLICT (fb_token_id, ad_account_id) DO UPDATE SET
				name = EXCLUDED.name,
				fb_owner_name = EXCLUDED.fb_owner_name,
				fb_owner_id = EXCLUDED.fb_owner_id,
				business_name = EXCLUDED.business_name,
				owner_business_id = EXCLUDED.owner_business_id,
				account_status = EXCLUDED.account_status,
				status_label = EXCLUDED.status_label,
				amount_spent = EXCLUDED.amount_spent,
				spend_cap = EXCLUDED.spend_cap,
				balance = EXCLUDED.balance,
				daily_spend_limit = EXCLUDED.daily_spend_limit,
				admin_name = EXCLUDED.admin_name,
				hidden_admins = EXCLUDED.hidden_admins,
				other_admin_names = EXCLUDED.other_admin_names,
				timezone_name = EXCLUDED.timezone_name,
				timezone_offset = EXCLUDED.timezone_offset,
				country_code = EXCLUDED.country_code,
				is_personal = EXCLUDED.is_personal,
				funding_source = EXCLUDED.funding_source,
				disable_reason = EXCLUDED.disable_reason,
				disable_reason_label = EXCLUDED.disable_reason_label,
				next_bill_date = EXCLUDED.next_bill_date,
				is_prepay = EXCLUDED.is_prepay,
				owner_role = EXCLUDED.owner_role,
				last_refresh_at = EXCLUDED.last_refresh_at,
				updated_at = EXCLUDED.updated_at`,
			userID, tenantID, tokenID, acc.ID, acc.AccountID, acc.Name,
			acc.FbOwnerName, acc.FbOwnerID, acc.BusinessName, acc.OwnerBusinessID,
			acc.AccountStatus, acc.StatusLabel, acc.Platform,
			acc.AmountSpent, acc.Currency, acc.SpendCap, acc.Balance, acc.DailySpendLimit,
			acc.AdminName, acc.HiddenAdmins, string(otherAdminNamesJSON),
			acc.TimezoneName, acc.TimezoneOffset, acc.CountryCode, acc.IsPersonal,
			acc.FundingSource, acc.DisableReason, acc.DisableReasonLabel,
			acc.NextBillDate, acc.CreatedTime, acc.IsPrepay, acc.OwnerRole, now, now,
		)
		if err != nil {
			log.Printf("[FB-CACHE] 保存广告账户缓存失败 (token_id=%d, ad_account=%s): %v", tokenID, acc.ID, err)
			return err
		}
	}

	return nil
}

// UpdateAdAccountRemark 更新广告账户的本地备注（仅存缓存表，FB 刷新不覆盖）
func (s *FbCacheService) UpdateAdAccountRemark(userID uint, tenantID *uint, adAccountID, remark string) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	tag, err := db.Pool.Exec(ctx,
		`UPDATE fb_ad_accounts_cache
		 SET remark = $3, updated_at = NOW()
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND ad_account_id = $4`,
		userID, tenantID, remark, adAccountID,
	)
	if err != nil {
		return fmt.Errorf("更新备注失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("账户不存在，请先刷新列表")
	}

	return nil
}

// ==================== 刷新状态管理 ====================

// StartRefresh 开始刷新任务
func (s *FbCacheService) StartRefresh(userID uint, tenantID *uint, refreshType string) (int, error) {
	if db.Pool == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	var id int
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO fb_refresh_status (user_id, tenant_id, refresh_type, status, started_at)
		 VALUES ($1, $2, $3, 'running', NOW())
		 RETURNING id`,
		userID, tenantID, refreshType,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("创建刷新任务失败: %w", err)
	}

	return id, nil
}

// CompleteRefresh 完成刷新任务
func (s *FbCacheService) CompleteRefresh(refreshID int, errMsg string) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	status := "completed"
	if errMsg != "" {
		status = "failed"
	}

	_, err := db.Pool.Exec(ctx,
		`UPDATE fb_refresh_status 
		 SET status = $1, completed_at = NOW(), error_message = $2
		 WHERE id = $3`,
		status, errMsg, refreshID,
	)
	return err
}

// cleanStaleRefreshes 清理卡死的刷新任务（服务重启等导致 running 状态残留）
func (s *FbCacheService) cleanStaleRefreshes(userID uint, tenantID *uint, refreshType string) {
	if db.Pool == nil {
		return
	}
	ctx := context.Background()
	_, _ = db.Pool.Exec(ctx,
		`UPDATE fb_refresh_status
		 SET status = 'failed', completed_at = NOW(), error_message = '刷新任务中断（超时未完成）'
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND refresh_type = $3
		   AND status = 'running' AND started_at < NOW() - INTERVAL '10 minutes'`,
		userID, tenantID, refreshType,
	)
}

// GetRefreshStatus 获取最近一次刷新状态
func (s *FbCacheService) GetRefreshStatus(userID uint, tenantID *uint, refreshType string) (*models.FbRefreshStatusResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	s.cleanStaleRefreshes(userID, tenantID, refreshType)

	ctx := context.Background()
	var status string
	var completedAt *time.Time
	var errMsg string

	err := db.Pool.QueryRow(ctx,
		`SELECT status, completed_at, COALESCE(error_message, '')
		 FROM fb_refresh_status
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND refresh_type = $3
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID, tenantID, refreshType,
	).Scan(&status, &completedAt, &errMsg)
	if err != nil {
		// 没有记录，返回空状态
		return &models.FbRefreshStatusResponse{
			Status:    "none",
			IsRunning: false,
		}, nil
	}

	isRunning := status == "running"

	return &models.FbRefreshStatusResponse{
		Status:    status,
		IsRunning: isRunning,
		Error:     errMsg,
	}, nil
}

// ShouldRefresh 判断是否应该启动后台刷新：无进行中任务 且 距上次刷新超过冷却期
// 防止"列表请求→触发刷新→前端轮询完成后重新加载→又触发刷新"的死循环
func (s *FbCacheService) ShouldRefresh(userID uint, tenantID *uint, refreshType string, cooldown time.Duration) bool {
	if db.Pool == nil {
		return false
	}
	if s.IsRefreshing(userID, tenantID, refreshType) {
		return false
	}
	ctx := context.Background()
	var lastStarted time.Time
	err := db.Pool.QueryRow(ctx,
		`SELECT started_at FROM fb_refresh_status
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND refresh_type = $3
		 ORDER BY created_at DESC LIMIT 1`,
		userID, tenantID, refreshType,
	).Scan(&lastStarted)
	if err != nil {
		return true // 从未刷新过
	}
	return time.Since(lastStarted) > cooldown
}

// IsRefreshing 检查是否正在刷新
func (s *FbCacheService) IsRefreshing(userID uint, tenantID *uint, refreshType string) bool {
	if db.Pool == nil {
		return false
	}

	s.cleanStaleRefreshes(userID, tenantID, refreshType)

	ctx := context.Background()
	var count int
	err := db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM fb_refresh_status
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND refresh_type = $3 AND status = 'running'`,
		userID, tenantID, refreshType,
	).Scan(&count)
	if err != nil {
		return false
	}

	return count > 0
}

// ==================== BM（Business Manager）缓存 ====================

// GetCachedBms 从缓存表获取 BM 列表
func (s *FbCacheService) GetCachedBms(userID uint, tenantID *uint) (*models.FbBmListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT c.bm_id, c.name, c.fb_owner_name, c.fb_owner_id, c.status_label,
		        c.push_status, c.remark, c.owner_role, c.verification_status,
		        c.admin_count, c.pending_admin_count, c.admin_names, c.partner_count, c.ad_account_count,
		        c.created_time, c.last_refresh_at
		 FROM fb_bm_cache c
		 JOIN fb_tokens t ON t.id = c.fb_token_id AND t.status = 1
		 WHERE c.user_id = $1 AND c.tenant_id IS NOT DISTINCT FROM $2
		 ORDER BY c.created_time DESC`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询BM缓存失败: %w", err)
	}
	defer rows.Close()

	var list []models.FbBmListItem

	for rows.Next() {
		var (
			item             models.FbBmListItem
			adminNamesStr    string
			pushStatus       *string
			remark           *string
			ownerRole        *string
			verification     *string
			createdTime      *string
			statusLabel      *string
		)
		if err := rows.Scan(
			&item.BmID, &item.Name, &item.FbOwnerName, &item.FbOwnerID, &statusLabel,
			&pushStatus, &remark, &ownerRole, &verification,
			&item.AdminCount, &item.PendingAdminCount, &adminNamesStr, &item.PartnerCount, &item.AdAccountCount,
			&createdTime, &item.LastRefreshAt,
		); err != nil {
			log.Printf("[FB-CACHE] 扫描BM缓存行失败: %v", err)
			continue
		}

		if statusLabel != nil {
			item.StatusLabel = *statusLabel
		}
		if pushStatus != nil {
			item.PushStatus = *pushStatus
		}
		if remark != nil {
			item.Remark = *remark
		}
		if ownerRole != nil {
			item.OwnerRole = *ownerRole
		}
		if verification != nil {
			item.VerificationStatus = *verification
		}
		if createdTime != nil {
			item.CreatedTime = *createdTime
		}

		if adminNamesStr != "" {
			json.Unmarshal([]byte(adminNamesStr), &item.AdminNames)
		}
		if item.AdminNames == nil {
			item.AdminNames = []string{}
		}

		list = append(list, item)
	}

	if list == nil {
		list = []models.FbBmListItem{}
	}

	return &models.FbBmListResponse{
		List:  list,
		Total: len(list),
	}, nil
}

// SaveBmsCache 保存 BM 列表到缓存表（upsert 不覆盖 remark / push_status 本地字段）
func (s *FbCacheService) SaveBmsCache(userID uint, tenantID *uint, bms []models.FbBmListItem) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	for _, bm := range bms {
		tokenID := int(bm.TokenID)
		if tokenID == 0 || bm.BmID == "" {
			continue
		}
		adminNamesJSON, _ := json.Marshal(bm.AdminNames)

		_, err := db.Pool.Exec(ctx,
			`INSERT INTO fb_bm_cache
				(user_id, tenant_id, fb_token_id, bm_id, name,
				 fb_owner_name, fb_owner_id, status_label,
				 owner_role, verification_status,
				 admin_count, pending_admin_count, admin_names, partner_count, ad_account_count,
				 created_time, last_refresh_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
			ON CONFLICT (fb_token_id, bm_id) DO UPDATE SET
				name = EXCLUDED.name,
				fb_owner_name = EXCLUDED.fb_owner_name,
				fb_owner_id = EXCLUDED.fb_owner_id,
				status_label = EXCLUDED.status_label,
				owner_role = EXCLUDED.owner_role,
				verification_status = EXCLUDED.verification_status,
				admin_count = EXCLUDED.admin_count,
				pending_admin_count = EXCLUDED.pending_admin_count,
				admin_names = EXCLUDED.admin_names,
				partner_count = EXCLUDED.partner_count,
				ad_account_count = EXCLUDED.ad_account_count,
				created_time = EXCLUDED.created_time,
				last_refresh_at = EXCLUDED.last_refresh_at,
				updated_at = EXCLUDED.updated_at`,
			userID, tenantID, tokenID, bm.BmID, bm.Name,
			bm.FbOwnerName, bm.FbOwnerID, bm.StatusLabel,
			bm.OwnerRole, bm.VerificationStatus,
			bm.AdminCount, bm.PendingAdminCount, string(adminNamesJSON), bm.PartnerCount, bm.AdAccountCount,
			bm.CreatedTime, now, now,
		)
		if err != nil {
			log.Printf("[FB-CACHE] 保存BM缓存失败 (token_id=%d, bm=%s): %v", tokenID, bm.BmID, err)
			return err
		}
	}

	return nil
}

// UpdateBmRemark 更新 BM 的本地备注（仅存缓存表，FB 刷新不覆盖）
func (s *FbCacheService) UpdateBmRemark(userID uint, tenantID *uint, bmID, remark string) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	tag, err := db.Pool.Exec(ctx,
		`UPDATE fb_bm_cache
		 SET remark = $3, updated_at = NOW()
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND bm_id = $4`,
		userID, tenantID, remark, bmID,
	)
	if err != nil {
		return fmt.Errorf("更新备注失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("BM 不存在，请先刷新列表")
	}

	return nil
}

// ==================== FB 公共主页缓存 ====================

// GetCachedPages 从缓存表获取公共主页列表
func (s *FbCacheService) GetCachedPages(userID uint, tenantID *uint) (*models.FbPageListResponse, error) {
	if db.Pool == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	rows, err := db.Pool.Query(ctx,
		`SELECT c.page_id, c.name, c.link, c.fb_owner_name, c.fb_owner_id,
		        c.category, c.fan_count, c.followers_count, c.is_published, c.verification_status,
		        c.website, c.phone, c.email, c.address, c.admin_names,
		        c.bm_name, c.ad_perm, c.profanity_filter, c.blocked_count,
		        c.push_status, c.remark, c.last_refresh_at
		 FROM fb_pages_cache c
		 JOIN fb_tokens t ON t.id = c.fb_token_id AND t.status = 1
		 WHERE c.user_id = $1 AND c.tenant_id IS NOT DISTINCT FROM $2
		 ORDER BY c.name`,
		userID, tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("查询主页缓存失败: %w", err)
	}
	defer rows.Close()

	var list []models.FbPageItem

	for rows.Next() {
		var (
			item          models.FbPageItem
			adminNamesStr string
		)
		if err := rows.Scan(
			&item.PageID, &item.Name, &item.Link, &item.FbOwnerName, &item.FbOwnerID,
			&item.Category, &item.FanCount, &item.FollowersCount, &item.IsPublished, &item.VerificationStatus,
			&item.Website, &item.Phone, &item.Email, &item.Address, &adminNamesStr,
			&item.BusinessName, &item.AdPerm, &item.ProfanityFilter, &item.BlockedCount,
			&item.PushStatus, &item.Remark, &item.LastRefreshAt,
		); err != nil {
			log.Printf("[FB-CACHE] 扫描主页缓存行失败: %v", err)
			continue
		}

		if adminNamesStr != "" {
			json.Unmarshal([]byte(adminNamesStr), &item.AdminNames)
		}
		if item.AdminNames == nil {
			item.AdminNames = []string{}
		}

		list = append(list, item)
	}

	if list == nil {
		list = []models.FbPageItem{}
	}

	return &models.FbPageListResponse{
		List:  list,
		Total: len(list),
	}, nil
}

// SavePagesCache 保存公共主页列表到缓存表（upsert 不覆盖 remark / push_status 本地字段）
func (s *FbCacheService) SavePagesCache(userID uint, tenantID *uint, pages []models.FbPageItem) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	for _, page := range pages {
		tokenID := int(page.TokenID)
		if tokenID == 0 || page.PageID == "" {
			continue
		}
		adminNamesJSON, _ := json.Marshal(page.AdminNames)

		_, err := db.Pool.Exec(ctx,
			`INSERT INTO fb_pages_cache
				(user_id, tenant_id, fb_token_id, page_id, name, link,
				 fb_owner_name, fb_owner_id, category, fan_count, followers_count,
				 is_published, verification_status, website, phone, email, address,
				 admin_names, bm_name, ad_perm, profanity_filter, blocked_count,
				 last_refresh_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
			ON CONFLICT (fb_token_id, page_id) DO UPDATE SET
				name = EXCLUDED.name,
				link = EXCLUDED.link,
				fb_owner_name = EXCLUDED.fb_owner_name,
				fb_owner_id = EXCLUDED.fb_owner_id,
				category = EXCLUDED.category,
				fan_count = EXCLUDED.fan_count,
				followers_count = EXCLUDED.followers_count,
				is_published = EXCLUDED.is_published,
				verification_status = EXCLUDED.verification_status,
				website = EXCLUDED.website,
				phone = EXCLUDED.phone,
				email = EXCLUDED.email,
				address = EXCLUDED.address,
				admin_names = EXCLUDED.admin_names,
				bm_name = EXCLUDED.bm_name,
				ad_perm = EXCLUDED.ad_perm,
				profanity_filter = EXCLUDED.profanity_filter,
				blocked_count = EXCLUDED.blocked_count,
				last_refresh_at = EXCLUDED.last_refresh_at,
				updated_at = EXCLUDED.updated_at`,
			userID, tenantID, tokenID, page.PageID, page.Name, page.Link,
			page.FbOwnerName, page.FbOwnerID, page.Category, page.FanCount, page.FollowersCount,
			page.IsPublished, page.VerificationStatus, page.Website, page.Phone, page.Email, page.Address,
			string(adminNamesJSON), page.BusinessName, page.AdPerm, page.ProfanityFilter, page.BlockedCount,
			now, now,
		)
		if err != nil {
			log.Printf("[FB-CACHE] 保存主页缓存失败 (token_id=%d, page=%s): %v", tokenID, page.PageID, err)
			return err
		}
	}

	return nil
}

// UpdatePageRemark 更新主页的本地备注（仅存缓存表，FB 刷新不覆盖）
func (s *FbCacheService) UpdatePageRemark(userID uint, tenantID *uint, pageID, remark string) error {
	if db.Pool == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx := context.Background()
	tag, err := db.Pool.Exec(ctx,
		`UPDATE fb_pages_cache
		 SET remark = $3, updated_at = NOW()
		 WHERE user_id = $1 AND tenant_id IS NOT DISTINCT FROM $2 AND page_id = $4`,
		userID, tenantID, remark, pageID,
	)
	if err != nil {
		return fmt.Errorf("更新备注失败: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("主页不存在，请先刷新列表")
	}

	return nil
}

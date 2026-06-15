package services

import (
	"context"
	"time"

	"github.com/githubi2/FbAi/art-design-server/db"
)

// SessionService 会话服务
type SessionService struct{}

var DefaultSessionService = &SessionService{}

// Create 创建会话（同时删除该用户所有旧会话，实现 SSO 单点登录）
// tenantID: 租户上下文。管理员登录时为 nil，租户用户登录时设为用户所属租户ID
func (s *SessionService) Create(userID uint, token, refreshToken string, expiresAt time.Time, tenantID *uint) error {
	if db.Pool == nil {
		return nil // 无数据库时不报错
	}

	ctx := context.Background()

	// 单点登录：仅管理员用户删除旧会话（租户用户不需要 SSO）
	// 租户用户 tenantID != nil 时允许多处登录
	if tenantID == nil {
		_, _ = db.Pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	}

	// 创建新会话
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO sessions (user_id, token, refresh_token, tenant_id, expires_at, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, token, refreshToken, tenantID, expiresAt, time.Now(),
	)
	return err
}

// Validate 验证 token，返回 userID。过期或不存在返回 0, false
func (s *SessionService) Validate(token string) (uint, bool) {
	if db.Pool == nil {
		return 0, false
	}

	ctx := context.Background()
	var userID uint
	var expiresAt time.Time

	err := db.Pool.QueryRow(ctx,
		`SELECT user_id, expires_at FROM sessions WHERE token = $1`, token,
	).Scan(&userID, &expiresAt)
	if err != nil {
		return 0, false
	}

	// 检查是否过期
	if time.Now().After(expiresAt) {
		// 删除过期会话
		_, _ = db.Pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
		return 0, false
	}

	return userID, true
}

// InvalidateUser 失效指定用户的所有会话（管理员修改密码时调用）
func (s *SessionService) InvalidateUser(userID uint) error {
	if db.Pool == nil {
		return nil
	}

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// GetTenantID 从 token 获取当前租户上下文
func (s *SessionService) GetTenantID(token string) (*uint, error) {
	if db.Pool == nil {
		return nil, nil
	}

	ctx := context.Background()
	var tenantID *uint
	err := db.Pool.QueryRow(ctx,
		`SELECT tenant_id FROM sessions WHERE token = $1 AND expires_at > $2`,
		token, time.Now(),
	).Scan(&tenantID)
	if err != nil {
		return nil, err
	}
	return tenantID, nil
}

// SetTenantID 更新会话的租户上下文（租户切换）
func (s *SessionService) SetTenantID(token string, tenantID *uint) error {
	if db.Pool == nil {
		return nil
	}

	ctx := context.Background()
	_, err := db.Pool.Exec(ctx,
		`UPDATE sessions SET tenant_id = $1 WHERE token = $2`,
		tenantID, token,
	)
	return err
}

// CleanExpired 清理所有过期会话（可定期调用）
func (s *SessionService) CleanExpired() (int64, error) {
	if db.Pool == nil {
		return 0, nil
	}

	ctx := context.Background()
	result, err := db.Pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < $1`, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

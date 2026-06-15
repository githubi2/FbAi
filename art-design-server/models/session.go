package models

import "time"

// Session 用户会话模型（数据库 sessions 表）
type Session struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"userId"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	TenantID     *uint     `json:"tenantId"`   // 当前租户上下文 (NULL=全局视角)
	ExpiresAt    time.Time `json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

-- 009_fb_bm_cache: BM（Business Manager）列表缓存表
-- 与 fb_ad_accounts_cache 同策略：GET 纯读缓存，POST 显式后台刷新，remark/push_status 本地字段刷新不覆盖

CREATE TABLE IF NOT EXISTS fb_bm_cache (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    tenant_id INT REFERENCES tenants(id),
    fb_token_id INT NOT NULL REFERENCES fb_tokens(id) ON DELETE CASCADE,
    bm_id VARCHAR(64) NOT NULL,               -- Business ID
    name VARCHAR(256) DEFAULT '',             -- BM 名称
    fb_owner_name VARCHAR(128) DEFAULT '',    -- 授权 FB 账号名
    fb_owner_id VARCHAR(64) DEFAULT '',
    status_label VARCHAR(20) DEFAULT '正常',   -- 状态：API 可达=正常
    push_status VARCHAR(20) DEFAULT '',       -- 本地推送状态（刷新不覆盖）
    remark VARCHAR(255) DEFAULT '',           -- 本地备注（刷新不覆盖）
    owner_role VARCHAR(20) DEFAULT '',        -- 授权用户在 BM 中的角色（business_users.role: ADMIN/EMPLOYEE）
    verification_status VARCHAR(32) DEFAULT '', -- 认证状态：verified/not_verified/...
    admin_count INT DEFAULT 0,                -- business_users 中 role=ADMIN 的数量
    admin_names TEXT DEFAULT '[]',            -- 管理员名字 JSON array
    partner_count INT DEFAULT 0,              -- 合作伙伴数 = owned_businesses + agencies
    ad_account_count INT DEFAULT 0,           -- owned_ad_accounts + client_ad_accounts
    created_time VARCHAR(32) DEFAULT '',      -- BM 创建时间（FB 原始字符串）
    last_refresh_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 唯一索引：每个 fb_token + bm_id 一条缓存
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_bm_cache_bm
    ON fb_bm_cache(fb_token_id, bm_id);

CREATE INDEX IF NOT EXISTS idx_fb_bm_cache_user
    ON fb_bm_cache(user_id, tenant_id);

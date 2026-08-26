-- 007_fb_cache_tables: FB账号和广告账户缓存表
-- 实现缓存优先加载策略：先返回DB数据，异步刷新FB API

-- FB账号缓存表
CREATE TABLE IF NOT EXISTS fb_accounts_cache (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    tenant_id INT REFERENCES tenants(id),
    fb_token_id INT NOT NULL REFERENCES fb_tokens(id) ON DELETE CASCADE,
    fb_user_id VARCHAR(64) NOT NULL,
    fb_user_name VARCHAR(128) DEFAULT '',
    label VARCHAR(64) DEFAULT '',
    scopes TEXT DEFAULT '[]',           -- JSON array
    expires_at TIMESTAMPTZ,
    days_until_expiry INT DEFAULT 0,
    has_ad_perm BOOLEAN DEFAULT false,
    account_status VARCHAR(20) DEFAULT '正常',  -- 正常/已过期/异常
    bm_count INT DEFAULT 0,
    personal_ad_count INT DEFAULT 0,
    bm_ad_count INT DEFAULT 0,
    data_error TEXT DEFAULT '',
    last_refresh_at TIMESTAMPTZ,        -- 最后一次成功刷新时间
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 唯一索引：每个用户+fb_token 一条缓存
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_accounts_cache_token
    ON fb_accounts_cache(fb_token_id);

CREATE INDEX IF NOT EXISTS idx_fb_accounts_cache_user
    ON fb_accounts_cache(user_id, tenant_id);

-- 广告账户缓存表
CREATE TABLE IF NOT EXISTS fb_ad_accounts_cache (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    tenant_id INT REFERENCES tenants(id),
    fb_token_id INT NOT NULL REFERENCES fb_tokens(id) ON DELETE CASCADE,
    ad_account_id VARCHAR(64) NOT NULL,  -- act_xxx 格式
    account_id VARCHAR(64),              -- 数字ID
    name VARCHAR(256) DEFAULT '',
    fb_owner_name VARCHAR(128) DEFAULT '',
    fb_owner_id VARCHAR(64) DEFAULT '',
    business_name VARCHAR(256) DEFAULT '',
    owner_business_id VARCHAR(64) DEFAULT '',
    account_status INT DEFAULT 0,
    status_label VARCHAR(50) DEFAULT '',
    platform VARCHAR(20) DEFAULT 'facebook',
    amount_spent DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'USD',
    spend_cap DECIMAL(15,2) DEFAULT 0,
    balance DECIMAL(15,2) DEFAULT 0,
    daily_spend_limit DECIMAL(15,2) DEFAULT 0,
    admin_name VARCHAR(128) DEFAULT '',
    hidden_admins INT DEFAULT 0,
    other_admin_names TEXT DEFAULT '[]',  -- JSON array
    timezone_name VARCHAR(64) DEFAULT '',
    timezone_offset DECIMAL(5,2) DEFAULT 0,
    country_code VARCHAR(10) DEFAULT '',
    is_personal INT DEFAULT 0,
    funding_source VARCHAR(256) DEFAULT '',
    disable_reason INT DEFAULT 0,
    disable_reason_label VARCHAR(128) DEFAULT '',
    next_bill_date VARCHAR(32) DEFAULT '',
    created_time VARCHAR(32) DEFAULT '',
    last_refresh_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 唯一索引：每个 fb_token + ad_account_id 一条缓存
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_ad_accounts_cache_acct
    ON fb_ad_accounts_cache(fb_token_id, ad_account_id);

CREATE INDEX IF NOT EXISTS idx_fb_ad_accounts_cache_user
    ON fb_ad_accounts_cache(user_id, tenant_id);

-- 刷新状态跟踪表（用于前端轮询）
CREATE TABLE IF NOT EXISTS fb_refresh_status (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    tenant_id INT REFERENCES tenants(id),
    refresh_type VARCHAR(20) NOT NULL,   -- 'accounts' / 'ad_accounts' / 'all'
    status VARCHAR(20) DEFAULT 'pending', -- pending/running/completed/failed
    started_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fb_refresh_status_user
    ON fb_refresh_status(user_id, tenant_id, created_at DESC);

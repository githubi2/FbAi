-- ============================================================
-- Facebook OAuth Token 存储表
-- 用于存储用户的 Facebook 授权信息
-- ============================================================

-- 修复 sequence（如有需要）
SELECT setval('fb_tokens_id_seq', (SELECT COALESCE(MAX(id), 0) FROM fb_tokens));

-- 创建 fb_tokens 表
CREATE TABLE IF NOT EXISTS fb_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    fb_user_id VARCHAR(64) DEFAULT '',           -- Facebook 用户 ID
    fb_user_name VARCHAR(128) DEFAULT '',        -- Facebook 用户名
    access_token TEXT NOT NULL,                  -- Facebook access token
    token_type VARCHAR(32) DEFAULT 'bearer',     -- token 类型
    expires_at TIMESTAMP,                        -- token 过期时间
    scopes TEXT[] DEFAULT '{}',                  -- 授权的权限范围
    bm_list JSONB DEFAULT '[]',                  -- 可访问的 Business Manager 列表（缓存）
    ad_accounts JSONB DEFAULT '[]',              -- 广告账户列表（缓存）
    selected_ad_account_id VARCHAR(64) DEFAULT '', -- 当前选中的广告账户
    status INT DEFAULT 1,                        -- 1=有效 0=失效
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 每个用户只能有一条有效的 Facebook 授权
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_tokens_user_id ON fb_tokens(user_id);

-- 013_fb_pixels_cache: FB 像素缓存表
-- 与 fb_pages_cache 同策略：GET 纯读缓存，POST 显式后台刷新，remark 本地字段刷新不覆盖

CREATE TABLE IF NOT EXISTS fb_pixels_cache (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    tenant_id INT REFERENCES tenants(id),
    fb_token_id INT NOT NULL REFERENCES fb_tokens(id) ON DELETE CASCADE,
    pixel_id VARCHAR(64) NOT NULL,          -- 像素数字ID
    name VARCHAR(256) DEFAULT '',           -- 像素名称
    ad_account_id VARCHAR(64) DEFAULT '',   -- 所属广告账号 act_xxx
    ad_account_name VARCHAR(256) DEFAULT '',-- 所属广告账号名称
    owner_bm_id VARCHAR(64) DEFAULT '',     -- 所属 BM ID
    owner_bm_name VARCHAR(256) DEFAULT '',  -- 所属 BM 名称
    creator_name VARCHAR(128) DEFAULT '',   -- 创建者名称
    is_unavailable INT DEFAULT 0,           -- 1=不可用 0=正常
    creation_time TIMESTAMPTZ,              -- 像素创建时间
    last_fired_time TIMESTAMPTZ,            -- 最近一次上报事件时间
    role_names TEXT DEFAULT '[]',           -- 当前用户在像素上的角色/权限 JSON array
    admin_names TEXT DEFAULT '[]',          -- 管理员名单 JSON array
    shared_agencies TEXT DEFAULT '[]',      -- 共享合作伙伴（agency）名单 JSON array
    remark VARCHAR(255) DEFAULT '',         -- 备注（本地字段）
    last_refresh_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 唯一索引：每个 fb_token + pixel_id 一条缓存
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_pixels_cache_pixel
    ON fb_pixels_cache(fb_token_id, pixel_id);

CREATE INDEX IF NOT EXISTS idx_fb_pixels_cache_user
    ON fb_pixels_cache(user_id, tenant_id);

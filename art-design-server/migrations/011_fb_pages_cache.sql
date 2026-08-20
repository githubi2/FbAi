-- 011_fb_pages_cache: FB 公共主页缓存表
-- 与 fb_bm_cache 同策略：GET 纯读缓存，POST 显式后台刷新，remark/push_status 本地字段刷新不覆盖

CREATE TABLE IF NOT EXISTS fb_pages_cache (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    tenant_id INT REFERENCES tenants(id),
    fb_token_id INT NOT NULL REFERENCES fb_tokens(id) ON DELETE CASCADE,
    page_id VARCHAR(64) NOT NULL,          -- 主页数字ID
    name VARCHAR(256) DEFAULT '',          -- 主页名称
    link VARCHAR(512) DEFAULT '',          -- 主页链接
    fb_owner_name VARCHAR(128) DEFAULT '', -- 所属 FB 账号名
    fb_owner_id VARCHAR(64) DEFAULT '',    -- 所属 FB 账号ID
    category VARCHAR(256) DEFAULT '',      -- 主页分类
    fan_count INT DEFAULT 0,               -- 点赞数
    followers_count INT DEFAULT 0,         -- 粉丝数
    is_published INT DEFAULT 1,            -- 发布状态 1=已发布 0=未发布
    verification_status VARCHAR(32) DEFAULT '', -- 主页认证
    website VARCHAR(512) DEFAULT '',
    phone VARCHAR(64) DEFAULT '',
    email VARCHAR(256) DEFAULT '',
    address VARCHAR(512) DEFAULT '',
    admin_names TEXT DEFAULT '[]',         -- 管理员名单 JSON array
    push_status VARCHAR(20) DEFAULT '',    -- 推送状态（本地字段）
    remark VARCHAR(255) DEFAULT '',        -- 备注（本地字段）
    last_refresh_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 唯一索引：每个 fb_token + page_id 一条缓存
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_pages_cache_page
    ON fb_pages_cache(fb_token_id, page_id);

CREATE INDEX IF NOT EXISTS idx_fb_pages_cache_user
    ON fb_pages_cache(user_id, tenant_id);

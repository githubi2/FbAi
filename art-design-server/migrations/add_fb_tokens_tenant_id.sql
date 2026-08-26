-- ============================================================
-- 为 fb_tokens 表添加 tenant_id 实现租户隔离
-- ============================================================

-- 1. 添加 tenant_id 列（NULL = 超级管理员）
ALTER TABLE fb_tokens ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id);

-- 2. 重建唯一索引（user_id 本身全局唯一，tenant_id 仅用于查询过滤）
DROP INDEX IF EXISTS idx_fb_tokens_user_id;
DROP INDEX IF EXISTS idx_fb_tokens_user_tenant;
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_tokens_user_id ON fb_tokens(user_id);

-- 3. 创建 tenant_id 查询索引
CREATE INDEX IF NOT EXISTS idx_fb_tokens_tenant_id ON fb_tokens(tenant_id);

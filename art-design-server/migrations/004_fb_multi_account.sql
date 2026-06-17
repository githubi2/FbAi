-- 004_fb_multi_account: 支持多 FB 账号授权
-- 旧约束：UNIQUE(user_id) — 一个用户只能授权一个 FB 账号
-- 新约束：UNIQUE(user_id, fb_user_id) WHERE status=1 — 同一用户可授权多个不同 FB 账号

-- 1. 删除旧的唯一约束
DROP INDEX IF EXISTS idx_fb_tokens_user_id;

-- 2. 新增部分唯一约束：同一用户+同一FB账号 的有效记录唯一
--    如果同一用户再次授权同一 FB 账号 → ON CONFLICT DO UPDATE（刷新 token）
--    如果同一用户授权不同 FB 账号 → INSERT 新记录
CREATE UNIQUE INDEX IF NOT EXISTS idx_fb_tokens_user_fb_active
  ON fb_tokens(user_id, fb_user_id) WHERE status = 1;

-- 3. 添加备注字段
ALTER TABLE fb_tokens ADD COLUMN IF NOT EXISTS label VARCHAR(64) DEFAULT '';

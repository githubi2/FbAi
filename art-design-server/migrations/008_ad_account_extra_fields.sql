-- 008_ad_account_extra_fields.sql
-- 广告账户缓存表新增：账户类型（预付费/后付费）、所有者角色、本地备注
-- 手动执行：psql 或任意 PG 客户端

ALTER TABLE fb_ad_accounts_cache
  ADD COLUMN IF NOT EXISTS is_prepay INT NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS owner_role VARCHAR(32) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS remark VARCHAR(255) NOT NULL DEFAULT '';

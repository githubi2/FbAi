-- 012_fb_pages_extra_fields: FB 公共主页缓存表扩展字段
-- 管理员改用主页访问口令后已可获取；新增 BM 归属 / 广告权限 / 不文明用语过滤 / 黑名单数

ALTER TABLE fb_pages_cache ADD COLUMN IF NOT EXISTS bm_name VARCHAR(256) DEFAULT '';        -- 所属 BM 名称（business 字段）
ALTER TABLE fb_pages_cache ADD COLUMN IF NOT EXISTS ad_perm INT DEFAULT -1;                 -- 广告权限 1=正常 0=无权限 -1=未知（tasks 含 ADVERTISE）
ALTER TABLE fb_pages_cache ADD COLUMN IF NOT EXISTS profanity_filter VARCHAR(32) DEFAULT ''; -- 隐藏不文明用语 none/medium/strong
ALTER TABLE fb_pages_cache ADD COLUMN IF NOT EXISTS blocked_count INT DEFAULT 0;            -- 黑名单数量（/blocked 边）

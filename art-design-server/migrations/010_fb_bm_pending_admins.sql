-- 010_fb_bm_pending_admins: BM 缓存表补充「邀请中管理员数」
-- 管理员总数 = 在职管理员(business_users, role=ADMIN) + 邀请中管理员(pending_users, role=ADMIN)

ALTER TABLE fb_bm_cache
  ADD COLUMN IF NOT EXISTS pending_admin_count INT DEFAULT 0;

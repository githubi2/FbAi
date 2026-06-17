-- Add BM管理 menu under 广告管理 (AdAccount, id=12)
-- 2026-06-17

-- Fix sequence
SELECT setval('menus_id_seq', (SELECT COALESCE(MAX(id), 0) FROM menus));

-- Insert child menu: BM管理
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order, hidden)
VALUES ('AdAccountBm', 12, 'bm', '/ad-account/bm/index', '', 'BM管理', 3, false);

-- Verify
SELECT id, name, parent_id, path, component, sort_order FROM menus WHERE name ILIKE '%AdAccount%' ORDER BY sort_order, id;

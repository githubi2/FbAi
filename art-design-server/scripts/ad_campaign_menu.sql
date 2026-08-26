-- 插入广告投放菜单（一级，位置：广告账户管理之后）
BEGIN;
INSERT INTO menus (name, title, parent_id, path, component, icon, sort_order, menu_type, hidden, status)
VALUES ('AdCampaign', '广告投放', 0, '/ad-campaign', '/ad-campaign/index', 'ri:megaphone-line', 7, 'menu', false, 1);

-- 后续菜单顺延：BM列表 7→8，主页列表 8→9，像素列表 9→10
UPDATE menus SET sort_order = 8 WHERE id = 16;  -- AdAccountBmList
UPDATE menus SET sort_order = 9 WHERE id = 18;  -- PageManageList
UPDATE menus SET sort_order = 10 WHERE id = 20; -- PixelManageList

-- R_SUPER 加入新菜单 ID
UPDATE roles SET menu_ids = array_append(menu_ids, (SELECT id FROM menus WHERE name = 'AdCampaign'))
WHERE id = 1 AND NOT (SELECT id FROM menus WHERE name = 'AdCampaign') = ANY(menu_ids);
COMMIT;

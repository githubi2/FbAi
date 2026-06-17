-- Change BM管理 from menu to directory
UPDATE menus SET menu_type = 'directory', component = '/index/index' WHERE id = 15;

-- Insert new BM列表 child menu
INSERT INTO menus (name, title, parent_id, path, component, icon, sort_order, hidden, menu_type, status, created_at, updated_at)
VALUES ('AdAccountBmList', 'BM列表', 15, 'list', '/ad-account/bm/index', '', 1, false, 'menu', 1, NOW(), NOW());

-- Verify
SELECT id, name, title, parent_id, path, component, menu_type FROM menus WHERE id >= 12 ORDER BY id;

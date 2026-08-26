-- Add PixelManage top-level directory + PixelManageList child menu
SELECT setval('menus_id_seq', (SELECT COALESCE(MAX(id), 0) FROM menus));

INSERT INTO menus (name, title, parent_id, path, component, icon, sort_order, hidden, menu_type, status, created_at, updated_at)
VALUES ('PixelManage', '像素管理', 0, '/pixel-manage', '/index/index', 'ri:crosshair-2-line', 8, false, 'directory', 1, NOW(), NOW());

INSERT INTO menus (name, title, parent_id, path, component, icon, sort_order, hidden, menu_type, status, created_at, updated_at)
VALUES ('PixelManageList', '像素列表', (SELECT id FROM menus WHERE name = 'PixelManage'), 'list', '/pixel-manage/index', '', 1, false, 'menu', 1, NOW(), NOW());

-- Update R_SUPER menu_ids to include new menus
UPDATE roles
SET menu_ids = array_cat(menu_ids, ARRAY[
    (SELECT id FROM menus WHERE name = 'PixelManage'),
    (SELECT id FROM menus WHERE name = 'PixelManageList')
]::int[])
WHERE id = 1;

-- Verify
SELECT id, name, title, parent_id, path, component, menu_type, sort_order FROM menus WHERE name IN ('PixelManage', 'PixelManageList');

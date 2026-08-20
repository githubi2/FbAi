-- Add PageManage top-level directory + PageManageList child menu
SELECT setval('menus_id_seq', (SELECT COALESCE(MAX(id), 0) FROM menus));

INSERT INTO menus (name, title, parent_id, path, component, icon, sort_order, hidden, menu_type, status, created_at, updated_at)
VALUES ('PageManage', '主页管理', 0, '/page-manage', '/index/index', 'ri:pages-line', 7, false, 'directory', 1, NOW(), NOW());

INSERT INTO menus (name, title, parent_id, path, component, icon, sort_order, hidden, menu_type, status, created_at, updated_at)
VALUES ('PageManageList', '主页列表', (SELECT id FROM menus WHERE name = 'PageManage'), 'list', '/page-manage/index', '', 1, false, 'menu', 1, NOW(), NOW());

-- Update R_SUPER menu_ids to include new menus
UPDATE roles
SET menu_ids = array_cat(menu_ids, ARRAY[
    (SELECT id FROM menus WHERE name = 'PageManage'),
    (SELECT id FROM menus WHERE name = 'PageManageList')
]::int[])
WHERE id = 1;

-- Verify
SELECT id, name, title, parent_id, path, component, menu_type, sort_order FROM menus WHERE name IN ('PageManage', 'PageManageList');
SELECT id, role_name, menu_ids FROM roles WHERE id = 1;

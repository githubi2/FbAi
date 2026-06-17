-- Add AdAccountManage menu (child of AdAccount id=12)
SELECT setval('menus_id_seq', (SELECT COALESCE(MAX(id), 0) FROM menus));

INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order, hidden)
VALUES ('AdAccountManage', 12, 'manage', '/ad-account/manage/index', '', '广告账户管理', 2, false);

-- Update R_SUPER menu_ids to include new menu
UPDATE roles SET menu_ids = '{1,3,2,4,5,6,7,12,13,14}' WHERE id = 1;

-- Show result
SELECT id, name, parent_id, path, sort_order FROM menus WHERE parent_id = 12 OR id = 12 ORDER BY parent_id, sort_order;
SELECT id, role_name, menu_ids FROM roles WHERE id = 1;

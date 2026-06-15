INSERT INTO menus (name, title, parent_id, path, component, menu_type, icon, sort_order, hidden, status)
VALUES ('Tenant', '租户管理', 2, 'tenant', '/system/tenant', 'menu', 'building', 6, false, 1)
ON CONFLICT DO NOTHING;

-- 更新 R_SUPER 的 menu_ids 包含新菜单
UPDATE roles SET menu_ids = array_append(menu_ids, 7) WHERE role_code = 'R_SUPER' AND NOT (7 = ANY(menu_ids));

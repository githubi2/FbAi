-- 插入租户系统管理菜单
INSERT INTO menus (id, name, title, parent_id, path, component, icon, hidden, menu_type, sort_order) VALUES
(8, 'TenantSystem', '租户系统管理', 0, '/tenant-system', '/index/index', 'ri:building-2-line', false, 'directory', 3),
(9, 'TenantUser', '用户管理', 8, 'user', '/tenant-system/user', 'ri:user-3-line', false, 'menu', 1),
(10, 'TenantRole', '角色管理', 8, 'role', '/tenant-system/role', 'ri:shield-user-line', false, 'menu', 2),
(11, 'TenantMenu', '菜单管理', 8, 'menu', '/tenant-system/menu', 'ri:menu-line', false, 'menu', 3)
ON CONFLICT (id) DO NOTHING;

-- 更新现有租户管理员角色的 menu_ids: {1,3,8,9,10,11}
-- Dashboard(id=1), Console(id=3), TenantSystem(id=8), TenantUser(id=9), TenantRole(id=10), TenantMenu(id=11)
UPDATE roles SET menu_ids = '{1,3,8,9,10,11}' WHERE role_code = 'T8_R_ADMIN';

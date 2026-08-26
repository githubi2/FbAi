-- 插入广告账户管理菜单（不指定 ID，让序列自动生成）
-- 父菜单: 广告管理
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order, hidden)
VALUES ('AdAccount', 0, '/ad-account', '/index/index', 'ri:advertisement-line', '广告管理', 4, false);

-- 子菜单: 账户列表
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order, hidden)
VALUES ('AdAccountList', (SELECT id FROM menus WHERE name = 'AdAccount'), 'list', '/ad-account/index', '', '账户列表', 1, false);

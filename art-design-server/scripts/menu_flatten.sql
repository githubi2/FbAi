-- 菜单扁平化改造：广告管理/BM管理/主页管理/像素管理 目录移除，子页面提升为一级菜单
BEGIN;

-- 1. 提升子级为一级菜单（parent_id=0 + 独立一级路径 + 原组件）
UPDATE menus SET parent_id = 0, path = '/ad-account',       component = '/ad-account/index',        icon = 'ri:advertisement-line', sort_order = 5 WHERE id = 13;
UPDATE menus SET parent_id = 0, path = '/ad-account-manage', component = '/ad-account/manage/index', icon = 'ri:advertisement-line', sort_order = 6 WHERE id = 14;
UPDATE menus SET parent_id = 0, path = '/bm-manage',         component = '/ad-account/bm/index',     icon = 'ri:building-2-line',     sort_order = 7 WHERE id = 16;
UPDATE menus SET parent_id = 0, path = '/page-manage',       component = '/page-manage/index',       icon = 'ri:pages-line',          sort_order = 8 WHERE id = 18;
UPDATE menus SET parent_id = 0, path = '/pixel-manage',      component = '/pixel-manage/index',      icon = 'ri:crosshair-2-line',    sort_order = 9 WHERE id = 20;

-- 2. 删除目录
DELETE FROM menus WHERE id IN (12, 15, 17, 19);

-- 3. 角色 menu_ids 移除目录 ID（子级保留）
UPDATE roles SET menu_ids = array_remove(menu_ids, 12) WHERE 12 = ANY(menu_ids);
UPDATE roles SET menu_ids = array_remove(menu_ids, 15) WHERE 15 = ANY(menu_ids);
UPDATE roles SET menu_ids = array_remove(menu_ids, 17) WHERE 17 = ANY(menu_ids);
UPDATE roles SET menu_ids = array_remove(menu_ids, 19) WHERE 19 = ANY(menu_ids);

COMMIT;

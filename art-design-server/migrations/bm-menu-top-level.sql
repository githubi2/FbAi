-- Move BM管理 from child of 广告管理(id=12) to top-level directory
UPDATE menus SET parent_id = 0, path = '/bm-manage', sort_order = 6 WHERE id = 15;

-- Verify
SELECT id, name, title, parent_id, path, component, menu_type, sort_order FROM menus WHERE id >= 12 ORDER BY id;

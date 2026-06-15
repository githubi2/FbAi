INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (7, 2, '个人中心', 'UserCenter', 'user-center', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'UserCenter', title = '个人中心', path = 'user-center', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (8, 0, '结果页', 'Result', '/result', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Result', title = '结果页', path = '/result', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (9, 8, '成功页', 'ResultSuccess', 'success', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'ResultSuccess', title = '成功页', path = 'success', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (10, 8, '失败页', 'ResultFail', 'fail', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'ResultFail', title = '失败页', path = 'fail', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (11, 0, '异常页', 'Exception', '/exception', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception', title = '异常页', path = '/exception', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (12, 11, '403', 'Exception403', '403', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception403', title = '403', path = '403', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (13, 11, '404', 'Exception404', '404', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception404', title = '404', path = '404', updated_at = NOW();

INSERT INTO menus (id, parent_id, title, name, path, created_at, updated_at) VALUES
  (14, 11, '500', 'Exception500', '500', NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception500', title = '500', path = '500', updated_at = NOW();

-- Fix name mismatch: Menu → Menus
UPDATE menus SET name = 'Menus' WHERE id = 6;

-- Add missing menus for frontend routes
INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (7, 'UserCenter', 'user-center', 2, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'UserCenter', path = 'user-center', parent_id = 2, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (8, 'Result', '/result', 0, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Result', path = '/result', parent_id = 0, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (9, 'ResultSuccess', 'success', 8, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'ResultSuccess', path = 'success', parent_id = 8, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (10, 'ResultFail', 'fail', 8, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'ResultFail', path = 'fail', parent_id = 8, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (11, 'Exception', '/exception', 0, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception', path = '/exception', parent_id = 0, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (12, 'Exception403', '403', 11, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception403', path = '403', parent_id = 11, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (13, 'Exception404', '404', 11, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception404', path = '404', parent_id = 11, updated_at = NOW();

INSERT INTO menus (id, name, path, parent_id, created_at, updated_at) VALUES
  (14, 'Exception500', '500', 11, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET name = 'Exception500', path = '500', parent_id = 11, updated_at = NOW();

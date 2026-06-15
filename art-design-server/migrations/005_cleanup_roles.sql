DELETE FROM roles WHERE role_code IN ('U_ADMIN', 'U_USER');
INSERT INTO roles (id, role_name, role_code, description, status, created_at, updated_at) 
VALUES (2, '管理员', 'R_ADMIN', '管理员权限', 1, now(), now())
ON CONFLICT (id) DO UPDATE SET role_name='管理员', role_code='R_ADMIN';
INSERT INTO roles (id, role_name, role_code, description, status, created_at, updated_at) 
VALUES (3, '普通用户', 'R_USER', '普通用户权限', 1, now(), now())
ON CONFLICT (id) DO UPDATE SET role_name='普通用户', role_code='R_USER';
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE code IN (
  'system:user:list','system:user:create','system:user:edit','system:user:delete',
  'system:role:list','system:role:create','system:role:edit','system:role:delete',
  'system:menu:list','dashboard:view','dashboard:export'
) ON CONFLICT DO NOTHING;
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions WHERE code IN ('dashboard:view') ON CONFLICT DO NOTHING;

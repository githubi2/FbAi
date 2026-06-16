-- 为已有租户管理员角色添加 system:menu:list 权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.tenant_id IS NOT NULL
  AND r.role_code LIKE 'T%_R_ADMIN'
  AND p.code = 'system:menu:list'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );

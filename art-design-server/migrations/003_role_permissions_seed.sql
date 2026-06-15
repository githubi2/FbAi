-- ============================================
-- 为现有角色分配权限
-- ============================================

-- R_SUPER (role_id=1) — 所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- R_ADMIN (role_id=2) — 系统管理权限（不含租户管理）
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions 
WHERE code IN (
    'system:user:list', 'system:user:create', 'system:user:edit', 'system:user:delete',
    'system:role:list', 'system:role:create', 'system:role:edit', 'system:role:delete',
    'system:menu:list',
    'dashboard:view', 'dashboard:export'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- R_USER (role_id=3) — 最小权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions 
WHERE code IN ('dashboard:view')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ============================================
-- 创建索引
-- ============================================
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant_id ON sessions(tenant_id);

-- ============================================
-- RLS 策略（需要 postgres 用户执行）
-- ============================================

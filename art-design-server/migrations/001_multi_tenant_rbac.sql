-- ============================================
-- 多租户 + RBAC 权限系统 — 数据库迁移
-- ============================================

BEGIN;

-- ============================================
-- 1. 创建 tenants 表
-- ============================================
CREATE TABLE IF NOT EXISTS tenants (
    id            SERIAL PRIMARY KEY,
    name          VARCHAR(128) NOT NULL,
    code          VARCHAR(64) UNIQUE NOT NULL,
    status        INT DEFAULT 1,
    contact_name  VARCHAR(64) DEFAULT '',
    contact_phone VARCHAR(20) DEFAULT '',
    contact_email VARCHAR(128) DEFAULT '',
    description   VARCHAR(256) DEFAULT '',
    created_at    TIMESTAMP DEFAULT now(),
    updated_at    TIMESTAMP DEFAULT now()
);

-- ============================================
-- 2. 为已有表添加 tenant_id 字段
-- ============================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id);
ALTER TABLE roles ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id);
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS tenant_id INT REFERENCES tenants(id);

-- 已有角色为系统全局角色，tenant_id 保持 NULL
-- 已有管理员用户 tenant_id 保持 NULL（全局管理员）

-- ============================================
-- 3. 创建 permissions 表（权限点）
-- ============================================
CREATE TABLE IF NOT EXISTS permissions (
    id          SERIAL PRIMARY KEY,
    code        VARCHAR(128) UNIQUE NOT NULL,
    name        VARCHAR(128) NOT NULL,
    module      VARCHAR(64) NOT NULL,
    action      VARCHAR(32) NOT NULL,
    description VARCHAR(256) DEFAULT '',
    created_at  TIMESTAMP DEFAULT now()
);

-- ============================================
-- 4. 创建 role_permissions 关联表
-- ============================================
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- ============================================
-- 5. 种子权限点数据
-- ============================================
INSERT INTO permissions (code, name, module, action, description) VALUES
-- system 模块
('system:user:list',     '查看用户列表',   'system', 'list',   '查看用户列表权限'),
('system:user:create',   '创建用户',       'system', 'create', '创建用户权限'),
('system:user:edit',     '编辑用户',       'system', 'edit',   '编辑用户权限'),
('system:user:delete',   '删除用户',       'system', 'delete', '删除用户权限'),
('system:role:list',     '查看角色列表',   'system', 'list',   '查看角色列表权限'),
('system:role:create',   '创建角色',       'system', 'create', '创建角色权限'),
('system:role:edit',     '编辑角色',       'system', 'edit',   '编辑角色权限'),
('system:role:delete',   '删除角色',       'system', 'delete', '删除角色权限'),
('system:menu:list',     '查看菜单列表',   'system', 'list',   '查看菜单列表权限'),
('system:menu:create',   '创建菜单',       'system', 'create', '创建菜单权限'),
('system:menu:edit',     '编辑菜单',       'system', 'edit',   '编辑菜单权限'),
('system:menu:delete',   '删除菜单',       'system', 'delete', '删除菜单权限'),
('system:tenant:list',   '查看租户列表',   'system', 'list',   '查看租户列表权限'),
('system:tenant:create', '创建租户',       'system', 'create', '创建租户权限'),
('system:tenant:edit',   '编辑租户',       'system', 'edit',   '编辑租户权限'),
('system:tenant:delete', '删除租户',       'system', 'delete', '删除租户权限'),
-- dashboard 模块
('dashboard:view',       '查看仪表盘',     'dashboard', 'view',   '查看仪表盘权限'),
('dashboard:export',     '导出报表',       'dashboard', 'export', '导出报表权限')
ON CONFLICT (code) DO NOTHING;

-- ============================================
-- 6. 为现有角色分配权限
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
-- 7. 启用 RLS（Row-Level Security）
-- ============================================
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;

-- RLS 策略：当前租户只能看自己的数据
-- app.current_tenant_id 未设置时（超级管理员），策略不生效
CREATE POLICY tenant_isolation_users ON users
    FOR ALL
    USING (
        current_setting('app.current_tenant_id', true) = '' 
        OR tenant_id = current_setting('app.current_tenant_id')::INT
    )
    WITH CHECK (
        current_setting('app.current_tenant_id', true) = '' 
        OR tenant_id = current_setting('app.current_tenant_id')::INT
    );

CREATE POLICY tenant_isolation_roles ON roles
    FOR ALL
    USING (
        current_setting('app.current_tenant_id', true) = '' 
        OR tenant_id = current_setting('app.current_tenant_id')::INT
    )
    WITH CHECK (
        current_setting('app.current_tenant_id', true) = '' 
        OR tenant_id = current_setting('app.current_tenant_id')::INT
    );

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_sessions_tenant_id ON sessions(tenant_id);

COMMIT;

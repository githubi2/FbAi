# 多租户 + RBAC 权限系统 — 2026-06-16

## 概述
实现了完整的多租户（Multi-Tenant）管理和 RBAC（基于角色的访问控制）权限系统。

## 修改文件

### 数据库迁移
| 文件 | 变更 |
|------|------|
| `migrations/001_multi_tenant_rbac.sql` | 新增 tenants/permissions/role_permissions 表 |
| `migrations/002_seed_permissions.sql` | 种子 18 个权限点 |
| `migrations/003_role_permissions_seed.sql` | 为默认角色分配权限 |
| `migrations/004_add_tenant_menu.sql` | 添加租户管理菜单 |

### 后端新增
| 文件 | 用途 |
|------|------|
| `models/tenant.go` | Tenant、Permission 模型 + 请求/响应结构体 |
| `services/tenant_service.go` | 租户 CRUD + 事务创建（含角色和用户） + 权限查询 |
| `handlers/tenant_handler.go` | 租户 API 处理器 + 租户切换 |
| `middleware/rbac.go` | RequirePermission 权限中间件 + TenantContext 租户上下文 |

### 后端修改
| 文件 | 变更 |
|------|------|
| `models/user.go` | User 增加 TenantID；UserInfoResponse 增加 permissions/tenantId/tenantName |
| `models/role.go` | Role 增加 TenantID；Create/UpdateRequest 增加 PermissionIDs |
| `models/session.go` | Session 增加 TenantID |
| `services/session_service.go` | Create 接受 tenantID；新增 GetTenantID/SetTenantID |
| `services/user_service.go` | 所有查询增加 tenant_id；GetPasswordHash 返回 tenantID |
| `services/role_service.go` | 所有查询增加 tenant_id；Create/Update 处理 permissionIDs |
| `services/menu_service.go` | listFallback 增加 Tenant 菜单 |
| `handlers/auth_handler.go` | Login 适配新签名；GetUserInfo 返回 permissions + tenant |
| `middleware/auth.go` | ValidateUser 返回 tenantID；GenerateToken 接受 tenantID |
| `routes/router.go` | 全部路由加 RequirePermission 守卫；新增租户路由 |

### 前端新增
| 文件 | 用途 |
|------|------|
| `src/api/tenant.ts` | 租户 API 调用层 |
| `src/views/system/tenant/index.vue` | 租户管理列表页 |
| `src/views/system/tenant/modules/TenantForm.vue` | 创建/编辑租户弹窗 |
| `src/components/common/TenantSwitcher.vue` | 顶部租户切换下拉组件 |
| `src/store/modules/tenant.ts` | 租户状态管理（Pinia + persistedstate） |

### 前端修改
| 文件 | 变更 |
|------|------|
| `src/types/api/api.d.ts` | 新增 Tenant 命名空间类型 |
| `src/router/modules/system.ts` | 新增 /system/tenant 路由 |
| `src/router/core/MenuProcessor.ts` | MENU_I18N_MAP + MENU_ICON_MAP 增加 Tenant |
| `src/locales/langs/zh.json` | 新增 menus.system.tenant |
| `src/locales/langs/en.json` | 新增 menus.system.tenant |
| `src/components/core/layouts/art-header-bar/index.vue` | 集成 TenantSwitcher |

## API 端点

| 方法 | 路径 | 权限 | 说明 |
|------|------|------|------|
| GET | /api/v1/tenants | system:tenant:list | 租户列表 |
| GET | /api/v1/tenants/:id | system:tenant:list | 租户详情 |
| POST | /api/v1/tenants | system:tenant:create | 创建租户（含管理员账号） |
| PUT | /api/v1/tenants/:id | system:tenant:edit | 更新租户 |
| DELETE | /api/v1/tenants/:id | system:tenant:delete | 删除租户 |
| POST | /api/v1/tenants/switch | AuthRequired | 切换租户上下文 |
| GET | /api/v1/tenants/current | AuthRequired | 当前租户上下文 |

## 权限点（18个）

- system:user:list/create/edit/delete
- system:role:list/create/edit/delete
- system:menu:list/create/edit/delete
- system:tenant:list/create/edit/delete
- dashboard:view/export

## 设计决策

1. **权限与菜单分离**：menu_ids 控制侧边栏可见性，permissions 控制 API 操作 + 按钮可见性
2. **租户隔离双保险**：Service 层过滤 + PostgreSQL RLS 策略
3. **管理员 SSO + 租户多登**：管理员登录踢旧会话，租户用户允许多处登录
4. **创建租户事务**：一次请求创建租户 + 角色 + 管理员账号 + 权限分配

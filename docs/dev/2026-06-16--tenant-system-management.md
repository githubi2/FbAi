# 租户系统管理 — 2026-06-16

## 概述

新增独立于超级管理员"系统管理"的"租户系统管理"菜单。租户管理员可在此管理本租户内的用户、角色和查看菜单权限。

## 核心设计

| 项目 | 超级管理员 | 租户管理员 |
|------|-----------|-----------|
| 看到的菜单 | 系统管理（全局） | 租户系统管理（租户内） |
| 用户管理 | 所有用户 | 仅本租户用户（RLS 过滤） |
| 角色管理 | 所有角色 | 仅本租户角色（RLS 过滤） |
| 菜单管理 | 完整 CRUD | **只读**（菜单为全局资源，不可修改） |
| 权限分配 | 所有菜单可选 | 仅租户管理员可见菜单的子集 |

## 修改文件

### 数据库

| 变更 | 说明 |
|------|------|
| `menus` 表新增 4 条记录 | id=8 TenantSystem, 9 TenantUser, 10 TenantRole, 11 TenantMenu |
| `roles.menu_ids` 更新 | 租户管理员角色 menu_ids: {1,3,8,9,10,11}（Dashboard, Console + 租户系统管理） |
| `role_permissions` 更新 | 租户管理员新增 `system:menu:list` 权限 |

### 后端 (art-design-server)

| 文件 | 变更 |
|------|------|
| `services/tenant_service.go` | 新租户创建时分配新 menu_ids；T_ADMIN 新增 system:menu:list 权限 |
| `handlers/role_handler.go` | `GetMenus`: 租户上下文中过滤 allMenus（只显示当前用户可见菜单）; `Create`: 自动绑定 tenant_id |
| `services/role_service.go` | 新增 `GetUserMenuIDs(userID)` 方法 |
| `migrations/add_tenant_system_menus.sql` | 菜单和角色更新的 SQL 迁移 |
| `migrations/add_tenant_menu_permission.sql` | 已有租户权限更新 SQL |

### 前端 (art-design-pro)

| 文件 | 变更 |
|------|------|
| `src/router/modules/tenantSystem.ts` | **新增** 租户系统管理路由模块 |
| `src/router/modules/index.ts` | 注册 tenantSystemRoutes |
| `src/views/tenant-system/user/index.vue` | **新增** 租户用户管理页面 |
| `src/views/tenant-system/user/modules/user-dialog.vue` | **新增** 租户用户弹窗 |
| `src/views/tenant-system/user/modules/user-search.vue` | **新增** 租户用户搜索 |
| `src/views/tenant-system/role/index.vue` | **新增** 租户角色管理页面 |
| `src/views/tenant-system/role/modules/role-edit-dialog.vue` | **新增** 租户角色编辑弹窗 |
| `src/views/tenant-system/role/modules/role-permission-dialog.vue` | **新增** 租户角色权限弹窗（菜单树已过滤） |
| `src/views/tenant-system/menu/index.vue` | **新增** 租户菜单查看页面（只读） |
| `src/router/core/MenuProcessor.ts` | 添加 TenantSystem/TenantUser/TenantRole/TenantMenu 的 i18n 和 icon 映射 |
| `src/locales/langs/zh.json` | 添加 `menus.tenantSystem.*` 中文翻译 |
| `src/locales/langs/en.json` | 添加 `menus.tenantSystem.*` 英文翻译 |

## 菜单树

```
租户管理员可见:
  仪表盘 (Dashboard)
    └─ 工作台 (Console)
  租户系统管理 (TenantSystem)
    ├─ 用户管理 (TenantUser)
    ├─ 角色管理 (TenantRole)
    └─ 菜单管理 (TenantMenu)  [只读]

超级管理员可见（不变）:
  仪表盘 (Dashboard)
    └─ 工作台 (Console)
  系统管理 (System)
    ├─ 用户管理 (User)
    ├─ 角色管理 (Role)
    ├─ 菜单管理 (Menus)
    └─ 租户管理 (Tenant)
```

## API 变更

### 新增/修改的 API 行为

| 端点 | 变更 |
|------|------|
| `GET /api/v1/roles/:id/menus` | 租户上下文下 `allMenus` 只返回当前用户角色可见的菜单 |
| `POST /api/v1/roles` | 租户上下文下自动绑定 `tenant_id` |

## 验证方法

1. 登录超级管理员 (admin/admin123)：侧边栏应显示"系统管理"（含用户/角色/菜单/租户管理），不显示"租户系统管理"
2. 登录租户管理员：侧边栏应显示"租户系统管理"（含用户/角色/菜单管理），不显示"系统管理"
3. 租户管理员 → 用户管理：应只显示本租户的用户，可新增/编辑/删除
4. 租户管理员 → 角色管理：应只显示本租户的角色，可新增/编辑子角色，权限分配菜单树仅限于可见菜单
5. 租户管理员 → 菜单管理：应只读显示可见菜单树，不能修改

## 关键原则遵循

- **Rule 15 (Feature Isolation)**: 新建 `views/tenant-system/` 目录，不修改 `views/system/`
- **Rule 14 (Modularization)**: 每个功能独立文件，页面 < 250 行
- **Rule 7.6 (Table Standard)**: 使用 ArtTable + ArtTableHeader + useTable
- **Rule 7.7 (Button Style)**: 页面顶部 plain button + v-ripple，弹窗 footer 标准按钮

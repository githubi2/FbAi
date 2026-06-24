# 共享角色改造

**日期**: 2026-06-24  
**类型**: refactor  
**影响范围**: 后端 tenant_service.go + 前端 role/index.vue

## 问题
多租户架构下，每创建一个租户就自动创建 2 个角色（租户管理员 + 普通用户），导致角色表中出现同名角色（如两个"租户管理员"），用户难以区分。

## 根因
`tenant_service.go` 的 `Create()` 方法在事务中创建角色，`role_name` 固定为"租户管理员"/"普通用户"，仅通过 `role_code`（`T{id}_R_ADMIN`）区分。

## 方案
改为全局共享角色：
- `R_TENANT_ADMIN` — 租户管理员（共享，tenant_id=NULL）
- `R_TENANT_USER` — 普通用户（共享，tenant_id=NULL）

创建租户时不再建角色，改为查询已有的共享角色 ID。

## 修改文件

| 文件 | 改动 |
|------|------|
| `art-design-server/services/tenant_service.go` | 删除角色创建+权限分配逻辑（~50行），改为 `SELECT id FROM roles WHERE role_code='R_TENANT_ADMIN'` |
| `art-design-pro/src/views/system/role/index.vue` | 移除租户列、fetchGetTenantList 导入、tenantNameMap |
| `art-design-pro/src/types/api/api.d.ts` | 添加 `tenantId?` 字段到 RoleListItem |

## 数据库迁移
通过 API 执行：
1. 创建共享角色 R_TENANT_ADMIN (id=28) + R_TENANT_USER (id=29)
2. 用户 yixiu(17) test2(33) 的 role_id 迁移到 28
3. 删除旧租户角色 21/22/26/27
4. 分配权限：R_TENANT_ADMIN 获 11 个权限点，R_TENANT_USER 获 1 个

## 验证
- 角色管理页面从 5 条降至 3 条，无重复
- yixiu（租户管理员）登录正常，菜单正确

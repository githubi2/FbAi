# 租户登录菜单缺失修复 — 2026-06-16

## 问题

租户管理员账号 `yixiu/admin123` 登录后：
1. 出现 **"抱歉，服务器出错了"**（首次：menu_ids 完全为空）
2. 看到仪表盘和工作台，但**缺少"用户管理"菜单**（二次：孤儿菜单问题）

## 根因分析（三层问题）

### 第一层：租户角色 menu_ids 为空（已修复）

`tenant_service.go` 创建租户角色时，INSERT 语句没有 `menu_ids` 字段 → 默认 `{}` → `/api/v1/auth/menus` 返回 `null` → 前端动态路由注册失败 → 500。

### 第二层：用户 role_id 指向错误角色

用户 `yixiu` 的 `role_id=2`（全局 R_ADMIN），而非 `role_id=19`（租户 T7_R_ADMIN）。
且全局角色 2 被错误重命名为"租户管理员"（原为"管理员"）。

### 第三层：孤儿菜单 — 子菜单的父菜单不在 menu_ids 中

角色 2 的 `menu_ids={1,3,4}`：
- 1=Dashboard, 3=Console(父=1), 4=User(父=2)
- **System(id=2) 不在 menu_ids 中** → User(id=4) 的父菜单缺失 → User 成为孤儿 → 不显示

## 修复

### 1. 代码修复

| 文件 | 修改 |
|------|------|
| `services/tenant_service.go` | 租户角色 INSERT 增加 `menu_ids` 列 |
| `services/menu_service.go` | `TreeByIDs()` 自动补全祖先菜单，防止孤儿 |

**TreeByIDs 孤儿预防逻辑**：
```go
// 对每个目标菜单，向上遍历添加所有祖先到 idSet
for _, id := range menuIDs {
    for mid := uint(id); mid != 0; {
        idSet[mid] = true
        parent := menuMap[mid]
        mid = parent.ParentID
    }
}
```

### 2. 数据库修复

```sql
-- 恢复角色名称
UPDATE roles SET role_name = '管理员' WHERE id = 2;
UPDATE roles SET role_name = '普通用户' WHERE id = 3;

-- 修复角色 2 的 menu_ids（添加 System 父菜单）
UPDATE roles SET menu_ids = '{1,2,3,4}' WHERE id = 2;

-- 修复 yixiu 的 role_id
UPDATE users SET role_id = 19 WHERE id = 16;
```

## 菜单分配方案

| 角色 | menu_ids | 菜单内容 |
|------|----------|---------|
| R_SUPER | {1,2,3,4,5,6,7} | 全部 |
| R_ADMIN | {1,2,3,4} | Dashboard, System, Console, User |
| R_USER | {} | 无 |
| T{id}_R_ADMIN | {1,2,3,4,5} | Dashboard, System, Console, User, Role |
| T{id}_R_USER | {1,3} | Dashboard, Console |

## 验证结果

```
yixiu (租户管理员, roleId=19):
  Dashboard → Console
  System   → User, Role        ✅

admin (超级管理员):
  Dashboard → Console
  System   → User, Role, Menus, Tenant  ✅
```

## 修改文件清单

| 文件 | 修改 |
|------|------|
| `art-design-server/services/tenant_service.go` | 租户角色 INSERT 增加 `menu_ids` |
| `art-design-server/services/menu_service.go` | `TreeByIDs()` 自动补全祖先 |
| DB `roles` 表 | 修复角色名称 + menu_ids |
| DB `users` 表 | 修复 yixiu 的 role_id |

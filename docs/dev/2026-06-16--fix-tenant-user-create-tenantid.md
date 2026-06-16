# Fix: 租户系统管理用户创建未绑定 tenant_id — 2026-06-16

## 问题描述

租户管理员在「租户系统管理 → 用户管理」中创建用户时，新用户 `tenant_id` 为 NULL，导致：
1. 新创建的用户不在租户用户列表中显示（因为列表按 `tenant_id` 过滤）
2. 新创建的用户出现在超级管理员的全局用户列表中
3. 租户数据隔离失效

## 根因

`UserHandler.Create` 缺少从 Gin 上下文自动绑定 `tenantID` 的逻辑。对比 `RoleHandler.Create`（已正确实现）：

```go
// RoleHandler.Create ✅ 正确实现
if req.TenantID == nil {
    if tid, exists := c.Get("tenantID"); exists {
        if t, ok := tid.(*uint); ok && t != nil {
            req.TenantID = t
        }
    }
}

// UserHandler.Create ❌ 修复前缺失这段逻辑
```

租户管理员创建用户时，前端不发送 `tenantId` 字段（也不应该发送），但后端 `UserHandler.Create` 没有从 `TenantContext` 中间件设置的上下文 (`c.Get("tenantID")`) 中获取租户 ID，导致 `CreateUserRequest.TenantID` 始终为 nil，用户以全局身份创建。

## 修改内容

### 1. UserHandler.Create — 自动绑定 tenantID

**文件**: `handlers/user_handler.go`

在 `ShouldBindJSON` 之后、`Create` 之前，添加与 `RoleHandler.Create` 相同的自动绑定逻辑（第 68-75 行）。

### 2. UserHandler.Update — 租户隔离检查

**文件**: `handlers/user_handler.go`

添加 `checkTenantAccess()` 方法（第 125-148 行），在 Update 和 Delete 操作前验证：
- 超级管理员（tenantID=nil）可操作所有用户
- 租户用户只能操作同租户的用户（`targetUser.TenantID == requestTenantID`）
- 跨租户操作返回 403 Forbidden

### 3. 清理孤儿用户

删除了 `w_568724`（id=25），该用户由租户管理员创建但 `tenant_id=NULL`、`role_id=25`（不存在的角色）。

## 验证

| 测试场景 | 结果 |
|---------|------|
| 租户用户创建用户 → tenantId 自动绑定 | ✅ tenantId=8 |
| 租户用户列表 → 只显示本租户用户 | ✅ 3 users (all tenant_id=8) |
| 租户用户删除超级管理员 → 拒绝 | ✅ 403 Forbidden |
| 租户用户更新超级管理员 → 拒绝 | ✅ 403 Forbidden |
| 租户用户删除本租户用户 → 允许 | ✅ 200 OK |
| 超级管理员创建全局用户 → 正常 | ✅ tenantId=null |
| 超级管理员删除租户用户 → 正常 | ✅ 200 OK |

## 影响范围

- 租户系统管理的用户管理功能现在正确隔离数据
- 已创建的孤儿用户（如 w_568724）需手动清理
- 向前兼容：前端无需修改，API 行为透明

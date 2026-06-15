# Fix Role Permission Dialog — 2026-06-16

## Problem
角色管理的「菜单权限」编辑弹窗所有角色都不显示已勾选的菜单权限，保存功能也未实现。

## Root Cause Analysis

| # | Bug | File:Line | Impact |
|---|-----|-----------|--------|
| 1 | `:default-checked-keys="[1,2,3]"` 硬编码 + `node-key="name"` 类型不匹配 | `role-permission-dialog.vue:15,17` | 数字 key `[1,2,3]` 匹配不到字符串 name（如 `"Dashboard"`），所有角色都不显示勾选 |
| 2 | `watch` 弹窗打开时未调用 API 加载角色权限 | `role-permission-dialog.vue:142-149` | 只有 `TODO` 注释 + `console.log`，从未加载后端数据 |
| 3 | `savePermission` 未实现 | `role-permission-dialog.vue:163-167` | 只有 `TODO` 注释，仅弹出虚拟 toast 不调后端 |
| 4 | `getAllNodeKeys` 用 `name` 而非与 `node-key` 一致的 `id` | `role-permission-dialog.vue:208-218` | 全选/取消全选功能使用错误的 key |

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/views/system/role/modules/role-permission-dialog.vue` | Whole file rewrite of watch/save/getAllNodeKeys + node-key change | Fix all 4 bugs |

### Detailed Changes

1. **`node-key="name"` → `node-key="id"`** (line 15)
   - 后端 `roleMenus` 返回的是 menu ID（数字），tree 的 `id` 字段与之一致
   
2. **移除硬编码 `:default-checked-keys="[1,2,3]"`** (was line 17)
   - 改用 `setCheckedKeys()` 动态设置从后端加载的权限
   
3. **实现 `watch` 权限加载** (lines 139-157)
   - 弹窗打开时调用 `fetchGetRoleMenus(roleId)` 获取 `roleMenus`
   - `await nextTick()` 等待 tree 渲染后 `setCheckedKeys(roleMenus)`
   
4. **实现 `savePermission`** (lines 168-201)
   - 收集 `tree.getCheckedKeys()`，过滤出数字 key（纯菜单 ID，排除 auth 子节点字符串 key）
   - 调用 `fetchUpdateRole(roleId, { roleName, roleCode, description, status, menuIds })` 保存
   
5. **修复 `getAllNodeKeys`** (lines 236-251)
   - 从 `node.name` 改用 `node.id`，与 `node-key="id"` 一致
   - 返回类型从 `string[]` 改为 `(string | number)[]`

6. **修复 `handleClose`** (line 165)
   - 增加 `isSelectAll.value = false` 重置全选按钮状态

## API Used

- `GET /api/v1/roles/:id/menus` — 获取角色的菜单权限（返回 `{ allMenus, roleMenus }`）
- `PUT /api/v1/roles/:id` — 更新角色（含 `menuIds` 字段）

## Verification

```bash
# TypeScript check
npx vue-tsc --noEmit  # ✅ passed

# ESLint check
npx eslint src/views/system/role/modules/role-permission-dialog.vue  # ✅ passed
```

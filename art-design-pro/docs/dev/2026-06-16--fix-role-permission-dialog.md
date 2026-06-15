# Fix Role Permission Dialog — 2026-06-16

## Problem
角色管理的「菜单权限」编辑弹窗所有角色都不显示已勾选的菜单权限，保存功能也未实现。

## Root Cause Analysis (Round 1)

| # | Bug | Impact |
|---|-----|--------|
| 1 | `:default-checked-keys="[1,2,3]"` + `node-key="name"` 类型不匹配 | 数字 key 匹配不到字符串 name，所有角色不显示勾选 |
| 2 | `watch` 弹窗打开时未调用 API 加载角色权限 | 只有 TODO + console.log |
| 3 | `savePermission` 未实现 | 只有 TODO，仅弹出虚拟 toast |
| 4 | `getAllNodeKeys` 用 `name` 但 node-key 改成了 `id` | 全选功能 key 不一致 |

## Root Cause Analysis (Round 2 — frontend mode issue)

**项目运行在 `VITE_ACCESS_MODE=frontend` 模式。**

第一次修复将 `node-key` 改为 `"id"`，但在 frontend 模式下：
- `menuList` 来自 `asyncRoutes`（前端路由定义）
- 前端路由 **没有 `id` 字段**（`AppRouteRecord.id` 是 optional 且未赋值）
- 树节点 key 全部为 `undefined` → `setCheckedKeys()` 永远无法匹配

## Final Fix

### 核心思路：ID ↔ Name 双向映射

后端存储用 `menuIds`（数字 ID），前端路由用 `name`（字符串）。通过 `GET /api/v1/roles/:id/menus` 返回的 `allMenus` 数据构建映射：

```
allMenus = [{ id: 1, name: "Dashboard" }, { id: 2, name: "System" }, ...]
         ↓
idToNameMap:  1 → "Dashboard", 2 → "System", ...
nameToIdMap:  "Dashboard" → 1, "System" → 2, ...
```

### 修改内容

| 变更 | 说明 |
|------|------|
| `node-key` 保持 `"name"` | 与前端路由 name 字段一致 |
| 弹窗打开时调用 `fetchGetRoleMenus` | 获取 `{ allMenus, roleMenus }` |
| 构建 `idToNameMap` / `nameToIdMap` | 基于 `allMenus` 数据 |
| 加载：`roleMenus(ID)` → `name` 列表 → `setCheckedKeys` | 后端 ID 转为树节点 name |
| 保存：勾选的 `name` → `ID` 列表 → `fetchUpdateRole` | 前端 name 转为后端 ID |
| `getAllNodeKeys` 使用 `name` | 与 `node-key="name"` 一致 |
| 修复 `handleClose` 重置 `isSelectAll` | 关闭弹窗时重置全选按钮 |

### 关键代码

```typescript
// 构建双向映射
const idToNameMap = ref<Map<number, string>>(new Map())
const nameToIdMap = ref<Map<string, number>>(new Map())

// 加载权限（ID → name → setCheckedKeys）
const roleMenus: number[] = res.roleMenus || []
const allMenus: any[] = (res as any).allMenus || []

allMenus.forEach((menu) => {
  idToNameMap.set(Number(menu.id), menu.name)
  nameToIdMap.set(menu.name, Number(menu.id))
})

const checkedNames = roleMenus.map(id => idToNameMap.get(id)!).filter(Boolean)
treeRef.value?.setCheckedKeys(checkedNames)

// 保存（name → ID → update API）
const checkedKeys: string[] = tree.getCheckedKeys()
const menuIds = checkedKeys
  .map(name => nameToIdMap.get(name))
  .filter((id): id is number => id != null)

await fetchUpdateRole(roleId, { ..., menuIds })
```

## API Used

- `GET /api/v1/roles/:id/menus` — 返回 `{ allMenus, roleMenus }`
- `PUT /api/v1/roles/:id` — 更新角色（含 `menuIds` 字段）

## Verification

```bash
npx vue-tsc --noEmit   # ✅
pnpm build             # ✅
```

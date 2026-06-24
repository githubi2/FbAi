# 菜单编辑表单修复 — 2026-06-16

## 问题 1：菜单名称不支持中文输入

菜单管理编辑页的「菜单名称」输入框校验规则 `/^[a-zA-Z][a-zA-Z0-9_]*$/` 只允许英文字母/数字/下划线，不支持中文。DB `menus.title` 存储的就是中文显示名称（如"广告管理"、"系统管理"）。

### 根因

原表单只有一个 `name` 字段（标签"菜单名称"），提交时同时赋值给 `title` 和 `name`：
```typescript
title: formData.name,  // 显示名称 → 需要中文
name: formData.name,   // 路由标识 → 需要英文
```
由于 `name` 校验禁止中文，无法输入中文菜单名。

### 修复

拆分表单为两个独立字段：`title`（菜单名称，支持中文）和 `name`（路由标识，英文限制）。

---

## 问题 2：编辑目录菜单后类型变成「菜单」

修改菜单名称保存后，「目录」类型变成了「菜单」类型。

### 根因

1. **`menuTreeToAppRoute` 丢失 `menuType`**：后端返回的 `menuType` 在转换时没有映射到前端 `meta`
2. **`handleSubmit` 更新时没传 `menuType`**：后端收到空值默认填 `"menu"`，覆盖了 `"directory"`
3. **`handleSubmit` 新建时硬编码 `menuType: 'menu'`**：无法创建目录类型

### 修复

- `menuTreeToAppRoute` 映射 `menuType` 到 `meta.menuType`
- `handleSubmit` 编辑时优先用 `editData.meta.menuType`（后端原值）
- `handleSubmit` 新建时用 `formData.menuType`（用户选择）

---

## 问题 3：缺少目录类型和上级菜单选择

菜单编辑弹窗只有「菜单」和「按钮」两种类型，无法创建/编辑目录。且没有选择上级菜单的功能。

### 修复

1. **「目录」单选按钮**：新增 `ElRadioButton value="directory"`
2. **上级菜单选择器**：使用 `ElTreeSelect`，传入当前菜单树数据，支持搜索和清除（顶级菜单）
3. **目录专用表单**：简化为核心字段（名称、路由、路径、组件、图标、排序、启用/隐藏）
4. **新建时类型可切换**：`handleAddMenu` 不再锁定类型，用户可自由选择目录/菜单/按钮
5. **编辑时读取后端类型**：`loadFormData` 根据 `meta.menuType` 自动设置表单类型

---

## 修改内容

### Modified

| 文件 | 修改 | 原因 |
|------|------|------|
| `src/views/system/menu/modules/menu-dialog.vue` | ① Props 加 `menuTree`，`type` 加 `directory`；② 模板加「目录」radio + `ElTreeSelect` 上级菜单 slot；③ `MenuFormData` 加 `parentId`；④ 表单类型加 `'directory'`；⑤ `formItems` 加 `directory` 分支（简化字段）+ 所有类型显示上级菜单；⑥ `loadFormData` 读后端 `menuType` 并设置；⑦ `dialogTitle` 支持目录；⑧ `disableMenuType` 简化 | 支持目录创建/编辑和上级菜单选择 |
| `src/views/system/menu/index.vue` | ① `menuTreeToAppRoute` 映射 `menuType`；② `MenuFormData` 加 `title`/`menuType`；③ `handleSubmit` 分别传 `title`/`name`/`menuType`/`parentId`；④ `dialogType` 类型加 `directory`；⑤ `handleEditMenu` 根据后端类型设 `dialogType`；⑥ `handleAddMenu` 不锁定类型；⑦ 传 `menuTree` prop | 完整的目录/菜单/按钮 CRUD + 上级菜单 |

### 校验规则对照

| 字段 | 旧规则 | 新规则 |
|------|--------|--------|
| `title`（菜单名称） | 不存在 | `required`, `min:2, max:64`（**支持中文**） |
| `name`（路由标识） | `/^[a-zA-Z][a-zA-Z0-9_]*$/` | 不变 |
| `label`（权限标识） | `required`, `max:64` | **已删除** |
| `parentId`（上级菜单） | 不存在 | **新增**（ElTreeSelect，clearable） |

### 表单字段对照（按类型）

| 字段 | 目录 | 菜单 | 按钮 |
|------|:--:|:--:|:--:|
| 上级菜单 | ✅ | ✅ | — |
| 名称 (title) | ✅ | ✅ | ✅ (authName) |
| 路由标识 (name) | ✅ | ✅ | ✅ (authLabel) |
| 路由地址 (path) | ✅ | ✅ | — |
| 组件路径 (component) | ✅ | ✅ | — |
| 图标 | ✅ | ✅ | — |
| 排序 | ✅ | ✅ | ✅ |
| 外部链接/link | — | ✅ | — |
| 角色权限/roles | — | ✅ | — |
| 徽章/缓存等 switches | — | ✅ | — |
| 启用/隐藏 | ✅ | ✅ | — |

---

## 验证

- ✅ ESLint: 0 errors
- ✅ TypeScript: `vue-tsc --noEmit` 通过
- ✅ Prettier: 自动格式化后通过

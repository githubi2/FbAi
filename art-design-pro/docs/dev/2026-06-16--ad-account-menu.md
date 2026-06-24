# 广告账户管理菜单 & 页面 — 2026-06-16

## 概述
新增「广告管理」菜单及「账户列表」页面，展示广告投放账户的 mock 数据表格。

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/router/modules/index.ts` | 注册 `adAccountRoutes` 到 `routeModules` | 新路由模块引入 |
| `src/router/core/MenuProcessor.ts` | 添加 `AdAccount`/`AdAccountList` 的 MENU_I18N_MAP 和 MENU_ICON_MAP | 后端菜单渲染需要 i18n + icon 映射 |
| `src/locales/langs/zh.json` | 添加 `menus.adAccount.title` / `menus.adAccount.list` | 中文菜单显示 |
| `src/locales/langs/en.json` | 添加 `menus.adAccount.title` / `menus.adAccount.list` | 英文菜单显示 |

## Added

| File | Purpose |
|------|---------|
| `src/views/ad-account/index.vue` | 广告账户列表页面，含 8 条 mock 数据（Google/Facebook/TikTok/百度/Amazon 等平台） |
| `src/router/modules/adAccount.ts` | 路由模块：父菜单 `/ad-account` + 子页面 `list` |

## Database Changes

```sql
-- menus 表插入 2 条记录
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order)
VALUES ('AdAccount', 0, '/ad-account', '/index/index', 'ri:advertisement-line', '广告管理', 4);
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order)
VALUES ('AdAccountList', 12, 'list', '/ad-account/index', '', '账户列表', 1);

-- R_SUPER 角色 menu_ids 新增 {12, 13}
UPDATE roles SET menu_ids = '{1,3,2,4,5,6,7,12,13}' WHERE id = 1;
```

## 页面功能

- 表格列：序号、账户名称、账户ID、投放平台、状态（启用/暂停/禁用）、账户余额、日预算、创建时间、操作
- 操作按钮：编辑、充值、删除（当前均弹提示，待后续接入 API）
- 遵循项目规范：使用 `ArtTable` + `ArtTableHeader` + `useTable` + `ArtButtonTable`
- 按钮风格：顶部 `<ElButton v-ripple>` plain 按钮

## 验证结果

- ✅ `pnpm build` 通过（17.21s）
- ✅ `GET /api/v1/auth/menus` 返回包含 AdAccount(id=12) + AdAccountList(id=13)
- ✅ 完整菜单树：Dashboard(1) → 3, System(2) → 4,5,6,7, AdAccount(12) → 13

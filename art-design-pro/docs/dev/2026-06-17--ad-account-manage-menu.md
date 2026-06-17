# 广告账户管理菜单 — 2026-06-17

## 概述

在「广告管理」(AdAccount) 父菜单下新增「广告账户管理」(AdAccountManage) 子页面。

## 菜单结构

```
广告管理 (AdAccount, id=12)
├── FB账户列表 (AdAccountList, id=13) — 现有
└── 广告账户管理 (AdAccountManage, id=14) — 新增 ✨
```

## 修改清单

| 文件 | 变更 | 原因 |
|------|------|------|
| `src/views/ad-account/manage/index.vue` | 新建 | 广告账户管理列表页（ArtTable + 弹窗 CRUD，mock 数据） |
| `src/router/modules/adAccount.ts` | 添加子路由 `AdAccountManage` | 注册新页面的路由 |
| `src/locales/langs/zh.json` | 添加 `menus.adAccount.manage` 等 12 个 i18n 键 | 中文显示 |
| `src/locales/langs/en.json` | 添加对应英文翻译 | 国际化 |
| `src/router/core/MenuProcessor.ts` | 添加 `AdAccountManage` 到 `MENU_I18N_MAP` 和 `MENU_ICON_MAP` | 侧边栏正确显示 i18n 标题和图标 |
| `art-design-server/services/menu_service.go` | 添加 fallback 条目 (ID=18) | 后端兜底菜单数据 |
| `art-design-server/migrations/add_ad_account_manage_menu.sql` | 新建迁移文件 | 数据库 INSERT + 角色更新 |

## 数据库变更

```sql
-- 插入新子菜单 (父菜单 AdAccount id=12)
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order, hidden)
VALUES ('AdAccountManage', 12, 'manage', '/ad-account/manage/index', '', '广告账户管理', 2, false);

-- 更新超级管理员菜单权限
UPDATE roles SET menu_ids = '{1,3,2,4,5,6,7,12,13,14}' WHERE id = 1;
```

## 验证

- ✅ `GET /api/v1/auth/menus` 返回 AdAccountManage (id=14)
- ✅ 完整菜单树：AdAccount → [AdAccountList, AdAccountManage]
- ✅ `vue-tsc --noEmit` 无 ad-account 相关 TS 错误
- ✅ R_SUPER role menu_ids 包含 14
- ✅ 后端 fallback 包含 ID=18 条目

## 页面功能

- 广告账户列表（ArtTable + useTable）
- 平台标签（Facebook / Google Ads / TikTok）
- 状态标签（启用/禁用）
- 消耗金额显示
- 新增/编辑弹窗（ElDialog + ElForm）
- 删除确认（ElMessageBox）
- 当前使用 mock 数据，待后续接入真实 API

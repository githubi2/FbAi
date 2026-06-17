# BM管理 菜单 — 2026-06-17

## 概述
在「广告管理」子菜单下新增「BM管理」页面，用于展示所有已授权 Facebook 账号下的 Business Manager 列表。

## Modified

| File | Change | Reason |
|------|--------|--------|
| `src/router/modules/adAccount.ts` | 新增 `AdAccountBm` 子路由 | Pattern C: 在已有父菜单下添加子路由 |
| `src/router/core/MenuProcessor.ts` | 添加 `MENU_I18N_MAP.AdAccountBm` 和 `MENU_ICON_MAP.AdAccountBm` | 后端菜单标题中文化和图标映射 |
| `src/locales/langs/zh.json` | 添加 `menus.adAccount.bm`、`menus.adAccount.columns.bmId`、`menus.adAccount.bmNoData` | 国际化中文 |
| `src/locales/langs/en.json` | 添加 `menus.adAccount.bm`、`menus.adAccount.columns.bmId`、`menus.adAccount.bmNoData` | 国际化英文 |
| `art-design-server/services/menu_service.go` | `listFallback()` 新增 `AdAccountBm` (id=19) | 后端 fallback 菜单列表 |

## Added

| File | Purpose |
|------|---------|
| `src/views/ad-account/bm/index.vue` | BM管理页面 — 使用 ArtTable + useTable 标准模式，当前为 mock 数据 |
| `art-design-server/migrations/add_ad_account_bm_menu.sql` | 数据库迁移脚本 |

## Database Changes

```sql
-- Insert BM管理 sub-menu under AdAccount (id=12)
INSERT INTO menus (name, parent_id, path, component, icon, title, sort_order, hidden)
VALUES ('AdAccountBm', 12, 'bm', '/ad-account/bm/index', '', 'BM管理', 3, false);
-- → id=15

-- Update R_SUPER role menu_ids to include BM管理
UPDATE roles SET menu_ids = '{1,3,2,4,5,6,7,12,13,14,15}' WHERE id = 1;
```

## Menu Structure

```
AdAccount (id=12, 广告管理)
├── AdAccountList (id=13, FB账号列表)    sort=1
├── AdAccountManage (id=14, 广告账户管理)  sort=2
└── AdAccountBm (id=15, BM管理)           sort=3  ← NEW
```

## Verification

- ✅ `pnpm build` — 无新增 TypeScript 错误（已有错误来自未修改文件）
- ✅ `GET /api/v1/auth/menus` — AdAccountBm (id=15) 正确出现在 AdAccount 子菜单中
- ✅ Full menu tree: Dashboard → System → AdAccount → AdAccountList/AdAccountManage/AdAccountBm
- ✅ R_SUPER `menu_ids` 含 15
- ✅ 后端 `listFallback()` 已更新 (id=19)

## TODO

- [ ] 后端 API：实现 `GET /api/v1/fb/bm/list` 聚合所有 fb_tokens 的 bm_list 数据
- [ ] 前端：将 BM管理 页面从 mock 数据改为调用真实 API

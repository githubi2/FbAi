# BM管理菜单重构 — 2026-06-18

## 概述

将广告管理目录下的 **BM管理** 从普通菜单项改为目录菜单，内部新增 **BM列表** 子菜单，展示原BM管理页面内容。

## 修改前

```
📁 广告管理 (AdAccount, parent_id=12)
  📄 FB账户列表 (AdAccountList, id=13)
  📄 广告账户管理 (AdAccountManage, id=14)
  📄 BM管理 (AdAccountBm, id=15, menu_type=menu)  ← 直接菜单
```

## 修改后

```
📁 广告管理 (AdAccount, parent_id=12)
  📄 FB账户列表 (AdAccountList, id=13)
  📄 广告账户管理 (AdAccountManage, id=14)
  📁 BM管理 (AdAccountBm, id=15, menu_type=directory)  ← 改为目录
    📄 BM列表 (AdAccountBmList, id=16, menu_type=menu)  ← 新增子菜单
```

## Modified

| File | Change | Reason |
|------|--------|--------|
| DB `menus` 表 | id=15: `menu_type` → `directory`, `component` → `/index/index`; 新增 id=16 | 菜单结构调整 |
| `src/router/modules/adAccount.ts` | `bm` 从 flat menu 改为 directory + children | 前端路由匹配新结构 |
| `src/router/core/MenuProcessor.ts` | MENU_I18N_MAP + MENU_ICON_MAP 新增 AdAccountBmList | 后端菜单标题/图标映射 |
| `src/locales/langs/zh.json` | 新增 `menus.adAccount.bmList: "BM列表"` | 中文菜单标题 |
| `src/locales/langs/en.json` | 新增 `menus.adAccount.bmList: "BM List"` | 英文菜单标题 |
| `art-design-server/services/menu_service.go` | listFallback: id=19 改为 directory, 新增 id=20 (BM列表) | 后端降级菜单数据 |
| `art-design-server/seed.sql` | R_SUPER menu_ids 增加 15,16 | 种子数据同步 |
| DB `roles` 表 | R_SUPER (id=1) menu_ids 增加 16 | 超级管理员可见新菜单 |

## Why

用户需求：BM管理应该是目录，内部包含BM列表页。原BM管理页面内容（BM列表展示）移到新的BM列表子菜单中，保持功能不变。

## 验证

- [x] API `/api/v1/auth/menus` 返回正确的树结构（BM管理=目录，BM列表=子菜单）
- [x] 无孤儿菜单（所有子菜单父菜单均在 menu_ids 中）
- [x] ESLint 通过（前端改动文件）
- [ ] 前端 UI 验证（用户浏览器测试）

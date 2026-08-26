# 菜单扁平化：广告/BM/主页/像素 二级菜单提升为一级 — 2026-08-26

## 需求

侧边栏菜单扁平化：移除"广告管理/BM管理/主页管理/像素管理"4 个一级目录，
其下 5 个子页面提升为一级菜单（点击直达页面，无下拉箭头）。仪表盘、系统管理保持不变。

## 最终菜单结构

```
仪表盘（目录，含工作台）
系统管理（目录，含用户/角色/菜单/租户）
FB账户列表   → /ad-account        → views/ad-account/index
广告账户管理 → /ad-account-manage → views/ad-account/manage/index
BM列表       → /bm-manage         → views/ad-account/bm/index
主页列表     → /page-manage       → views/page-manage/index
像素列表     → /pixel-manage      → views/pixel-manage/index
```

## Modified

| 文件 | 变更 |
|------|------|
| `art-design-server/services/menu_service.go` | `listFallback()` 同步：广告管理/BM管理目录删除，AdAccountList/AdAccountManage/AdAccountBmList 提升为一级（DB 不可用时兜底一致） |
| `art-design-server/scripts/menu_flatten.sql` | 新增：DB 迁移脚本（事务） |

## Database 变更（menu_flatten.sql 已执行）

1. `UPDATE menus` × 5：id 13/14/16/18/20 → `parent_id=0` + 独立一级 path + icon + sort_order
2. `DELETE FROM menus` × 4：id 12（广告管理）/15（BM管理）/17（主页管理）/19（像素管理）
3. `UPDATE roles` × 4：R_SUPER 等角色 `menu_ids` 移除 12/15/17/19（子级 13/14/16/18/20 保留）

迁移后 R_SUPER menu_ids = `{1,3,2,4,5,6,7,13,14,16,18,20}`

## 关键技术决策

- **一级菜单直接是页面，前端 0 代码修改**：
  - `RouteTransformer.isFirstLevelRoute()`：一级路由无 children → 自动用 Layout 包裹 + 页面组件作为唯一 child（已确认源码逻辑）
  - `SidebarSubmenu`：无 children 的一级菜单渲染为 `ElMenuItem`（可点击，无箭头）
  - `MENU_I18N_MAP` / `MENU_ICON_MAP` 已含 AdAccountList/AdAccountManage/AdAccountBmList/PageManageList/PixelManageList 映射
- **path 唯一性**：`extractFirstSegment()` 一级路由只取 path 第一段，故广告账户管理用单段 `/ad-account-manage`（`/ad-account/manage` 会与 `/ad-account` 冲突）；其余复用原顶级路径（/bm-manage、/page-manage、/pixel-manage）

## 验证（浏览器实测）

1. 后端菜单树 API：5 个一级页面菜单 child=0 ✅
2. admin 登录 → 侧边栏 7 个顶级：仪表盘/系统管理/FB账号列表/广告账户管理/BM列表/主页列表/像素列表 ✅
3. 逐个点击：FB账号列表（表格数据正常）✅ 广告账户管理 ✅ BM列表 ✅ 主页列表 ✅ 像素列表 ✅
4. 菜单管理页树状结构正确 ✅

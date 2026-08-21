# 主页管理（FB 公共主页列表）

日期：2026-08-21

## 功能

新增一级菜单「主页管理」→「主页列表」，展示所有已授权 FB 账号下的公共主页（Facebook Pages），支持搜索/筛选、后台异步刷新、备注编辑。

## 表格列

按需求截图实现：主页名称、推送状态、备注、主页ID、创建时间、创建渠道、主页状态、申诉时间、发布状态、允许评论、隐藏不文明用语、主页认证、广告权限、BM、管理员、屏蔽词设置、黑名单列表、地址、电话、邮箱、网址、主页分类、主页类型、点赞数、粉丝数、操作。（无序号、收藏列）

**FB 官方 API 未提供的字段**显示 `—`：创建时间、创建渠道、主页状态、申诉时间、允许评论、隐藏不文明用语、广告权限、BM、屏蔽词设置、黑名单列表、主页类型。

实际对接的 FB 字段（`/me/accounts`）：name、link、category、fan_count、followers_count、is_published、verification_status、website、phone、emails、location；管理员名单走 `/{page-id}/roles` 边（失败容忍）。

## 后端

- 迁移：`migrations/011_fb_pages_cache.sql`（fb_pages_cache 表，remark/push_status 为本地字段，刷新不覆盖）
- 菜单：`migrations/add_page_manage_menu.sql`（PageManage 顶级目录 + PageManageList 子菜单，R_SUPER menu_ids 追加）
- 接口：
  - `GET /api/v1/fb/pages` — 缓存直出
  - `POST /api/v1/fb/pages/refresh-all` — 后台异步刷新（5 分钟冷却，幂等），refresh_type = `pages`
  - `PUT /api/v1/fb/pages/:id/remark` — 更新本地备注
- 代码：`models/fb.go`（FbPageItem/FbPageListResponse）、`services/fb_service.go`（GetPageList）、`services/fb_cache_service.go`（GetCachedPages/SavePagesCache/UpdatePageRemark）、`handlers/fb_handler.go`、`routes/router.go`
- 工具：`scripts/runsql`（用 .env 的 DATABASE_URL 手动执行迁移 SQL）

## 前端

- 页面：`src/views/page-manage/index.vue`（ArtTable + ArtTableHeader + useTable，客户端分页+筛选，刷新状态轮询）
- 列配置：`src/views/page-manage/columns.ts`（列工厂拆出，保持单文件 <300 行）
- API：`src/api/facebook.ts`（fetchFbPages / fetchFbRefreshPages / fetchFbUpdatePageRemark；fetchRefreshStatus 类型加 `pages`）
- 菜单映射：`src/router/core/MenuProcessor.ts`（MENU_I18N_MAP / MENU_ICON_MAP 加 PageManage / PageManageList）
- i18n：`menus.pageManage.*`（zh/en）

## 同会话顺带修复

1. 增加授权弹窗：检测按钮等结果返回后再跳结果页；结果改为一个账号一个方块的卡片（状态/失败账号/当前账号/返回信息）；检测完成后按钮变「重新检测」
2. 广告授权接口：FB /users 边 role 数字映射（ADMIN→1001/ADVERTISER→1002/ANALYST→1003）；BM 名下账户自动走 assigned_users 边（business 参数 + tasks 数组）；用户名主页地址自动解析为数字 UID

## 2026-08-21 追加：主页扩展字段实测与接入

用真实 token 逐个实测 Graph API 后的结论：

| 列 | 接口 | 结果 |
|---|---|---|
| 管理员 | `/{page}/roles`（必须**主页访问口令**，用户口令报 190/2069032） | ✅ 取 tasks 含 MANAGE 的用户 |
| 黑名单列表 | `/{page}/blocked`（主页口令） | ✅ 显示数量 |
| 隐藏不文明用语 | `/{page}/settings` 的 `PROFANITY_FILTER`（none/medium/strong） | ✅ |
| 广告权限 | `/me/accounts` 的 `tasks` 含 `ADVERTISE` | ✅ 1=正常 0=无权限 -1=未知 |
| BM | Page 的 `business` 字段 | ✅（未绑定 BM 显示 —） |
| 创建时间 | Page 无 `created_time` 字段（#100 报错实测） | ❌ 显示 — |
| 屏蔽词设置 | 无接口（`/moderation_keywords` 报 Unknown path） | ❌ 显示 — |
| 创建渠道/主页状态/申诉时间/允许评论/主页类型 | Graph API 无对应字段 | ❌ 显示 — |

变更：迁移 `012_fb_pages_extra_fields.sql`（bm_name/ad_perm/profanity_filter/blocked_count），
`GetPageList` 改用 /me/accounts 返回的主页访问口令调 roles/blocked/settings。

# 广告投放监控（只读）— 2026-08-27

## 需求

新增"广告投放"一级菜单（位置：广告账户管理之后），按广告账户筛选，只读监控 FB 广告投放三层结构（广告系列/广告组/广告），含近 7 天数据统计。不创建广告（后续单独开发）。

## 后端（Go + Gin，v26.0）

| 文件 | 变更 |
|------|------|
| `models/fb.go` | 新增 FbInsight/FbCampaign/FbAdSet/FbAd + 列表响应结构 |
| `services/fb_service.go` | 新增 `resolveTokenByAccountID`（按广告账户解析用户 token，校验归属）+ `GetCampaigns`（含 insights 合并）+ `GetAdSets` + `GetAds` |
| `handlers/fb_handler.go` | 新增 CampaignList/AdSetList/AdList 3 个 handler（userID+tenantID 鉴权，缺参 400） |
| `routes/router.go` | 注册 3 条 GET 路由 |
| `services/menu_service.go` | fallback 添加 AdCampaign 一级菜单 |
| `scripts/ad_campaign_menu.sql` | DB 迁移：插入广告投放菜单 + R_SUPER menu_ids + 后续菜单顺延 |

### API

| 端点 | FB 调用 |
|------|---------|
| `GET /api/v1/fb/campaigns?accountId=act_xxx` | `/{act}/campaigns` (12 字段) + `/{act}/insights?date_preset=last_7d&level=campaign`（每账户 1 次合并，避免 N+1） |
| `GET /api/v1/fb/campaigns/:id/adsets?accountId=` | `/{campaign_id}/adsets` |
| `GET /api/v1/fb/adsets/:id/ads?accountId=` | `/{adset_id}/ads`（含 creative{id,name}） |

Token 解析：`fb_ad_accounts_cache.ad_account_id` → `fb_token_id` → `fb_tokens.access_token`，强制 user_id/tenant_id 匹配（账户不属于用户 → 400 错误）。

## 前端

| 文件 | 变更 |
|------|------|
| `src/views/ad-campaign/index.vue` | 三标签页（系列/组/广告）+ 账户下拉筛选 + 3×useTable 客户端分页 |
| `src/views/ad-campaign/columns.ts` | 3 组列配置（状态彩标/预算/统计列）+ 操作列下钻 |
| `src/api/facebook.ts` | fetchFbCampaigns/FbAdSets/FbAds + 类型 |
| `src/router/core/MenuProcessor.ts` | AdCampaign → i18n + icon（ri:megaphone-line） |
| `src/locales/langs/zh.json`/`en.json` | menus.adCampaign.* |

## 踩坑记录（重要）

1. **TDZ 崩溃**：`columnsFactory`（经 `useTableColumns`）在 setup 时**立即执行**，直接引用后定义的 `onViewAdSets` 抛 `Cannot access before initialization` → 改函数声明（提升），运行时才解析依赖。症状：白屏 + [VueError]。`isAccountDisabled` 同理（必须 function 声明）。
2. **`replaceSearchParams` 不触发请求**：useTable 无 watch(searchParams)，替换参数后**必须手动调用返回的 `getData()`**（即 getDataByPage，重置第 1 页）。旧 bm/index.vue 的 handleSearch 只调 replaceSearchParams（未调 getData），其搜索可能同样无效——后续可修。
3. **insights 空数据**：FB `/insights` 对部分账户返回空（数据源侧，非代码问题）→ 前端显示 —，符合设计。
4. **"账户已停用"状态不来自 API**（2026-08-27 用户纠错）：FB API 的 campaign/adset/ad `status`/`effective_status` 对停用账户仍返回 ACTIVE；FB 后台"账户已停用"是其 UI 按**广告账户状态**计算的。修复：前端用所选账户的 `accountStatus != 1 || disableReason > 0` 推断（来自 ad-accounts/detail），状态列统一显示"账户已停用"（danger）。
5. **直接切标签页无上下文 → 数据空**（2026-08-27 用户反馈）：`watch(activeTab)` 只在 activeCampaignId 已存在时加载，用户直接点"广告组/广告"标签时空白。修复：切 tab 自动取**第一行**作为下钻上下文（campaign→adset→ad 逐级），无需先点行内按钮。
6. **广告组/广告数据不全——只按单个 campaign/adset 查询**（2026-08-27 用户纠错，第二轮）：原实现先查 campaign 再查其 adsets（只返回第一个系列的数据，广告漏掉另一半）。修复：改为**账户级聚合**——`GET /{act}/adsets`（一次取全部广告组，含 `campaign{id,name}` 归属）与 `GET /{act}/ads`（一次取全部广告，含 `campaign{id,name},adset{id,name}` 归属），与 FB 后台"广告组/广告"全量视图一致。新增 `GET /api/v1/fb/adsets?accountId=` / `GET /api/v1/fb/ads?accountId=` 端点（旧 campaign/adset 级端点保留兼容）。前端广告组/广告 tab 已改为账户级拉取，并新增"所属系列/所属广告组"列。
7. **多账户全量 + 空态修复**（2026-08-27 用户反馈，第三轮）：
   - **默认全量**：进入页面即聚合**全部授权广告账户**数据（`accountIds` 缺省=全部，`resolveAccounts` 从 fb_ad_accounts_cache 取全部账户凭据）；筛选支持**多选**（`accountIds=act_x,act_y`，前端 ElSelect multiple + clearable，placeholder"不选=全部账户，可多选"）。
   - **账户列**：三表首列新增"账户"（账户名 + BM 名），按行判定停用状态（`isAccountDisabled(row)` 用 row.accountId 查账户状态，不再全局）。
   - **空态单图标**：ArtTable 自带 `#empty` 插槽（"暂无数据"），页面手动加的 `ElEmpty` 导致双空态——移除 ad-campaign（3 tab）/bm/page-manage/manage 支付弹窗的手动 ElEmpty，改用 ArtTable `:empty-text`（新增 noPaymentRecords 等 i18n 键）。
   - **限速雪崩修复**：多账户全量时每账户 2 次 FB 调用，4s 限速 + 前端 15s 超时 → 请求积压雪崩（单请求 70s+）。修复：`FB_RATE_LIMIT_MS=1500`（.env，限速器原生支持）、axios 3 个聚合端点 `timeout: 180000`、adset/ad 表格 `immediate: false`（切 tab 才加载，避免 3 端点并发排队）。
   
   实测：全量 campaigns 16.6s 返回（3 账户 ×（campaigns+insights）× 1.5s + FB 往返）；浏览器全流程验证通过（默认全量 2 条 + 账户列 + 多选筛选空态单图标 + 重置回全量）。

### 官方文档状态依据（2026-08-27 实抓 v26.0 广告组字段表）

- `effective_status` enum：`{ACTIVE, PAUSED, DELETED, CAMPAIGN_PAUSED, ARCHIVED, IN_PROCESS, WITH_ISSUES}` —— **不含 ACCOUNT_DISABLED**
- `status`/`configured_status` enum：`{ACTIVE, PAUSED, DELETED, ARCHIVED}` —— 实体自身设置状态
- **结论**：campaign/adset/ad 的状态枚举无法反映"账户停用"（实测停用账户下仍返回 ACTIVE）；"账户已停用"是 FB 后台 UI 按 **Ad Account 的 account_status**（1=活跃/2=禁用…）计算。本项目用 `accountStatus != 1 || disableReason > 0` 推断，符合官方口径。

## 验证

- `go build` ✅ + 真实 API 实测：campaigns 2 条（肽-减肥-通投-18+ 系列）、adsets 1 条、ads 4 条（含创意）✅
- `vue-tsc --noEmit` 无类型错误 ✅
- 浏览器实测全链路：菜单→选账户→系列→广告组→广告 三级下钻 ✅（截图确认）
- 修复后实测（2026-08-27）：状态列"账户已停用"红标 ✅；直接切"广告组/广告"标签自动加载数据 ✅

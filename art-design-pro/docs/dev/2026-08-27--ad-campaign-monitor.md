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

## 验证

- `go build` ✅ + 真实 API 实测：campaigns 2 条（肽-减肥-通投-18+ 系列）、adsets 1 条、ads 4 条（含创意）✅
- `vue-tsc --noEmit` 无类型错误 ✅
- 浏览器实测全链路：菜单→选账户→系列→广告组→广告 三级下钻 ✅（截图确认）
- 修复后实测（2026-08-27）：状态列"账户已停用"红标 ✅；直接切"广告组/广告"标签自动加载数据 ✅

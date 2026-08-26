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

1. **TDZ 崩溃**：`columnsFactory`（经 `useTableColumns`）在 setup 时**立即执行**，直接引用后定义的 `onViewAdSets` 抛 `Cannot access before initialization` → 改函数声明（提升），运行时才解析依赖。症状：白屏 + [VueError]。
2. **`replaceSearchParams` 不触发请求**：useTable 无 watch(searchParams)，替换参数后**必须手动调用返回的 `getData()`**（即 getDataByPage，重置第 1 页）。旧 bm/index.vue 的 handleSearch 只调 replaceSearchParams（未调 getData），其搜索可能同样无效——后续可修。
3. **insights 空数据**：FB `/insights` 对部分账户返回空（数据源侧，非代码问题）→ 前端显示 —，符合设计。

## 验证

- `go build` ✅ + 真实 API 实测：campaigns 2 条（肽-减肥-通投-18+ 系列）、adsets 1 条、ads 4 条（含创意）✅
- `vue-tsc --noEmit` 无类型错误 ✅
- 浏览器实测全链路：菜单→选账户→系列→广告组→广告 三级下钻 ✅（截图确认）

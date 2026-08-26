# FB Graph API 版本迁移 v22.0 → v26.0 — 2026-08-26

## Modified

| 文件 | 变更 | 原因 |
|------|------|------|
| `art-design-server/.env` | `FB_GRAPH_VERSION=v22.0` → `v26.0`（gitignored，不进仓库） | 升级 Graph API 版本到最新 |

## Why

1. **业务规划**：后续要基于 Graph API 开发广告投放业务（campaign/adset/creative），需保持在 Meta 最新 API 版本上，避免旧版本行为差异与缺乏文档支持。
2. **成本极低**：`services/fb_service.go` 第 55-58 行已支持 `FB_GRAPH_VERSION` 环境变量（默认 v22.0），无需改代码，改一个环境变量即可。
3. **版本生命周期**：v22.0 于 2025-01-21 发布，2027-05-20 停用；v26.0 为当前最新（2026-07-29 发布，停用待定）。

## 迁移前验证（2026-08-26，真实 token 实测）

| 端点/字段 | v22.0 | v26.0 | 结论 |
|-----------|-------|-------|------|
| `/me/adaccounts`（13 字段） | ✅ | ✅ 逐字一致 | 兼容 |
| `/{act_id}`：funding_source_details / disable_reason / is_personal / business_country_code | ✅ | ✅ 一致 | 兼容 |
| `/me/businesses`（id,name,created_time,verification_status,permitted_roles） | ✅ | ✅ 一致 | 兼容 |
| `daily_spend_limit` / `next_bill_date` | ❌ #100 | ❌ #100 | 两版本均不可用（已知） |
| Business 对象状态字段（status/blocked/is_disabled/…21+ 候选） | ❌ 不存在 | ❌ 不存在 | BM 无状态字段确认 |
| OAuth authorize / access_token | — | 正常 | 未观测到变更 |

## Changelog 审查（v23-v26 官方变更均与项目无关）

- v23.0：Advantage+ 系列、Targeting、Bidding、Special Ad Categories（广告投放创建类）
- v24.0：Ad Creative、Messenger lead ads、Advantage+ 废弃、Custom Audiences（创建类）
- v25.0：Page/Post/Video/Stories Insights 指标废弃、metadata 参数废弃、Webhooks mTLS
- v26.0：Commerce Order API 废弃、Rights Manager 字段迁移、legacy protocol（pretty/debug）废弃

项目仅使用：OAuth 登录 + 读取广告账户列表/BM 列表/详情，全部字段在两版本一致，无受影响项。

## 迁移步骤（已执行）

1. `python` 原地替换 `.env` 第 16 行 → `FB_GRAPH_VERSION=v26.0`（未触碰密码行）
2. 杀旧 server.exe → 确认 9090 端口释放 → 重启 `./server.exe`（环境变量即读即用，无需重新编译）
3. 回归验证：
   - `GET /api/v1/ping` → pong ✅
   - `POST /api/v1/auth/login` → token ✅
   - `POST /api/v1/fb/bm-list/refresh` → started ✅
   - 后端日志确认实际请求：`[FbRateLimiter] "/v26.0/2264541517695037/business_users"` 等 ✅
   - `GET /api/v1/fb/refresh-status` → isRunning=false, error="" ✅
   - `GET /api/v1/fb/bm-list` → 数据完整（admins=3, adAccounts=3）✅

## 注意事项

- `.env` 在 `.gitignore` 中，本变更不进仓库；需在部署环境同步设置 `FB_GRAPH_VERSION=v26.0`。
- 后续广告业务开发应直接基于 v26.0 编写，并用 `fields` 逐个验证字段可用性（参考 fbai-standards `fb-api-field-debugging.md`）。
- 已知待办（未包含在本迁移）：BM 状态列封禁检测（owned_ad_accounts 全禁用推断），另行处理。

# FB 账号列表 — 数据拉取失败状态显示 — 2025-06-17

## 需求
FB 账号列表页，如果某条数据的拉取（FB API 调用）失败，将状态改为"异常"并显示失败原因。

## Modified

### 后端

| File | Change | Reason |
|------|--------|--------|
| `migrations/006_fb_last_error.sql` | 新增 `last_error TEXT` + `last_error_at TIMESTAMPTZ` 列 | 持久化记录每次 FB API 调用的失败信息 |
| `models/fb.go` — `FbToken` | 添加 `LastError` + `LastErrorAt` 字段 | Go 结构体映射新列 |
| `models/fb.go` — `FbAccountListItem` | 添加 `DataError string` 字段；`AccountStatus` 支持 "异常" | 前端列表项携带错误信息 |
| `services/fb_service.go` — `ListAccounts` | SQL 查询增加 `last_error`；状态判断优先检查 `lastError` | 有错误记录时状态为"异常" |
| `services/fb_service.go` — `GetAdAccountsDetail` | FB API 调用失败时 `setLastError()`；成功时 `clearLastError()` | 自动记录/清除每个 token 的错误 |
| `services/fb_service.go` — `RefreshAccountStats` | 失败时 `setLastError()`；成功时 `clearLastError()` | 手动刷新时记录/清除错误 |
| `services/fb_service.go` | 新增 `setLastError()` + `clearLastError()` 辅助方法 | 统一错误记录逻辑 |

### 前端

| File | Change | Reason |
|------|--------|--------|
| `src/api/facebook.ts` — `FbAccount` | 添加 `dataError: string` | 接收后端返回的错误信息 |
| `src/views/ad-account/index.vue` | 状态列增加"异常"分支（`warning` 标签+`ElTooltip` 显示错误） | 用户可悬停查看失败原因 |
| `src/locales/langs/zh.json` | 添加 `status.error: "异常"` | i18n 中文 |
| `src/locales/langs/en.json` | 添加 `status.error: "Error"` | i18n 英文 |

## 状态优先级
`异常` > `已过期` > `正常`

- 有 `last_error` 记录 → "异常"（warning 标签，悬停显示错误详情）
- token 已过期 → "已过期"（danger 标签）
- 其他 → "正常"（success 标签）

## 错误记录时机
1. **广告账户详情聚合**（`GET /api/v1/fb/ad-accounts/detail`）：遍历各 token 调 FB API，单个失败即记录 `last_error`
2. **手动刷新统计**（`POST /api/v1/fb/accounts/:id/refresh`）：调用 FB API 失败时记录；成功时清除
3. **账号列表查询**（`GET /api/v1/fb/accounts`）：仅读取 `last_error` 展示，不主动调 FB API（避免拖慢页面）

## Why
之前账号列表只能显示"正常"或"已过期"（基于 token 到期时间），无法感知 token 已被吊销、权限不足等运行时错误。现在任何 FB API 调用失败都会记录错误，前端列表页直接展示"异常"状态和具体原因。

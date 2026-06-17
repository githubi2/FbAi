# FB 多账号授权改造 — 2026-06-17

## 需求
- 所有用户可授权多个 FB 账号（不再限制一个用户只能授权一个）
- FB 账号列表改为显示已授权的 FB 账号（而非其下的广告账户）
- 保持多租户数据隔离

## 数据库变更

| 变更 | SQL |
|------|-----|
| 删除旧唯一约束 | `DROP INDEX idx_fb_tokens_user_id` |
| 新增部分唯一约束 | `CREATE UNIQUE INDEX idx_fb_tokens_user_fb_active ON fb_tokens(user_id, fb_user_id) WHERE status = 1` |
| 新增备注字段 | `ALTER TABLE fb_tokens ADD COLUMN label VARCHAR(64) DEFAULT ''` |

迁移文件：`art-design-server/migrations/004_fb_multi_account.sql`

### 唯一约束说明
- `(user_id, fb_user_id) WHERE status=1`：同一用户授权同一 FB 账号 → DO UPDATE（刷新 token）
- 同一用户授权不同 FB 账号 → INSERT 新记录
- Pending 记录（status=0）不受约束限制

## 后端变更

### 修改文件

| 文件 | 改动 |
|------|------|
| `models/fb.go` | ① `FbToken` 新增 `Label` 字段 ② 新增 `FbAccountListItem`、`FbAccountListResponse`、`FbUpdateLabelRequest` 类型 |
| `services/fb_service.go` | ① `SaveToken`：改为 `ON CONFLICT (user_id, fb_user_id) WHERE status=1` ② `Disconnect`：改为按主键 ID 断开 ③ 新增 `DisconnectAll`、`GetTokenByID`、`ListAccounts`、`UpdateLabel`、`RefreshAccountStats` ④ `GetToken`：SELECT 增加 `label` 字段 ⑤ `GetAuthURL`：pending 记录直接 INSERT（不用 ON CONFLICT） |
| `handlers/fb_handler.go` | ① `Disconnect`：支持路径参数 `:id` ② 新增 `ListAccounts`、`UpdateLabel`、`RefreshStats` ③ 新增 `parseUint` 辅助函数 ④ import 增加 `strconv` |
| `routes/router.go` | 新增路由组 `fb/accounts`：`GET ""`, `DELETE "/:id"`, `PUT "/:id/label"`, `POST "/:id/refresh"` |

### 新增 API

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/fb/accounts` | FB 账号列表（含 BM/广告账户统计） |
| `DELETE` | `/api/v1/fb/accounts/:id` | 断开指定 FB 账号 |
| `PUT` | `/api/v1/fb/accounts/:id/label` | 更新备注 |
| `POST` | `/api/v1/fb/accounts/:id/refresh` | 刷新 BM 和广告账户缓存 |

### 列表返回字段

```json
{
  "id": 1,
  "fbUserId": "122196014900875069",
  "fbUserName": "Fshsh Ushsh",
  "label": "主账号",
  "scopes": ["ads_read", "ads_management", "business_management"],
  "expiresAt": "2026-08-16T...",
  "createdAt": "2026-06-17T...",
  "daysUntilExpiry": 60,
  "hasAdPerm": true,
  "accountStatus": "正常",
  "bmCount": 0,
  "personalAdCount": 1,
  "bmAdCount": 0
}
```

## 前端变更

### 修改文件

| 文件 | 改动 |
|------|------|
| `api/facebook.ts` | 新增 `FbAccount`、`FbAccountListResponse` 类型；新增 `fetchFbAccountList`、`fetchFbDisconnectAccount`、`fetchFbUpdateLabel`、`fetchFbRefreshStats` 函数 |
| `views/ad-account/index.vue` | **完全重写**：从"连接状态面板 + 广告账户表格"改为"FB 账号列表 + 连接按钮" |
| `locales/langs/zh.json` | 新增 18 个 i18n key（columns、status、adPerm、label 相关） |
| `locales/langs/en.json` | 同上（英文对应） |

### 表格列

| 列 | prop | 格式 |
|----|------|------|
| 序号 | type: 'index' | — |
| 账户名称 | accountName | `label ? "label (fbUserName)" : fbUserName` |
| 账户 ID | fbUserId | 纯文本 |
| 状态 | accountStatus | 绿色 Tag"正常" / 红色 Tag"已过期" |
| 广告权限 | hasAdPerm | 绿色 Tag"已授权" / 灰色 Tag"无权限" |
| BM | bmCount | 数字 |
| 广告户 | adAccounts | `BM: 2，个人: 1` |
| 授权有效期 | validity | 按天显示，>30天绿 / 7-30天橙 / <7天红 / 已过期红 Tag |
| 授权时间 | createdAt | `YYYY-MM-DD HH:mm` |
| 操作 | actions | 编辑备注 / 刷新统计 / 断开连接（ArtButtonTable） |

### 操作按钮

| 按钮 | 图标 | 功能 |
|------|------|------|
| 编辑备注 | ri:pencil-line | 弹窗编辑 label（最长64字符，实时字数统计） |
| 刷新统计 | ri:refresh-line | 调用 Graph API 更新 BM 和广告账户缓存 |
| 断开连接 | ri:delete-bin-5-line | 二次确认后软删除（status→0） |

### 连接流程变化

```
旧：授权 → 覆盖旧记录 → 始终只有一条
新：授权 → 同一FB账号刷新token / 不同FB账号新增记录
轮询：检查 accounts 总数是否增加（而非检查 connected 状态）
```

## 数据隔离

沿用现有多租户模式，所有查询使用 `tenant_id IS NOT DISTINCT FROM` 过滤：
- `user_id` + `tenant_id` 组合确定数据归属
- 每个租户的用户只能查看本租户内的 FB 账号
- 超级管理员（tenant_id=NULL）可查看所有

## 向后兼容

| 旧 API | 状态 |
|--------|------|
| `GET /api/v1/fb/status` | ✅ 保留（返回第一个有效连接） |
| `GET /api/v1/fb/ad-accounts` | ✅ 保留（返回第一个有效连接的广告账户） |
| `DELETE /api/v1/fb/disconnect` | ✅ 兼容（不传 id 时断开所有） |
| `GET /api/v1/fb/auth-url` | ✅ 不变 |
| `POST /api/v1/fb/callback` | ✅ 不变（回调自然走新 SaveToken 逻辑） |

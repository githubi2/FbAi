# 单点登录 + 记住密码 + 管理员改密码踢人 — 2026-06-16

## 功能概述

1. **记住密码**：登录时勾选"记住密码"，token 有效期 3 天；不勾选则 24 小时。
2. **单点登录 (SSO)**：同一账号在新设备/浏览器登录时，旧会话自动失效。
3. **管理员改密码踢人**：超级管理员修改其他用户密码后，目标用户所有登录会话立即失效，强制退出。

## 修改文件

### 数据库

| 操作 | 对象 | 说明 |
|------|------|------|
| CREATE TABLE | `sessions` | 新建会话表：id, user_id, token, refresh_token, expires_at, created_at |
| CREATE INDEX | `idx_sessions_token` | token 查询索引 |
| CREATE INDEX | `idx_sessions_user_id` | 用户会话查询索引 |
| CREATE INDEX | `idx_sessions_expires_at` | 过期清理索引 |

### 后端 (Go/Gin)

| 文件 | 操作 | 说明 |
|------|------|------|
| `models/session.go` | **新增** | Session 结构体 |
| `models/user.go` | 修改 | LoginRequest 新增 `rememberMe` 字段 |
| `services/session_service.go` | **新增** | SessionService：Create/Validate/InvalidateUser/CleanExpired |
| `middleware/auth.go` | **重写** | 移除内存 `tokenStore`，改用 DB sessions 表；GenerateToken 接受 rememberMe；新增 InvalidateUserSessions |
| `handlers/auth_handler.go` | 修改 | Login 传递 `req.RememberMe` 到 GenerateToken；新增 fmt/time 导入 |
| `handlers/user_handler.go` | 修改 | Update 时如果修改密码，调用 InvalidateUserSessions |

### 前端 (Vue 3)

| 文件 | 操作 | 说明 |
|------|------|------|
| `src/types/api/api.d.ts` | 修改 | LoginParams 新增 `rememberMe: boolean` |
| `src/store/modules/user.ts` | 修改 | 新增 `rememberMe`、`tokenExpiresAt` 状态；`setToken` 接受 isRememberMe；新增 `checkTokenExpiry()` 方法 |
| `src/views/auth/login/index.vue` | 修改 | 发送 `rememberMe: rememberPassword` 到 API |
| `src/router/guards/beforeEach.ts` | 修改 | `handleLoginStatus` 中调用 `checkTokenExpiry()` 主动检查过期 |

## 设计决策

### 为什么不用 JWT？
当前阶段使用简单的 token 字符串存入 DB sessions 表。这样：
- 服务端可以主动失效会话（SSO、管理员踢人）
- 不需要 JWT 的 refresh token 轮换逻辑
- 后续可平滑升级为 JWT + sessions 表混合方案

### 为什么在 SessionService.Create 中实现 SSO？
`Create` 方法在插入新会话前先 `DELETE FROM sessions WHERE user_id = $1`，保证每个用户只有最新一个 session 有效。这是最简单的单点登录实现。

### Token 过期机制
- **后端主动**：`Validate` 查询时检查 `expires_at`，过期则删除行返回 false → 401
- **前端被动**：HTTP 拦截器捕获 401 → 自动 `logOut()`
- **前端主动**：路由守卫每次跳转调用 `checkTokenExpiry()`，基于本地 `tokenExpiresAt` 提前判断

### 为什么在 handler 层失效 session 而不是 service 层？
`user_service.Update` 是通用方法（创建、更新都复用它），而"改密码踢人"是 handler 层的业务逻辑。保持 service 层纯净，业务逻辑在 handler 层处理。

## 验证结果

| 测试场景 | 结果 |
|----------|------|
| 登录 (rememberMe=true) → token 有效 | ✅ |
| 登录 (rememberMe=false) → token 有效 | ✅ |
| SSO 重新登录 → 旧 token 失效 | ✅ |
| 管理员改密码 → 目标用户 token 失效 | ✅ |
| 新密码登录 → 成功 | ✅ |
| 前端类型检查 `vue-tsc --noEmit` | ✅ |
| 后端编译 `go build` | ✅ |

## 数据库 DDL

```sql
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(512) NOT NULL UNIQUE,
    refresh_token VARCHAR(512),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

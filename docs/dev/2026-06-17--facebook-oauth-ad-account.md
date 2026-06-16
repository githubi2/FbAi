# Facebook OAuth 授权 + 广告账户管理 — 2026-06-17

## 概述

实现 Facebook OAuth 2.0 服务端授权流程，后台通过 Facebook Marketing API 获取并管理广告账户和 Business Manager。

## 流程

```
后台生成授权链接 → 用户在浏览器中打开链接 → Facebook OAuth 授权
    → 回调到后端 /api/v1/fb/callback → 自动换取 access_token
    → 加密存入数据库 → 前端通过后端代理调用 Facebook API
```

## 新增文件

| 文件 | 用途 |
|------|------|
| `art-design-server/models/fb.go` | Facebook 数据模型（FbToken, FbAdAccount, FbBusinessManager） |
| `art-design-server/services/fb_service.go` | OAuth 流程 + Facebook Graph API 调用服务 |
| `art-design-server/handlers/fb_handler.go` | HTTP 处理器（auth-url, callback, status, ad-accounts, disconnect） |
| `art-design-server/migrations/add_fb_tokens.sql` | 数据库迁移（fb_tokens 表） |
| `art-design-pro/src/api/facebook.ts` | 前端 API 层（类型定义 + 请求函数） |

## 修改文件

| 文件 | 变更 | 原因 |
|------|------|------|
| `art-design-server/routes/router.go` | 添加 `/api/v1/fb/*` 路由组 + `/api/v1/fb/callback` 回调 | 注册 Facebook API 端点 |
| `art-design-server/.env.example` | 添加 FB_APP_ID, FB_APP_SECRET, FB_REDIRECT_URI, FB_GRAPH_VERSION | Facebook 应用配置 |
| `art-design-pro/src/views/ad-account/index.vue` | 替换 mock 数据为真实 Facebook API 集成，添加连接状态面板和 OAuth 流程 | Facebook 授权 + 真实数据 |
| `art-design-pro/src/locales/langs/zh.json` | 添加 adAccount 部分的 i18n keys（connectFb, disconnectFb 等） | 国际化 |
| `art-design-pro/src/locales/langs/en.json` | 添加对应英文翻译 | 国际化 |

## API 端点

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| `GET` | `/api/v1/fb/auth-url` | Bearer Token | 获取 Facebook OAuth 授权链接（state 编码 userID） |
| `GET` | `/api/v1/fb/callback` | 无（Facebook 回调） | OAuth 回调，code→token，重定向到前端 |
| `GET` | `/api/v1/fb/status` | Bearer Token | 获取连接状态（connected, fbUserName, expiresAt） |
| `GET` | `/api/v1/fb/ad-accounts` | Bearer Token | 获取广告账户列表 + BM 列表 |
| `DELETE` | `/api/v1/fb/disconnect` | Bearer Token | 断开 Facebook 连接 |

## 数据库变更

```sql
CREATE TABLE fb_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    fb_user_id VARCHAR(64),
    fb_user_name VARCHAR(128),
    access_token TEXT NOT NULL,     -- Facebook access token
    token_type VARCHAR(32),
    expires_at TIMESTAMP,
    scopes TEXT[],
    bm_list JSONB,                  -- BM 列表缓存
    ad_accounts JSONB,              -- 广告账户缓存
    selected_ad_account_id VARCHAR(64),
    status INT DEFAULT 1,           -- 1=有效 0=失效
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
CREATE UNIQUE INDEX idx_fb_tokens_user_id ON fb_tokens(user_id);
```

## OAuth State 设计

- **编码格式**: `hex(userID):hex(nonce)`（例: `1:a1b2c3d4e5f6...`）
- **存储**: 临时写入 `fb_tokens.access_token` 字段（`pending:<state>`），5 分钟过期
- **CSRF 防护**: state 参数两次验证 — ① 解析 userID hex ② 匹配 DB 中的 pending 记录

## 前置条件

需要在 [Meta 开发者平台](https://developers.facebook.com) 完成：
1. 创建 Facebook 应用
2. 添加「Facebook 登录」产品
3. 配置 OAuth 回调 URL: `http://localhost:9090/api/v1/fb/callback`（生产环境需改为 HTTPS 域名）
4. 添加权限: `ads_management`, `ads_read`, `business_management`
5. 将 App ID 和 App Secret 填入 `.env` 的 `FB_APP_ID` 和 `FB_APP_SECRET`

## 验证结果

- ✅ 后端编译成功（`go build -o server.exe ./main.go`）
- ✅ 前端 TypeScript 类型检查（`pnpm build` — ad-account 相关错误全部清除）
- ✅ API 端点测试通过:
  - `GET /api/v1/fb/auth-url` → 提示需要配置 FB_APP_ID
  - `GET /api/v1/fb/status` → 返回 `connected: false`
  - `GET /api/v1/fb/ad-accounts` → 返回未授权错误
- ✅ 数据库迁移执行成功（`fb_tokens` 表已创建）

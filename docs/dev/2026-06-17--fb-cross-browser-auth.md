# Facebook OAuth 跨浏览器授权 + 短链接 — 2026-06-17

## 问题

1. 用户点击「连接 Facebook」后，`window.open()` 直接在**同一浏览器**中打开 Facebook 授权页。
   但如果用户的 Facebook 账号登录在**其他浏览器**中，就无法完成授权。
2. Facebook OAuth 授权链接非常长（~300 字符），不便于复制和分享。

## 需求

1. 生成一个独立的授权链接，用户可以**复制**到已登录 Facebook 的浏览器中打开并授权。
2. 授权完成后，后台管理系统**自动检测**授权状态。
3. 生成**短链接**，方便复制、发送到其他设备。

---

## 修改内容

### 后端：`art-design-server/models/fb.go`

`FbAuthURLResponse` 新增 `shortUrl` 字段：
```go
type FbAuthURLResponse struct {
    AuthURL  string `json:"authUrl"`
    ShortURL string `json:"shortUrl"`  // 新增
}
```

### 后端：`art-design-server/services/fb_service.go`

新增短链接相关功能：

| 方法 | 说明 |
|------|------|
| `GetShortAuthURL(userID, serverHost)` | 生成完整授权链接 + 8 位随机 token 的短链接 |
| `ResolveShortToken(token)` | 根据 token 获取完整授权链接（5 分钟有效） |
| `cleanExpiredShortTokens()` | 后台清理过期 token |

**存储**：内存 `sync.Map`，5 分钟自动过期（与 Facebook state 过期时间一致）。

**短链接格式**：`http://{serverHost}/api/v1/fb/go/{8位随机token}`
例如：`http://localhost:9090/api/v1/fb/go/a1b2c3d4`

### 后端：`art-design-server/handlers/fb_handler.go`

| Handler | 变更 |
|---------|------|
| `AuthURL` | 改用 `GetShortAuthURL()`，返回 `{authUrl, shortUrl}` |
| `Callback` | （上轮）不再重定向 localhost:3006，返回 HTML 成功页面 |
| `ShortRedirect` | **新增** — `GET /api/v1/fb/go/:token`，重定向到完整 FB 授权链接 |
| - | 过期时返回友好的 HTML 提示页面 |

### 后端：`art-design-server/routes/router.go`

新增路由（无需登录）：
```go
v1.GET("/fb/go/:token", handlers.DefaultFbHandler.ShortRedirect)
```

### 前端：`art-design-pro/src/api/facebook.ts`

`fetchFbAuthUrl()` 返回类型新增 `shortUrl`：
```ts
request.get<{ authUrl: string; shortUrl: string }>
```

### 前端：`art-design-pro/src/views/ad-account/index.vue`

**对话框 UI 变更：**

```
┌─────────────────────────────────────┐
│  Facebook 授权                       │
│                                     │
│  复制以下链接，在已登录 Facebook     │
│  的浏览器中打开并授权：              │
│                                     │
│  短链接（推荐）                      │
│  ┌──────────────────────┐ ┌──────┐ │
│  │ http://localhost:9090 │ │ 复制 │ │
│  │ /api/v1/fb/go/abc123 │ │ 链接 │ │
│  └──────────────────────┘ └──────┘ │
│                                     │
│  ▶ 查看完整链接                     │
│                                     │
│  [在当前浏览器打开] 或复制到其他      │
│                                     │
│  ℹ 等待授权完成...每3秒自动检测      │
│                                     │
│                          [取消]     │
└─────────────────────────────────────┘
```

**核心逻辑变更：**

| 变更 | 说明 |
|------|------|
| `shortUrl` / `fullAuthUrl` | 分别存储短链接和完整链接 |
| `copyShortUrl()` | 复制短链接（主操作） |
| `copyFullUrl()` | 复制完整链接（折叠区内） |
| `fallbackCopy()` | 统一的降级复制方案（Clipboard API + execCommand） |
| `openAuthUrl()` | 用完整链接打开（更可靠） |

### 国际化：`zh.json` / `en.json`

新增 2 个键：

| 键 | 中文 | English |
|----|------|---------|
| `menus.adAccount.shortLinkLabel` | 短链接（推荐） | Short Link (Recommended) |
| `menus.adAccount.fullLinkLabel` | 查看完整链接 | View Full Link |

---

## 完整流程

```
点击「连接 Facebook」
  → GET /api/v1/fb/auth-url → 获取 {authUrl, shortUrl}
  → 弹窗展示短链接 + 复制按钮
  → 开始轮询 GET /api/v1/fb/status（每3秒）
  → 用户复制短链接 → FB 浏览器打开
    → GET /api/v1/fb/go/:token → 307 Redirect → Facebook OAuth 授权页
    → 用户授权
    → Facebook 回调 GET /api/v1/fb/callback → 后端保存 token
    → 返回 HTML 成功页面 ✅
  → 前端轮询检测到 connected = true
  → 弹窗自动关闭 → 刷新广告账户列表 ✅
```

---

## 注意事项

1. **短链接 5 分钟有效**，与 Facebook state 过期时间一致。
2. **短链接基于 serverHost**（`c.Request.Host`），自动适配当前服务地址。
3. **`/api/v1/fb/go/:token` 无需登录**，任何浏览器都能访问。
4. 完整链接仍保留在折叠面板中，以备不时之需。
5. 跨机器授权仍需 `FB_REDIRECT_URI` 可被授权浏览器访问。

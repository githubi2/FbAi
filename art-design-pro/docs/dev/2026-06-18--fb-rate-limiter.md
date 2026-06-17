# FB 请求限速队列 — Dev Doc

**日期**: 2026-06-18  
**类型**: feat  
**影响范围**: 前端 FB 请求 + Go 后端 → Facebook Graph API 请求

## 双层限速架构

```
浏览器前端 (Vue)                    Go 后端 (Gin)
     │                                   │
     │  ① 前端限速 (4s 间隔)              │  ② 后端限速 (4s 间隔)
     │  http/index.ts URL 匹配           │  fbGet / exchangeToken / getFbUserInfo
     │  → fbRateLimiter.schedule()       │  → DefaultFbRateLimiter.Do()
     │                                   │
     ▼                                   ▼
Vite 代理 ────────────────────────→ Gin Handler → FbService → Facebook API
```

## 修改内容

### 前端新增

| 文件 | 说明 |
|---|---|
| `src/utils/http/rateLimiter.ts` | 通用请求限速队列工具（TypeScript） |

### 前端修改

| 文件 | 变更 |
|---|---|
| `src/utils/http/index.ts` | 在 `request()` 函数中集成 FB 端点限速 |

### 后端新增

| 文件 | 说明 |
|---|---|
| `art-design-server/services/fb_rate_limiter.go` | Go 后端 FB 请求限速队列 |

### 后端修改

| 文件 | 变更 |
|---|---|
| `art-design-server/services/fb_service.go` | `fbGet()`、`exchangeLongLivedToken()`、`getFbUserInfo()` 三个方法接入限速器 |

---

## 变更详情

### 前端 — `rateLimiter.ts` + `http/index.ts`

- 类 `RateLimiter`：内存队列，确保请求间隔 >= `minInterval`（默认 4s）
- 环境变量 `VITE_FB_RATE_LIMIT_MS` 可自定义间隔
- 支持 `AbortSignal` 取消排队中的请求
- 开发环境日志：`[RateLimiter] "/api/v1/fb/xxx" 排队等待 NNNms`
- `http/index.ts` 中通过 `FB_API_PREFIX` 常量匹配 `/api/v1/fb/` 前缀，自动走限速队列

### 后端 — `fb_rate_limiter.go` + `fb_service.go`

- 类 `FbRateLimiter`：单 goroutine + channel 串行队列（128 缓冲）
- 默认间隔 4 秒，环境变量 `FB_RATE_LIMIT_MS` 可自定义
- `Do(ctx, endpoint, fn)` 方法：加入队列，阻塞等待执行完成
- 日志：`[FbRateLimiter] "/v22.0/me/adaccounts" 排队等待 3.662s`
- 接入点：
  1. `fbGet()` — 所有 GET 请求走此方法（广告账户、BM、支付等）
  2. `exchangeLongLivedToken()` — OAuth 短期 token 换长期 token
  3. `getFbUserInfo()` — 获取 FB 用户 id/name

---

## 设计理由

- **双重保护**：前端限速防止用户快速点击；后端限速确保 Go → Facebook 的真实事先请求间隔也受控
- **不使用 Redis**：前端无法连 Redis；后端单实例场景下 goroutine 队列已足够
- **4 秒间隔**：经实测足够宽松，防止 Facebook 人机检测

## 验证结果

### 前端
- ✅ `[RateLimiter] "/api/v1/fb/accounts" 排队等待 3968ms`
- ✅ 零 JS 运行时错误

### 后端
- ✅ 编译通过，服务正常启动
- ✅ 连续 3 次请求耗时：4.31s / 7.99s / 7.98s（间隔约 4s）
- ✅ 服务端日志确认：
  ```
  [FbRateLimiter] "/v22.0/me/businesses" 排队等待 3.401s
  [FbRateLimiter] "/v22.0/me/adaccounts" 排队等待 3.662s
  ```

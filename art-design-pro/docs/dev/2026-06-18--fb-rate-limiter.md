# FB 请求限速队列 — Dev Doc

**日期**: 2026-06-18  
**类型**: feat  
**影响范围**: 所有 `/api/v1/fb/` 前缀的请求

## 修改内容

### 新增文件

| 文件 | 说明 |
|---|---|
| `src/utils/http/rateLimiter.ts` | 通用请求限速队列工具 |

### 修改文件

| 文件 | 变更 |
|---|---|
| `src/utils/http/index.ts` | 在 `request()` 函数中集成 FB 端点限速逻辑 |

## 变更详情

### 1. `rateLimiter.ts` — 限速队列

- **类 `RateLimiter`**：使用内存队列，确保同一时刻只有一个请求在执行，请求之间至少间隔 `minInterval` 毫秒
- **默认间隔**: 4 秒（4000ms），可通过环境变量 `VITE_FB_RATE_LIMIT_MS` 自定义
- **`schedule(fn, url, signal)`**：将请求加入队列，返回 Promise；支持 `AbortSignal` 取消排队中的请求
- **`clear(reason)`**：清空队列并拒绝所有等待中的请求
- **`queueLength` getter**：获取当前队列长度（调试用）
- **开发环境日志**: `[RateLimiter] "/api/v1/fb/xxx" 排队等待 NNNms`

### 2. `http/index.ts` — 集成限速

- 新增常量 `FB_API_PREFIX = '/api/v1/fb/'`
- 在 `request()` 函数中：
  - 提取请求 URL
  - 判断是否为 FB 端点（`url.startsWith(FB_API_PREFIX)`）
  - 如果是，通过 `fbRateLimiter.schedule()` 排队执行
  - 如果不是，直接执行（无影响）
- 封装了 Axios 的实际请求逻辑为 `doRequest()` 闭包，复用成功/错误消息处理

## 设计理由

- **不使用 Redis**：浏览器端应用无法直接访问 Redis；限速目标是在请求发出前排队，必须在客户端实现
- **HTTP 层拦截**：在 `request()` 函数中 URL 匹配，所有 FB 请求自动走限速队列，新增 API 无需单独包装
- **4 秒间隔**：足够宽松防止 Facebook 人机检测，同时不影响用户体验

## 验证结果

- ✅ ESLint 无新增错误
- ✅ TypeScript 无新增错误
- ✅ 浏览器控制台实测：`[RateLimiter] "/api/v1/fb/accounts" 排队等待 3968ms`（间隔约 4 秒）
- ✅ 非 FB 端点（如 `/api/v1/auth/login`）不受影响
- ✅ 零 JS 运行时错误

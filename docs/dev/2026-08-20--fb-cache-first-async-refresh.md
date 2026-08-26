# FB 列表缓存优先 + 后台异步刷新修复 — 2026-08-20

## 问题

用户反馈：FB 账号列表 / 广告账户管理页面刷新后要等很久才显示数据。

## 诊断

- `GET /fb/accounts` 本身已走缓存（毫秒级），但存在 4 个缺陷：
  1. **后台刷新不调 FB API**：`refreshAccountsCache` 只重读 `fb_tokens` 的 JSONB，新授权账号的 BM/广告户统计永远为 0
  2. **OAuth 回调不触发刷新**：`Callback` 只 `SaveToken`，新账号统计为空
  3. **detail 无缓存时同步阻塞**：`AdAccountsDetail` 首次访问同步调 Graph API（每广告账户 2-3 次请求 × 4 秒限速，几十秒）
  4. **running 任务卡死**：服务重启后 `fb_refresh_status` 残留 running 记录，`IsRefreshing` 永真 → 后台刷新永久停摆
  5. **多账号缓存错配**：`SaveAccountsCache/SaveAdAccountsCache` 把全量数据按循环到的每个 tokenID 各存一遍（UPSERT 互踩）

## 修改文件

| 文件 | 变更 |
|------|------|
| `models/fb.go` | `FbAdAccountDetail` 新增 `TokenID uint \`json:"-"\``（缓存关联用，不返回前端） |
| `services/fb_service.go` | `GetAdAccountsDetail` 设置 `detail.TokenID = tokenID` |
| `services/fb_cache_service.go` | ① `SaveAccountsCache` 去掉 tokenID 参数，改用 `acc.ID`（即 fb_tokens.id）逐条关联 ② `SaveAdAccountsCache` 改用 `acc.TokenID` ③ 新增 `cleanStaleRefreshes`：running 超 10 分钟自动标记 failed，`IsRefreshing`/`GetRefreshStatus` 调用前自动清理 ④ `GetCachedAccounts` 的 `CreatedAt` 改用 `c.created_at`（原误用 last_refresh_at，授权时间列显示成了刷新时间） |
| `handlers/fb_handler.go` | ① `refreshAccountsCache` 重写：逐 token 调 `RefreshAccountStats`（真实 Graph API，走限速队列）→ 重读 → 存缓存 ② `AdAccountsDetail` 无缓存时直接返回空列表（异步刷新已在后台跑，前端轮询 refresh-status 自动重载） ③ `Callback` 授权成功后 `go h.refreshAccountsCache()` ④ `RefreshStats` 手动刷新后同步更新缓存表 ⑤ `saveAccountsToCache`/`saveAdAccountsToCache` 简化为单次调用 |

## 最终数据流

```
用户进入页面
  ├─ 有缓存 → 缓存表直出（<20ms）          ← 用户立即看到数据
  ├─ 无缓存 → accounts 读 fb_tokens（快）/ detail 返回空
  └─ 同时后台 goroutine：
       逐 token 调 Graph API（4s 限速）→ 回写 fb_tokens + 缓存表
前端每 2s 轮询 /fb/refresh-status → 刷新完成自动重载表格
```

## 二次修复：刷新-轮询死循环（同日 17:40）

**现象**：页面表格持续转圈 10 秒以上，`refresh-status` 每 5 秒轮询永不停止，`/fb/accounts` 反复重载。

**根因**：`ListAccounts`/`AdAccountsDetail` 每次请求都触发后台刷新 → 前端轮询发现 isRunning → 刷新完成后 `refreshData()` 重新加载 → 又触发新刷新 → 无限循环。表格永远处于"刚刷新完又重载"的状态。

**修复**：新增 `FbCacheService.ShouldRefresh(userID, tenantID, type, cooldown)` —— 无进行中任务 **且距上次刷新超过 5 分钟冷却期** 才启动后台刷新。两个列表 handler 改用此判断。OAuth 回调里的主动刷新不受冷却限制。

**验证**：连续 3 次请求 `isRunning` 恒为 false；日志恢复安静（之前每 5 秒一次轮询）；接口 2-10ms。

## 三次修复：前端静默重载（同日 17:55）

**现象**：每次进页面请求两次 accounts 接口，表格转圈要等两次都返回。

**根因**：后台刷新运行期间，前端轮询完成后调用 `refreshData()` → 重新触发表格 loading → 二次转圈，用户感知为"等很久"。

**修复**（前端两个页面）：
- `ad-account/index.vue`、`ad-account/manage/index.vue`：轮询完成后的 `refreshData()` 改为 `silentReload()` —— 直接请求接口替换 `data.value`，不触发 loading
- 账号列表页 `totalAccounts` 从 `computed(pagination.total)` 改为 `ref`，由 useTable `hooks.onSuccess` 和 silentReload 同步更新
- 效果：首次请求（缓存数据）→ 表格立即显示；后台刷新完成后数据**无感知替换**

## 四次重构：读写分离 + 显式刷新触发（同日 18:15）

**问题**：刷新触发藏在 GET 列表接口里（隐式副作用），前端刷新检查散落三处（onMounted / fetch 函数内 / 轮询），职责混乱。

**重构后契约**：

| 接口 | 行为 |
|------|------|
| `GET /fb/accounts` | **纯读**：缓存直出 → 无缓存读 fb_tokens（不调 FB API）。零副作用 |
| `POST /fb/accounts/refresh-all` | **显式触发**后台刷新；5 分钟冷却期内/刷新中为 no-op，幂等 |
| `GET /fb/ad-accounts/detail` | 同上纯读；无缓存返回空列表 |
| `POST /fb/ad-accounts/refresh-all` | 同上显式触发 |
| `GET /fb/refresh-status?type=` | 查询刷新状态（前端轮询用） |

**前端单向流程**（两个页面一致）：
```
onMounted → useTable 自动 GET（立即显示缓存数据）
         → POST refresh-all（触发后台刷新）
         → 若 isRunning → 每 2s 轮询 refresh-status
         → 完成 → silentReload() 静默替换数据（不转圈）
```
fetch 函数内不再夹带刷新状态检查（单一职责）。

**附带**：`vite.config.ts` 增加 `server.warmup.clientFiles`（main.ts + 两个广告页面 + 用户管理页），dev 服务器启动时预编译核心模块，缓解冷启动白屏。

| 场景 | 结果 |
|------|------|
| `GET /fb/accounts` 响应时间 | ✅ 18ms（缓存直出） |
| `GET /fb/ad-accounts/detail` 响应时间 | ✅ 6.7ms（缓存直出） |
| 后台刷新状态流转 | ✅ running → completed |
| 刷新后统计数据更新 | ✅ bmCount 0→1，personalAdCount 0→1，bmAdCount 0→3 |
| 授权时间列显示 | ✅ 显示真实授权时间 |
| 后端编译 `go build` | ✅ |

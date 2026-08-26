# FbAi 项目记忆档案

> 生成时间：2026-08-19。本文件汇总项目进度、功能清单、架构决策与当前状态，供后续会话快速恢复上下文。

---

## 一、项目定位

**FbAi** — Facebook 广告账户管理平台。基于 `art-design-pro` 模板（Vue 3）+ 自研 Go 后端，核心场景：

1. 多租户 SaaS 管理后台（用户/角色/菜单/租户 + RBAC 权限）
2. Facebook OAuth 授权（一个用户可授权多个 FB 账号）
3. 广告账户（Ad Account）与 BM（Business Manager）管理：列表、授权/删除成员权限、支付记录、批量操作
4. FB API 调用保护：跨进程限速队列（Go + Python 共用 Redis）

## 二、架构

| 包 | 技术栈 | 端口 | 入口 |
|---|---|---|---|
| `art-design-pro/` | Vue 3 + Vite 7 + TS + Element Plus + Pinia + Tailwind 4 | 3006 | `src/main.ts` |
| `art-design-server/` | Go 1.25 + Gin + pgx v5 (PostgreSQL) | 9090 | `main.go` |

- 前端 Vite 代理 `/api` → `localhost:9090`；启动：`./start.bat`（Windows，开发模式）/ `./start.sh`
- **日常使用（秒开）**：`./start-prod.bat` 或 `pnpm start:prod` — 生产构建 + `vite preview`（481 个 dev 模块请求 → 几个打包文件，解决 dev 模式 F5 白屏 ~10s 问题）；改了前端代码需删 `dist/` 重新构建
- `vite.config.ts` 已配 `preview.proxy`（/api → 9090，回退默认 localhost:9090）
- 默认登录：`admin` / `admin123`
- 数据库：PostgreSQL 17，库名 `fbai`；`DATABASE_URL` 缺失时回退内存模式
- 隧道工具：根目录有 `ngrok.exe` / `bore.exe`（用于 FB OAuth 回调公网暴露）

## 三、功能进度时间线

### 2026-06-16 — 基座搭建
- **后端初建**：Gin + pgx v5，routes → handlers → services → models 分层（`art-design-server/docs/dev/2026-06-16--initial-backend-setup.md`）
- **PostgreSQL 登录**：去掉登录页演示账号下拉框；登录走 DB 验证（`docs/dev/2026-06-16--remove-login-select-and-postgres-login.md`）
- **SSO + 记住密码 + 改密踢人**：`sessions` 表存 token（非 JWT）；记住密码=3 天，否则 24h；新登录删除旧会话；管理员改密码强制目标用户下线（`docs/dev/2026-06-16--sso-remember-me.md`）
- **多租户 RBAC**：`tenants`/`permissions`(18 权限点）/`role_permissions` 表；`TenantContext` 中间件设 PG 会话变量；双层隔离 = 服务层 `WHERE tenant_id`（主）+ RLS（辅，roles 表 RLS 已禁用）（`docs/dev/2026-06-16--multi-tenant-rbac.md`）
- **租户系统管理**：租户管理页面 + 顶部租户切换组件 + 创建租户事务（租户+角色+管理员账号）（`docs/dev/2026-06-16--tenant-system-management.md`）
- **输入校验管理** + 修复租户创建用户 tenant_id 绑定、登录 500、菜单 i18n/图标映射等（多个 fix 文档）

### 2026-06-17 — Facebook OAuth 核心
- **FB OAuth 授权 + 广告账户管理**（`docs/dev/2026-06-17--facebook-oauth-ad-account.md`）
  - 流程：后端生成授权链接 → FB 授权 → 回调 `/api/v1/fb/callback` → 换 access_token → 加密存 `fb_tokens` 表 → 后端代理调 Graph API
  - state 设计：`hex(userID):hex(nonce)`，pending 记录 5 分钟过期，双重验证防 CSRF
  - 需在 Meta 开发者平台配置：App、Facebook 登录产品、回调 URL、权限 `ads_management` / `ads_read` / `business_management`
- **FB 多账号授权**（`docs/dev/2026-06-17--fb-multi-account.md`）
  - 一个用户可授权多个 FB 账号；唯一约束 `(user_id, fb_user_id) WHERE status=1`（同账号刷 token，不同账号新增）
  - FB 账户列表页完全重写：表格显示状态/广告权限/BM 数/广告户数/有效期，操作=编辑备注/刷新统计/断开连接（软删除）
  - 新增 API：`GET/DELETE /api/v1/fb/accounts[/:id]`、`PUT /:id/label`、`POST /:id/refresh`
- **Business Login config_id 授权**（`docs/dev/2026-06-17--fb-business-login-config-id.md`）
- **跨浏览器授权**（`docs/dev/2026-06-17--fb-cross-browser-auth.md`）+ 授权结果页样式
- 前端：广告账户管理页、列扩展、菜单对接（`art-design-pro/docs/dev/2026-06-17--*`）

### 2026-06-18 — 表格增强 + 限速
- **BM 管理菜单重构**：BM管理改为目录菜单，新增「BM列表」子菜单（id=16）（`docs/dev/2026-06-18--bm-menu-restructure.md`）
- **管理员标签弹窗**：admin/hiddenAdmin 标签值为 0 时直接提示不弹窗；零值检查 + 步骤式 UI；后端新增 `otherAdminNames` 字段
- **FB 请求限速队列**（`art-design-pro/docs/dev/2026-06-18--fb-rate-limiter.md` + `docs/features/fb-redis-rate-limiter.md`）
  - Go 端：所有 FB 接口排队，默认间隔 4 秒（`FB_RATE_LIMIT_MS`），双层限速
  - Redis 分布式：Go 与 Python AI 自动化共享 `fb:next_available_at` key + 相同 Lua 原子脚本；无 Redis 时 Go 自动降级内存限速
- 表格分页搜索、管理员标签值零值检查等

### 2026-06-24 — 权限管理弹窗 + 共享角色
- **授权弹窗**（AddAuthDialog）：对接真实 FB API，三种角色类型下拉；步骤指示器 UI 经过多轮对齐修复
- **删除授权弹窗 + 后端**（`docs/dev/2026-06-24--delete-auth-backend.md`）
  - `POST /api/v1/fb/ad-accounts/remove-user`，5 种删除模式：deleteTheirs / deleteExceptTheirs / deleteExceptSelf / deleteSelf / deleteBM
- **广告账户列表**：复选框多选 + 批量操作按钮组
- **共享角色重构**（`art-design-pro/docs/dev/2026-06-24--shared-roles-refactor.md`）
  - 角色改全局共享：`R_TENANT_ADMIN`（id=28，11 权限点）、`R_TENANT_USER`（id=29，1 权限点），tenant_id=NULL
  - 创建租户不再建角色，改为查共享角色；旧租户角色 21/22/26/27 已删除
- AGENTS.md 新增 Rule 44：每次改完都要跑测试

## 四、数据库表

| 表 | 说明 |
|---|---|
| `users` | tenant_id=NULL 为超级管理员；role_name 不能为空 |
| `roles` | role_code 全局唯一（如 `R_SUPER`、`R_TENANT_ADMIN`）；menu_ids int[] 控制菜单可见性 |
| `menus` | 后端驱动动态菜单（`VITE_ACCESS_MODE=backend`）；图标需 `ri:` 前缀 |
| `sessions` | token 认证 + SSO；管理员单点，租户可多登 |
| `tenants` / `permissions` / `role_permissions` | 多租户 + 18 权限点 |
| `fb_tokens` | FB OAuth token 加密存储；`label` 备注；`bm_list`/`ad_accounts` JSONB 缓存 |
| `fb_accounts_cache` / `fb_ad_accounts_cache` / `fb_refresh_status` | ⚠️ 未提交的缓存方案（见 WIP） |

## 五、FB API 端点（`/api/v1/fb`，均需 Bearer，callback 除外）

| 端点 | 说明 |
|---|---|
| `GET /fb/auth-url` | 生成 OAuth 授权链接 |
| `GET /fb/callback` | FB 回调（免认证），换 token 存库后重定向前端 |
| `POST /fb/data-deletion` | FB 数据删除回调（合规要求） |
| `GET /fb/go/:token` | 短链接重定向 |
| `GET /fb/status` / `GET /fb/ad-accounts` | 旧版兼容接口（返回第一个有效连接） |
| `GET /fb/ad-accounts/detail` | 广告账户详情 |
| `GET /fb/ad-accounts/:id/payments` | 支付记录 |
| `POST /fb/ad-accounts/assign-user` / `remove-user` | 分配/删除广告账户成员权限 |
| `POST /fb/users/lookup` | FB 用户查询 |
| `GET/DELETE/PUT/POST /fb/accounts...` | 多账号管理：列表、断开、改备注、刷新统计 |
| `GET /fb/refresh-status` | 刷新任务状态（前端轮询用） |

## 六、缓存优先架构（2026-08-20 已修复并验证）

详见 `docs/dev/2026-08-20--fb-cache-first-async-refresh.md`。

- 两个列表接口（`/fb/accounts`、`/fb/ad-accounts/detail`）均为**缓存直出（<20ms）纯读接口，零副作用**
- **读写分离**：后台刷新由 `POST /fb/accounts/refresh-all` / `POST /fb/ad-accounts/refresh-all` 显式触发（5 分钟冷却，幂等）
- 后台刷新真实调 Graph API（走 4s 限速队列），结果回写 `fb_tokens` + 缓存表
- 前端单向流程：GET 立即显示 → POST 触发刷新 → 轮询 refresh-status → 完成后 `silentReload()` 静默替换（不转圈）
- OAuth 授权成功后服务端自动后台拉取统计（不经冷却）
- `fb_refresh_status` 卡死的 running 任务 10 分钟自动标记失败
- vite.config 已配 `server.warmup`（预编译核心页面，缓解 dev 冷启动白屏）
- `fb_refresh_status` 卡死的 running 任务 10 分钟自动标记失败
- 缓存按每条的 `fb_token_id` 关联（修复了多账号错配）

**仍未提交的 WIP**：`AddToBmDialog.vue`（添加到 BM 弹窗）、`add_ad_account_menu.sql` 等；本次修复的 4 个文件也未提交。

## 七、关键规则速记（详见 AGENTS.md）

- 后端改代码必须重编译重启 `server.exe`（无热重载）；Go 代理 `GOPROXY=https://goproxy.cn,direct`
- SQL 只用参数化查询；用户输入前后端双重验证
- 响应格式 `{ code, msg, data }`；前端请求自动解包 data
- 表格必须 `ArtTable` + `ArtTableHeader` + `useTable()`；按钮用 `ArtButtonTable`
- 所有用户可见文本必须 `$t()` 国际化；模板单根元素；文件 <300 行
- 每个功能必须写文档 `docs/dev/YYYY-MM-DD--<feature>.md`
- 菜单-数据库同步：前端路由 name 必须与 `menus` 表匹配；超级管理员菜单用 `TreeByIDs`
- Git：禁止直接提 master、禁止 `git add .`、禁止绕过 pre-commit、禁提 `.env`
- 限速间隔参数 Go/Python 必须一致（默认 4000ms）

## 八、目录速查

```
E:\FbAi\
├── AGENTS.md                  # 项目规则（每次必读）
├── start.bat / start.sh       # 一键启动
├── docs/                      # 根级开发文档（后端+跨端功能）
├── art-design-pro/            # 前端
│   ├── AGENTS.md              # 前端专属规则
│   ├── docs/dev|features/     # 前端功能文档
│   └── src/views/             # ad-account(含bm、manage)、system、tenant-system、auth...
├── art-design-server/         # 后端
│   ├── migrations/            # 16 个 SQL 迁移（手动执行）
│   ├── seed.sql               # 种子数据
│   ├── services/fb_service.go # FB 核心服务（1942 行）
│   └── handlers/fb_handler.go # FB API（814 行）
└── fb_test.py / token.txt     # Python 自动化测试脚本
```

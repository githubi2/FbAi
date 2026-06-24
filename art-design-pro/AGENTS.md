# FbAi — art-design-pro: AI Agent Rules & Context

> **Target Audience**: AI coding agents (mimo, Cursor, Copilot, Claude Code, etc.) **Version**: 2026-06-16 **Repo**: https://github.com/githubi2/FbAi **Upstream**: https://github.com/Daymychen/art-design-pro **Working Directory**: `E:\FbAi` (all terminal commands run from here; `cd art-design-pro` for frontend, `cd art-design-server` for backend)

---

## 1. Tech Stack

| Layer           | Technology                                | Version |
| --------------- | ----------------------------------------- | ------- |
| Framework       | Vue 3 (Composition API, `<script setup>`) | 3.5.x   |
| Language        | TypeScript (strict mode)                  | 5.6.x   |
| Build           | Vite                                      | 7.x     |
| UI Library      | Element Plus                              | 2.11.x  |
| CSS             | Tailwind CSS 4 + SCSS                     | 4.1.x   |
| State           | Pinia + pinia-plugin-persistedstate       | 3.x     |
| Router          | Vue Router                                | 4.5.x   |
| Charts          | ECharts                                   | 6.x     |
| HTTP            | Axios                                     | 1.12.x  |
| i18n            | Vue I18n (zh-CN, en)                      | 9.x     |
| Package manager | pnpm                                      | ≥8.8    |
| Backend         | Go + Gin + PostgreSQL 17                  | —       |

Other notable deps: crypto-js, xlsx, file-saver, wangeditor (rich text), xgplayer (video), qrcode.vue, vue-draggable-plus, @vueuse/core, @iconify/vue.

---

## 2. Project Structure

```
E:\FbAi\
├── art-design-pro/          # Frontend (Vue 3 + Element Plus)
│   ├── src/
│   │   ├── api/             # API layer (auth.ts, system-manage.ts, etc.)
│   │   ├── assets/          # Static assets
│   │   │   ├── images/      # (@imgs/)
│   │   │   ├── icons/       # (@icons/)
│   │   │   └── styles/      # (@styles/) — el-ui.scss (global EP tuning)
│   │   ├── components/      # Shared components (ArtButtonTable, ArtTableHeader, etc.)
│   │   ├── config/          # App config (fastEnter.ts)
│   │   ├── directives/      # Custom directives (v-auth, v-roles, v-ripple)
│   │   ├── hooks/           # Composables (useTable, useAuth, useDebounce, etc.)
│   │   ├── locales/         # i18n (zh.json, en.json) + index.ts config
│   │   ├── plugins/         # (@plugins/)
│   │   ├── router/          # Vue Router (modules: dashboard, system, tenantSystem)
│   │   ├── store/           # Pinia stores (useXxxStore pattern)
│   │   ├── types/           # TypeScript types (api/api.d.ts, common/response.ts)
│   │   ├── utils/           # Utilities (ui/, sys/, table/tableConfig.ts)
│   │   └── views/           # Pages
│   │       ├── auth/        # Login (ArtDragVerify slider!), Register, Forgot Password
│   │       ├── dashboard/   # Console dashboard (charts, stats)
│   │       ├── system/      # User, Role, Menu, Tenant management (super admin)
│   │       ├── tenant-system/ # Tenant-scoped User, Role, Menu management
│   │       ├── exception/   # 403, 404, 500
│   │       ├── outside/     # Iframe embedding
│   │       └── result/      # Success/Fail result pages
│   ├── docs/
│   │   ├── dev/             # Per-change development documentation
│   │   └── features/        # Feature-level documentation
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
│
└── art-design-server/       # Backend (Go + Gin)
    ├── main.go              # Entry: router setup, db.Connect(), r.Run()
    ├── routes/              # Route definitions (thin)
    ├── handlers/            # HTTP handlers (bind → service → respond)
    ├── services/            # Business logic
    ├── models/              # Data structs, DB models
    ├── db/                  # pgxpool connection (postgres.go)
    ├── middleware/           # Auth, TenantContext, RequirePermission
    ├── crypto/               # bcrypt HashPassword + CheckPassword
    ├── migrations/          # SQL migration files
    ├── scripts/             # Seed scripts
    └── .env                 # DATABASE_URL (gitignored!)
```

### Path Aliases (tsconfig paths)

| Alias       | Maps to                                           |
| ----------- | ------------------------------------------------- |
| `@/`        | `src/`                                            |
| `@views/`   | `src/views/`                                      |
| `@imgs/`    | `src/assets/images/`                              |
| `@icons/`   | `src/assets/icons/`                               |
| `@utils/`   | `src/utils/`                                      |
| `@stores/`  | `src/store/`                                      |
| `@plugins/` | `src/plugins/`                                    |
| `@styles/`  | `src/assets/styles/`                              |
| `@hooks/`   | `src/hooks/` (manual convention, not in tsconfig) |

---

## 3. Environment & Configuration

| Item | Value |
| --- | --- |
| Frontend dev port | `http://localhost:3006` |
| Backend port | `http://localhost:9090` |
| Access mode | `VITE_ACCESS_MODE=backend` (DB-driven menus via `/api/v1/auth/menus`) |
| Production API | `.env.production`: `VITE_API_URL = /` (reverse proxy, NOT apifoxmock) |
| Dev proxy | Vite proxy → `localhost:9090` |
| Mock status | **Removed** (2026-06-16). `src/mock/` deleted entirely. All data from real backend. |
| DB | PostgreSQL 17, host `127.0.0.1:5432`, db `fbai`, user `fbai` |
| psql path | `"C:\Program Files\PostgreSQL\17\bin\psql.exe"` (not in PATH; use full path) |
| DSN | `postgres://fbai:***@127.0.0.1:5432/fbai?sslmode=disable` (in `art-design-server/.env`, gitignored) |
| Admin login | `admin` / `admin123` |
| Seed demo users | `alex`, `sophia`, `liam`, `olivia`, `emma` — all password `123456` |
| Node | ≥ 20.19.0, pnpm ≥ 8.8 |
| Go proxy (China) | `GOPROXY=https://goproxy.cn,direct` |

### Key Commands

```bash
# Frontend
cd E:/FbAi/art-design-pro
pnpm dev          # Start dev server (auto-opens browser, :3006)
pnpm build        # Type-check (vue-tsc) + production build (vite)
pnpm serve        # Preview production build
pnpm lint         # ESLint check
pnpm fix          # ESLint auto-fix
pnpm commit       # Interactive conventional commit (cz-git)

# Backend
cd E:/FbAi/art-design-server
GOPROXY=https://goproxy.cn,direct go build -o server.exe ./main.go
./server.exe &    # Start (background, port :9090)
```

### Red Lines (MUST NEVER VIOLATE)

- ❌ Never commit to `master` directly — use feature branches
- ❌ Never push without passing pre-commit hooks (Husky + lint-staged)
- ❌ Never hardcode API keys, passwords, or DSN strings in source code — use `.env` files
- ❌ Never use `any` for API responses — always type them
- ❌ Never reinvent Element Plus components
- ❌ Never `git add .` (blind staging) — stage specific files
- ❌ Never force-push (`--force`) to master

---

## 4. MANDATORY RULES (43 Rules)

### 🔴 Rule 0: CodeGraph First — Query Before You Code

**Before writing or modifying ANY code**, you MUST use CodeGraph to understand the codebase.

CodeGraph has pre-indexed **221 files, 2,565 nodes, and 6,266 relationships**. Query it to discover:

- **Where** a symbol is defined and used (paths)
- **Who calls** a function and **what it calls** (callers/callees)
- **Impact** — which files break if you change something

**MCP Tools** (use in-session):

| Tool                | Purpose                                  | When to use                |
| ------------------- | ---------------------------------------- | -------------------------- |
| `codegraph_search`  | Search symbols by name/type              | Starting any task          |
| `codegraph_explore` | Deep dive: symbols + source + call paths | Understanding an area      |
| `codegraph_node`    | One symbol's full context                | Quick lookup               |
| `codegraph_callers` | All callers of a symbol                  | Before modifying           |
| `codegraph_callees` | Everything a symbol calls                | Understanding dependencies |
| `codegraph_impact`  | What breaks if you change this           | Before refactoring         |
| `codegraph_files`   | Project file structure                   | Navigating                 |

**CLI Fallback** (when MCP tools unavailable):

```bash
codegraph query "symbolName"      # Search symbols
codegraph explore "auth login"    # Deep dive on an area
codegraph callers "fetchLogin"    # All callers
codegraph impact "BaseResponse"   # Impact analysis
codegraph status                  # Index stats
```

**Required Workflow**: Receive task → Query CodeGraph → Understand impact → THEN code. Never skip.

### API & Data Layer (Rules 1-3)

1. **Base response format**: `{ code: number, msg: string, data: T }` — defined in `src/types/common/response.ts`
2. **Network requests auto-unwrap** `data` from response — destructure directly: `const { token } = await fetchLogin(...)`
3. **Table pagination**: use field mapping from `src/utils/table/tableConfig.ts` — `recordFields`, `totalFields`, `currentFields`, `sizeFields`, `paginationKey`. When backend uses different field names, add them to arrays — never hardcode in components.

### Permissions (Rules 4-7)

4. Two modes controlled by `.env` `VITE_ACCESS_MODE`: `frontend` (roles-based) or `backend` (menu-based)
5. **Current mode**: `backend` — menus come from DB via `GET /api/v1/auth/menus` (role-filtered by `menu_ids`). Super admin gets menus filtered by R_SUPER role's `menu_ids` — NOT all menus via `Tree()`.
6. Button permissions: use `hasAuth()` composable from `@/hooks/core/useAuth`, or `v-auth` / `v-roles` directives
7. **Permission Flow (Login → Dashboard)** — THE most breakage-prone path:

```
Login Submit → POST /api/v1/auth/login → { token, refreshToken }
  → store token, set isLogin=true
  → router.push('/')
  → beforeEach guard fires:
    ├── isLogin? yes
    ├── routes registered? no → enter handleDynamicRoutes()
    │   ├── fetchUserInfo() → GET /api/user/info → { roles, buttons, userId, ... }
    │   │   └── ❌ If 404/no-roles → throw → routeInitFailed → Exception500
    │   ├── getMenuList() →
    │   │   ├── frontend mode: filter asyncRoutes by roles
    │   │   └── backend mode: fetch from /api/v1/auth/menus (role-filtered)
    │   ├── register dynamic routes
    │   └── validate target path permissions
    └── routes registered? yes → proceed to dashboard
```

**Key files**: `src/router/guards/beforeEach.ts` (lines 256-365), `src/api/auth.ts` (fetchGetUserInfo), `src/types/api/api.d.ts` (line 78-85)

### Auth System Details

- **Login page**: username + password only (no role selector). **ArtDragVerify slider** — user MUST drag slider before clicking login; otherwise form submission blocked with red border. This is UX feature, not a bug.
- **Remember Me**: checked = 3 days token expiry, unchecked = 24 hours. Enforced server-side via `sessions.expires_at`.
- **Single Sign-On (SSO)**: new login for same user deletes ALL old sessions (`SessionService.Create`).
- **Admin password change → kick**: when super admin updates another user's password, all sessions for that user are deleted.
- **Login page guard**: already-logged-in users visiting `/auth/login` are redirected to `/`.
- **Route guard init queue**: never `next(false)` when init is in progress; use Promise-based queue so pending navigations wait for init to complete, then replay.

### Component & UI (Rules 8-14)

8. **Element Plus covers 80%** of UI needs — ALWAYS use `<el-*>` components, NEVER reinvent. Check https://element-plus.org/zh-CN/component/overview.html first.
9. Element Plus UI tuning at `src/assets/styles/el-ui.scss` — do NOT override globally elsewhere.
10. **Single root element** in every Vue template — `<Transition>` animation requires it; multi-root causes blank pages.
11. New Element Plus deps: add to `optimizeDeps.include` in `vite.config.ts` to prevent dev reloads. Restart dev server after adding.
12. **All tables MUST use `ArtTable` + `ArtTableHeader` + `useTable()` hook** — never raw `<ElTable>`. Copy pattern from `src/views/system/user/index.vue` (canonical reference).
13. **Table operation buttons MUST use `ArtButtonTable`** icon-only buttons (types: `add`, `edit`, `delete`, `view`, `more`). Pass custom icon via `icon` prop. Never `<ElButton>` text buttons in table columns. Width: 3 buttons → `160`, 2 buttons → `120`.
14. **Button styling rules**: - Page header: `<ElButton v-ripple>文字</ElButton>` (plain, no `type="primary"`) - Dialog footer: cancel = `<ElButton>取消</ElButton>`, confirm = `<ElButton type="primary">确定</ElButton>` - Table operations: `ArtButtonTable` component - ❌ Never `type="success"`, `type="warning"`, `type="info"` anywhere outside dialogs 14a. **Table column width**: use `minWidth` for data columns (responsive), `width` only for `selection`, `index`, `fixed: 'right'` columns. Column formatters: use native `new Date()` (no dayjs/moment in project).

### Internationalization (Rules 15-16)

15. All user-facing text MUST use `$t('key.path')` — NEVER hardcode Chinese/English strings.
16. Language files: `src/locales/langs/zh.json` and `en.json`; config at `src/locales/index.ts` (Composition API mode, zh fallback).

### Code Quality (Rules 17-19)

17. Pre-commit auto-fix via Husky + lint-staged: ESLint → Prettier → Stylelint (DO NOT bypass). Config:
    - `*.{js,ts}`: eslint --fix + prettier --write
    - `*.{vue,html}`: eslint --fix + prettier --write + stylelint --fix
    - `*.{scss,css}`: stylelint --fix + prettier --write
18. Commits: conventional via `pnpm commit`; types: feat/fix/docs/style/refactor/perf/test/build/ci/revert/chore/wip
19. Never commit without passing lint-staged.

### System Components (Rules 20-21)

20. Built-in available: icon selector, image crop, Excel import/export (xlsx), video player (xgplayer), rich text editor (wangeditor), watermark, QR code (qrcode.vue), drag (vue-draggable-plus), file download (file-saver), encryption (crypto-js), event bus (mitt), progress bar (nprogress).
21. System templates: cards, banners, charts (ECharts), maps, chat, calendar, pricing.

### Development Documentation (Rules 22-23)

22. Every code change MUST be documented. Create `docs/dev/YYYY-MM-DD--<feature-name>.md` with: what changed (file paths + before/after), what was added (new files/APIs/components), why (rationale).
23. Major features get `docs/features/<feature-name>.md` with design decisions, API signatures, component hierarchy, and usage examples.

### Modularization (Rules 24-27)

24. Every feature MUST be a self-contained module/component. Never dump everything into one file.
25. One Vue component per concern. Extract composables into `src/hooks/`. Extract utility functions into `src/utils/`.
26. If a file exceeds ~300 lines, split it. Component >300 lines → extract sub-components into `modules/`. Script >150 lines → extract composable. Template >200 lines → break into sub-components.
27. **No 屎山 (shit mountain)** — refactor immediately when a file grows out of control.

### Feature Isolation (Rules 28-30)

28. When a new feature is SIMILAR to an existing one but not identical: **COPY the original, modify the copy.**
29. Never modify an existing stable module to handle two similar-but-different features. No `if (isNewFeature)` branches in stable code.
30. Shared logic goes into a shared util/hook — NOT into either feature's module.

### Backend — Gin Framework (Rules 31-40)

31. Router: `gin.Default()` (Logger + Recovery) for production, `gin.New()` for tests
32. Handlers: always `func(c *gin.Context)` — never use bare `http.HandlerFunc`
33. JSON binding: use `c.ShouldBindJSON(&struct)` with struct tags `json:"field" binding:"required"` — Gin auto-validates. Never `c.BindJSON()` (crashes).
34. Responses: `c.JSON(http.StatusXxx, gin.H{...})` — always use `net/http` status constants, never magic numbers.
35. Route groups: `v1 := r.Group("/api/v1")` — group by version/domain, apply shared middleware.
36. Middleware: `r.Use(middleware)` — auth, logging, CORS; use `c.Next()` to continue, `c.Abort()` to stop.
37. Error handling: rely on Gin's built-in Recovery middleware — do NOT write custom panic handlers.
38. Project structure: `routes/` → `handlers/` → `services/` → `models/` — each layer thin and focused.
39. Port: default `:8080` via `r.Run()` — never hardcode in handler logic. FbAi uses `:9090`.
40. Database: PostgreSQL via `pgx` driver (`github.com/jackc/pgx/v5`) — connection pool (`pgxpool`), env-based DSN (`os.Getenv("DATABASE_URL")`), never hardcode credentials.

### Input Validation — Frontend AND Backend (Rule 41)

41. **All user input MUST be validated on BOTH frontend AND backend.**
    - Frontend: Element Plus `FormRules` on every `<ElForm>`. Validate before submit: `formRef.value.validate()`.
    - Backend: Gin `binding:` struct tags on every request struct (`required`, `min`, `max`, `email`, `url`, `oneof`, `gte`, `lte`).
    - Business validation AFTER `ShouldBindJSON` (cross-field checks, uniqueness, permission checks).
    - **SQL injection**: parameterized queries ONLY (`$1`, `$2`). Never string concatenation.
    - **XSS**: Vue auto-escapes `{{ }}`. Never `v-html` with user input. If unavoidable, sanitize first.

### GitHub Auto-Push (Rules 42-43)

42. After every completed feature or bugfix: Code → verify (lint, build) → commit → push to GitHub. Never end a session with un-pushed verified code.
43. Commit granularity: one commit per logical change. No "WIP" commits on master. Always use `pnpm commit` for conventional format.

### Test After Every Change (Rule 44)

44. **每次改完代码都要跑一次测试**。无论是前端还是后端，每次修改完成后必须执行验证：
    - **前端改动**：`pnpm lint`（ESLint 检查）+ 浏览器实际测试
    - **后端改动**：`go build`（编译通过）+ API 端点测试（curl/urllib 验证返回值）
    - **前后端联调**：启动后端 → 前端发起真实请求 → 检查响应格式和业务逻辑
    - **绝不能跳过测试直接提交**。即使改动看起来"很简单"，也要跑一遍验证。
    - 测试失败 → 修复 → 重新测试 → 通过后才能 commit + push。

---

## 5. Critical Pitfalls (TOP 35)

### Backend Pitfalls

1. **Backend struct change → MUST rebuild AND restart** server.exe. Gin's `ShouldBindJSON` silently drops unknown JSON fields — no error, no warning. Frontend `ElMessage.success()` masks this. Always rebuild + restart + curl-test after backend changes.

2. **Go proxy in China**: Default `proxy.golang.org` blocked. Always prefix: `GOPROXY=https://goproxy.cn,direct go build ...`

3. **Port conflicts**: Always check with `netstat -ano | grep ":9090"` before starting. Stale processes keep old binary running.

4. **Backend restart full workflow** (every step mandatory):

   ```bash
   # 1. Kill old server
   taskkill //F //IM server.exe 2>/dev/null
   # 2. Verify port is free (NO "LISTENING" lines)
   netstat -ano 2>/dev/null | grep ":9090" | grep LISTENING
   # 3. If still LISTENING, kill by PID
   taskkill //F //PID <pid>
   # 4. Start new build
   cd /e/FbAi/art-design-server && ./server.exe &
   # 5. Health check
   curl -s http://localhost:9090/api/v1/ping
   ```

5. **Terminal curl password/token masking**: `terminal()` silently replaces passwords/tokens in curl with `***`. Use `execute_code` with Python `urllib.request` for authenticated API tests.

6. **`patch replace_all: true` corrupts Go files**: Go structs have repeated field patterns (`Status int`, `SortOrder int`). `replace_all: true` matches across ALL structs, corrupting `CreatedAt`/`UpdatedAt` fields. Never use `replace_all: true` in Go model files. Use unique patterns with surrounding context.

7. **Backend struct tag field alignment (DB → Go → Frontend)**: When DB column names differ from frontend expectations, use Go struct tags to bridge:

   ```go
   Email string `json:"userEmail" gorm:"column:email"`     // DB: email, Frontend: userEmail
   Phone string `json:"userPhone" gorm:"column:phone"`     // DB: phone, Frontend: userPhone
   CreatedAt time.Time `json:"createTime" gorm:"column:created_at"`
   ```

8. **`ElSelect` `:value` undefined → silent failure**: When backend returns `id` but template uses `role.roleId`, every option's `:value` is `undefined` and clicking does nothing. Always normalize field names in data loader.

9. **Backend `MenuTree` format ≠ frontend `AppRouteRecord`**: `/api/v1/auth/menus` returns Go `MenuTree` with flat fields (`title`, `icon`, `hidden`), but frontend expects `AppRouteRecord` with nested `meta: { isIframe, title, icon, isHide }`. Without `transformBackendMenu()`, `route.meta` is `undefined` → `TypeError: Cannot read properties of undefined (reading 'isIframe')`.

### Permission / Menu Pitfalls

10. **User `role_name` must not be empty**: `user_service.go` originally hardcoded `role_name=''`. This causes `GET /api/user/info` to return `roles=null` → frontend skips `filterMenuByRoles()` → **ALL users see ALL menus**. Fix: Create/Update must look up `role_name` from role table; GetUserInfoHandler must fallback to `role_id` lookup.

11. **Orphan menus — child in `menu_ids` but parent missing**: `TreeByIDs()` filters by idSet then builds tree. If a child menu (e.g., User id=4, parent=System id=2) is in `menu_ids` but its parent is NOT, the child silently disappears. **Fix**: `TreeByIDs()` must auto-include all ancestor menus by walking up the parent chain. **Verification**: always print FULL tree with all children, not just top-level count.

12. **Super admin `Tree()` returns ALL menus**: Previously super admin used `Tree()` (returns all menus including tenant-only). **Fix**: super admin must also use `TreeByIDs(role.MenuIDs)` so menus are filtered by R_SUPER role's explicit `menu_ids`. Never use `Tree()` as a bypass.

13. **Menu name synchronization (DB ↔ Frontend routes)**: Every frontend route `name` MUST have matching row in DB `menus` table. Audit:

    ```bash
    grep -r 'name:' src/router/modules/ | grep -oP "name:\\s*'[^']+'" | sort -u
    # Compare against:
    psql -c "SELECT name FROM menus ORDER BY id;"
    ```

14. **Menu icons need `ri:` prefix**: DB stores plain names like `"dashboard"`, but `ArtSvgIcon` needs `ri:dashboard-line`. Must map in `MenuProcessor.transformBackendMenu()`.

15. **Menu titles need i18n key mapping**: `/api/v1/auth/menus` returns raw titles from DB. Sidebar `formatMenuTitle()` only translates titles starting with `menus.`. Must map raw names to i18n keys in `MENU_I18N_MAP` inside `MenuProcessor.transformBackendMenu()`.

16. **PostgreSQL `menus.title` is NOT NULL**: INSERT must include a `title` value. Audit with `\d menus`.

17. **`VITE_ACCESS_MODE` change needs frontend restart**: Changing `.env` does NOT hot-reload with Vite. Kill and restart `pnpm dev`.

18. **ElTree `node-key` type must match checked keys AND data source**: Element Plus Tree with `show-checkbox` — type mismatch silently fails.

    | Mode       | Data source                | Has `id`? | `node-key`         |
    | ---------- | -------------------------- | --------- | ------------------ |
    | `frontend` | `asyncRoutes` (route defs) | **NO**    | `"name"`           |
    | `backend`  | DB via API                 | **YES**   | `"id"` or `"name"` |

    When backend stores numeric IDs but frontend uses `"name"` as node-key, build bidirectional ID↔name map:

    ```typescript
    const res = await fetchGetRoleMenus(roleId)
    const idToName = new Map(res.allMenus.map((m) => [m.id, m.name]))
    const names = res.roleMenus.map((id) => idToName.get(id)).filter(Boolean)
    treeRef.value?.setCheckedKeys(names)
    ```

### Multi-Tenant Pitfalls

19. **`roles` table RLS DISABLED** (2026-06-16): `ALTER TABLE roles DISABLE ROW LEVEL SECURITY`. Reason: pgxpool connection pool reuses connections → `set_config('app.current_tenant_id')` doesn't propagate between connections → super admin saw only 3 global roles instead of all 5. Primary defense: **explicit `WHERE tenant_id = $1` in service-layer SQL**.

20. **RLS + pgxpool connection pool is unreliable — rely on application-level WHERE clauses**: PostgreSQL RLS depends on session-level `current_setting('app.custom_param')`. With `pgxpool`, connections are reused across requests, and a stale GUC value from a previous tenant request can leak into a super admin query. **Solution**: explicit `WHERE tenant_id = $1` in every service-layer SQL query. RLS is secondary best-effort only.

21. **`TenantContext` middleware must ALWAYS set `app.current_tenant_id`** — even for super admin (`tenantID=nil` → set to empty string `''`). Previously skipped for super admin, leaving stale GUC from previous connection. Fix:

    ```go
    val := ""
    if tenantID != nil { val = fmt.Sprintf("%d", *tenantID) }
    _, _ = db.Pool.Exec(c.Request.Context(),
        "SELECT set_config('app.current_tenant_id', $1, true)", val)
    ```

22. **RLS `current_setting` empty string cast crash**: `''::INT` throws `invalid input syntax for type integer`. Must wrap with `NULLIF`:

    ```sql
    -- ✅ Correct
    USING (NULLIF(current_setting('app.current_tenant_id', true), '')::INT = tenant_id)
    ```

23. **Tenant `role_code` MUST be unique across all tenants**: Using global codes like `"R_USER"` causes `duplicate key violates unique constraint`. Use tenant-prefixed codes:

    ```go
    roleCode := fmt.Sprintf("T%d_R_ADMIN", tenantID)   // → "T5_R_ADMIN"
    roleCode := fmt.Sprintf("T%d_R_USER", tenantID)     // → "T5_R_USER"
    ```

24. **Tenant role creation MUST include `menu_ids`**: `tenant_service.go` INSERT without `menu_ids` → defaults to `{}` → `TreeByIDs([])` → `data: null` → frontend route registration fails → Exception500.
    - **Tenant admin role `menu_ids`**: `{1, 3, 8, 9, 10, 11}` (Dashboard, Console, TenantSystem + children — NO global System menu)
    - **Tenant user role `menu_ids`**: `{1, 3}` (Dashboard, Console only)

25. **Handler `tenantID` auto-bind — every Create handler MUST copy `RoleHandler.Create` pattern**: `UserHandler.Create` was originally MISSING this, causing tenant users created with `tenant_id=NULL` (global scope) — a silent data isolation failure:

    ```go
    // After ShouldBindJSON, MUST add this block:
    if req.TenantID == nil {
        if tid, exists := c.Get("tenantID"); exists {
            if t, ok := tid.(*uint); ok && t != nil {
                req.TenantID = t
            }
        }
    }
    ```

26. **Tenant data isolation: rely on explicit service-layer WHERE clauses, NOT just RLS**: Every service method querying tenant-scoped tables must:
    - Accept `tenantID *uint` parameter from handler (read from `c.Get("tenantID")`)
    - Add `WHERE tenant_id = $n` when `tenantID != nil`
    - Fall back to no filter when `tenantID == nil` (global/super admin view)
    - **Role list for tenant context**: `WHERE tenant_id = $1` (NOT `WHERE tenant_id IS NULL OR tenant_id = $1`) — tenant admins shouldn't see global roles.

### Frontend / Windows Pitfalls

27. **`write_file`/`patch` on Windows — always use absolute paths**: MSYS-style paths like `/e/FbAi/...` resolve to `E:\e\FbAi\...` (double drive letter). Files land in wrong directory, `read_file` returns "not found". **Always use**: `E:\FbAi\art-design-pro\...`

28. **`write_file`/`patch` silently corrupt credentials**: Replaces passwords with `***` in file content. Use `terminal` to create/edit `.env` files. NEVER hardcode credentials in Go source code — they will be silently corrupted. Read DSN via `os.Getenv`.

29. **`useTable` requires paginated API — flat arrays break type inference**: `useTable` infers `TRecord` from `InferApiResponse<TApiFn>` expecting `PaginatedResponse`. If API returns flat array, TS infers `TRecord = never`. Fix: wrap in paginated adapter:

    ```typescript
    const fetchPaged = async (_p: any) => {
      const list = await fetchGetXxxList()
      return { list: list || [], total: list.length, page: 1, size: list.length }
    }
    useTable({ core: { apiFn: fetchPaged, ... } })
    ```

30. **Password minimum 6 characters**: Backend `binding:"min=6"`. Test passwords must be ≥6 chars.

31. **psql inline commands with Chinese characters fail in MSYS bash**: `psql -c "INSERT ... VALUES ('中文')"` produces UTF-8 encoding error. Always write Chinese-content SQL to a `.sql` file and execute with `psql -f file.sql`.

32. **Token generation — NEVER use `string(rune())`**: `string(rune(userID))` produces Unicode control characters (U+0001) that break JSON serialization and HTTP headers. Always use `fmt.Sprintf`.

33. **Go toolchain auto-upgrade**: Gin v1.12+ requires Go ≥ 1.25. If host has Go 1.21, `go build` auto-downloads go1.25 toolchain (100MB+ on first build).

34. **Windows PostgreSQL pg_hba.conf dual-address**: `localhost` resolves to `::1` (IPv6) first. Update BOTH `127.0.0.1/32` and `::1/128` lines in pg_hba.conf, or connect via `-h 127.0.0.1`.

35. **No date library — use native `new Date()`**: The project has no dayjs/moment. Parse ISO strings and format manually:
    ```typescript
    const d = new Date(row.createTime)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    ```

---

## 6. Database Schema (Core Tables)

| Table | Key Fields | Notes |
| --- | --- | --- |
| `users` | id, user_name, password (bcrypt), email, phone, role_id, role_name, tenant_id, status, created_at | tenant_id=NULL for global admin |
| `roles` | id, role_name, role_code, menu_ids (int[]), tenant_id, status, created_at | role_code UNIQUE; **RLS DISABLED** on this table |
| `menus` | id, name, title, parent_id, path, component, icon, sort_order, hidden, created_at | title is NOT NULL |
| `sessions` | id, user_id, token, refresh_token, expires_at, tenant_id | SSO: new login deletes old sessions |
| `tenants` | id, name, code, contact_name, contact_phone, status, created_at | — |
| `permissions` | id, code, module, action, description, created_at | 18 permission points: `system:user:*` (5), `system:role:*` (4), `system:menu:*` (4), `system:tenant:*` (3), `dashboard:*` (2) |
| `role_permissions` | role_id, permission_id | M:N junction; R_SUPER=18 perms, R_ADMIN=11, R_USER=1 |

---

## 7. API Endpoint Map (Frontend ↔ Backend)

| Frontend Function | URL | Go Handler | Status |
| --- | --- | --- | --- |
| `fetchLogin()` | `POST /api/v1/auth/login` | `AuthHandler.Login` | ✅ |
| `fetchGetUserInfo()` | `GET /api/user/info` | `UserHandler.GetInfo` | ✅ |
| `fetchGetMenuList()` | `GET /api/v1/auth/menus` | `AuthHandler.GetMenus` (role-filtered) | ✅ |
| `fetchGetUserList()` | `GET /api/user/list` → `/api/v1/users` | `UserHandler.List` | ✅ |
| `fetchCreateUser()` | `POST /api/v1/users` | `UserHandler.Create` | ✅ |
| `fetchUpdateUser()` | `PUT /api/v1/users/:id` | `UserHandler.Update` | ✅ |
| `fetchDeleteUser()` | `DELETE /api/v1/users/:id` | `UserHandler.Delete` | ✅ |
| `fetchGetRoleList()` | `GET /api/role/list` → `/api/v1/roles` | `RoleHandler.List` | ✅ |
| `fetchCreateRole()` | `POST /api/v1/roles` | `RoleHandler.Create` | ✅ |
| `fetchUpdateRole()` | `PUT /api/v1/roles/:id` | `RoleHandler.Update` | ✅ |
| `fetchDeleteRole()` | `DELETE /api/v1/roles/:id` | `RoleHandler.Delete` | ✅ |
| `fetchGetRoleMenus()` | `GET /api/v1/roles/:id/menus` | `RoleHandler.GetMenus` | ✅ |
| `fetchCreateMenu()` | `POST /api/v1/menus` | `MenuHandler.Create` | ✅ |
| `fetchUpdateMenu()` | `PUT /api/v1/menus/:id` | `MenuHandler.Update` | ✅ |
| `fetchDeleteMenu()` | `DELETE /api/v1/menus/:id` | `MenuHandler.Delete` | ✅ |

**Backward-compatible API aliases** (legacy frontend paths preserved):

- `GET /api/user/list` → delegates to `/api/v1/users`
- `GET /api/role/list` → delegates to `/api/v1/roles`
- `GET /api/v3/system/menus/simple` → delegates to menu tree endpoint

---

## 8. Key Architecture Decisions

- **Composition API with `<script setup>`** — clean, tree-shakeable components
- **Element Plus auto-import** — no manual component registration (unplugin-vue-components + unplugin-auto-import)
- **SCSS module + Tailwind hybrid** — SCSS for component styles, Tailwind for layout/utilities
- **Pinia with persistedstate** — state survives page reloads via localStorage
- **Backend access mode**: `VITE_ACCESS_MODE=backend` — menus dynamically served from DB, filtered by user's role `menu_ids`
- **Multi-tenant architecture**: dual-layer isolation — service-level `WHERE tenant_id` (primary) + RLS (best-effort, disabled on roles table)
- **Feature isolation**: tenant-system pages are copies of system pages (`views/system/` → `views/tenant-system/`), not modifications — independent maintenance
- **`roles` table RLS DISABLED**: pgxpool connection pool + `set_config()` unreliable → application-level WHERE clauses are the sole defense
- **`users` table RLS retained** as secondary defense, but all queries also use explicit `WHERE tenant_id`

---

## 9. Module Dependency Flow

```
auth (login) → dashboard (console) → system (users/roles/menus/tenant)  [超级管理员]
                                   → tenant-system (users/roles/menu)   [租户管理员]
                                   → user-center
                                   → exception pages
```

---

## 10. Pre-Code Checklist

Before writing ANY code, verify:

- [ ] **CodeGraph**: Have I queried paths, callers, callees, and impact?
- [ ] **Access Mode**: `VITE_ACCESS_MODE=backend` — menus come from DB, not frontend routes
- [ ] **Menu-DB sync**: If adding/modifying frontend routes, does DB `menus` have matching `name`? Are i18n keys mapped?
- [ ] **DB constraints**: Check `\d <table>` for NOT NULL columns before INSERT
- [ ] **Element Plus**: Does Element Plus already provide this component?
- [ ] **Table pattern**: Using `ArtTable` + `ArtTableHeader` + `useTable()`? Never raw `<ElTable>`
- [ ] **Button style**: Header = plain `<ElButton v-ripple>`, dialog = plain cancel + primary confirm
- [ ] **i18n**: All user-facing text using `$t()`?
- [ ] **Single root**: Template has single root element?
- [ ] **API format**: Response format `{ code, msg, data }` respected?
- [ ] **Permissions**: Using `hasAuth()` or `v-auth`/`v-roles`?
- [ ] **Modularization**: Each file under 300 lines? Logic extracted to hooks/utils?
- [ ] **Feature isolation**: Similar feature? COPY, don't MODIFY
- [ ] **Documentation**: Created `docs/dev/YYYY-MM-DD--<feature>.md`?
- [ ] **Input validation**: Both frontend (`FormRules`) AND backend (`binding:` tags) on all user inputs?
- [ ] **Tenant context**: New handler has `tenantID` auto-bind? Service has `WHERE tenant_id` filter?
- [ ] **Backend restart**: If backend code changed, rebuilt + killed old process + verified port free + restarted + curl tested?
- [ ] **Git push**: Feature complete and verified? Push NOW.
- [ ] **Lint**: Will pre-commit hooks pass (ESLint + Prettier + Stylelint)?

---

## 11. Authoritative Documentation Links

- **art-design-pro docs**: https://www.artd.pro/docs/zh/guide/introduce.html
  - 必读: https://www.artd.pro/docs/zh/guide/must-read.html
  - 组件库: https://www.artd.pro/docs/zh/guide/essentials/element-plus.html
  - 规范: https://www.artd.pro/docs/zh/guide/project/standard.html
  - 权限: https://www.artd.pro/docs/zh/guide/in-depth/permission.html
  - 国际化: https://www.artd.pro/docs/zh/guide/in-depth/locale.html
- **Element Plus**: https://element-plus.org/zh-CN/component/overview.html
- **Gin (backend)**: https://gin-gonic.com/zh-cn/docs/

---

## 12. Vue Component Template

```vue
<script setup lang="ts">
  // 1. Imports (Vue, libs, components)
  // 2. i18n: const { t } = useI18n()
  // 3. Props/Emits
  // 4. Composables/Stores (useAuth from @/hooks/core/useAuth for permissions)
  // 5. Reactive state
  // 6. Computed
  // 7. Methods
  // 8. Lifecycle hooks
</script>

<template>
  <!-- MUST have SINGLE root element (multi-root → blank pages from Transition) -->
  <div>
    <!-- ALL text uses $t(): {{ $t('key.path') }} -->
    <!-- ALL UI uses Element Plus <el-*> components -->
    <!-- Permissions: v-if="hasAuth('authMark')" or v-auth="'authMark'" -->
  </div>
</template>

<style lang="scss" scoped>
  // Component-scoped SCSS
  // Use Element Plus CSS variables: var(--el-color-primary)
  // Global Element Plus tuning is in src/assets/styles/el-ui.scss — do NOT override elsewhere
</style>
```

---

## 13. Standard Table Page Pattern

```vue
<template>
  <div class="xxx-page art-full-height">
    <!-- Search bar (optional) -->
    <XxxSearch v-model="queryParams" @search="getData" />

    <ElCard class="art-table-card">
      <!-- Table header: column visibility + action buttons -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton @click="showAdd" v-ripple>新增</ElButton>
            <ElButton @click="handleBatchDelete" v-ripple>批量删除</ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- Table: use ArtTable, NOT ElTable -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- Dialog (extracted to separate component) -->
    <XxxDialog
      v-model:visible="dialogVisible"
      :mode="dialogMode"
      :row="currentRow"
      @success="refreshData"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { h } from 'vue'

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchXxxList,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'selection', width: 55 },
        { type: 'index', label: '#', width: 60 },
        { prop: 'name', label: '名称', minWidth: 150 },
        { prop: 'status', label: '状态', minWidth: 100 },
        {
          label: '操作',
          width: 160,
          fixed: 'right',
          formatter: (row: any) =>
            h('div', [
              h(ArtButtonTable, { type: 'edit', onClick: () => openDialog('edit', row) }),
              h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) }),
              h(ArtButtonTable, {
                type: 'view',
                icon: 'ri:shield-keyhole-line',
                onClick: () => openPerm(row)
              })
            ])
        }
      ]
    }
  })

  onMounted(() => getData())
</script>
```

---

## 14. Permission Button Patterns

```vue
<!-- Method 1: hasAuth() composable -->
<ElButton v-if="hasAuth('add')">添加</ElButton>

<!-- Method 2: v-auth directive (backend mode) -->
<ElButton v-auth="'add'">添加</ElButton>

<!-- Method 3: v-roles directive -->
<ElButton v-roles="['R_SUPER', 'R_ADMIN']">按钮</ElButton>
```

```typescript
import { useAuth } from '@/hooks/core/useAuth'
const { hasAuth } = useAuth()
```

---

## 15. Gin Backend Patterns

### Full Handler Template (with tenant auto-bind + validation)

```go
// Request struct with validation
type CreateUserRequest struct {
    UserName string `json:"userName" binding:"required,min=2,max=50"`
    Password string `json:"password" binding:"required,min=6,max=100"`
    Email    string `json:"email" binding:"required,email"`
    RoleID   uint   `json:"roleId" binding:"required,gt=0"`
    Status   int    `json:"status" binding:"oneof=0 1"`
    TenantID *uint  `json:"tenantId"`  // nullable — auto-bound by handler
}

// Handler
func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, models.Error(400, "参数校验失败: "+err.Error()))
        return
    }

    // ✅ CRITICAL: Auto-bind tenant_id from Gin context
    if req.TenantID == nil {
        if tid, exists := c.Get("tenantID"); exists {
            if t, ok := tid.(*uint); ok && t != nil {
                req.TenantID = t
            }
        }
    }

    // Business validation after ShouldBindJSON
    if exists := services.DefaultUserService.ExistsByName(req.UserName); exists {
        c.JSON(http.StatusConflict, models.Error(409, "用户名已存在"))
        return
    }

    user, err := services.DefaultUserService.Create(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, models.Error(500, "创建失败: "+err.Error()))
        return
    }

    c.JSON(http.StatusCreated, models.Success(user))
}
```

### Route Group Setup

```go
func SetupRoutes(r *gin.Engine) {
    // Public
    r.POST("/api/v1/auth/login", handlers.DefaultAuthHandler.Login)

    // Protected
    v1 := r.Group("/api/v1")
    v1.Use(middleware.AuthRequired())
    v1.Use(middleware.TenantContext())       // Sets c.Get("tenantID")
    {
        // Users
        v1.GET("/users", middleware.RequirePermission("system:user:list"), handlers.DefaultUserHandler.List)
        v1.POST("/users", middleware.RequirePermission("system:user:create"), handlers.DefaultUserHandler.Create)
        v1.PUT("/users/:id", middleware.RequirePermission("system:user:edit"), handlers.DefaultUserHandler.Update)
        v1.DELETE("/users/:id", middleware.RequirePermission("system:user:delete"), handlers.DefaultUserHandler.Delete)

        // Roles
        v1.GET("/roles", middleware.RequirePermission("system:role:list"), handlers.DefaultRoleHandler.List)
        v1.POST("/roles", middleware.RequirePermission("system:role:create"), handlers.DefaultRoleHandler.Create)
        v1.PUT("/roles/:id", middleware.RequirePermission("system:role:edit"), handlers.DefaultRoleHandler.Update)
        v1.DELETE("/roles/:id", middleware.RequirePermission("system:role:delete"), handlers.DefaultRoleHandler.Delete)
        v1.GET("/roles/:id/menus", handlers.DefaultRoleHandler.GetMenus)

        // Menus
        v1.GET("/menus", middleware.RequirePermission("system:menu:list"), handlers.DefaultMenuHandler.List)
        v1.POST("/menus", middleware.RequirePermission("system:menu:create"), handlers.DefaultMenuHandler.Create)
        v1.PUT("/menus/:id", middleware.RequirePermission("system:menu:edit"), handlers.DefaultMenuHandler.Update)
        v1.DELETE("/menus/:id", middleware.RequirePermission("system:menu:delete"), handlers.DefaultMenuHandler.Delete)

        // Auth
        v1.GET("/auth/menus", handlers.DefaultAuthHandler.GetMenus)
    }

    // Backward-compatible aliases
    r.GET("/api/user/list", middleware.AuthRequired(), handlers.DefaultUserHandler.List)
    r.GET("/api/role/list", middleware.AuthRequired(), handlers.DefaultRoleHandler.List)
}
```

### Service Layer Tenant Filter Pattern

```go
func (s *UserService) List(tenantID *uint, page, size int, keyword string) ([]models.User, int64) {
    // ✅ Always explicitly filter by tenant_id, never rely on RLS alone
    if tenantID == nil {
        // Global view (super admin): all users
        countSQL = `SELECT COUNT(*) FROM users WHERE user_name ILIKE $1`
    } else {
        // Tenant view: only users in this tenant
        countSQL = `SELECT COUNT(*) FROM users WHERE user_name ILIKE $1 AND tenant_id = $2`
    }
    // ...
}
```

---

## 16. Database Connection Pattern

```go
// db/postgres.go
package db

import (
    "context"
    "os"
    "github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Connect() error {
    dsn := os.Getenv("DATABASE_URL") // NEVER hardcode — will be corrupted by Hermes tools
    var err error
    Pool, err = pgxpool.New(context.Background(), dsn)
    return err
}

func Close() {
    Pool.Close()
}
```

```go
// main.go
func main() {
    if err := db.Connect(); err != nil {
        log.Fatal("DB connect failed:", err)
    }
    defer db.Close()

    r := gin.Default()
    routes.SetupRoutes(r)
    r.Run(":9090")
}
```

**Key DB facts**:

- DSN format: `postgres://fbai:***@127.0.0.1:5432/fbai?sslmode=disable` (in `art-design-server/.env`, gitignored)
- Always use `127.0.0.1` (not `localhost`) to avoid IPv6 pg_hba mismatch on Windows
- Credentials in `.env` file (gitignored), NEVER in Go source code — `write_file`/`patch` will corrupt them
- `\.env.example` committed for reference (without real credentials)
- bcrypt password hashing in `art-design-server/crypto/password.go` (shared by services + middleware)

---

## 17. Tenant System Management Patterns

### Feature Isolation: Copy, Don't Modify

```
src/views/system/           (super admin — stable, tested, deployed)
    ├── user/index.vue
    ├── role/index.vue
    └── menu/index.vue

        ↓ COPY (not modify)

src/views/tenant-system/    (tenant admin — independent copy)
    ├── user/index.vue      (modified: tenant-scoped API, no global roles in select)
    ├── role/index.vue      (modified: sub-role creation, permission subset)
    └── menu/index.vue      (modified: read-only view, menus are global resource)
```

### Tenant Creation Transaction (Go)

```go
func (s *TenantService) Create(req *CreateTenantRequest) (*models.Tenant, error) {
    tx, _ := db.Pool.Begin(context.Background())
    defer tx.Rollback(context.Background())

    // 1. Insert tenant
    var tenantID uint
    tx.QueryRow(context.Background(),
        `INSERT INTO tenants (name, code, contact_name, contact_phone) VALUES ($1,$2,$3,$4) RETURNING id`,
        req.Name, req.Code, req.ContactName, req.ContactPhone,
    ).Scan(&tenantID)

    // 2. Create tenant admin role — MUST include menu_ids
    tx.Exec(context.Background(),
        `INSERT INTO roles (role_name, role_code, menu_ids, tenant_id, status)
         VALUES ($1, $2, $3, $4, 1)`,
        "租户管理员",
        fmt.Sprintf("T%d_R_ADMIN", tenantID),     // ✅ unique across tenants
        "{1,3,8,9,10,11}",                        // ✅ MUST include menu_ids
        tenantID,
    )

    // 3. Create tenant user role
    tx.Exec(context.Background(),
        `INSERT INTO roles (role_name, role_code, menu_ids, tenant_id, status)
         VALUES ($1, $2, $3, $4, 1)`,
        "普通用户",
        fmt.Sprintf("T%d_R_USER", tenantID),      // ✅ unique across tenants
        "{1,3}",                                   // ✅ Dashboard + Console only
        tenantID,
    )

    // 4. Create admin user account
    // ...

    tx.Commit(context.Background())
    return tenant, nil
}
```

### Tenant Role `menu_ids` Reference

| Role | menu_ids | Menus Included |
| --- | --- | --- |
| Global R_SUPER | `{1,2,3,4,5,6,7,8,9,10,11,...}` | All menus |
| Global R_ADMIN | `{1,2,3,4,5,6,7}` | Dashboard + System (full) |
| Global R_USER | `{1,3}` | Dashboard + Console |
| Tenant Admin (`T*_R_ADMIN`) | `{1,3,8,9,10,11}` | Dashboard + Console + TenantSystem (full) |
| Tenant User (`T*_R_USER`) | `{1,3}` | Dashboard + Console only |

---

## 18. Menu Transformation Patterns

### `MenuProcessor.transformBackendMenu()` — Full Pattern

This is THE critical bridge between Go backend `MenuTree` (flat fields) and frontend `AppRouteRecord` (nested `meta`). Without it, the sidebar crashes.

```typescript
// In src/router/menus/MenuProcessor.ts or similar
private readonly MENU_I18N_MAP: Record<string, string> = {
  Dashboard: 'menus.dashboard.title',
  Console: 'menus.dashboard.console',
  System: 'menus.system.title',
  User: 'menus.system.user',
  Role: 'menus.system.role',
  Menu: 'menus.system.menu',
  Tenant: 'menus.system.tenant',
  TenantSystem: 'menus.tenantSystem.title',
  TenantUser: 'menus.tenantSystem.user',
  TenantRole: 'menus.tenantSystem.role',
  TenantMenu: 'menus.tenantSystem.menu',
  Exception: 'menus.exception.title',
  Result: 'menus.result.title',
  UserCenter: 'menus.userCenter.title',
}

private readonly MENU_ICON_MAP: Record<string, string> = {
  dashboard: 'ri:dashboard-line',
  console: 'ri:pie-chart-line',
  system: 'ri:settings-3-line',
  user: 'ri:user-settings-line',
  role: 'ri:admin-line',
  menu: 'ri:menu-line',
  tenant: 'ri:building-line',
  exception: 'ri:error-warning-line',
  result: 'ri:checkbox-circle-line',
}

private transformBackendMenu(menus: any[]): AppRouteRecord[] {
  return menus.map((menu) => ({
    id: menu.id,
    name: menu.name,
    path: menu.path || '',
    component: menu.component || '',
    meta: {
      title: this.MENU_I18N_MAP[menu.name] || menu.title || menu.name || '',
      icon: this.MENU_ICON_MAP[menu.icon] || (menu.icon ? `ri:${menu.icon}` : ''),
      isHide: menu.hidden || false,
      isIframe: false,
    },
    children: menu.children?.length ? this.transformBackendMenu(menu.children) : undefined,
  })) as AppRouteRecord[]
}
```

---

## 19. Field Alignment & Data Transformation Patterns

### DB → Go → Frontend Bridging

```go
// Go model — struct tags bridge DB column names to frontend field names
type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    UserName  string    `json:"userName" gorm:"column:user_name;size:64"`     // DB: user_name
    Email     string    `json:"userEmail" gorm:"column:email;size:128"`       // DB: email → frontend: userEmail
    Phone     string    `json:"userPhone" gorm:"column:phone;size:20"`        // DB: phone → frontend: userPhone
    RoleID    uint      `json:"roleId" gorm:"column:role_id"`                 // DB: role_id → frontend: roleId
    CreatedAt time.Time `json:"createTime" gorm:"column:created_at"`          // DB: created_at → frontend: createTime
    Status    int       `json:"status" gorm:"column:status;default:1"`        // DB: status (int) → frontend: status (int)
}
```

### Frontend Data Transformer (enum/type conversion)

```typescript
// In page component — transform backend data to frontend format
const { data } = useTable({
  core: {
    apiFn: fetchXxxList,
    transform: {
      dataTransformer: (records) =>
        records.map((item: any) => ({
          ...item,
          roleId: item.id || item.roleId, // field rename
          enabled: item.status === 1 // int → boolean
        }))
    }
  }
})
```

### ElSelect Gotcha — Always Normalize Field Names

```vue
<!-- ❌ BROKEN — backend returns 'id' but template uses 'role.roleId' → :value=undefined -->
<ElOption v-for="role in roleList" :key="role.id" :label="role.roleName" :value="role.roleId" />

<!-- ✅ FIX — normalize in data loader before passing to template -->
<ElOption v-for="role in roleList" :key="role.id" :label="role.roleName" :value="role.id" />
```

---

## 20. Backend Build & Restart Workflow

```bash
# Full verified restart sequence — every step is mandatory
cd E:/FbAi/art-design-server

# 1. Build with China proxy
GOPROXY=https://goproxy.cn,direct go build -o server.exe ./main.go

# 2. Kill ALL existing server.exe processes
taskkill //F //IM server.exe 2>/dev/null

# 3. Verify port is free (must show NO lines with "LISTENING")
netstat -ano 2>/dev/null | grep ":9090" | grep LISTENING
# If any LISTENING lines appear, kill by PID:
# taskkill //F //PID <pid>

# 4. Start new server in background
./server.exe &

# 5. Wait 1-2 seconds, then health check
sleep 2
curl -s http://localhost:9090/api/v1/ping
# Expected: {"code":200,"msg":"pong","data":null}

# 6. Login test (use Python, not curl — curl is token-masked by Hermes)
python3 -c "
import urllib.request, json
req = urllib.request.Request('http://localhost:9090/api/v1/auth/login',
    data=json.dumps({'userName':'admin','password':'admin123'}).encode(),
    headers={'Content-Type':'application/json'})
resp = json.loads(urllib.request.urlopen(req).read())
print('Login OK, token:', resp['data']['token'][:20] + '...')
"
```

---

## 21. Common Reusable Patterns

### Password Input with Generate + Show/Hide

```vue
<ElInput v-model="form.password" type="password" show-password>
  <template #append>
    <ElButton @click="generatePassword">生成</ElButton>
  </template>
</ElInput>
```

### Backend Optional Password Update (conditional SQL)

```go
if req.Password != "" {
    hashedPwd, _ := crypto.HashPassword(req.Password)
    // Only update password if provided
    db.Pool.Exec(ctx,
        `UPDATE users SET user_name=$1, email=$2, password=$3 WHERE id=$4`,
        req.UserName, req.Email, hashedPwd, id,
    )
} else {
    // Password field omitted
    db.Pool.Exec(ctx,
        `UPDATE users SET user_name=$1, email=$2 WHERE id=$3`,
        req.UserName, req.Email, id,
    )
}
```

### Table Column Width Guidelines

```typescript
columnsFactory: () => [
  { type: 'selection', width: 55 }, // fixed
  { type: 'index', label: '#', width: 60 }, // fixed
  { prop: 'userName', label: '用户名', minWidth: 120 }, // responsive
  { prop: 'email', label: '邮箱', minWidth: 180 }, // responsive
  { prop: 'createTime', label: '创建时间', minWidth: 170 }, // responsive
  { label: '操作', width: 160, fixed: 'right' } // fixed right
]
```

### Native Date Formatting (no dayjs/moment)

```typescript
function formatDateTime(val: string): string {
  const d = new Date(val)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
```

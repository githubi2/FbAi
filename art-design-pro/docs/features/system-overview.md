# FbAi 项目系统梳理 — 代码逻辑 / 模块调用关系 / 页面风格

> 生成日期：2026-08-26
> 适用范围：`art-design-pro`（Vue 3 前端）+ `art-design-server`（Go/Gin 后端）
> 本文基于对当前代码库的静态阅读整理，是理解本项目整体结构的唯一权威入口文档。

---

## 1. 项目总体架构

双包仓库（独立包，无 monorepo 工具）：

| 包 | 技术栈 | 端口 | 职责 |
|---|---|---|---|
| `art-design-pro/` | Vue 3.5 + Vite 7 + TS 5.6 + Element Plus 2.11 + Pinia + Tailwind 4 | 3006 | 管理后台前端 |
| `art-design-server/` | Go 1.25 + Gin + pgx v5 (PostgreSQL 17) | 9090 | REST API 后端 |

- 前端 Vite dev server 代理 `/api` → `http://localhost:9090`
- 数据流向：**Vue 视图 → api/ 层 → utils/http（axios） → Gin 后端 → PostgreSQL**
- 访问模式：`VITE_ACCESS_MODE=backend` —— 菜单由数据库驱动（`GET /api/v1/auth/menus`），非前端路由硬编码

---

## 2. 前端代码逻辑与模块调用关系

### 2.1 启动链路（`src/main.ts`）

```
main.ts
├── createApp(App)
├── initStore(app)          # Pinia + persistedstate（版本化存储键 sys-v{ver}-{storeId}）
├── initRouter(app)         # createWebHashHistory + 静态路由 + 前后置守卫
├── setupGlobDirectives(app)# v-auth / v-roles / v-ripple / v-highlight
├── setupErrorHandle(app)   # 全局错误处理
└── app.use(language)       # vue-i18n（zh/en）
```

`App.vue`：`ElConfigProvider`（locale、z-index=3000、card shadow: never）包裹 `RouterView`；
`onBeforeMount` 初始化主题，`onMounted` 检查存储兼容性 + 系统升级 + 路由过渡开关。

### 2.2 路由体系（`src/router/`）

```
router/
├── index.ts                # createRouter + initRouter（注册前后置守卫）
├── routes/
│   ├── staticRoutes.ts     # 静态路由：Login / Register / ForgetPassword / 布局壳
│   └── asyncRoutes.ts      # 前端模式备用路由（含 meta.roles）
├── modules/                # 业务路由模块（聚合于 index.ts）
│   ├── dashboard.ts        # 工作台（Dashboard → Console）
│   ├── system.ts           # 系统管理（User / Role / Menu / Tenant / UserCenter）
│   ├── tenantSystem.ts     # 租户系统（TenantUser / TenantRole / TenantMenu）
│   ├── adAccount.ts        # FB 广告账户（List / Manage）
│   ├── bmManage.ts         # BM 管理
│   ├── pageManage.ts       # 主页管理
│   ├── pixelManage.ts      # 像素管理
│   ├── result.ts           # 成功/失败结果页
│   └── exception.ts        # 403 / 404 / 500
├── guards/
│   ├── beforeEach.ts       # 全局前置守卫（核心：登录态 + 动态路由注册 + 权限）
│   └── afterEach.ts        # 后置守卫（进度条结束等）
└── core/                   # 路由核心类
    ├── RouteRegistry.ts        # 动态路由注册/卸载
    ├── MenuProcessor.ts        # 菜单获取/转换/路径规范化（模块调用中心）
    ├── RouteTransformer.ts     # 路由元数据转换
    ├── RouteValidator.ts       # 路由合法性校验（component 检查等）
    ├── RoutePermissionValidator.ts # 目标路径权限校验
    ├── ComponentLoader.ts      # 字符串路径 → 动态 import 组件
    └── IframeRouteManager.ts   # iframe 外链路由管理
```

### 2.3 路由守卫流程（`beforeEach.ts` — 最易出问题的链路）

```
导航到任意页面
  → NProgress.start()
  → 1. handleLoginStatus：未登录且非静态路由 → 跳 Login（带 redirect）；已登录访问 /auth/login → 跳 /；token 过期 → logOut
  → 2. routeInitFailed？ → 直接 500 页面（防死循环）
  → 3. 路由未注册且已登录 → handleDynamicRoutes()：
       a. fetchGetUserInfo() → GET /api/user/info（存 userStore.info）
       b. menuProcessor.getMenuList() → backend 模式：GET /api/v1/auth/menus
          → transformBackendMenu()（name→MENU_I18N_MAP 标题 + MENU_ICON_MAP 图标）
          → filterEmptyMenus() → normalizeMenuPaths（相对路径拼完整路径 + 推导 redirect）
       c. validateMenuList + routeRegistry.register(menuList)
       d. menuStore.setMenuList + IframeRouteManager.save + worktab 校验
       e. RoutePermissionValidator.validatePath(to.path) → 无权限跳 homePath，有权限 replace 导航
     （初始化进行中时：waitForRouteInit() 排队，完成后 next({path, replace:true})，避免 next(false) 吞点击）
  → 4. 根路径 / → homePath 重定向
  → 5. 已匹配 → setWorktab + setPageTitle + next()
  → 6. 未匹配 → Exception404
```

出错分支：401 → 放行登出；其他错误 → `routeInitFailed=true` → `Exception500`（需刷新/重登重置）。

### 2.4 状态管理（Pinia，`src/store/modules/`）

| Store | 职责 | 持久化 |
|---|---|---|
| `user.ts` | 登录态、token（rememberMe=3天/否则24h 过期时间戳）、用户信息、语言、锁屏、登出清理 | ✅ localStorage `user` |
| `menu.ts` | 菜单列表、homePath、动态路由移除函数 | ✅ |
| `setting.ts` | 主题/布局等系统设置 | ✅ |
| `worktab.ts` | 多标签页（opened / keepAliveExclude） | ✅ |
| `tenant.ts` | 当前租户切换（TenantSwitcher） | ✅ |
| `table.ts` | 表格配置/列缓存 | ✅ |

**登出链路**：`userStore.logOut()` → 保存 lastUserId → 清空用户态 → 移除 iframe 缓存 → `menuStore.setHomePath('')` → `resetRouterState(500)`（卸载动态路由）→ 跳 Login。
**换用户登录**：`checkAndClearWorktabs()` 用 lastUserId 对比，不同用户则清空标签页。

### 2.5 API 层与 HTTP 封装

```
api/（按域拆分的薄函数）
├── auth.ts               # login / userinfo
├── system-manage.ts      # 用户/角色/菜单 CRUD（后端真实 API）
├── tenant.ts             # 租户 CRUD + 切换
├── facebook.ts           # FB OAuth / 广告账户 / BM / 页面 / 像素（全部 FB 端点）
└── index.ts
```

`utils/http/index.ts`（axios 实例，`baseURL=VITE_API_URL`）：
- **请求拦截器**：自动附加 `Authorization: Bearer <token>`；POST/PUT 无 data 时把 `params` 自动转成 body（兼容 `params:` 写法）。
- **响应拦截器**：`code===200` 解包返回 `res.data.data`（**调用方直接解构 data**）；`code!==200` 抛 `HttpError`（`{code,msg}`）；业务码 401 → 防抖登出（3s 窗口）。
- 超时 15s，失败重试默认关（MAX_RETRIES=0）。
- `showSuccessMessage/showErrorMessage` 控制消息提示。
- **FB 端点不再走前端限速队列**（2026-07 起：前端直接读后端缓存，毫秒级；真正的 FB API 由后端后台串行刷新）。

### 2.6 Hooks（`src/hooks/core/`）

| Hook | 用途 |
|---|---|
| `useTable.ts` | **表格核心方案**：API 请求/分页/搜索防抖/列管理/缓存/5 种刷新策略，自动类型推导 |
| `useTableColumns.ts` | 列配置工厂（动态显隐/排序/持久化） |
| `useTableHeight.ts` | 表格高度自适应 |
| `useAuth.ts` | `hasAuth(authMark)` 按钮权限判断（backend 模式 authList） |
| `useAppMode.ts` | `isFrontendMode` 判断 VITE_ACCESS_MODE |
| `useTheme.ts` | 主题初始化/切换（明暗） |
| `useCommon.ts` | homePath 等公共计算 |
| `useHeaderBar.ts` / `useLayoutHeight.ts` / `useFastEnter.ts` / `useCeremony.ts` / `useChart.ts` | 布局/杂项 |

### 2.7 组件体系（`src/components/core/` — 项目自研 Art 组件）

| 目录 | 内容 | 代表组件 |
|---|---|---|
| `layouts/` | 布局全家桶 | ArtHeaderBar、ArtMenus（侧边栏）、ArtPageContent、ArtWorkTab、ArtBreadcrumb、ArtGlobalSearch、ArtSettingsPanel、ArtGlobalComponent、ArtFastEnter、ArtScreenLock |
| `tables/` | 表格 | **ArtTable**、ArtTableHeader（列筛选+刷新） |
| `forms/` | 表单 | ArtButtonTable（图标操作按钮）、ArtWangEditor |
| `charts/` | 图表封装 | ArtLineChart、ArtBarChart、ArtRingChart、ArtRadarChart、ArtKLineChart 等（基于 ECharts 6） |
| `cards/` | 卡片 | ArtStatsCard、ArtLineChartCard、ArtProgressCard、ArtImageCard 等 |
| `base/` | 基础 | ArtLogo、ArtSvgIcon、ArtBackToTop |
| `views/` | 业务视图组件 | ArtResultPage、ArtException、AuthTopBar、LoginLeftView |
| `text-effect/` | 文字动效 | ArtCountTo（数字滚动）、ArtTextScroll、ArtFestivalTextScroll |
| `media/` | 媒体 | ArtVideoPlayer（xgplayer）、ArtCutterImg |
| `others/` | 其他 | ArtWatermark、ArtMenuRight、ArtFireworksEffect |
| `theme/` | 主题 | ThemeSvg |
| `widget/` | 小部件 | ArtIconButton |
| `banners/` | 横幅 | — |

### 2.8 视图模块（`src/views/`）

```
views/
├── auth/           # login（含 ArtDragVerify 滑块）/ register / forget-password
├── index/          # 布局壳（sidebar + header + content + global）
├── dashboard/console/  # 工作台：CardList + ActiveUser + SalesOverview + NewUser + Dynamic + TodoList + AboutProject
├── system/         # 超级管理员：user(CRUD+搜索+弹窗) / role(含权限树弹窗) / menu(拖拽排序+弹窗) / tenant(租户CRUD) / user-center
├── tenant-system/  # 租户管理员（system 的隔离副本）：user / role / menu（只读）
├── ad-account/     # FB：index（账号列表）+ manage（20列广告账户管理+AddAuth/DeleteAuth/AddToBm 弹窗）+ bm（BM列表）
├── page-manage/    # 主页管理（FB 页面）
├── pixel-manage/   # 像素管理（FB 像素）
├── result/         # success / fail
├── exception/      # 403 / 404 / 500
└── outside/        # Iframe 外链
```

---

## 3. 后端代码逻辑与模块调用关系

### 3.1 启动链路（`main.go`）

```
main.go
├── godotenv.Load()            # 加载 .env（DATABASE_URL / REDIS_URL / FB_* / SERVER_PORT）
├── config.DefaultConfig()     # 读环境变量
├── db.Connect(DSN)            # pgxpool 连接池（失败回退内存模式）
├── services.TryUpgradeToRedis() # 有 REDIS_URL 升级为分布式 FB 限速
└── routes.SetupRouter().Run(:9090)
```

### 3.2 分层与调用方向

```
routes/（URL → Handler 映射，唯一 router 文件）
   ↓
middleware/（CORS → AuthRequired → TenantContext → RequirePermission）
   ↓
handlers/（ShouldBindJSON 校验 → 调 Service → models.Success/Error 响应）
   ↓
services/（业务逻辑 + 参数化 SQL，WHERE tenant_id 过滤）
   ↓
models/（结构体 + 常量） + db/（pgxpool）
```

### 3.3 中间件链（对每个受保护请求依次执行）

| 中间件 | 文件 | 作用 |
|---|---|---|
| `CORS()` | `middleware/cors.go` | 跨域 |
| `AuthRequired()` | `middleware/auth.go` | 取 `Authorization: Bearer <token>` → `SessionService.Validate` → `c.Set("userID")` |
| `TenantContext()` | 同文件 | 从 session 取 tenantID → 设置 PostgreSQL GUC `app.current_tenant_id`（超管设 `''`）→ `c.Set("tenantID")` |
| `RequirePermission("system:user:create")` | `middleware/rbac.go` | 按用户角色查 `role_permissions` 校验权限点 |

### 3.4 路由分组（`routes/router.go`）

```
/api/v1/ping              健康检查（无鉴权）
/api/v1/auth/login        登录（无鉴权）
/api/v1/fb/callback       FB OAuth 回调（无鉴权）
/api/v1/fb/go/:token      FB 短链接 302（无鉴权）
/api/v1/privacy-policy    隐私政策（无鉴权）
/api/v1/fb/data-deletion  FB 数据删除回调（无鉴权）
── 兼容旧路由 ──
/api/user/info  /api/user/list  /api/role/list  /api/v3/system/menus/simple
── 受保护组（AuthRequired + TenantContext）──
/auth/userinfo  /auth/menus
/tenants        CRUD + switch + current + permissions
/users          CRUD + batch-delete
/roles          CRUD + /:id/menus
/menus          CRUD + /tree
/fb/*           auth-url / status / ad-accounts(+detail/payments/assign-user/remove-user) / accounts(+label/refresh)
                bm-list / pages / pixels / users/lookup / disconnect / refresh-status
```

### 3.5 服务层（`services/`）

| 服务 | 关键方法 | 说明 |
|---|---|---|
| `user_service.go` | GetPasswordHash / GetByID / List(tenantID,page,size,keyword) / Create(自动绑定 role_name) / Update / Delete | 用户 CRUD，带租户过滤 |
| `role_service.go` | List / GetByID / Create / Update(含 menuIds) / Delete | 角色 CRUD |
| `menu_service.go` | List / Tree / **TreeByIDs**(自动补祖先防孤儿菜单) / Create / Update / Delete / ListFallback | 菜单树构建（核心） |
| `tenant_service.go` | Create(查共享角色 R_TENANT_ADMIN 而非新建) / List / Switch 等 | 租户管理 + 共享角色方案 |
| `session_service.go` | Create(SSO 踢旧) / Validate / InvalidateUser / DeleteExpired | 会话（token 认证核心） |
| `fb_service.go` | 授权/换 token/广告账户聚合/支付记录/分配用户/删除权限 | FB Graph API 代理（SOCKS5/HTTP 代理支持） |
| `fb_cache_service.go` | 缓存读写（fb_accounts_cache / fb_ad_accounts_cache / fb_refresh_status） | 缓存优先 + 异步刷新 |
| `fb_rate_limiter.go` / `fb_rate_limiter_redis.go` | 内存队列（4s间隔）/ Redis 原子时隙 | FB 限速 |

### 3.6 数据库表（PostgreSQL 17，DB: `fbai`）

`users`、`roles`、`menus`、`sessions`、`tenants`、`permissions`、`role_permissions`、`fb_tokens`、`fb_accounts_cache`、`fb_ad_accounts_cache`、`fb_refresh_status`（11 张）。

关键点：
- `users.role_name` 不能为空（否则前端 roles 为 null → 所有用户看到所有菜单）
- `roles` 表 RLS 已禁用；靠服务层 `WHERE tenant_id` 过滤（主要防御）
- `menus.title` NOT NULL；`scopes` 用 PostgreSQL `TEXT[]`

---

## 4. 前后端完整调用链（登录 → 工作台示例）

```
[用户点击登录]
前端 login/index.vue handleSubmit
  → api/auth.ts fetchLogin({userName,password,rememberMe})
  → axios POST /api/v1/auth/login
  → AuthHandler.Login
      → middleware.ValidateUser（UserService.GetPasswordHash + crypto.CheckPassword）
      → GenerateToken(userID, rememberMe, tenantID) → SessionService.Create（SSO：删除旧会话）
      → 返回 { token, refreshToken, userInfo }
  → userStore.setToken(rememberMe→3天/24h 过期) → setLoginStatus(true) → router.push('/')

[路由守卫接管]
  → fetchGetUserInfo() → GET /api/user/info → GetUserInfoHandler → UserService.GetByID → 返回 roles/permissions/tenantId
  → MenuProcessor.getMenuList() → GET /api/v1/auth/menus
      → AuthHandler.GetMenus → RoleService.GetByID(menu_ids) → MenuService.TreeByIDs(menu_ids)（超管同样走 TreeByIDs）
      → 返回 MenuTree[] → transformBackendMenu（MENU_I18N_MAP/MENU_ICON_MAP 映射）→ 注册动态路由
  → RoutePermissionValidator 校验目标路径 → 进入 /dashboard/console
```

---

## 5. 系统页面风格（设计体系）

### 5.1 整体布局

```
┌────────────┬──────────────────────────────┐
│            │  ArtHeaderBar（顶栏）          │
│ ArtMenus   ├──────────────────────────────┤
│ 侧边菜单    │  ArtWorkTab（标签页）           │
│            │  ArtPageContent（内容区）       │
│            │                              │
└────────────┴──────────────────────────────┘
全局覆盖层：ArtGlobalComponent（快捷键/烟花/通知/LockScreen 等）
```

- 壳：`views/index/index.vue`（`#app-sidebar` + `#app-header` + `#app-content` + `#app-global`）
- 侧边栏可折叠、菜单支持 i18n 标题 + `ri:` 前缀图标
- 顶栏含面包屑、全局搜索、通知、多语言、明暗主题、设置面板（ArtSettingsPanel：主题色/布局/圆角等）

### 5.2 主题与颜色（`el-ui.scss` + `useTheme`）

- 主色：`--main-color: var(--el-color-primary)`，由设置面板动态切换
- 圆角：`--el-border-radius-base: calc(var(--custom-radius)/3 + 2px)`，全局组件高度 36px
- 支持明暗两套主题（`html.dark`），暗色下 ElCard 背景跟随 `--default-box-color`
- 全局样式分层：`reset.scss → app.scss → el-ui.scss → el-dark.scss → dark.scss → 路由过渡/主题切换动画`；Tailwind 仅用于布局工具类

### 5.3 表格页范式（**所有数据表格页统一**）

以 `views/system/user/index.vue` 为范本，结构固定：

```
<div class="xxx-page art-full-height">          ← 单根元素 + 全高
  <XxxSearch v-model="searchForm" @search @reset/>   ← 可选搜索栏
  <ElCard class="art-table-card">                ← 卡片容器（shadow: never）
    <ArtTableHeader v-model:columns @refresh>    ← 头部：列筛选 + #left 插槽放操作按钮
      <ElButton v-ripple>新增用户</ElButton>      ← 页面头部按钮：无 type + v-ripple
    <ArtTable :data :columns :pagination />      ← 统一表格（非 ElTable）
    <XxxDialog v-model:visible :type :data />    ← 编辑弹窗
  </ElCard>
</div>
```

`script setup` 固定模式：
1. `useTable({ core: { apiFn, apiParams:{current:1,size:20}, columnsFactory }, transform:{ dataTransformer } })`
2. `columnsFactory` 中：selection/index 列用 `width`，数据列用 `minWidth`，操作列 `width:120~160, fixed:'right'`
3. 单元格用 `formatter` + `h(ElTag)` / `h(ArtButtonTable)` 渲染
4. 状态枚举映射（如 `status===1 → {type:'success', text:'启用'}`）
5. 弹窗提交：`handleDialogSubmit` 区分 add/edit，成功后 `ElMessage.success` + `refreshData()`
6. 删除：`ElMessageBox.confirm` 二次确认

### 5.4 按钮规范（三种场景，全局统一）

| 位置 | 样式 |
|---|---|
| 页面头部/表格左上方 | `<ElButton v-ripple>文字</ElButton>`（无 type） |
| 弹窗底部 | 取消 = `<ElButton>取消</ElButton>`；确定 = `<ElButton type="primary">确定</ElButton>` |
| 表格操作列 | `ArtButtonTable` 图标按钮（add/edit/delete/view/more 内置类型 + 自定义 icon） |

### 5.5 弹窗模式

- 表单弹窗：`ElForm` + `rules` 前后端双重校验；标题/按钮/消息全部 `$t()`（zh.json/en.json）
- 复杂弹窗（FB 授权/删除）：Step 步骤式（`ElCheckboxGroup` + `ElCheckbox` + 确认按钮禁用态）、tab 双栏、绿色主操作按钮等，风格遵循 `references/dialog-patterns.md`
- 权限树弹窗：`ElTree show-checkbox` + `node-key` 与数据源类型一致（backend 模式用 id 或 name 映射）

### 5.6 页面风格速览

| 页面 | 风格 |
|---|---|
| **登录页** | 左右分栏：左侧 LoginLeftView（插画/标语），右侧表单（标题+副标题+输入框+**ArtDragVerify 滑块验证**+记住密码+登录按钮）；支持明暗主题 |
| **工作台** | 卡片网格：顶部 CardList 统计卡 → `ElRow/ElCol` 栅格（响应式 :sm/:md/:lg）排布图表卡（ActiveUser 环形图、SalesOverview 折线图、NewUser 柱状图、Dynamic 数字、TodoList 待办、AboutProject 卡片） |
| **系统管理** | 标准表格页范式（搜索框+卡片+表格+弹窗），菜单管理用 SortableJS 拖拽排序 |
| **FB 广告账户** | 20 列宽表格（`minWidth` 数据列 + 横向滚动），多弹窗（授权/删除/添加BM），缓存优先 + 刷新状态提示（"数据更新中..."） |
| **异常/结果页** | ArtException / ArtResultPage 统一组件（图标+标题+描述+按钮） |

---

## 6. 关键约定（编码时必须遵守）

1. 响应格式 `{ code, msg, data }`，请求自动解包 data
2. 表格统一 `ArtTable + ArtTableHeader + useTable()`，禁止裸 `ElTable`
3. 所有用户可见文本走 `$t()`，i18n 键先查 `src/locales/langs/zh.json`
4. 模板单根元素；Element Plus 优先（80% UI）；新 EP 依赖加 `optimizeDeps.include`
5. 前端路由 `name` 必须与数据库 `menus.name` 同步；新增菜单需同步 i18n 映射 + 菜单图标 `ri:` 前缀
6. 后端分层 `routes → handlers → services → models`，Gin `ShouldBindJSON` + `binding` 标签双重校验，参数化查询防 SQL 注入
7. 租户数据隔离：服务层显式 `WHERE tenant_id`（主要防御），避免仅依赖 RLS
8. 相似功能：复制模块修改副本（功能隔离），不污染原模块
9. 文档：每次变更写 `docs/dev/YYYY-MM-DD--<feature>.md`
10. 验证后立即提交并推送 GitHub（conventional commits）

---

## 7. 附：快速索引

- 页面风格范本：`src/views/system/user/index.vue`
- 路由守卫入口：`src/router/guards/beforeEach.ts`
- 菜单转换：`src/router/core/MenuProcessor.ts`
- HTTP 封装：`src/utils/http/index.ts`
- 后端路由表：`art-design-server/routes/router.go`
- 菜单树构建：`art-design-server/services/menu_service.go`
- 会话/认证：`art-design-server/services/session_service.go` + `middleware/auth.go`
- 完整 API 映射：`docs/features/` 与此文档配套的 `fbai-context` 参考资料

# FbAi — Agent Guide

## 架构概览

双包仓库（独立包，无 monorepo 工具）：

| 包名 | 技术栈 | 端口 | 入口 |
|---|---|---|---|
| `art-design-pro/` | Vue 3 + Vite 7 + TypeScript + Element Plus + Pinia + Tailwind CSS 4 | 3006 | `src/main.ts` |
| `art-design-server/` | Go 1.25 + Gin + pgx v5 (PostgreSQL) | 9090 | `main.go` |

后端提供 API，前端 Vite 开发服务器代理 `/api` → `localhost:9090`。

## 启动命令

```bash
# 一键启动
./start.bat          # Windows
./start.sh           # Linux/macOS

# 手动启动
cd art-design-server && go run main.go    # 需要 GOPROXY=https://goproxy.cn,direct
cd art-design-pro && pnpm dev
```

默认登录：`admin` / `admin123`

---

## 后端关键规则

### 配置与数据库
- 配置文件：`art-design-server/.env`（从 `.env.example` 复制）
- 数据库：PostgreSQL。`DATABASE_URL` 缺失时回退到内存模式（无持久化）
- 迁移文件：`art-design-server/migrations/` — 手动执行，不自动运行
- 种子数据：`art-design-server/seed.sql` — 演示用户和角色

### 多租户 RBAC
- `TenantContext` 中间件设置 PostgreSQL 会话变量 `app.current_tenant_id`
- 超级管理员 = `tenant_id IS NULL` 的用户
- **主要防御**：服务层 `WHERE tenant_id = $1`，不要仅依赖 RLS

### 认证系统
- 基于 Token（非 JWT），存储在 `sessions` 表
- 请求头：`Authorization: Bearer <token>`
- 登录：`POST /api/v1/auth/login` → 返回 token

### API 路由
- 新 API：`/api/v1/...`（推荐使用）
- 旧 API：`/api/user/...` 和 `/api/role/...`（兼容保留）

### 后端代码规则
- 请求绑定：`c.ShouldBindJSON(&struct)` + 结构体标签 `binding:"required"`
- 响应格式：`c.JSON(http.StatusXxx, gin.H{...})`
- 项目结构：`routes/` → `handlers/` → `services/` → `models/`
- **所有用户输入必须前后端双重验证**
- **SQL 注入防护**：只使用参数化查询 (`$1`, `$2`)，禁止字符串拼接

---

## 前端关键规则

### 环境配置
- 包管理器：pnpm (>=8.8.0)，Node >=20.19.0
- 环境变量：`.env`（基础），`.env.development`，`.env.production`
- 前缀：`VITE_`
- 代理目标：`VITE_API_PROXY_URL`（默认 `http://localhost:9090`）

### 路径别名
| 别名 | 映射到 |
|---|---|
| `@/` | `src/` |
| `@views/` | `src/views/` |
| `@utils/` | `src/utils/` |
| `@stores/` | `src/store/` |
| `@imgs/` | `src/assets/images/` |
| `@icons/` | `src/assets/icons/` |
| `@styles/` | `src/assets/styles/` |

### 自动导入
- Vue、Vue Router、Pinia、VueUse API 自动导入
- 无需手动 `import { ref, computed, useRouter }` 等
- 组件自动导入：`unplugin-vue-components`

### SCSS 全局变量
- `@styles/core/el-light.scss` 和 `@styles/core/mixin.scss` 通过 `vite.config.ts` 注入到所有 SCSS 文件
- 直接使用其 mixin/变量

### 路由
- Hash 模式：`createWebHashHistory`
- 静态路由：`src/router/routes/`
- 动态菜单：后端驱动，`GET /api/v1/auth/menus`

---

## 代码风格规则

### ESLint + Prettier
| 规则 | 说明 |
|---|---|
| 单引号 | 使用单引号包裹字符串 |
| 无分号 | 语句末尾不加分号 |
| 100字符宽度 | 每行最大100字符 |
| 无尾逗号 | `trailingComma: "none"` |
| Vue缩进 | `<script>` 和 `<style>` 标签内缩进 |
| 禁止var | 必须使用 `let` 或 `const` |
| 允许any | TypeScript 允许使用 `any` 类型 |
| 格式缩进 | 2空格缩进，不使用Tab |

### 提交规范 (Commitlint)
允许的提交类型：
- `feat` — 新增功能
- `fix` — 修复缺陷
- `docs` — 文档变更
- `style` — 代码格式（不影响功能）
- `refactor` — 代码重构
- `perf` — 性能优化
- `test` — 测试相关
- `build` — 构建流程/依赖变更
- `ci` — CI配置变更
- `revert` — 回滚commit
- `chore` — 辅助工具更改
- `wip` — 开发中

交互式提交：`pnpm commit`

---

## 前端强制规则 (43条)

### API 与数据层
1. 响应格式统一：`{ code: number, msg: string, data: T }`
2. 网络请求自动解包 `data`，直接解构：`const { token } = await fetchLogin(...)`
3. 表格分页使用 `src/utils/table/tableConfig.ts` 中的字段映射

### 权限系统
4. 两种模式：`frontend`（前端路由）/ `backend`（后端菜单）
5. 当前使用 `backend` 模式 — 菜单来自数据库
6. 按钮权限：使用 `hasAuth()` 或 `v-auth` / `v-roles` 指令
7. 权限流程：登录 → 获取用户信息 → 获取菜单 → 注册路由

### 组件与UI
8. **Element Plus 覆盖 80% UI** — 优先使用 `<el-*>` 组件
9. **所有表格必须使用** `ArtTable` + `ArtTableHeader` + `useTable()` hook
10. 表格操作按钮必须使用 `ArtButtonTable` 组件
11. 按钮样式：
    - 页面头部：`<ElButton v-ripple>文字</ElButton>`（无类型）
    - 弹窗底部：取消 = `<ElButton>取消</ElButton>`，确定 = `<ElButton type="primary">确定</ElButton>`
    - 表格操作：`ArtButtonTable` 组件
12. Vue模板**必须有单根元素**
13. 新 Element Plus 依赖：添加到 `vite.config.ts` 的 `optimizeDeps.include`
14. 表格列宽度：数据列用 `minWidth`，固定列用 `width`

### 国际化
15. 所有用户可见文本**必须使用** `$t('key.path')` — 禁止硬编码中英文
16. 语言文件：`src/locales/langs/zh.json` 和 `en.json`

### 代码质量
17. 提交前自动执行：ESLint → Prettier → Stylelint
18. 禁止绕过 lint-staged 直接提交
19. 每个功能必须有文档：`docs/dev/YYYY-MM-DD--<feature>.md`

### 模块化
20. 每个功能必须是独立模块/组件
21. 单Vue组件单职责，逻辑提取到 hooks/utils
22. 文件超过300行应拆分
23. **相似功能**：复制原功能修改副本，不要修改现有模块

### 后端 (Gin)
24. 请求绑定：使用 `c.ShouldBindJSON(&struct)` + 结构体标签
25. 响应格式：`c.JSON(http.StatusXxx, gin.H{...})`
26. 项目结构：`routes/` → `handlers/` → `services/` → `models/`
27. 数据库：PostgreSQL via `pgx`，使用连接池，DSN从环境变量读取
28. **所有用户输入必须前后端双重验证**
29. **SQL 注入防护**：只使用参数化查询

### 开发流程
30. 代码 → 验证(lint/build) → 提交 → 推送到GitHub
31. 每次逻辑变更一个提交
32. 会话结束前必须推送已验证的代码

---

## 关键注意事项

### 后端
1. **后端结构变更后必须重启** — `server.exe` 不会热重载
2. **Go代理**：中国地区必须设置 `GOPROXY=https://goproxy.cn,direct`
3. **端口冲突**：启动前检查 `netstat -ano | grep ":9090"`
4. **后端重启流程**：
   ```bash
   taskkill //F //IM server.exe 2>/dev/null
   GOPROXY=https://goproxy.cn,direct go build -o server.exe ./main.go
   ./server.exe &
   curl -s http://localhost:9090/api/v1/ping
   ```

### 前端
5. **Windows路径**：使用绝对路径 (`E:\FbAi\...`)，不用MSYS路径
6. **环境变量文件**：使用 `terminal` 编辑 `.env`，避免工具自动替换密码
7. **VITE_ACCESS_MODE 变更需要重启前端**
8. **前端路由名必须与数据库菜单名同步**

### 权限与菜单
9. **用户 `role_name` 不能为空** — 否则所有用户看到所有菜单
10. **孤儿菜单**：子菜单在 `menu_ids` 但父菜单不在 → 子菜单消失
11. **超级管理员菜单**：必须使用 `TreeByIDs(role.MenuIDs)` 而非 `Tree()`
12. **菜单图标需要 `ri:` 前缀**
13. **菜单标题需要 i18n 键映射**

### 多租户
14. **`roles` 表 RLS 已禁用** — 主要依赖服务层 WHERE 子句
15. **租户角色必须包含 `menu_ids`** — 否则登录时 500 错误
16. **租户 `role_code` 必须全局唯一** — 使用 `T{id}_R_ADMIN` 格式
17. **所有 Create handler 必须自动绑定 tenant_id**

### 数据类型
18. **密码最少6字符** — 后端 `binding:"min=6"`
19. **无日期库** — 使用原生 `new Date()` 格式化
20. **`useTable` 需要分页 API** — 扁平数组会导致类型推断失败

---

## 数据库核心表

| 表名 | 关键字段 | 说明 |
|---|---|---|
| `users` | id, user_name, password (bcrypt), email, phone, role_id, role_name, tenant_id, status | tenant_id=NULL 为全局管理员 |
| `roles` | id, role_name, role_code, menu_ids (int[]), tenant_id, status | role_code 唯一；**RLS 已禁用** |
| `menus` | id, name, title, parent_id, path, component, icon, sort_order, hidden | title NOT NULL |
| `sessions` | id, user_id, token, refresh_token, expires_at, tenant_id | SSO：新登录删除旧会话 |
| `tenants` | id, name, code, contact_name, contact_phone, status | — |
| `permissions` | id, code, module, action, description | 18个权限点 |
| `role_permissions` | role_id, permission_id | 多对多关联 |

---

## API 端点映射

| 前端函数 | URL | Go Handler |
|---|---|---|
| `fetchLogin()` | `POST /api/v1/auth/login` | `AuthHandler.Login` |
| `fetchGetUserInfo()` | `GET /api/user/info` | `UserHandler.GetInfo` |
| `fetchGetMenuList()` | `GET /api/v1/auth/menus` | `AuthHandler.GetMenus` |
| `fetchGetUserList()` | `GET /api/user/list` → `/api/v1/users` | `UserHandler.List` |
| `fetchCreateUser()` | `POST /api/v1/users` | `UserHandler.Create` |
| `fetchUpdateUser()` | `PUT /api/v1/users/:id` | `UserHandler.Update` |
| `fetchDeleteUser()` | `DELETE /api/v1/users/:id` | `UserHandler.Delete` |
| `fetchGetRoleList()` | `GET /api/role/list` → `/api/v1/roles` | `RoleHandler.List` |
| `fetchCreateRole()` | `POST /api/v1/roles` | `RoleHandler.Create` |
| `fetchUpdateRole()` | `PUT /api/v1/roles/:id` | `RoleHandler.Update` |
| `fetchDeleteRole()` | `DELETE /api/v1/roles/:id` | `RoleHandler.Delete` |
| `fetchGetRoleMenus()` | `GET /api/v1/roles/:id/menus` | `RoleHandler.GetMenus` |
| `fetchCreateMenu()` | `POST /api/v1/menus` | `MenuHandler.Create` |
| `fetchUpdateMenu()` | `PUT /api/v1/menus/:id` | `MenuHandler.Update` |
| `fetchDeleteMenu()` | `DELETE /api/v1/menus/:id` | `MenuHandler.Delete` |

---

## 架构决策

- **Composition API + `<script setup>`** — 干净、可树摇的组件
- **Element Plus 自动导入** — 无需手动注册组件
- **SCSS 模块 + Tailwind 混合** — SCSS 用于组件样式，Tailwind 用于布局/工具类
- **Pinia + persistedstate** — 状态通过 localStorage 持久化
- **后端访问模式**：`VITE_ACCESS_MODE=backend` — 菜单从数据库动态加载
- **多租户架构**：双层隔离 — 服务层 `WHERE tenant_id`（主要）+ RLS（尽力）
- **功能隔离**：租户系统页面是系统页面的副本，独立维护

---

## 代码检查清单

编写代码前，验证：
- [ ] 是否查询了代码依赖和影响范围？
- [ ] 菜单-数据库同步：新增/修改前端路由时，数据库 `menus` 是否有匹配的 `name`？
- [ ] 数据库约束：INSERT 前检查 `\d <table>` 的 NOT NULL 列
- [ ] Element Plus：是否已有此组件？
- [ ] 表格模式：使用 `ArtTable` + `ArtTableHeader` + `useTable()`？
- [ ] 按钮样式：头部 = 无类型按钮，弹窗 = 取消+主按钮
- [ ] 国际化：所有用户文本使用 `$t()`？
- [ ] 单根：模板有单根元素？
- [ ] API格式：响应格式 `{ code, msg, data }` 正确？
- [ ] 权限：使用 `hasAuth()` 或 `v-auth`/`v-roles`？
- [ ] 模块化：每个文件 < 300 行？逻辑提取到 hooks/utils？
- [ ] 功能隔离：相似功能？复制，不修改
- [ ] 文档：创建 `docs/dev/YYYY-MM-DD--<feature>.md`？
- [ ] 输入验证：前后端都有验证？
- [ ] 租户上下文：新 handler 有 `tenantID` 自动绑定？服务有 `WHERE tenant_id` 过滤？
- [ ] 后端重启：代码变更后是否重启并验证？
- [ ] Git推送：功能完成并验证？立即推送

---

## 禁止事项

- ❌ 禁止直接提交到 `master` — 使用功能分支
- ❌ 禁止绕过 pre-commit hooks 提交
- ❌ 禁止在源代码中硬编码 API 密钥、密码或 DSN
- ❌ 禁止在 API 响应中使用 `any` 类型
- ❌ 禁止重新实现 Element Plus 组件
- ❌ 禁止使用 `git add .`（盲添加）— 添加特定文件
- ❌ 禁止强制推送到 master (`--force`)
- ❌ 禁止将 `.env` 文件提交到 Git — 凭据泄露风险

---

## 文档链接

- **art-design-pro 文档**: https://www.artd.pro/docs/zh/guide/introduce.html
- **Element Plus**: https://element-plus.org/zh-CN/component/overview.html
- **Gin 后端**: https://gin-gonic.com/zh-cn/docs/

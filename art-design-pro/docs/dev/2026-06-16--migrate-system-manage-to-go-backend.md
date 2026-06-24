# 用户/角色/菜单管理 — 数据迁移至 Go 后端 — 2026-06-16

## 概述

将前端用户管理、角色管理、菜单管理的全部数据从内存 mock 迁移到 Go 后端 + PostgreSQL，实现完整的 CRUD 操作和前后端数据联通。

## 修改的文件

### Go 后端 (art-design-server)

| 文件 | 变更 | 原因 |
| --- | --- | --- |
| `models/role.go` | 重写：字段对齐 DB schema (`roleName`, `roleCode`, `description`, `menuIds`, `status`) | 匹配 PostgreSQL `roles` 表 |
| `models/menu.go` | 重写：字段对齐 DB schema (`title`, `name`, `path`, `sortOrder`, `menuType`, `hidden`) | 匹配 PostgreSQL `menus` 表 |
| `models/user.go` | 更新 JSON tag：`userEmail`, `userPhone`, `createTime`, `updateTime` | 匹配前端期望的字段名 |
| `services/role_service.go` | 完全重写：从内存 map 迁移到 PostgreSQL 查询 | 数据持久化 |
| `services/menu_service.go` | 完全重写：从内存 map 迁移到 PostgreSQL 查询 | 数据持久化 |
| `services/user_service.go` | 添加 `createTime`/`updateTime` 返回字段；密码哈希改用 `crypto` 包 | 字段对齐；解决 import cycle |
| `handlers/role_handler.go` | 更新字段引用；Status 默认值 1 | 适配新 model |
| `handlers/menu_handler.go` | 更新字段引用；Status 默认值 1 | 适配新 model |
| `handlers/auth_handler.go` | 更新角色匹配逻辑（`r.RoleName` 替代 `r.Name`） | 适配新 model |
| `middleware/auth.go` | 密码验证改用 `crypto.CheckPassword` | 提取公共 crypto 包 |
| `routes/router.go` | 添加向后兼容路由：`/api/user/list`, `/api/role/list`, `/api/v3/system/menus/simple` | 兼容前端旧 API 路径 |

### 新增文件

| 文件                 | 用途                                                 |
| -------------------- | ---------------------------------------------------- |
| `crypto/password.go` | 密码哈希/验证工具（bcrypt），独立包避免 import cycle |
| `seed.sql`           | 数据库种子脚本（菜单 + 角色权限 + 演示用户）         |

### 前端 (art-design-pro)

| 文件 | 变更 | 原因 |
| --- | --- | --- |
| `src/api/system-manage.ts` | 添加完整的 CRUD 函数：`fetchCreateUser`, `fetchUpdateUser`, `fetchDeleteUser`, `fetchCreateRole`, `fetchUpdateRole`, `fetchDeleteRole`, `fetchCreateMenu`, `fetchUpdateMenu`, `fetchDeleteMenu` | 支撑页面的增删改查操作 |
| `src/types/api/api.d.ts` | 添加 `CreateUserParams`, `UpdateUserParams`, `CreateRoleParams`, `UpdateRoleParams` 类型 | 类型约束 |
| `src/views/system/user/index.vue` | 数据转换器适配后端字段；CRUD 操作连接真实 API | 消除 mock data 依赖 |
| `src/views/system/user/modules/user-dialog.vue` | 重写表单：支持完整字段（密码、昵称、邮箱、手机、状态） | 配合后端 CRUD |
| `src/views/system/role/index.vue` | 数据转换器适配后端字段（`id`→`roleId`, `status`→`enabled`）；删除操作连接真实 API | 消除 mock data 依赖 |
| `src/views/system/role/modules/role-edit-dialog.vue` | 创建/编辑操作连接真实 API（`fetchCreateRole`/`fetchUpdateRole`） | 角色 CRUD |
| `src/views/system/menu/index.vue` | 添加 `menuTreeToAppRoute` 转换函数（后端 MenuTree→前端 AppRouteRecord）；CRUD 操作连接真实 API | 菜单 CRUD + 数据格式适配 |

## 设计决策

### 1. 密码哈希处理

- 从 `middleware` 包提取到独立的 `crypto` 包，解决 `services → middleware → services` 的 import cycle
- 用户创建时自动 bcrypt 哈希密码（12 rounds）
- 演示用户密码统一为 `123456`

### 2. 字段命名对齐

- 后端 JSON 输出直接匹配前端期望字段名（`userEmail`, `userPhone`, `createTime`, `updateTime`）
- 数据库列名保持 snake_case（PostgreSQL 惯例），通过 GORM `column` tag 映射
- 前端数据转换器处理 `id→roleId`, `status(int)→enabled(bool)` 等格式转换

### 3. 向后兼容路由

保留旧版 API 路径避免破坏现有调用：

- `/api/user/list` → `/api/v1/users`
- `/api/role/list` → `/api/v1/roles`
- `/api/v3/system/menus/simple` → `/api/v1/menus/tree`

### 4. 种子数据

- **用户**: admin + 5 个演示用户（alex, sophia, liam, olivia, emma）
- **角色**: 超级管理员(R_SUPER)、管理员(R_ADMIN)、普通用户(R_USER)
- **菜单**: 6 个菜单项（Dashboard → Console, System → User/Role/Menu）
- 角色菜单权限: R_SUPER 拥有全部菜单, R_ADMIN 拥有前3个, R_USER 仅有 Dashboard

## API 端点清单

| 方法   | 路径                          | 说明                   |
| ------ | ----------------------------- | ---------------------- |
| POST   | `/api/v1/auth/login`          | 登录                   |
| GET    | `/api/user/info`              | 用户信息（兼容前端）   |
| GET    | `/api/user/list`              | 用户列表（兼容旧路径） |
| GET    | `/api/role/list`              | 角色列表（兼容旧路径） |
| GET    | `/api/v3/system/menus/simple` | 菜单树（兼容旧路径）   |
| GET    | `/api/v1/users`               | 用户列表（分页）       |
| GET    | `/api/v1/users/:id`           | 用户详情               |
| POST   | `/api/v1/users`               | 创建用户               |
| PUT    | `/api/v1/users/:id`           | 更新用户               |
| DELETE | `/api/v1/users/:id`           | 删除用户               |
| GET    | `/api/v1/roles`               | 角色列表               |
| GET    | `/api/v1/roles/:id`           | 角色详情               |
| GET    | `/api/v1/roles/:id/menus`     | 角色菜单权限           |
| POST   | `/api/v1/roles`               | 创建角色               |
| PUT    | `/api/v1/roles/:id`           | 更新角色               |
| DELETE | `/api/v1/roles/:id`           | 删除角色               |
| GET    | `/api/v1/menus`               | 菜单列表（平铺）       |
| GET    | `/api/v1/menus/tree`          | 菜单树                 |
| GET    | `/api/v1/menus/:id`           | 菜单详情               |
| POST   | `/api/v1/menus`               | 创建菜单               |
| PUT    | `/api/v1/menus/:id`           | 更新菜单               |
| DELETE | `/api/v1/menus/:id`           | 删除菜单               |

## 验证结果

- ✅ Go backend 编译通过
- ✅ 全部 API 端点测试通过（直接调用和 Vite proxy）
- ✅ 前端 TypeScript 类型检查通过
- ✅ 数据库种子数据正确
- ✅ 用户列表返回 6 个用户（分页）
- ✅ 角色列表返回 3 个角色
- ✅ 菜单树返回 2 个根节点
- ✅ 用户信息接口正确返回 roles 和 userId

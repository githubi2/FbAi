# art-design-server Backend — Initial Setup — 2026-06-16

## Overview

创建 Gin 框架的 Go 后端项目 `art-design-server`，与前端 `art-design-pro` 完全分离，提供认证、用户管理、角色管理、菜单管理的 RESTful API。

## Tech Stack

| 层 | 技术 | 版本 |
|---|---|---|
| 语言 | Go | 1.25 (toolchain) / 1.21 (host) |
| 框架 | Gin | v1.12.0 |
| 加密 | bcrypt (golang.org/x/crypto) | v0.53.0 |
| 数据存储 | 内存 (map) | 开发阶段 |

## Project Structure

```
art-design-server/
├── main.go                  # 入口：启动 Gin 服务（端口 9090）
├── config/
│   └── config.go            # 应用配置（Server/Database/JWT）
├── models/
│   ├── response.go          # 标准响应 BaseResponse（对齐前端）
│   ├── user.go              # 用户模型 + 请求体
│   ├── role.go              # 角色模型 + 请求体
│   └── menu.go              # 菜单模型 + 菜单树
├── middleware/
│   ├── cors.go              # CORS 跨域（允许 localhost:3006）
│   └── auth.go              # 认证中间件 + Token 管理
├── handlers/
│   ├── auth_handler.go      # 登录 / 用户信息 / 菜单权限
│   ├── user_handler.go      # 用户 CRUD
│   ├── role_handler.go      # 角色 CRUD + 菜单分配
│   └── menu_handler.go      # 菜单 CRUD + 树
├── services/
│   ├── user_service.go      # 用户业务逻辑
│   ├── role_service.go      # 角色业务逻辑
│   └── menu_service.go      # 菜单业务逻辑
└── routes/
    └── router.go            # 路由定义 + 分组
```

## API Endpoints

### Public
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/ping` | 健康检查 |
| POST | `/api/v1/auth/login` | 登录获取 token |

### Authenticated (Bearer token)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/userinfo` | 当前用户信息 |
| GET | `/api/v1/auth/menus` | 当前用户菜单树 |
| GET | `/api/v1/users` | 用户列表（分页） |
| GET | `/api/v1/users/:id` | 用户详情 |
| POST | `/api/v1/users` | 创建用户 |
| PUT | `/api/v1/users/:id` | 更新用户 |
| DELETE | `/api/v1/users/:id` | 删除用户 |
| POST | `/api/v1/users/batch-delete` | 批量删除 |
| GET | `/api/v1/roles` | 角色列表 |
| GET | `/api/v1/roles/:id` | 角色详情 |
| GET | `/api/v1/roles/:id/menus` | 角色菜单权限 |
| POST | `/api/v1/roles` | 创建角色 |
| PUT | `/api/v1/roles/:id` | 更新角色 |
| DELETE | `/api/v1/roles/:id` | 删除角色 |
| GET | `/api/v1/menus` | 菜单平铺列表 |
| GET | `/api/v1/menus/tree` | 菜单树 |
| POST | `/api/v1/menus` | 创建菜单 |
| PUT | `/api/v1/menus/:id` | 更新菜单 |
| DELETE | `/api/v1/menus/:id` | 删除菜单 |

## Response Format

All responses follow the standard format aligned with the frontend:

```json
{
  "code": 200,
  "msg": "success",
  "data": {}
}
```

## Test Results

All 8 API tests passed:

| Test | Result |
|------|--------|
| GET /api/v1/ping | 200 OK |
| POST /api/v1/auth/login | 200 OK + token |
| GET /api/v1/users (auth) | 200 OK, total=2 |
| GET /api/v1/menus/tree | 200 OK, 2 items |
| GET /api/v1/roles | 200 OK, 2 items |
| POST /api/v1/users | 200 OK, created |
| GET /api/v1/users (after) | 200 OK, total=3 |
| GET /api/v1/users (no auth) | 401 Unauthorized |

## Key Decisions

1. **端口 9090**: 原定 8080 被占用，改用 9090
2. **Token 生成**: 最初使用 `string(rune(userID))` 产生控制字符导致 JSON/HTTP 头问题，改为 `fmt.Sprintf("token_%d_art-design-%d", userID, time.Now().UnixNano())`
3. **Go 代理**: 国内网络需要 `GOPROXY=https://goproxy.cn,direct` 拉取依赖
4. **内存存储**: 开发阶段使用内存 map + sync.RWMutex，后续替换为数据库（GORM + SQLite/MySQL）
5. **前后端分离**: `art-design-server/` 与 `art-design-pro/` 完全独立，在同一 Git 仓库下不同目录

## Next Steps

- [ ] 集成 GORM + 数据库（SQLite 开发 / MySQL 生产）
- [ ] JWT 替换简单 Token
- [ ] 密码加密存储（bcrypt 已集成，需接入注册流程）
- [ ] 接口文档（Swagger）
- [ ] 单元测试 + 集成测试
- [ ] Docker 部署支持
